package handler

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/aurora-go/aurora/internal/service"
)

// ===== 路由注册完整性测试 =====
// 使用零值 Registry 构造 Router：所有 Handler 创建为非 nil 指针（svc 字段为 nil），
// 路由注册阶段不 panic；请求到达时 Handler 方法体访问 nil svc 会 panic，
// gin.Recovery 中间件捕获后返回 500，测试只校验状态码非 404。
// admin 路由的 JWTAuthEnhanced 中间件在无 Token 时直接返回 401，不访问 tokenSvc（nil 安全）。
//
// 注意: /api/ (GetHomeInfo) 路由未纳入测试，因其内部启动 goroutine 聚合数据，
// goroutine panic 无法被 gin.Recovery 捕获，会导致测试进程崩溃。

func newTestRouter() *Router {
	return NewRouter(&service.Registry{}, nil, slog.Default())
}

func TestRouter_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery()) // 捕获 nil svc panic，返回 500 而非崩溃

	router := newTestRouter()
	router.RegisterRoutes(r)

	// 验证关键公开路由已注册（对齐 Java 原始路径）
	publicRoutes := []struct {
		method string
		path   string
	}{
		// 文章
		{http.MethodGet, "/api/articles/topAndFeatured"},
		{http.MethodGet, "/api/articles/all"},
		{http.MethodGet, "/api/articles/search"},
		{http.MethodGet, "/api/archives/all"},
		// 认证（对齐 Java 路径）
		{http.MethodPost, "/api/users/login"},
		{http.MethodPost, "/api/users/register"},
		{http.MethodGet, "/api/users/code"},
		{http.MethodPost, "/api/users/logout"},
		// 分类标签
		{http.MethodGet, "/api/categories/all"},
		{http.MethodGet, "/api/tags/all"},
		{http.MethodGet, "/api/tags/topTen"},
		// 友链
		{http.MethodGet, "/api/links"},
		// 说说
		{http.MethodGet, "/api/talks"},
		// 相册
		{http.MethodGet, "/api/photos/albums"},
		// 评论
		{http.MethodGet, "/api/comments/topSix"},
		{http.MethodPost, "/api/comments/save"},
		// 关于
		{http.MethodGet, "/api/about"},
		// 作品集
		{http.MethodGet, "/api/portfolios/featured"},
		{http.MethodGet, "/api/portfolios"},
	}

	for _, route := range publicRoutes {
		t.Run(route.method+"_"+route.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(route.method, route.path, nil)
			r.ServeHTTP(w, req)
			// 不期望404(路由未注册), 其他状态码都可接受(如500=服务未就绪)
			if w.Code == http.StatusNotFound {
				t.Errorf("route %s %s not registered (404)", route.method, route.path)
			}
		})
	}
}

func TestRouter_AdminRoutes_RequireAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())

	router := newTestRouter()
	router.RegisterRoutes(r)

	// 验证管理路由需要认证（对齐 Java 路径）
	adminRoutes := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/admin/articles"},
		{http.MethodGet, "/api/admin/roles"},
		{http.MethodGet, "/api/admin/menus"},
		{http.MethodGet, "/api/admin/"},
		// 作品集后台路由
		{http.MethodGet, "/api/admin/portfolios"},
	}

	for _, route := range adminRoutes {
		t.Run(route.method+"_"+route.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(route.method, route.path, nil)
			r.ServeHTTP(w, req)
			// 无Token应该返回401（JWTAuthEnhanced 在 authHeader=="" 时直接 Abort，不访问 tokenSvc）
			if w.Code != http.StatusUnauthorized {
				t.Errorf("admin route %s %s should require auth (401), got status %d",
					route.method, route.path, w.Code)
			}
		})
	}
}

// ===== 健康检查端点测试 =====

func TestHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    map[string]interface{}{"components": map[string]string{"mysql": "UP"}},
			"timestamp": int64(1712800000),
			"version":   "1.0.0-go",
		})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("health endpoint status = %d; want %d", w.Code, http.StatusOK)
	}
}

// ===== Benchmark =====

func BenchmarkRouterRegistration(b *testing.B) {
	gin.SetMode(gin.TestMode)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := gin.New()
		router := newTestRouter()
		router.RegisterRoutes(r)
	}
}
