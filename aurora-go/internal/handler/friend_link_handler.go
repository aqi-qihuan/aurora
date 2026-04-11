package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/util"
)

// FriendLinkHandler 友链管理处理器（对标 Java FriendLinkController）
type FriendLinkHandler struct {
	// friendLinkService service.FriendLinkService
}

func NewFriendLinkHandler() *FriendLinkHandler { return &FriendLinkHandler{} }

// ListFriendLinks 获取友链列表（前台，仅审核通过的）
// GET /api/links
func (h *FriendLinkHandler) ListFriendLinks(c *gin.Context) {
	// TODO: P0-5 查询 status=APPROVED 的友链
	util.ResponseSuccess(c, []interface{}{})
}

// SaveFriendLink 申请友链（前台，需后台审核）
// POST /api/links
// 对标 FriendLinkController.saveLink() - 申请模式
func (h *FriendLinkHandler) SaveFriendLink(c *gin.Context) {
	var friendLinkVO dto.FriendLinkVO
	if err := c.ShouldBindJSON(&friendLinkVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	_ = friendLinkVO // TODO: P0-5 保存为PENDING状态

	zap.L().Info("New friend link application received", zap.Any("link", friendLinkVO))
	util.ResponseSuccess(c, "友链申请已提交，请等待审核")
}

// ==================== 后台管理端点 ====================

// ListAdminFriendLinks 后台友链列表（含全部状态）
// GET /api/admin/links
func (h *FriendLinkHandler) ListAdminFriendLinks(c *gin.Context) {
	pageNum, pageSize := util.PageQuery(c)
	util.ResponseSuccessWithPage(c, []interface{}{}, int64(0), pageNum, pageSize)
}

// UpdateFriendLink 更新友链信息（后台）
// PUT /api/admin/links/:id
func (h *FriendLinkHandler) UpdateFriendLink(c *gin.Context) {
	var friendLinkVO dto.FriendLinkVO
	if err := c.ShouldBindJSON(&friendLinkVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	_ = friendLinkVO
	util.ResponseSuccess(c, nil)
}

// ReviewFriendLink 审核友链申请
// PUT /api/admin/links/:id/review
// 对标 FriendLinkController.reviewLink()
func (h *FriendLinkHandler) ReviewFriendLink(c *gin.Context) {
	var reviewVO dto.ReviewVO
	if err := c.ShouldBindJSON(&reviewVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}

	idStr := c.Param("id")
	_, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的友链ID"))
		return
	}

	zap.L().Debug("Review friend link", zap.String("id", idStr), zap.Bool("isApproved", reviewVO.IsApproved))
	util.ResponseSuccess(c, "审核完成")
}

// DeleteFriendLink 删除友链
// DELETE /api/admin/links/:id
func (h *FriendLinkHandler) DeleteFriendLink(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	zap.L().Debug("Delete friend link", zap.Int64("id", id))
	util.ResponseSuccess(c, "友链已删除")
}

// ToggleOnline 切换友链上线/下线状态
// PUT /api/admin/links/:id/toggle
func (h *FriendLinkHandler) ToggleOnline(c *gin.Context) {
	idStr := c.Param("id")
	_, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的友链ID"))
		return
	}
	util.ResponseSuccess(c, nil)
}
