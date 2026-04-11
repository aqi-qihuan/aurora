package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/aurora-go/aurora/internal/service"
	"github.com/aurora-go/aurora/internal/util"
)

// AuroraInfoHandler 首页/后台信息聚合处理器
// 对标 Java AuroraInfoController + BlogInfoController
type AuroraInfoHandler struct {
	svc *service.AuroraInfoService
}

func NewAuroraInfoHandler(svc *service.AuroraInfoService) *AuroraInfoHandler {
	return &AuroraInfoHandler{svc: svc}
}

// GetHomeInfo 获取前台首页数据聚合（文章/分类/标签/友链/统计）
// GET /api/home/info
func (h *AuroraInfoHandler) GetHomeInfo(c *gin.Context) {
	info, err := h.svc.GetHomeInfo(c.Request.Context())
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, info)
}

// GetAdminInfo 获取后台管理首页统计数据
// GET /api/admin/info (需JWT)
func (h *AuroraInfoHandler) GetAdminInfo(c *gin.Context) {
	dashboard, err := h.svc.GetAdminDashboard(c.Request.Context())
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, dashboard)
}
