package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/aurora-go/aurora/internal/service"
	"github.com/aurora-go/aurora/internal/util"
)

// AuroraInfoHandler 首页/后台信息聚合处理器
// 对标 Java AuroraInfoController + BlogInfoController
type AuroraInfoHandler struct {
	svc          *service.AuroraInfoService
	statsService *service.RedisStatsService
}

func NewAuroraInfoHandler(svc *service.AuroraInfoService, statsService *service.RedisStatsService) *AuroraInfoHandler {
	return &AuroraInfoHandler{svc: svc, statsService: statsService}
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
// GET /api/admin/ (需JWT)
func (h *AuroraInfoHandler) GetAdminInfo(c *gin.Context) {
	dashboard, err := h.svc.GetAdminDashboard(c.Request.Context())
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, dashboard)
}

// Report 上报访客信息
// POST /api/report
// 对标 Java AuroraInfoController.report (记录IP到Redis)
func (h *AuroraInfoHandler) Report(c *gin.Context) {
	// 获取客户端IP
	ip := util.GetClientIP(c)
	
	// 记录独立访客到 Redis
	if h.statsService != nil {
		go func() {
			_ = h.statsService.RecordUniqueVisitor(c.Request.Context(), ip)
		}()
	}
	
	util.ResponseSuccess(c, nil)
}
