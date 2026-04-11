package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/util"
)

// CommentHandler 评论处理器（对标 Java CommentController）
// 端点: 8个 (评论CRUD + 回复 + 审核 + 通知)
type CommentHandler struct {
	// commentService service.CommentService
}

func NewCommentHandler() *CommentHandler { return &CommentHandler{} }

// ListComments 获取文章评论列表（前台）
// GET /api/articles/:articleId/comments
// 对标 CommentController.listCommentsByArticleId()
func (h *CommentHandler) ListComments(c *gin.Context) {
	articleIdStr := c.Param("articleId")
	_, err := strconv.ParseInt(articleIdStr, 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的文章ID"))
		return
	}
	pageNum, pageSize := util.PageQuery(c)
	util.ResponseSuccessWithPage(c, []interface{}{}, int64(0), pageNum, pageSize)
}

// AddComment 发表评论/回复（支持嵌套回复）
// POST /api/articles/:articleId/comments
// 对标 CommentController.saveComment() - 含嵌套逻辑
func (h *CommentHandler) AddComment(c *gin.Context) {
	var commentVO dto.CommentVO
	if err := c.ShouldBindJSON(&commentVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	_ = commentVO // TODO: P0-5 保存评论 → 异步发送MQ邮件通知

	zap.L().Info("New comment added", zap.Any("comment", commentVO))
	util.ResponseSuccess(c, "评论发表成功")
}

// ReplyComment 回复指定评论
// POST /api/comments/:id/reply
func (h *CommentHandler) ReplyComment(c *gin.Context) {
	var replyVO dto.ReplyVO
	if err := c.ShouldBindJSON(&replyVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	_ = replyVO

	util.ResponseSuccess(c, "回复成功")
}

// LikeComment 点赞评论
// POST /api/comments/:id/like
func (h *CommentHandler) LikeComment(c *gin.Context) {
	idStr := c.Param("id")
	_, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的评论ID"))
		return
	}
	// TODO: P0-5 Redis ZINCRBY 计数 或 DB更新

	util.ResponseSuccess(c, map[string]interface{}{"likes": 1})
}

// ==================== 后台管理端点 ====================

// ListAdminComments 后台评论列表（含审核状态筛选）
// GET /api/admin/comments
func (h *CommentHandler) ListAdminComments(c *gin.Context) {
	var condition dto.ConditionVO
	c.ShouldBindQuery(&condition)
	_ = condition
	pageNum, pageSize := util.PageQuery(c)
	util.ResponseSuccessWithPage(c, []interface{}{}, int64(0), pageNum, pageSize)
}

// UpdateCommentReview 审核评论（通过/拒绝）
// PUT /api/admin/comments/:id/review
// 对标 CommentController.updateCommentReview()
func (h *CommentHandler) UpdateCommentReview(c *gin.Context) {
	var reviewVO dto.ReviewVO
	if err := c.ShouldBindJSON(&reviewVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	_ = reviewVO

	util.ResponseSuccess(c, nil)
}

// DeleteComment 删除评论
// DELETE /api/admin/comments/:ids
// 对标 CommentController.deleteComments()
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	idsStr := c.Query("ids") // 支持批量删除: "1,2,3"
	if idsStr == "" {
		idStr := c.Param("id")
		idsStr = idStr
	}
	_ = idsStr // TODO: P0-5 批量软删除

	zap.L().Debug("Delete comments", zap.String("ids", idsStr))
	util.ResponseSuccess(c, "评论已删除")
}

// GetCommentStats 获取评论统计信息（总数/待审核数）
// GET /api/admin/comments/stats
func (h *CommentHandler) GetCommentStats(c *gin.Context) {
	util.ResponseSuccess(c, map[string]interface{}{
		"total":       0,
		"pendingReview": 0,
		"approved":    0,
		"rejected":    0,
	})
}
