package middleware

import (
	"github.com/gin-gonic/gin"
)

// NoCache 禁用浏览器和CDN缓存的中间件
// 确保前端每次路由切换时都请求最新数据
// 对标Java Spring Boot的 @CacheControl 注解
func NoCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 禁止所有类型的缓存
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate, private, max-age=0")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		
		c.Next()
	}
}
