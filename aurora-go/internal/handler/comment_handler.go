package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/service"
	"github.com/aurora-go/aurora/internal/util"
	"github.com/aurora-go/aurora/internal/vo"
)

// CommentHandler 评论处理器（对标 Java CommentController）
type CommentHandler struct {
	svc *service.CommentService
}

func NewCommentHandler(svc *service.CommentService) *CommentHandler {
	return &CommentHandler{svc: svc}
}

// ListComments 获取文章评论列表（前台）
// GET /api/articles/:id/comments
func (h *CommentHandler) ListComments(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的文章ID"))
		return
	}
	list, err := h.svc.GetCommentsByArticle(c.Request.Context(), uint(id))
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, list)
}

// AddComment 发表评论/回复
// POST /api/articles/:id/comments
func (h *CommentHandler) AddComment(c *gin.Context) {
	var commentVO vo.CommentVO
	if err := c.ShouldBindJSON(&commentVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	// 从上下文中获取用户ID
	userID, _ := c.Get("userId")
	uid := uint(0)
	if id, ok := userID.(uint); ok {
		uid = id
	}
	clientIP := c.ClientIP()

	result, err := h.svc.CreateComment(c.Request.Context(), uid, commentVO, clientIP)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// ReplyComment 回复指定评论
// POST /api/comments/:id/reply
func (h *CommentHandler) ReplyComment(c *gin.Context) {
	var replyVO vo.CommentVO
	if err := c.ShouldBindJSON(&replyVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	userID, _ := c.Get("userId")
	uid := uint(0)
	if id, ok := userID.(uint); ok {
		uid = id
	}
	if parentID, err := strconv.ParseUint(c.Param("id"), 10, 64); err == nil {
		replyVO.ParentID = uint(parentID)
	}
	clientIP := c.ClientIP()

	result, err := h.svc.CreateComment(c.Request.Context(), uid, replyVO, clientIP)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// LikeComment 点赞评论
// POST /api/comments/:id/like
func (h *CommentHandler) LikeComment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的评论ID"))
		return
	}
	if err := h.svc.LikeComment(c.Request.Context(), uint(id)); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "点赞成功")
}

// ==================== 后台管理端点 ====================

// ListAdminComments 后台评论列表（含审核状态筛选）
// GET /api/admin/comments
func (h *CommentHandler) ListAdminComments(c *gin.Context) {
	var condition dto.ConditionVO
	c.ShouldBindQuery(&condition)
	pageNum, pageSize := util.PageQuery(c)
	page := dto.PageVO{PageNum: pageNum, PageSize: pageSize}

	result, err := h.svc.ListAdminComments(c.Request.Context(), condition, page)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// UpdateCommentReview 审核评论（通过/拒绝）
// PUT /api/admin/comments/:id/review
func (h *CommentHandler) UpdateCommentReview(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的评论ID"))
		return
	}
	var reviewVO struct {
		IsReview int8 `json:"isReview"`
	}
	if err := c.ShouldBindJSON(&reviewVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	if err := h.svc.ReviewComment(c.Request.Context(), uint(id), reviewVO.IsReview); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, nil)
}

// DeleteComment 删除评论
// DELETE /api/admin/comments/:ids
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	idsStr := c.Param("ids")
	if idsStr == "" {
		idsStr = c.Query("ids")
	}
	if idsStr == "" {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("请选择要删除的评论"))
		return
	}
	parts := strings.Split(idsStr, ",")
	for _, p := range parts {
		id, err := strconv.ParseUint(strings.TrimSpace(p), 10, 64)
		if err != nil {
			continue
		}
		_ = h.svc.DeleteComment(c.Request.Context(), uint(id))
	}
	util.ResponseSuccess(c, "评论已删除")
}

// GetCommentStats 获取评论统计信息
// GET /api/admin/comments/stats
func (h *CommentHandler) GetCommentStats(c *gin.Context) {
	stats, err := h.svc.GetCommentStats(c.Request.Context())
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, stats)
}
