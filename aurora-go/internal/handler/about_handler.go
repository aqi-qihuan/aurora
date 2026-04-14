package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/aurora-go/aurora/internal/service"
	"github.com/aurora-go/aurora/internal/util"
)

// AboutHandler 关于页面处理器（对标 Java AboutController）
type AboutHandler struct {
	svc *service.AboutService
}

func NewAboutHandler(svc *service.AboutService) *AboutHandler {
	return &AboutHandler{svc: svc}
}

// GetAbout 获取关于页面内容（前台公开）
// GET /api/about
func (h *AboutHandler) GetAbout(c *gin.Context) {
	// 对标Java：返回AboutDTO对象（包含content字段）
	aboutDTO, err := h.svc.GetAboutDTO(c.Request.Context())
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, aboutDTO) // ✅ 返回 {"content":"xxx"}
}

// SaveOrUpdate 保存/更新关于页面（后台）
// POST /api/admin/about
// PUT /api/admin/about/:id
func (h *AboutHandler) SaveOrUpdate(c *gin.Context) {
	var body struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		util.ResponseError(c, err)
		return
	}

	if err := h.svc.UpdateAbout(c.Request.Context(), body.Content); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "关于页面已更新")
}
