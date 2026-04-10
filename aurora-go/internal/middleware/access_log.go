package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/service"
	"github.com/gin-gonic/gin"
)

// AccessLog 操作日志中间件 (对标Java @OptLog AOP)
// 记录用户操作到数据库(操作日志表)
//
// 使用方式: 在需要记录的路由组上使用
//
//	r.Use(middleware.AccessLog(registry))
func AccessLog(registry *service.Registry, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// 读取请求体(用于记录参数, 注意: 请求体只能读一次)
		var body []byte
		if c.Request.Body != nil {
			body, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body)) // 重置以便后续Handler使用
		}

		// 执行后续Handler
		c.Next()

		// 异步记录操作日志 (不阻塞响应)
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		go func() {
			// 只记录成功或特定状态的操作 (避免大量404/500日志)
			if statusCode >= 400 && statusCode != 401 && statusCode != 403 {
				return
			}
			// 不记录静态资源和健康检查
			if shouldSkipPath(path) {
				return
			}

			userID := GetUserID(c)
			durationMs := duration.Milliseconds()

			logEntry := &dto.OperationLogVO{
				UserID:   userID,
				Method:   method,
				URL:      path + "?" + c.Request.URL.RawQuery,
				Duration: &durationMs,
				IP:       c.ClientIP(),
			}

			// 截断过长的请求体
			bodyStr := string(body)
			if len(bodyStr) > 1000 {
				bodyStr = bodyStr[:1000] + "...(truncated)"
			}
			_ = bodyStr // TODO: OperationLogVO暂无RequestBody字段，待扩展

			if err := registry.OperationLog.Save(logEntry); err != nil {
				logger.Error("保存操作日志失败", "error", err, "path", path)
			}
		}()
	}
}

// shouldSkipPath 判断是否跳过记录的路径
func shouldSkipPath(path string) bool {
	skipPrefixes := []string{
		"/health",
		"/metrics",
		"/favicon.ico",
		"/static/",
		"/api/home/info",
		"/api/articles/search",
	}
	for _, prefix := range skipPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}
