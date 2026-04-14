package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/service"
	"github.com/gin-gonic/gin"
)

// Recovery 全局异常恢复中间件 (对标Java ExceptionLogAspect)
// 捕获所有panic，记录到异常日志表，并返回500
func Recovery(registry *service.Registry, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录 panic 详情
				stack := string(debug.Stack())
				logger.Error("PANIC RECOVERED",
					"error", err,
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
					"client_ip", c.ClientIP(),
				)

				// 异步保存异常日志到数据库（对标Java ExceptionLogAspect）
				go func() {
					// 读取请求体
					var bodyStr string
					if c.Request.Body != nil {
						// 注意：此时请求体可能已被其他中间件读取，这里尽量获取
						bodyStr = "[request body captured by other middleware]"
					}

					logEntry := &dto.ExceptionLogVO{
						URL:           c.Request.URL.Path,
						Method:        fmt.Sprintf("%s.%s", c.HandlerName(), c.Request.Method),
						RequestMethod: c.Request.Method,
						RequestParam:  bodyStr,
						OptDesc:       "系统异常",
						ExceptionInfo: fmt.Sprintf("%v\n\n%s", err, stack),
						IP:            c.ClientIP(),
					}

					if registry.ExceptionLog != nil {
						if saveErr := registry.ExceptionLog.SaveExceptionLog(c.Request.Context(), *logEntry); saveErr != nil {
							logger.Error("保存异常日志失败", "error", saveErr)
						}
					}
				}()

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "Internal Server Error",
				})
			}
		}()
		c.Next()
	}
}
