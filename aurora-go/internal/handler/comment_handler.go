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
	"github.com/aurora-go/aurora/internal/middleware"
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

// @Summary 获取评论列表
// @Tags 评论
// @Description 获取前台评论列表（分页）
// @Accept json
// @Produce json
// @Param type query int false "评论类型"
// @Param topicId query int false "主题ID"
// @Param current query int false "当前页码"
// @Param size query int false "每页数量"
// @Success 200 {object} object "成功返回评论列表"
// @Failure 500 {object} object "服务器内部错误"
// @Router /api/comments [get]
func (h *CommentHandler) ListComments(c *gin.Context) {
	// 对标Java: CommentVO commentVO 作为Query参数绑定
	var commentVO vo.CommentVO
	if err := c.ShouldBindQuery(&commentVO); err != nil {
		// Query参数可能为空，给默认值
		commentVO.Type = 1
		commentVO.Current = 1
		commentVO.Size = 10
	}
	
	// 兼容前端传字符串类型的topicId
	topicIdStr := c.Query("topicId")
	if topicIdStr != "" {
		if topicID, err := strconv.ParseUint(topicIdStr, 10, 64); err == nil {
			uid := uint(topicID)
			commentVO.TopicID = &uid
			slog.Info("查询评论列表，解析 topicId 成功", "type", commentVO.Type, "topicId", uid)
		}
	}

	result, err := h.svc.ListComments(c.Request.Context(), commentVO)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// @Summary 发表评论/回复
// @Tags 评论
// @Description 发表评论或回复评论（需要登录）
// @Accept json
// @Produce json
// @Param commentVO body vo.CommentVO true "评论内容"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/comments/save [post]
func (h *CommentHandler) AddComment(c *gin.Context) {
	// 检查是否登录
	if !middleware.RequireLogin(c) {
		util.ResponseError(c, errors.ErrUnauthorized.WithMsg("请先登录"))
		return
	}

	var commentVO vo.CommentVO
	if err := c.ShouldBindJSON(&commentVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	
	slog.Info("收到评论请求", "type", commentVO.Type, "topicIdStr", commentVO.TopicIDStr, "topicId", commentVO.TopicID, "content", commentVO.Content[:min(30, len(commentVO.Content))])
	
	// 兼容前端传字符串类型的topicId (对标Java前端 arr[2] 从URL解析的是字符串)
	// 修复问题2和3：正确处理 topicId
	if commentVO.TopicIDStr != "" {
		if topicID, err := strconv.ParseUint(commentVO.TopicIDStr, 10, 64); err == nil {
			uid := uint(topicID)
			commentVO.TopicID = &uid
			slog.Info("解析 topicId 成功", "topicId", uid)
		}
	}
	
	// 从上下文中获取用户信息ID (优先 user_info_id = t_user_info.id，用于关联查询)
	userID := middleware.GetUserInfoID(c)
	if userID == 0 {
		userID = middleware.GetUserID(c)
	}
	if userID == 0 {
		util.ResponseError(c, errors.ErrUnauthorized.WithMsg("未获取到用户信息"))
		return
	}
	clientIP := c.ClientIP()

	result, err := h.svc.CreateComment(c.Request.Context(), userID, commentVO, clientIP)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// @Summary 回复指定评论
// @Tags 评论
// @Description 回复指定评论（需要登录）
// @Accept json
// @Produce json
// @Param id path int true "父评论ID"
// @Param replyVO body vo.CommentVO true "回复内容"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/comments/{id}/reply [post]
func (h *CommentHandler) ReplyComment(c *gin.Context) {
	var replyVO vo.CommentVO
	if err := c.ShouldBindJSON(&replyVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	
	// 兼容前端传字符串类型的topicId
	if replyVO.TopicIDStr != "" {
		if topicID, err := strconv.ParseUint(replyVO.TopicIDStr, 10, 64); err == nil {
			uid := uint(topicID)
			replyVO.TopicID = &uid
		}
	}
	
	// 优先使用 user_info_id（t_user_info.id，用于关联查询），回退到 auth id
	userID := middleware.GetUserInfoID(c)
	if userID == 0 {
		userID = middleware.GetUserID(c)
	}
	if parentID, err := strconv.ParseUint(c.Param("id"), 10, 64); err == nil {
		replyVO.ParentID = uint(parentID)
	}
	clientIP := c.ClientIP()

	result, err := h.svc.CreateComment(c.Request.Context(), userID, replyVO, clientIP)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// @Summary 点赞评论
// @Tags 评论
// @Description 点赞指定评论
// @Accept json
// @Produce json
// @Param id path int true "评论ID"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/comments/{id}/like [post]
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

// @Summary 后台评论列表
// @Tags 评论
// @Description 后台评论列表（含审核状态筛选）
// @Accept json
// @Produce json
// @Param keyword query string false "搜索关键词"
// @Param current query int false "当前页码"
// @Param size query int false "每页数量"
// @Success 200 {object} object "成功返回评论列表"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/comments [get]
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

// @Summary 审核评论
// @Tags 评论
// @Description 批量审核评论（通过/拒绝）
// @Accept json
// @Produce json
// @Param body body object true "审核参数(isReview:审核状态, ids:评论ID列表)"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/comments/review [put]
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

// @Summary 删除评论
// @Tags 评论
// @Description 后台批量删除评论
// @Accept json
// @Produce json
// @Param ids body []uint true "评论ID列表"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/comments [delete]
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

// @Summary 获取前6条最新评论
// @Tags 评论
// @Description 获取前6条最新评论（用于侧边栏）
// @Accept json
// @Produce json
// @Success 200 {object} object "成功返回评论列表"
// @Failure 500 {object} object "服务器内部错误"
// @Router /api/comments/topSix [get]
func (h *CommentHandler) ListTopSixComments(c *gin.Context) {
	list, err := h.svc.GetLatestComments(c.Request.Context(), 6)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, list)
}

// @Summary 获取评论统计信息
// @Tags 评论
// @Description 获取后台评论统计信息
// @Accept json
// @Produce json
// @Success 200 {object} object "成功返回评论统计数据"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/comments/stats [get]
func (h *CommentHandler) GetCommentStats(c *gin.Context) {
	stats, err := h.svc.GetCommentStats(c.Request.Context())
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, stats)
}

// @Summary 获取评论回复列表
// @Tags 评论
// @Description 根据评论ID获取回复列表
// @Accept json
// @Produce json
// @Param id path int true "评论ID"
// @Success 200 {object} object "成功返回回复列表"
// @Failure 400 {object} object "请求参数错误"
// @Failure 500 {object} object "服务器内部错误"
// @Router /api/comments/{id}/replies [get]
func (h *CommentHandler) ListRepliesByCommentId(c *gin.Context) {
	commentId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的评论ID"))
		return
	}

	list, err := h.svc.ListRepliesByCommentId(c.Request.Context(), uint(commentId))
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, list)
}
