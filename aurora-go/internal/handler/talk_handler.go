package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/util"
)

// TalkHandler 说说处理器（对标 Java TalkController）
type TalkHandler struct {
	// talkService service.TalkService
}

func NewTalkHandler() *TalkHandler { return &TalkHandler{} }

// ListTalks 获取说说列表（前台，按时间倒序）
// GET /api/talks
func (h *TalkHandler) ListTalks(c *gin.Context) {
	pageNum, pageSize := util.PageQuery(c)
	util.ResponseSuccessWithPage(c, []interface{}{}, int64(0), pageNum, pageSize)
}

// GetTalkById 获取说说详情（含回复评论）
// GET /api/talks/:id
func (h *TalkHandler) GetTalkById(c *gin.Context) {
	idStr := c.Param("id")
	_, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的说说ID"))
		return
	}
	// TODO: P0-5 查询说说 + 嵌套评论
	util.ResponseSuccess(c, nil)
}

// SaveOrUpdate 保存/更新说说（后台）
// POST /api/admin/talks (新增)
// PUT /api/admin/talks/:id (更新)
func (h *TalkHandler) SaveOrUpdate(c *gin.Context) {
	var talkVO dto.TalkVO
	if err := c.ShouldBindJSON(&talkVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	_ = talkVO
	zap.L().Debug("Save talk", zap.Any("talk", talkVO))
	util.ResponseSuccess(c, nil)
}

// DeleteTalk 删除说说（后台）
// DELETE /api/admin/talks/:id
func (h *TalkHandler) DeleteTalk(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	zap.L().Debug("Delete talk", zap.Int64("id", id))
	util.ResponseSuccess(c, "说说已删除")
}

// LikeTalk 点赞说说
// POST /api/talks/:id/like
func (h *TalkHandler) LikeTalk(c *gin.Context) {
	idStr := c.Param("id")
	_, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的说说ID"))
		return
	}
	util.ResponseSuccess(c, map[string]interface{}{"likes": 1})
}

// AddTalkComment 对说说发表评论/回复
// POST /api/talks/:id/comments
func (h *TalkHandler) AddTalkComment(c *gin.Context) {
	var commentVO dto.CommentVO
	if err := c.ShouldBindJSON(&commentVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	_ = commentVO // TODO: P0-5 保存说说评论(复用Comment表+type字段)

	util.ResponseSuccess(c, "评论成功")
}
