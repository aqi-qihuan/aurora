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

// @Summary 获取相册列表
// @Tags 相册
// @Description 获取前台公开相册列表
// @Accept json
// @Produce json
// @Success 200 {object} object "成功返回相册列表"
// @Failure 500 {object} object "服务器内部错误"
// @Router /api/photos/albums [get]
func (h *PhotoAlbumHandler) ListAlbums(c *gin.Context) {
	list, err := h.svc.GetAlbums(c.Request.Context())
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, list)
}

// @Summary 获取相册详情
// @Tags 相册
// @Description 根据ID获取相册详情
// @Accept json
// @Produce json
// @Param id path int true "相册ID"
// @Success 200 {object} object "成功返回相册详情"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/photos/albums/{id}/info [get]
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

// @Summary 保存/更新相册
// @Tags 相册
// @Description 后台保存或更新相册
// @Accept json
// @Produce json
// @Param albumVO body vo.PhotoAlbumVO true "相册信息"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/photos/albums [post]
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

// @Summary 删除相册
// @Tags 相册
// @Description 后台删除相册
// @Accept json
// @Produce json
// @Param id path int true "相册ID"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/photos/albums/{id} [delete]
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

// @Summary 后台相册管理列表
// @Tags 相册
// @Description 后台获取相册列表（分页）
// @Accept json
// @Produce json
// @Param keyword query string false "搜索关键词"
// @Param current query int false "当前页码"
// @Param size query int false "每页数量"
// @Success 200 {object} object "成功返回相册列表"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/photos/albums [get]
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

// @Summary 获取后台相册列表信息
// @Tags 相册
// @Description 获取后台相册列表信息（用于下拉选择/移动照片）
// @Accept json
// @Produce json
// @Success 200 {object} object "成功返回相册列表"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/photos/albums/info [get]
func (h *PhotoAlbumHandler) ListAlbumInfos(c *gin.Context) {
	list, err := h.svc.GetAlbumInfos(c.Request.Context())
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, list)
}

// @Summary 上传相册封面
// @Tags 相册
// @Description 上传相册封面图片
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "封面图片文件"
// @Success 200 {object} object "成功返回图片URL"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/photos/albums/upload [post]
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
