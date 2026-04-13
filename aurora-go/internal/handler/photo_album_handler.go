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
	"github.com/aurora-go/aurora/internal/vo"
)

// PhotoAlbumHandler 相册管理处理器（对标 Java AlbumController）
type PhotoAlbumHandler struct {
	svc       *service.PhotoAlbumService
	uploadSvc *strategy.UploadService
}

func NewPhotoAlbumHandler(svc *service.PhotoAlbumService, uploadSvc *strategy.UploadService) *PhotoAlbumHandler {
	return &PhotoAlbumHandler{svc: svc, uploadSvc: uploadSvc}
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
// GET /api/admin/photos/albums/:id/info
func (h *PhotoAlbumHandler) GetAlbumById(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的相册ID"))
		return
	}
	result, err := h.svc.GetAlbumById(c.Request.Context(), uint(id))
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// SaveOrUpdate 保存/更新相册（后台）
// POST /api/admin/photos/albums
func (h *PhotoAlbumHandler) SaveOrUpdate(c *gin.Context) {
	var albumVO vo.PhotoAlbumVO
	if err := c.ShouldBindJSON(&albumVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}

	// 对标Java saveOrUpdatePhotoAlbum，根据ID判断新增/更新
	if err := h.svc.SaveOrUpdateAlbum(c.Request.Context(), albumVO); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "操作成功")
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

// ListAdminAlbums 后台相册管理列表
// GET /api/admin/photos/albums
func (h *PhotoAlbumHandler) ListAdminAlbums(c *gin.Context) {
	var condition dto.ConditionVO
	c.ShouldBindQuery(&condition)
	pageNum, pageSize := util.PageQuery(c)
	page := dto.PageVO{PageNum: pageNum, PageSize: pageSize}

	result, err := h.svc.GetAdminAlbums(c.Request.Context(), condition, page)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// ListAlbumInfos 获取后台相册列表信息（用于下拉选择/移动照片，对标Java listPhotoAlbumInfosAdmin）
// GET /api/admin/photos/albums/info
func (h *PhotoAlbumHandler) ListAlbumInfos(c *gin.Context) {
	list, err := h.svc.GetAlbumInfos(c.Request.Context())
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, list)
}

// UploadAlbumCover 上传相册封面（对标Java PhotoAlbumController.savePhotoAlbumCover）
// POST /api/admin/photos/albums/upload
func (h *PhotoAlbumHandler) UploadAlbumCover(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("请选择要上传的封面"))
		return
	}

	// 打开文件获取输入流
	src, err := file.Open()
	if err != nil {
		util.ResponseError(c, errors.ErrFileUploadFailed.WithMsg("打开文件失败"))
		return
	}
	defer src.Close()

	// 读取文件内容到内存（对标Java AbstractUploadStrategyImpl.uploadFile）
	data, err := io.ReadAll(src)
	if err != nil {
		util.ResponseError(c, errors.ErrFileUploadFailed.WithMsg("读取文件失败"))
		return
	}

	// 调用上传服务（内部含MD5去重 + MinIO上传）
	url, err := h.uploadSvc.UploadAlbumCover(c.Request.Context(), data, file.Filename)
	if err != nil {
		util.ResponseError(c, err)
		return
	}

	// 返回访问URL
	util.ResponseSuccess(c, url)
}
