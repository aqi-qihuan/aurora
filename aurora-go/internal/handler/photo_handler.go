package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/service"
	"github.com/aurora-go/aurora/internal/util"
)

// PhotoHandler 相册照片处理器（对标 Java PhotoController）
type PhotoHandler struct {
	svc *service.PhotoService
}

func NewPhotoHandler(svc *service.PhotoService) *PhotoHandler {
	return &PhotoHandler{svc: svc}
}

// ListPhotos 获取相册下的照片列表
// GET /api/albums/:id/photos
func (h *PhotoHandler) ListPhotos(c *gin.Context) {
	albumId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的相册ID"))
		return
	}
	list, err := h.svc.GetPhotosByAlbum(c.Request.Context(), uint(albumId))
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, list)
}

// UploadPhoto 上传照片到指定相册（支持批量）
// POST /api/admin/albums/:id/photos
func (h *PhotoHandler) UploadPhoto(c *gin.Context) {
	albumId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的相册ID"))
		return
	}
	form, err := c.MultipartForm()
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("文件上传失败"))
		return
	}
	files := form.File["files"]
	if len(files) == 0 {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("请选择要上传的照片"))
		return
	}

	// TODO: P0-9 上传到MinIO后获取URL列表
	// 暂时使用模拟URL
	urls := make([]string, len(files))
	for i := range files {
		urls[i] = "/uploads/" + files[i].Filename
	}

	photos, err := h.svc.UploadPhotos(c.Request.Context(), uint(albumId), urls)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, photos)
}

// DeletePhoto 删除照片
// DELETE /api/admin/photos/:id
func (h *PhotoHandler) DeletePhoto(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的照片ID"))
		return
	}
	if err := h.svc.DeletePhoto(c.Request.Context(), uint(id)); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "照片已删除")
}
