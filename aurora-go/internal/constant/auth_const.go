package constant

import (
	"strings"
)

// ===== 认证相关常量 (对齐Java AuthConstant) =====

const (
	// Token 过期时间
	TokenExpireTime     = 7 * 24 * 60 * 60 // Token过期时间: 7天(秒)
	TokenRenewThreshold = 20               // 续期阈值: 20分钟

	// HTTP Header
	TokenHeader  = "Authorization"
	TokenPrefix  = "Bearer "

	// Claim Keys
	JwtClaimUserID    = "user_id"
	JwtClaimIssueAt  = "iat"
	JwtClaimExpireAt = "exp"

	// 默认角色
	RoleAdmin = "admin"
	RoleUser  = "user"

	// 登录类型
	LoginTypeEmail = 1 // 邮箱登录
	LoginTypeQQ   = 3 // QQ登录
)

// 白名单路径 (不需要JWT认证的路径)
var PublicPaths = []string{
	"POST:/api/users/login",
	"POST:/api/users/register",
	"POST:/api/users/code",
	"GET:/api/articles",
	"GET:/api/articles/*",
	"GET:/api/categories",
	"GET:/api/tags",
	"GET:/api/tags/*",
	"GET:/api/friendLinks",
	"GET:/api/talks",
	"GET:/api/talks/*",
	"GET:/api/photos/albums",
	"GET:/api/photos/albums/*/photos",
	"GET:/api/about",
	"GET:/api/website/config",
	"GET:/api/home/info",
	"POST:/api/comments",         // 游客可评论
	"GET:/api/articles/search",    // 搜索公开
	"POST:/api/users/oauth/login", // OAuth登录回调
	"GET:/api/users/oauth/url",    // 获取OAuth授权URL
}

// IsPublicPath 检查路径是否在白名单中
func IsPublicPath(method, path string) bool {
	for _, pp := range PublicPaths {
		parts := splitPathPattern(pp)
		if len(parts) == 2 && parts[0] == method {
			if matchPattern(parts[1], path) {
				return true
			}
		}
	}
	return false
}

// splitPathPattern 分割 "METHOD:path" 为 [METHOD, path]
func splitPathPattern(pattern string) []string {
	idx := strings.Index(pattern, ":")
	if idx == -1 {
		return []string{pattern}
	}
	return []string{pattern[:idx], pattern[idx+1:]}
}

// matchPattern 简单的通配符匹配 (仅支持尾部 *, 如 /api/articles/*)
func matchPattern(pattern, path string) bool {
	if !strings.HasSuffix(pattern, "*") {
		return pattern == path
	}
	prefix := strings.TrimSuffix(pattern, "*")
	return strings.HasPrefix(path, prefix)
}
