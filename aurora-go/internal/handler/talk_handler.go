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

// ListTalks 获取说说列表（前台，按时间倒序）
// GET /api/talks
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

// GetTalkById 获取说说详情（前台）
// GET /api/talks/:id
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

// SaveOrUpdate 保存或更新说说（后台，统一接口）
// POST /api/admin/talks
// 对标Java版：根据ID有无自动判断新增/更新
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

// DeleteTalks 批量删除说说（后台）
// DELETE /api/admin/talks
// 对标Java版：接收ID数组批量删除
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

// ListAdminTalks 后台说说列表
// GET /api/admin/talks
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

// GetAdminTalkById 后台获取说说详情
// GET /api/admin/talks/:id
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

// UploadTalkImage 上传说说图片
// POST /api/admin/talks/images
// 对标Java: uploadStrategyContext.executeUploadStrategy(file, FilePathEnum.TALK.getPath())
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
