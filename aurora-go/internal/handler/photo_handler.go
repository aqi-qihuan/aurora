package handler

import (
	"io"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/service"
	"github.com/aurora-go/aurora/internal/strategy"
	"github.com/aurora-go/aurora/internal/util"
)

// PhotoHandler 相册照片处理器（对标 Java PhotoController）
type PhotoHandler struct {
	svc       *service.PhotoService
	uploadSvc *strategy.UploadService
}

func NewPhotoHandler(svc *service.PhotoService, uploadSvc *strategy.UploadService) *PhotoHandler {
	return &PhotoHandler{svc: svc, uploadSvc: uploadSvc}
}

// ListPhotos 获取相册下的照片列表（后台管理用，分页）
// GET /api/admin/photos
func (h *PhotoHandler) ListAdminPhotos(c *gin.Context) {
	var condition dto.ConditionVO
	c.ShouldBindQuery(&condition)
	pageNum, pageSize := util.PageQuery(c)
	page := dto.PageVO{PageNum: pageNum, PageSize: pageSize}

	result, err := h.svc.ListAdminPhotos(c.Request.Context(), condition, page)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// ListPhotosByAlbumId 根据相册id查看照片列表（前台）
// GET /api/albums/:albumId/photos
func (h *PhotoHandler) ListPhotosByAlbumId(c *gin.Context) {
	albumId, err := strconv.ParseUint(c.Param("albumId"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的相册ID"))
		return
	}
	result, err := h.svc.ListPhotosByAlbumId(c.Request.Context(), uint(albumId))
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// UploadPhoto 上传照片（对标Java PhotoController.savePhotoAlbumCover）
// POST /api/admin/photos/upload
func (h *PhotoHandler) UploadPhoto(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("请选择要上传的照片"))
		return
	}

	// 打开文件获取输入流
	src, err := file.Open()
	if err != nil {
		util.ResponseError(c, errors.ErrFileUploadFailed.WithMsg("打开文件失败"))
		return
	}
	defer src.Close()

	// 读取文件内容
	data, err := io.ReadAll(src)
	if err != nil {
		util.ResponseError(c, errors.ErrFileUploadFailed.WithMsg("读取文件失败"))
		return
	}

	// 调用上传服务（MD5去重 + MinIO上传）
	url, err := h.uploadSvc.UploadPhoto(c.Request.Context(), data, file.Filename)
	if err != nil {
		util.ResponseError(c, err)
		return
	}

	// 返回访问URL
	util.ResponseSuccess(c, url)
}

// DeletePhotos 批量删除照片
// DELETE /api/admin/photos
func (h *PhotoHandler) DeletePhotos(c *gin.Context) {
	var ids []uint
	if err := c.ShouldBindJSON(&ids); err != nil || len(ids) == 0 {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("请选择要删除的照片"))
		return
	}
	if err := h.svc.DeletePhotos(c.Request.Context(), ids); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "照片已删除")
}

// ListAdminPhotos 后台照片管理列表
// GET /api/admin/photos

// SavePhotos 保存照片
// POST /api/admin/photos
func (h *PhotoHandler) SavePhotos(c *gin.Context) {
	var body struct {
		AlbumIDStr string   `json:"albumId" binding:"required"`
		PhotoURLs  []string `json:"photoUrls" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}

	albumId, err := strconv.ParseUint(body.AlbumIDStr, 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的相册ID格式"))
		return
	}

	if err := h.svc.SavePhotos(c.Request.Context(), uint(albumId), body.PhotoURLs); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "照片保存成功")
}

// UpdatePhoto 更新照片信息
// PUT /api/admin/photos
func (h *PhotoHandler) UpdatePhoto(c *gin.Context) {
	var body struct {
		ID       uint   `json:"id" binding:"required"`
		PhotoName string `json:"photoName"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	if err := h.svc.UpdatePhoto(c.Request.Context(), body.ID, body.PhotoName); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "照片信息已更新")
}

// MovePhotosAlbum 移动照片到其他相册
// PUT /api/admin/photos/album
func (h *PhotoHandler) MovePhotosAlbum(c *gin.Context) {
	var body struct {
		AlbumIDStr string   `json:"albumId" binding:"required"`
		PhotoIDs   []uint   `json:"photoIds" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}

	albumId, err := strconv.ParseUint(body.AlbumIDStr, 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的相册ID格式"))
		return
	}

	if err := h.svc.UpdatePhotosAlbum(c.Request.Context(), uint(albumId), body.PhotoIDs); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "照片已移动")
}

// UpdatePhotoDelete 更新照片删除状态（逻辑删除/恢复）
// PUT /api/admin/photos/delete
func (h *PhotoHandler) UpdatePhotoDelete(c *gin.Context) {
	var body struct {
		IDs      []uint `json:"ids" binding:"required"`
		IsDelete int8   `json:"isDelete" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	if err := h.svc.UpdatePhotoDelete(c.Request.Context(), body.IDs, body.IsDelete); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "照片状态已更新")
}
