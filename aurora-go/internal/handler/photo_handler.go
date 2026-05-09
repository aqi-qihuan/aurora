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

// @Summary 后台照片管理列表
// @Tags 照片
// @Description 后台获取照片列表（分页）
// @Accept json
// @Produce json
// @Param keyword query string false "搜索关键词"
// @Param albumId query int false "相册ID"
// @Param current query int false "当前页码"
// @Param size query int false "每页数量"
// @Success 200 {object} object "成功返回照片列表"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/photos [get]
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

// @Summary 根据相册ID获取照片列表
// @Tags 照片
// @Description 根据相册ID获取照片列表
// @Accept json
// @Produce json
// @Param albumId path int true "相册ID"
// @Success 200 {object} object "成功返回照片列表"
// @Failure 400 {object} object "请求参数错误"
// @Failure 500 {object} object "服务器内部错误"
// @Router /api/albums/{albumId}/photos [get]
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

// @Summary 上传照片
// @Tags 照片
// @Description 上传照片文件
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "照片文件"
// @Success 200 {object} object "成功返回照片URL"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/photos/upload [post]
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

// @Summary 批量删除照片
// @Tags 照片
// @Description 后台批量删除照片
// @Accept json
// @Produce json
// @Param ids body []uint true "照片ID列表"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/photos [delete]
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

// @Summary 保存照片
// @Tags 照片
// @Description 保存照片到指定相册
// @Accept json
// @Produce json
// @Param body body object true "相册ID和照片URL列表(albumId, photoUrls)"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/photos [post]
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

// @Summary 更新照片信息
// @Tags 照片
// @Description 更新照片名称等信息
// @Accept json
// @Produce json
// @Param body body object true "照片ID和名称(id, photoName)"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/photos [put]
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

// @Summary 移动照片到其他相册
// @Tags 照片
// @Description 批量移动照片到指定相册
// @Accept json
// @Produce json
// @Param body body object true "目标相册ID和照片ID列表(albumId, photoIds)"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/photos/album [put]
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

// @Summary 更新照片删除状态
// @Tags 照片
// @Description 批量更新照片删除状态（逻辑删除/恢复）
// @Accept json
// @Produce json
// @Param body body object true "照片ID列表和删除状态(ids, isDelete)"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/photos/delete [put]
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
