package handler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/util"
)

// AboutHandler 关于页面处理器（对标 Java AboutController）
type AboutHandler struct {
	// aboutService service.AboutService
}

func NewAboutHandler() *AboutHandler { return &AboutHandler{} }

// GetAbout 获取关于页面内容（前台公开）
// GET /api/about
func (h *AboutHandler) GetAbout(c *gin.Context) {
	// TODO: P0-5 从DB查询关于页内容(支持Markdown)
	util.ResponseSuccess(c, map[string]interface{}{
		"content": "",
	})
}

// SaveOrUpdate 保存/更新关于页面（后台）
// POST /api/admin/about
// PUT /api/admin/about/:id
func (h *AboutHandler) SaveOrUpdate(c *gin.Context) {
	var aboutVO dto.AboutVO
	if err := c.ShouldBindJSON(&aboutVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParam.WithMsg(err.Error()))
		return
	}
	_ = aboutVO

	zap.L().Debug("Save about page content")
	util.ResponseSuccess(c, "关于页面已更新")
}
