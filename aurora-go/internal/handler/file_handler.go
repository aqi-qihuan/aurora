package handler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/util"
)

// FileHandler 文件上传处理器（对标 Java FileController + UploadStrategy）
type FileHandler struct {
	// uploadService service.UploadService
}

func NewFileHandler() *FileHandler { return &FileHandler{} }

// UploadFile 通用文件上传（支持图片/文档/视频）
// POST /api/admin/upload
// 对标 FileController.upload() - MinIO/OSS双策略自动选择
func (h *FileHandler) UploadFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("请选择要上传的文件"))
		return
	}
	defer file.Close()

	zap.L().Info("File upload", zap.String("filename", header.Filename), zap.Int64("size", header.Size))

	// TODO: P0-5:
	//   1. AnalyzeFile(file) → 检测类型/扩展名白名单
	//   2. 上传到 MinIO (或 OSS, 根据配置)
	//   3. 返回文件URL

	util.ResponseSuccess(c, map[string]interface{}{
		"url":      "",
		"filename": header.Filename,
		"size":     header.Size,
	})
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

	zap.L().Info("Batch file upload", zap.Int("count", len(files)))
	// TODO: P0-5 并发上传 goroutine pool

	results := make([]map[string]interface{}, len(files))
	for i := range files {
		results[i] = map[string]interface{}{
			"url":      "",
			"filename": files[i].Filename,
		}
	}

	util.ResponseSuccess(c, results)
}

// UploadImage Markdown编辑器专用图片上传
// POST /api/admin/upload/image
// 对标 FileController.uploadImage() - 返回Markdown格式链接
func (h *FileHandler) UploadImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.String(400, "图片上传失败")
		return
	}
	defer file.Close()

	zap.L().Debug("Image upload for markdown editor", zap.String("filename", header.Filename))

	// TODO: P0-5 仅允许图片格式 → 上传MinIO → 返回Markdown格式URL

	// 返回Markdown编辑器兼容的响应格式
	c.JSON(200, gin.H{
		"success": 1,
		"message": "上传成功",
		"url":     "", // 实际图片URL
	})
}
