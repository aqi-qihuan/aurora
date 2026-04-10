package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiter 基于Redis滑动窗口的接口限流中间件 (完整实现)
// 使用Lua脚本保证原子性, 避免竞态条件
//
// 限流算法: Sliding Window Log (滑动窗口日志)
// - 使用Redis ZSet记录每个请求的时间戳
// - 窗口期内请求数超过阈值则拒绝
// - 支持按IP + 路径维度限流
func RateLimiter(rdb *redis.Client, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if rdb == nil {
			c.Next() // Redis未配置时跳过限流
			return
		}

		cfg := getRateLimitConfig(c.Request.URL.Path)
		clientIP := c.ClientIP()

		allowed, remaining, err := checkRateLimit(
			c.Request.Context(),
			rdb,
			clientIP,
			c.Request.URL.Path,
			cfg,
			logger,
		)
		if err != nil {
			logger.Warn("限流检查失败，默认放行", "error", err, "ip", clientIP, "path", c.Request.URL.Path)
			c.Next()
			return
		}

		// 设置限流响应头(方便客户端展示剩余额度)
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.RequestsPerSecond))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(cfg.Window).Unix()))

		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "请求过于频繁，请稍后再试",
				"retry_after": fmt.Sprintf("%.0f", cfg.Window.Seconds()),
			})
			return
		}

		c.Next()
	}
}

// ===== 限流核心逻辑 =====

// checkRateLimit 执行滑动窗口限流检查
// 返回: 是否允许, 剩余额度, 错误
func checkRateLimit(ctx context.Context, rdb *redis.Client, clientIP string, path string, cfg RateLimiterConfig, logger *slog.Logger) (bool, int64, error) {
	key := rateLimitKey(clientIP, path)
	now := time.Now().UnixMilli()
	windowMs := cfg.Window.Milliseconds()

	result, err := rdb.Eval(ctx, slidingWindowLuaScript, []string{key},
		windowMs, cfg.RequestsPerSecond, now,
	).Int64()

	if err != nil {
		return false, 0, fmt.Errorf("执行限流脚本: %w", err)
	}

	// result = -1 表示超限, >0 表示当前窗口内请求数
	if result < 0 {
		logger.Debug("触发限流",
			"ip", clientIP,
			"path", path,
			"limit", cfg.RequestsPerSecond,
			"window", cfg.Window,
		)
		return false, 0, nil
	}

	remaining := int64(cfg.RequestsPerSecond) - result
	if remaining < 0 {
		remaining = 0
	}
	return true, remaining, nil
}

// slidingWindowLuaScript Redis Lua滑动窗口脚本 (原子操作)
const slidingWindowLuaScript = `
local key = KEYS[1]
local window = tonumber(ARGV[1])
local limit = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

-- 清理窗口外的过期记录
redis.call('ZREMRANGEBYSCORE', key, '-inf', now - window)

-- 获取当前窗口内的请求数
local count = redis.call('ZCARD', key)

if count < limit then
    -- 记录本次请求 (score=时间戳, member=时间戳:随机数防重复)
    redis.call('ZADD', key, now, now .. ':' .. math.random())
    redis.call('PEXPIRE', key, window)
    return count + 1
else
    return -1
end
`

// rateLimitKey 生成限流唯一Key
func rateLimitKey(clientIP string, path string) string {
	return fmt.Sprintf("ratelimit:%s:%s", clientIP, path)
}

// ===== 限流配置 =====

// RateLimiterConfig 单个路径的限流配置
type RateLimiterConfig struct {
	RequestsPerSecond int           `json:"requests_per_second"` // 窗口期最大请求数
	Window            time.Duration `json:"window"`               // 窗口大小
	Burst             int           `json:"burst"`                // 允许的突发请求(预留)
}

// DefaultRateLimits 默认限流规则表 (按路径匹配前缀)
var DefaultRateLimits = map[string]RateLimiterConfig{
	"/api/users/login":     {RequestsPerSecond: 5, Window: time.Minute},      // 登录: 5次/分钟
	"/api/users/register": {RequestsPerSecond: 3, Window: time.Minute},      // 注册: 3次/分钟
	"/api/users/code":     {RequestsPerSecond: 1, Window: time.Minute},      // 验证码: 1次/分钟
	"/api/users/oauth":    {RequestsPerSecond: 10, Window: time.Minute},     // OAuth: 10次/分钟
	"/api/comments":       {RequestsPerSecond: 10, Window: time.Minute},     // 评论: 10次/分钟(防灌水)
	"/api/admin/":         {RequestsPerSecond: 60, Window: time.Second},     // 管理后台: 60QPS
	"/api/articles/search": {RequestsPerSecond: 20, Window: time.Second},    // 搜索: 20QPS
	defaultPath:            {RequestsPerSecond: 100, Window: time.Second},   // 其他: 100QPS
}

const defaultPath = "_default_"

// getRateLimitConfig 根据路径获取限流配置 (最长前缀匹配)
func getRateLimitConfig(path string) RateLimiterConfig {
	for pattern, cfg := range DefaultRateLimits {
		if pattern == defaultPath {
			continue
		}
		// 前缀匹配
		if len(path) >= len(pattern) && path[:len(pattern)] == pattern {
			return cfg
		}
	}
	// 返回默认配置
	if def, ok := DefaultRateLimits[defaultPath]; ok {
		return def
	}
	return RateLimiterConfig{
		RequestsPerSecond: 100,
		Window:            time.Second,
	}
}
