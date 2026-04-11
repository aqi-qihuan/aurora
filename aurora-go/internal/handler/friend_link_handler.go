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

// FriendLinkHandler 友链管理处理器（对标 Java FriendLinkController）
type FriendLinkHandler struct {
	svc *service.FriendLinkService
}

func NewFriendLinkHandler(svc *service.FriendLinkService) *FriendLinkHandler {
	return &FriendLinkHandler{svc: svc}
}

// ListFriendLinks 获取友链列表（前台，仅审核通过的）
// GET /api/links
func (h *FriendLinkHandler) ListFriendLinks(c *gin.Context) {
	list, err := h.svc.GetApprovedLinks(c.Request.Context())
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, list)
}

// SaveFriendLink 申请友链（前台，需后台审核）
// POST /api/links
func (h *FriendLinkHandler) SaveFriendLink(c *gin.Context) {
	var friendLinkVO vo.FriendLinkVO
	if err := c.ShouldBindJSON(&friendLinkVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	userID, _ := c.Get("userId")
	uid := uint(0)
	if id, ok := userID.(uint); ok {
		uid = id
	}
	result, err := h.svc.ApplyFriendLink(c.Request.Context(), uid, friendLinkVO)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// ==================== 后台管理端点 ====================

// ListAdminFriendLinks 后台友链列表（含全部状态）
// GET /api/admin/links
func (h *FriendLinkHandler) ListAdminFriendLinks(c *gin.Context) {
	var condition dto.ConditionVO
	c.ShouldBindQuery(&condition)
	pageNum, pageSize := util.PageQuery(c)
	page := dto.PageVO{PageNum: pageNum, PageSize: pageSize}

	result, err := h.svc.ListAdminLinks(c.Request.Context(), condition, page)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// UpdateFriendLink 更新友链信息（后台）
// PUT /api/admin/links/:id
func (h *FriendLinkHandler) UpdateFriendLink(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的友链ID"))
		return
	}
	var friendLinkVO vo.FriendLinkVO
	if err := c.ShouldBindJSON(&friendLinkVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	if err := h.svc.UpdateFriendLink(c.Request.Context(), uint(id), friendLinkVO); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, nil)
}

// ReviewFriendLink 审核友链申请
// PUT /api/admin/links/:id/review
func (h *FriendLinkHandler) ReviewFriendLink(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的友链ID"))
		return
	}
	var body struct {
		IsApproved bool `json:"isApproved"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	status := int8(1) // 通过
	if !body.IsApproved {
		status = -1 // 拒绝
	}
	if err := h.svc.ReviewFriendLink(c.Request.Context(), uint(id), status); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "审核完成")
}

// DeleteFriendLink 删除友链
// DELETE /api/admin/links/:id
func (h *FriendLinkHandler) DeleteFriendLink(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的友链ID"))
		return
	}
	if err := h.svc.DeleteFriendLink(c.Request.Context(), uint(id)); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "友链已删除")
}

// ToggleOnline 切换友链上线/下线状态
// PUT /api/admin/links/:id/toggle
func (h *FriendLinkHandler) ToggleOnline(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的友链ID"))
		return
	}
	var body struct {
		Online bool `json:"online"`
	}
	c.ShouldBindJSON(&body)
	if err := h.svc.SetFriendLinkOnline(c.Request.Context(), uint(id), body.Online); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, nil)
}
