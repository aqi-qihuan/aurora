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

// @Summary 获取友链列表
// @Tags 友链
// @Description 获取前台友链列表
// @Accept json
// @Produce json
// @Success 200 {object} object "成功返回友链列表"
// @Failure 500 {object} object "服务器内部错误"
// @Router /api/links [get]
func (h *FriendLinkHandler) ListFriendLinks(c *gin.Context) {
	list, err := h.svc.ListFriendLinks(c.Request.Context())
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, list)
}

// @Summary 新增或更新友链
// @Tags 友链
// @Description 后台新增或更新友链信息
// @Accept json
// @Produce json
// @Param friendLinkVO body vo.FriendLinkVO true "友链信息"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/links [post]
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

// @Summary 后台友链列表
// @Tags 友链
// @Description 后台获取友链列表（分页，含全部状态）
// @Accept json
// @Produce json
// @Param keyword query string false "搜索关键词"
// @Param current query int false "当前页码"
// @Param size query int false "每页数量"
// @Success 200 {object} object "成功返回友链列表"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/links [get]
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

// @Summary 更新友链信息
// @Tags 友链
// @Description 后台更新友链信息
// @Accept json
// @Produce json
// @Param friendLinkVO body vo.FriendLinkVO true "友链信息"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/links [put]
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

// @Summary 批量删除友链
// @Tags 友链
// @Description 后台批量删除友链
// @Accept json
// @Produce json
// @Param ids body []uint true "友链ID列表"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/links [delete]
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
