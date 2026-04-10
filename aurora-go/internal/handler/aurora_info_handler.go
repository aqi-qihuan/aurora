package handler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/aurora-go/aurora/internal/util"
)

// AuroraInfoHandler 首页/后台信息聚合处理器
// 对标 Java AuroraInfoController + BlogInfoController
type AuroraInfoHandler struct {
	// auroraInfoService service.AuroraInfoService
	// redisService     service.RedisService
}

func NewAuroraInfoHandler() *AuroraInfoHandler { return &AuroraInfoHandler{} }

// GetHomeInfo 获取前台首页数据聚合（文章/分类/标签/友链/统计）
// GET /api/home/info
// 对标 AuroraInfoController.getHomeInfo()
func (h *AuroraInfoHandler) GetHomeInfo(c *gin.Context) {
	// TODO: P0-5 并发查询 goroutine errgroup:
	//   - 最新文章(10篇)
	//   - 推荐文章(6篇)
	//   - 分类列表(含文章数)
	//   - 热门标签(20个)
	//   - 友链列表
	//   - 页面统计(Redis ZSet/HyperLogLog)

	zap.L().Debug("Get home info")
	util.ResponseSuccess(c, map[string]interface{}{
		"articles":       []interface{}{},
		"featuredArticles": []interface{}{},
		"categories":      []interface{}{},
		"tags":            []interface{}{},
		"friendLinks":     []interface{}{},
		"pageViewCount":   0,
		"uniqueVisitorCount": 0,
	})
}

// GetAdminInfo 获取后台管理首页统计数据
// GET /api/admin/info (需JWT)
// 对标 AuroraInfoController.getAdminInfo() - 含Redis统计
func (h *AuroraInfoHandler) GetAdminInfo(c *gin.Context) {
	// TODO: P0-5 并发查询:
	//   - 文章总数 / 今日新增
	//   - 评论总数 / 待审核数
	//   - 用户总数 / 今日注册
	//   - 访客统计(HyperLogLog)
	//   - 文章浏览排行Top10(ZSet)
	//   - 访问地域分布(Geo)

	util.ResponseSuccess(c, map[string]interface{}{
		"articleStats": map[string]interface{}{
			"total":      0,
			"todayNew":   0,
			"published":  0,
			"draft":      0,
		},
		"commentStats": map[string]interface{}{
			"total":         0,
			"pendingReview": 0,
		},
		"userStats": map[string]interface{}{
			"total":    0,
			"todayNew": 0,
		},
		"visitStats": map[string]interface{}{
			"pvToday": 0,
			"uvToday": 0,
		},
		"topArticles": []interface{}{},
		"areaStats":   []interface{}{},
	})
}
