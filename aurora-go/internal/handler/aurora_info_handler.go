package handler

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"time"

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
// 完全对标 Java AuroraInfoServiceImpl.report()
// 逻辑: 计算访客唯一指纹(IP+浏览器+操作系统) → 判断是否为新访客 → 记录地域+浏览量+独立访客
func (h *AuroraInfoHandler) Report(c *gin.Context) {
	// 1. 获取客户端IP
	ip := util.GetClientIP(c)
	
	// 2. 解析UserAgent获取浏览器和操作系统
	userAgent := c.GetHeader("User-Agent")
	browser := extractBrowser(userAgent)
	os := extractOS(userAgent)
	
	// 3. 生成唯一指纹 (对标Java: ipAddress + browser.getName() + operatingSystem.getName())
	uuid := fmt.Sprintf("%s%s%s", ip, browser, os)
	
	// 4. 计算MD5指纹
	md5Hash := md5.Sum([]byte(uuid))
	fingerprint := hex.EncodeToString(md5Hash[:])
	
	// 5. 检查是否为独立访客（去重）
	if h.statsService != nil {
		// 使用独立 Context，避免 HTTP 请求结束后被取消
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		
		// 使用全局Set检查是否已访问过（对标Java redisService.sIsMember(UNIQUE_VISITOR, md5)）
		isNewVisitor, err := h.statsService.IsUniqueVisitor(ctx, fingerprint)
		if err != nil {
			// Redis 出错时，降级处理：仍然执行统计，但不去重
			slog.Warn("检查独立访客失败，降级处理", "ip", ip, "error", err.Error())
			isNewVisitor = true // 假设是新访客，继续执行统计
		}
		
		// 6. 如果是新访客，标记并记录统计信息
		if isNewVisitor {
			// 标记为已访问（对标Java redisService.sAdd(UNIQUE_VISITOR, md5)）
			_ = h.statsService.MarkUniqueVisitor(ctx, fingerprint)
			
			// 使用独立 Context 执行异步统计任务
			asyncCtx, asyncCancel := context.WithTimeout(context.Background(), 5*time.Second)
			go func() {
				defer asyncCancel()
				
				// 6.1 获取IP归属地并记录地域统计（对标Java redisService.hIncr(VISITOR_AREA, ipProvince, 1L)）
				ipSource := util.GetIPRegion(ip)
				province := extractProvince(ipSource)
				if province == "" {
					province = "未知"
				}
				if err := h.statsService.RecordVisitorArea(asyncCtx, province); err != nil {
					slog.Warn("记录访客地域失败", "province", province, "error", err.Error())
				}
				
				// 6.2 增加总浏览量 PV（对标Java redisService.incr(BLOG_VIEWS_COUNT, 1)）
				if err := h.statsService.IncrementTotalViews(asyncCtx); err != nil {
					slog.Warn("增加总浏览量失败", "error", err.Error())
				}
				
				// 6.3 记录按天的独立访客（用于定时任务归档）
				// 修复：使用 fingerprint（MD5）而不是 IP，与检查逻辑保持一致
				if err := h.statsService.RecordUniqueVisitorByFingerprint(asyncCtx, fingerprint); err != nil {
					slog.Warn("记录每日独立访客失败", "fingerprint", fingerprint, "error", err.Error())
				} else {
					slog.Info("记录每日独立访客成功", "fingerprint", fingerprint[:8]+"...")
				}
			}()
		}
	}
	
	util.ResponseSuccess(c, nil)
}

// extractBrowser 从UserAgent提取浏览器名称（对标Java Browser.getName()）
func extractBrowser(ua string) string {
	ua = strings.ToLower(ua)
	switch {
	case strings.Contains(ua, "edg") || strings.Contains(ua, "edge"):
		return "Edge"
	case strings.Contains(ua, "chrome"):
		return "Chrome"
	case strings.Contains(ua, "firefox"):
		return "Firefox"
	case strings.Contains(ua, "safari"):
		return "Safari"
	case strings.Contains(ua, "msie") || strings.Contains(ua, "trident"):
		return "IE"
	default:
		return "Other"
	}
}

// extractOS 从UserAgent提取操作系统（对标Java OperatingSystem.getName()）
func extractOS(ua string) string {
	ua = strings.ToLower(ua)
	switch {
	case strings.Contains(ua, "windows"):
		return "Windows"
	case strings.Contains(ua, "macintosh") || strings.Contains(ua, "mac os"):
		return "MacOS"
	case strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad"):
		return "iOS"
	case strings.Contains(ua, "android"):
		return "Android"
	case strings.Contains(ua, "linux"):
		return "Linux"
	default:
		return "Other"
	}
}

// extractProvince 从IP归属地提取省份
// 示例: "中国|0|浙江省|杭州市|电信" → "浙江省"
// 对标Java IpUtil.getIpProvince(ipSource)
func extractProvince(ipSource string) string {
	if ipSource == "" || ipSource == "内网IP" || ipSource == "局域网" {
		return ""
	}
	
	parts := strings.Split(ipSource, "|")
	// 格式: "国家|区域|省份|城市|运营商"
	// 索引:  0     1    2    3     4
	if len(parts) >= 3 && parts[2] != "0" {
		return parts[2]
	}
	
	// 如果没有省份信息，返回国家
	if len(parts) >= 1 && parts[0] != "0" {
		return parts[0]
	}
	
	return ""
}
