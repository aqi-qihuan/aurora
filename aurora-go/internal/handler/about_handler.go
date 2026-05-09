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

// @Summary 获取关于页面内容
// @Tags 关于
// @Description 获取前台公开的关于页面内容
// @Accept json
// @Produce json
// @Success 200 {object} object "成功返回关于页面内容"
// @Failure 500 {object} object "服务器内部错误"
// @Router /api/about [get]
func (h *AboutHandler) GetAbout(c *gin.Context) {
	// 对标Java：返回AboutDTO对象（包含content字段）
	aboutDTO, err := h.svc.GetAboutDTO(c.Request.Context())
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, aboutDTO) // ✅ 返回 {"content":"xxx"}
}

// @Summary 保存/更新关于页面
// @Tags 关于
// @Description 保存或更新关于页面内容（后台管理）
// @Accept json
// @Produce json
// @Param content body string true "关于页面内容"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/about [post]
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
