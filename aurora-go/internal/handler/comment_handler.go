package handler

import (
	"io"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"golang.org/x/exp/slog"

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

// ListComments 获取评论列表（前台，分页，对标 Java getComments）
// GET /api/comments?type=5&topicId=123&current=1&size=7
func (h *CommentHandler) ListComments(c *gin.Context) {
	// 对标Java: CommentVO commentVO 作为Query参数绑定
	var commentVO vo.CommentVO
	if err := c.ShouldBindQuery(&commentVO); err != nil {
		// Query参数可能为空，给默认值
		commentVO.Type = 1
		commentVO.Current = 1
		commentVO.Size = 10
	}

	result, err := h.svc.ListComments(c.Request.Context(), commentVO)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
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
	userID, _ := c.Get("user_id")
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
	userID, _ := c.Get("user_id")
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

// UpdateCommentReview 审核评论（通过/拒绝）- 支持批量
// PUT /api/admin/comments/review
func (h *CommentHandler) UpdateCommentReview(c *gin.Context) {
	var reviewVO struct {
		IsReview int8   `json:"isReview"`
		IDs      []uint `json:"ids"`
	}
	if err := c.ShouldBindJSON(&reviewVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("参数错误"))
		return
	}
	if len(reviewVO.IDs) == 0 {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("请选择要审核的评论"))
		return
	}

	if err := h.svc.BatchReviewComments(c.Request.Context(), reviewVO.IDs, reviewVO.IsReview); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.SuccessWithMessage(c, "审核成功", nil)
}

// DeleteComment 删除评论
// DELETE /api/admin/comments
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	var ids []uint

	// 1. 直接读取 Body 内容 (解决 Gin ShouldBindJSON 在 DELETE 请求中可能失效的问题)
	body, err := io.ReadAll(c.Request.Body)
	if err == nil && len(body) > 0 {
		// 尝试 1: 解析 { "data": [...] } 或 { "ids": [...] }
		var wrapper map[string]interface{}
		if json.Unmarshal(body, &wrapper) == nil {
			if raw, ok := wrapper["data"]; ok {
				if arr, ok := raw.([]interface{}); ok {
					for _, v := range arr {
						if id, ok := v.(float64); ok {
							ids = append(ids, uint(id))
						}
					}
				}
			} else if raw, ok := wrapper["ids"]; ok {
				if arr, ok := raw.([]interface{}); ok {
					for _, v := range arr {
						if id, ok := v.(float64); ok {
							ids = append(ids, uint(id))
						}
					}
				}
			}
		} else {
			// 尝试 2: 直接解析为数组 [id1, id2, ...]
			var directIDs []interface{}
			if json.Unmarshal(body, &directIDs) == nil {
				for _, v := range directIDs {
					if id, ok := v.(float64); ok {
						ids = append(ids, uint(id))
					}
				}
			}
		}
	}

	// 2. 兼容 Query 参数 (适配 ?ids=1,2,3)
	if len(ids) == 0 {
		idsStr := c.Query("ids")
		if idsStr != "" {
			parts := strings.Split(idsStr, ",")
			for _, p := range parts {
				id, err := strconv.ParseUint(strings.TrimSpace(p), 10, 64)
				if err == nil {
					ids = append(ids, uint(id))
				}
			}
		}
	}

	if len(ids) == 0 {
		slog.Warn("删除评论失败: 未获取到ID", "path", c.FullPath(), "body", string(body))
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("请选择要删除的评论"))
		return
	}

	if err := h.svc.BatchDeleteComments(c.Request.Context(), ids); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.SuccessWithMessage(c, "评论已删除", nil)
}

// ListTopSixComments 获取前6条最新评论（用于侧边栏）
// GET /api/comments/topSix
func (h *CommentHandler) ListTopSixComments(c *gin.Context) {
	list, err := h.svc.GetLatestComments(c.Request.Context(), 6)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, list)
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
