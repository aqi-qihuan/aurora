package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/service"
	"github.com/aurora-go/aurora/internal/util"
	"github.com/aurora-go/aurora/internal/vo"
)

// PhotoAlbumHandler 相册管理处理器（对标 Java AlbumController）
type PhotoAlbumHandler struct {
	svc *service.PhotoAlbumService
}

func NewPhotoAlbumHandler(svc *service.PhotoAlbumService) *PhotoAlbumHandler {
	return &PhotoAlbumHandler{svc: svc}
}

// ListAlbums 获取相册列表（前台，公开相册）
// GET /api/albums
func (h *PhotoAlbumHandler) ListAlbums(c *gin.Context) {
	list, err := h.svc.GetAlbums(c.Request.Context())
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, list)
}

// GetAlbumById 获取相册详情
// GET /api/albums/:id
func (h *PhotoAlbumHandler) GetAlbumById(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的相册ID"))
		return
	}
	// 后台使用分页查询获取详情，前台直接获取公开相册
	var condition dto.ConditionVO
	c.ShouldBindQuery(&condition)
	pageNum, pageSize := util.PageQuery(c)
	page := dto.PageVO{PageNum: pageNum, PageSize: pageSize}

	result, err := h.svc.GetAdminAlbums(c.Request.Context(), condition, page)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	_ = id
	util.ResponseSuccess(c, result)
}

// SaveOrUpdate 保存/更新相册（后台）
// POST /api/admin/albums (新增)
// PUT /api/admin/albums/:id (更新)
func (h *PhotoAlbumHandler) SaveOrUpdate(c *gin.Context) {
	var albumVO vo.PhotoAlbumVO
	if err := c.ShouldBindJSON(&albumVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}

	idStr := c.Param("id")
	if idStr != "" {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的相册ID"))
			return
		}
		if err := h.svc.UpdateAlbum(c.Request.Context(), uint(id), albumVO); err != nil {
			util.ResponseError(c, err)
			return
		}
		util.ResponseSuccess(c, nil)
		return
	}

	result, err := h.svc.CreateAlbum(c.Request.Context(), albumVO)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// DeleteAlbum 删除相册（后台）
// DELETE /api/admin/albums/:id
func (h *PhotoAlbumHandler) DeleteAlbum(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的相册ID"))
		return
	}
	if err := h.svc.DeleteAlbum(c.Request.Context(), uint(id)); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "相册已删除")
}
