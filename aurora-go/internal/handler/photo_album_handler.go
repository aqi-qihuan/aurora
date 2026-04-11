package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/util"
)

// PhotoAlbumHandler 相册管理处理器（对标 Java AlbumController）
type PhotoAlbumHandler struct {
	// photoAlbumService service.PhotoAlbumService
}

func NewPhotoAlbumHandler() *PhotoAlbumHandler { return &PhotoAlbumHandler{} }

// ListAlbums 获取相册列表（前台，公开相册）
// GET /api/albums
func (h *PhotoAlbumHandler) ListAlbums(c *gin.Context) {
	util.ResponseSuccess(c, []interface{}{})
}

// GetAlbumById 获取相册详情（含照片预览）
// GET /api/albums/:id
func (h *PhotoAlbumHandler) GetAlbumById(c *gin.Context) {
	idStr := c.Param("id")
	_, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的相册ID"))
		return
	}

	// TODO: P0-5 查询相册 + 照片列表 + 是否私密(需登录)
	util.ResponseSuccess(c, nil)
}

// SaveOrUpdate 保存/更新相册（后台）
// POST /api/admin/albums (新增)
// PUT /api/admin/albums/:id (更新)
func (h *PhotoAlbumHandler) SaveOrUpdate(c *gin.Context) {
	var albumVO dto.PhotoAlbumVO
	if err := c.ShouldBindJSON(&albumVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	_ = albumVO

	zap.L().Debug("Save album", zap.Any("album", albumVO))
	// TODO: P0-5 上传封面图到MinIO + 保存相册信息

	util.ResponseSuccess(c, nil)
}

// DeleteAlbum 删除相册（后台）
// DELETE /api/admin/albums/:id
func (h *PhotoAlbumHandler) DeleteAlbum(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	zap.L().Debug("Delete album", zap.Int64("id", id))
	// TODO: P0-5 删除相册下所有照片(MinIO+DB) + 删除相册记录

	util.ResponseSuccess(c, "相册已删除")
}
