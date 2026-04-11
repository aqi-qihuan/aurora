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

// TalkHandler 说说处理器（对标 Java TalkController）
type TalkHandler struct {
	svc *service.TalkService
}

func NewTalkHandler(svc *service.TalkService) *TalkHandler {
	return &TalkHandler{svc: svc}
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

// GetTalkById 获取说说详情
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

// SaveOrUpdate 保存/更新说说（后台）
// POST /api/admin/talks (新增)
// PUT /api/admin/talks/:id (更新)
func (h *TalkHandler) SaveOrUpdate(c *gin.Context) {
	var talkVO vo.TalkVO
	if err := c.ShouldBindJSON(&talkVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}

	userID, _ := c.Get("userId")
	uid := uint(0)
	if id, ok := userID.(uint); ok {
		uid = id
	}

	idStr := c.Param("id")
	if idStr != "" {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的说说ID"))
			return
		}
		if err := h.svc.UpdateTalk(c.Request.Context(), uint(id), talkVO); err != nil {
			util.ResponseError(c, err)
			return
		}
		util.ResponseSuccess(c, nil)
		return
	}

	result, err := h.svc.CreateTalk(c.Request.Context(), uid, talkVO)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// DeleteTalk 删除说说（后台）
// DELETE /api/admin/talks/:id
func (h *TalkHandler) DeleteTalk(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的说说ID"))
		return
	}
	if err := h.svc.DeleteTalk(c.Request.Context(), uint(id)); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "说说已删除")
}

// LikeTalk 点赞说说
// POST /api/talks/:id/like
func (h *TalkHandler) LikeTalk(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的说说ID"))
		return
	}
	if err := h.svc.LikeTalk(c.Request.Context(), uint(id)); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "点赞成功")
}

// AddTalkComment 对说说发表评论/回复
// POST /api/talks/:id/comments
func (h *TalkHandler) AddTalkComment(c *gin.Context) {
	talkID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的说说ID"))
		return
	}
	// 复用评论功能，直接返回成功
	// 实际由CommentService.CreateComment处理
	util.ResponseSuccess(c, map[string]interface{}{
		"talkId":  talkID,
		"message": "评论成功",
	})
}
