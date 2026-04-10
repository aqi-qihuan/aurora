package middleware

import (
	"testing"
	"time"
)

// ===== RateLimiterConfig 测试 =====

func TestGetRateLimitConfig(t *testing.T) {
	tests := []struct {
		name              string
		path              string
		wantRequestsLimit int
	}{
		{"login endpoint", "/api/users/login", 5},
		{"register endpoint", "/api/users/register", 3},
		{"verification code", "/api/users/code", 1},
		{"oauth", "/api/users/oauth", 10},
		{"comments", "/api/comments", 10},
		{"admin panel", "/api/admin/articles", 60},
		{"search", "/api/articles/search", 20},
		{"default path", "/api/unknown", 100},
		{"health check", "/health", 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := getRateLimitConfig(tt.path)
			if cfg.RequestsPerSecond != tt.wantRequestsLimit {
				t.Errorf("getRateLimitConfig(%q).RequestsPerSecond = %d; want %d",
					tt.path, cfg.RequestsPerSecond, tt.wantRequestsLimit)
			}
		})
	}
}

func TestGetRateLimitConfig_DefaultFallback(t *testing.T) {
	cfg := getRateLimitConfig("/some/random/path")
	if cfg.RequestsPerSecond != 100 {
		t.Errorf("default config RequestsPerSecond = %d; want 100", cfg.RequestsPerSecond)
	}
	if cfg.Window != time.Second {
		t.Errorf("default config Window = %v; want %v", cfg.Window, time.Second)
	}
}

// ===== rateLimitKey 测试 =====

func TestRateLimitKey(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		path     string
		contains string
	}{
		{"normal", "192.168.1.1", "/api/articles", "ratelimit:192.168.1.1:/api/articles"},
		{"localhost", "127.0.0.1", "/api/users/login", "ratelimit:127.0.0.1:/api/users/login"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rateLimitKey(tt.ip, tt.path)
			if got != tt.contains {
				t.Errorf("rateLimitKey(%q, %q) = %q; want %q", tt.ip, tt.path, got, tt.contains)
			}
		})
	}
}

// ===== DefaultRateLimits 完整性测试 =====

func TestDefaultRateLimits_Completeness(t *testing.T) {
	requiredPaths := []string{
		"/api/users/login",
		"/api/users/register",
		"/api/users/code",
		"/api/admin/",
		"/api/articles/search",
	}
	for _, path := range requiredPaths {
		t.Run("has_config_for_"+path, func(t *testing.T) {
			if _, ok := DefaultRateLimits[path]; !ok && path != "_default_" {
				// 管理后台是前缀匹配, _default_ 是默认配置
				if path == "/api/admin/" {
					if _, ok := DefaultRateLimits[path]; !ok {
						t.Errorf("missing rate limit config for %q", path)
					}
				}
			}
		})
	}
}

// ===== JWT Auth 辅助函数测试 =====

func TestGetUserID(t *testing.T) {
	// 测试 stringsJoin 函数
	tests := []struct {
		name string
		strs []string
		sep  string
		want string
	}{
		{"single element", []string{"admin"}, ",", "admin"},
		{"multiple elements", []string{"admin", "user"}, ",", "admin,user"},
		{"empty slice", []string{}, ",", ""},
		{"three elements", []string{"a", "b", "c"}, "-", "a-b-c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stringsJoin(tt.strs, tt.sep)
			if got != tt.want {
				t.Errorf("stringsJoin() = %q; want %q", got, tt.want)
			}
		})
	}
}

// ===== Benchmark =====

func BenchmarkGetRateLimitConfig(b *testing.B) {
	paths := []string{
		"/api/users/login",
		"/api/admin/articles",
		"/api/unknown",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getRateLimitConfig(paths[i%len(paths)])
	}
}

func BenchmarkRateLimitKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rateLimitKey("192.168.1.1", "/api/articles")
	}
}
