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

// TalkHandler 说说处理器（完全对标 Java TalkController）
type TalkHandler struct {
	svc     *service.TalkService
	fileSvc *service.FileService
}

func NewTalkHandler(svc *service.TalkService, fileSvc *service.FileService) *TalkHandler {
	return &TalkHandler{svc: svc, fileSvc: fileSvc}
}

// @Summary 获取说说列表
// @Tags 说说
// @Description 获取前台说说列表（分页，按时间倒序）
// @Accept json
// @Produce json
// @Param current query int false "当前页码"
// @Param size query int false "每页数量"
// @Success 200 {object} object "成功返回说说列表"
// @Failure 500 {object} object "服务器内部错误"
// @Router /api/talks [get]
func (h *TalkHandler) ListTalks(c *gin.Context) {
	pageNum, pageSize := util.PageQuery(c)
	page := dto.PageVO{PageNum: pageNum, PageSize: pageSize}

	result, err := h.svc.GetTalks(c.Request.Context(), page)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// @Summary 获取说说详情
// @Tags 说说
// @Description 根据ID获取说说详情
// @Accept json
// @Produce json
// @Param id path int true "说说ID"
// @Success 200 {object} object "成功返回说说详情"
// @Failure 400 {object} object "请求参数错误"
// @Failure 500 {object} object "服务器内部错误"
// @Router /api/talks/{id} [get]
func (h *TalkHandler) GetTalkById(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的说说ID"))
		return
	}
	result, err := h.svc.GetTalkByID(c.Request.Context(), uint(id))
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// @Summary 保存或更新说说
// @Tags 说说
// @Description 后台保存或更新说说
// @Accept json
// @Produce json
// @Param talkVO body vo.TalkVO true "说说内容"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/talks [post]
func (h *TalkHandler) SaveOrUpdate(c *gin.Context) {
	var talkVO vo.TalkVO
	if err := c.ShouldBindJSON(&talkVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}

	// 从Context获取用户信息ID（t_talk.user_id 应关联 t_user_info.id，不是 t_user_auth.id）
	// JWTAuthEnhanced注入了 user_info_id = UserInfo.id
	var userInfoID uint
	if uid, exists := c.Get("user_info_id"); exists {
		switch v := uid.(type) {
		case float64:
			userInfoID = uint(v)
		case int64:
			userInfoID = uint(v)
		case int:
			userInfoID = uint(v)
		case uint:
			userInfoID = v
		}
	}
	if userInfoID == 0 {
		util.ResponseError(c, errors.ErrUnauthorized.WithMsg("请先登录"))
		return
	}

	if err := h.svc.SaveOrUpdateTalk(c.Request.Context(), userInfoID, talkVO); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, nil)
}

// @Summary 批量删除说说
// @Tags 说说
// @Description 后台批量删除说说
// @Accept json
// @Produce json
// @Param ids body []uint true "说说ID列表"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/talks [delete]
func (h *TalkHandler) DeleteTalks(c *gin.Context) {
	var ids []uint
	if err := c.ShouldBindJSON(&ids); err != nil || len(ids) == 0 {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("请选择要删除的说说"))
		return
	}

	if err := h.svc.DeleteTalks(c.Request.Context(), ids); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "说说已删除")
}

// @Summary 后台说说列表
// @Tags 说说
// @Description 后台获取说说列表（分页）
// @Accept json
// @Produce json
// @Param keyword query string false "搜索关键词"
// @Param current query int false "当前页码"
// @Param size query int false "每页数量"
// @Success 200 {object} object "成功返回说说列表"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/talks [get]
func (h *TalkHandler) ListAdminTalks(c *gin.Context) {
	var condition dto.ConditionVO
	c.ShouldBindQuery(&condition)
	pageNum, pageSize := util.PageQuery(c)
	page := dto.PageVO{PageNum: pageNum, PageSize: pageSize}

	result, err := h.svc.ListAdminTalks(c.Request.Context(), condition, page)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// @Summary 后台获取说说详情
// @Tags 说说
// @Description 后台根据ID获取说说详情
// @Accept json
// @Produce json
// @Param id path int true "说说ID"
// @Success 200 {object} object "成功返回说说详情"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/talks/{id} [get]
func (h *TalkHandler) GetAdminTalkById(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的说说ID"))
		return
	}
	result, err := h.svc.GetAdminTalkByID(c.Request.Context(), uint(id))
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// @Summary 上传说说图片
// @Tags 说说
// @Description 上传说说中的图片
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "图片文件"
// @Success 200 {object} object "成功返回图片URL"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/talks/images [post]
func (h *TalkHandler) UploadTalkImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("请选择要上传的图片"))
		return
	}

	// 调用FileService上传到MinIO/本地存储
	url, err := h.fileSvc.UploadTalkImage(c.Request.Context(), file)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, url)
}
