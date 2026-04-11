package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// ===== 路由注册完整性测试 =====

func TestRouter_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	router := NewRouter(nil) // registry can be nil for route registration test
	router.RegisterRoutes(r)

	// 验证关键公开路由已注册
	publicRoutes := []struct {
		method string
		path   string
	}{
		// 文章
		{http.MethodGet, "/api/articles"},
		{http.MethodGet, "/api/articles/search"},
		{http.MethodGet, "/api/articles/topAndFeatured"},
		{http.MethodGet, "/api/articles/archives"},
		// 认证
		{http.MethodPost, "/api/auth/register"},
		{http.MethodPost, "/api/auth/login"},
		{http.MethodPost, "/api/auth/code"},
		// 分类标签
		{http.MethodGet, "/api/categories"},
		{http.MethodGet, "/api/tags"},
		// 友链
		{http.MethodGet, "/api/links"},
		// 说说
		{http.MethodGet, "/api/talks"},
		// 相册
		{http.MethodGet, "/api/albums"},
		// 首页信息
		{http.MethodGet, "/api/home/info"},
		// 网站配置
		{http.MethodGet, "/api/website/config"},
		// 关于
		{http.MethodGet, "/api/about"},
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

	router := NewRouter(nil)
	router.RegisterRoutes(r)

	// 验证管理路由需要认证
	adminRoutes := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/admin/articles"},
		{http.MethodGet, "/api/admin/roles"},
		{http.MethodGet, "/api/admin/menus"},
		{http.MethodGet, "/api/admin/info"},
	}

	for _, route := range adminRoutes {
		t.Run(route.method+"_"+route.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(route.method, route.path, nil)
			r.ServeHTTP(w, req)
			// 无Token应该返回401
			if w.Code != http.StatusUnauthorized {
				t.Errorf("admin route %s %s should require auth, got status %d",
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
		router := NewRouter(nil)
		router.RegisterRoutes(r)
	}
}
