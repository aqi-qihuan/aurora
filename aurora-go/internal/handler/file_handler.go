package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/service"
	"github.com/aurora-go/aurora/internal/util"
)

// FileHandler 文件上传处理器（对标 Java FileController + UploadStrategy）
type FileHandler struct {
	svc *service.FileService
}

func NewFileHandler(svc *service.FileService) *FileHandler {
	return &FileHandler{svc: svc}
}

// UploadFile 通用文件上传
// POST /api/admin/upload
func (h *FileHandler) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("请选择要上传的文件"))
		return
	}

	result, err := h.svc.UploadSingle(c.Request.Context(), file)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// BatchUpload 批量文件上传
// POST /api/admin/upload/batch
func (h *FileHandler) BatchUpload(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("文件上传失败"))
		return
	}
	files := form.File["files"]
	if len(files) == 0 {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("请选择要上传的文件"))
		return
	}

	results, err := h.svc.UploadBatch(c.Request.Context(), files)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, results)
}

// UploadImage Markdown编辑器专用图片上传
// POST /api/admin/upload/image
func (h *FileHandler) UploadImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.String(400, "图片上传失败")
		return
	}

	markdown, err := h.svc.UploadMarkdownImage(c.Request.Context(), file)
	if err != nil {
		c.String(400, err.Error())
		return
	}

	// 返回Markdown编辑器兼容的响应格式
	result, _ := h.svc.UploadSingle(c.Request.Context(), file)
	url := ""
	if result != nil {
		url = result.URL
	}
	c.JSON(200, gin.H{
		"success": 1,
		"message": "上传成功",
		"url":     url,
		"markdown": markdown,
	})
}
