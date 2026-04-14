package handler

import (
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

// ListFriendLinks 获取友链列表（前台，对标Java: listFriendLinks）
// GET /api/links
func (h *FriendLinkHandler) ListFriendLinks(c *gin.Context) {
	list, err := h.svc.ListFriendLinks(c.Request.Context())
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, list)
}

// SaveOrUpdateFriendLink 新增或更新友链（对标Java: saveOrUpdateFriendLink）
// POST /api/admin/links
func (h *FriendLinkHandler) SaveOrUpdateFriendLink(c *gin.Context) {
	var friendLinkVO vo.FriendLinkVO
	if err := c.ShouldBindJSON(&friendLinkVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	if err := h.svc.SaveOrUpdateFriendLink(c.Request.Context(), friendLinkVO); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, nil)
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
// PUT /api/admin/links
func (h *FriendLinkHandler) UpdateFriendLink(c *gin.Context) {
	var friendLinkVO vo.FriendLinkVO
	if err := c.ShouldBindJSON(&friendLinkVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	if err := h.svc.SaveOrUpdateFriendLink(c.Request.Context(), friendLinkVO); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, nil)
}

// DeleteFriendLink 批量删除友链
// DELETE /api/admin/links
// Java: @DeleteMapping("/admin/links") public ResultVO<?> deleteFriendLink(@RequestBody List<Integer> linkIdList)
// 前端 axios 发送的 body 是原始数组 [id1, id2, ...]
func (h *FriendLinkHandler) DeleteFriendLink(c *gin.Context) {
	var ids []uint
	if err := c.ShouldBindJSON(&ids); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("请提供要删除的友链ID列表"))
		return
	}
	if len(ids) == 0 {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("友链ID列表不能为空"))
		return
	}
	if err := h.svc.DeleteFriendLinks(c.Request.Context(), ids); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "友链已删除")
}
