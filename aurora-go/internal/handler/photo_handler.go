package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/util"
)

// PhotoHandler 相册照片处理器（对标 Java PhotoController）
type PhotoHandler struct {
	// photoService service.PhotoService
}

func NewPhotoHandler() *PhotoHandler { return &PhotoHandler{} }

// ListPhotos 获取相册下的照片列表
// GET /api/albums/:albumId/photos
func (h *PhotoHandler) ListPhotos(c *gin.Context) {
	albumIdStr := c.Param("albumId")
	_, err := strconv.ParseInt(albumIdStr, 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParam.WithMsg("无效的相册ID"))
		return
	}
	pageNum, pageSize := util.PageQuery(c)
	util.ResponseSuccessWithPage(c, []interface{}{}, int64(0), pageNum, pageSize)
}

// UploadPhoto 上传照片到指定相册（支持批量）
// POST /api/admin/albums/:albumId/photos
func (h *PhotoHandler) UploadPhoto(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParam.WithMsg("文件上传失败"))
		return
	}
	files := form.File["files"]
	if len(files) == 0 {
		util.ResponseError(c, errors.ErrInvalidParam.WithMsg("请选择要上传的照片"))
		return
	}

	zap.L().Info("Uploading photos", "count", len(files))
	// TODO: P0-5 遍历files → 上传MinIO → 保存Photo记录

	util.ResponseSuccess(c, map[string]interface{}{
		"uploadedCount": len(files),
		"urls":          []string{},
	})
}

// DeletePhoto 删除照片
// DELETE /api/admin/photos/:id
func (h *PhotoHandler) DeletePhoto(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	zap.L().Debug("Delete photo", zap.Int64("id", id))
	// TODO: P0-5 从MinIO删除 + DB删除记录

	util.ResponseSuccess(c, "照片已删除")
}
