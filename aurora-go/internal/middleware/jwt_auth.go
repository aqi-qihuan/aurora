package middleware

import (
	"log/slog"
	"net/http"

	"github.com/aurora-go/aurora/internal/constant"
	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/service"
	"github.com/gin-gonic/gin"
)

// JWTAuthEnhanced 增强版JWT认证中间件 (对标Java JwtAuthenticationTokenFilter)
//
// 完整流程:
//  1. 从Header提取Authorization: Bearer <token>
//  2. 调用TokenService解析JWT并从Redis获取用户详情
//  3. 自动续期(距离过期≤20分钟时刷新Redis TTL)
//  4. 将用户ID、角色、权限存入Gin Context
//  5. 白名单路径自动放行
func JWTAuthEnhanced(tokenSvc *service.TokenService, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		path := c.Request.URL.Path

		// 白名单检查
		if constant.IsPublicPath(method, path) {
			c.Next()
			return
		}

		authHeader := c.GetHeader(constant.TokenHeader)
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未登录或Token已过期",
			})
			return
		}

		tokenString := service.ExtractToken(authHeader)
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "认证格式错误，请使用 Bearer <token>",
			})
			return
		}

		// 解析Token并获取用户详情(含Redis Session查询)
		userDetail, err := tokenSvc.GetUserDetailDTO(tokenString)
		if err != nil {
			logger.Warn("Token验证失败", "error", err, "path", path)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Token无效或已过期，请重新登录",
			})
			return
		}

		if userDetail == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Session已过期，请重新登录",
			})
			return
		}

		// 检查账号是否被禁用
		if userDetail.IsDisable == 1 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "账号已被禁用，请联系管理员",
			})
			return
		}

		// 自动续期
		tokenSvc.RenewToken(userDetail)

		// 将用户信息注入Context (后续Handler/Middleware可直接使用)
		setUserContext(c, userDetail)

		logger.Debug("JWT认证成功",
			"user_id", userDetail.ID,
			"nickname", userDetail.Nickname,
			"path", path,
		)

		c.Next()
	}
}

// setUserContext 将用户信息写入Gin Context
func setUserContext(c *gin.Context, user *dto.UserDetailsDTO) {
	c.Set("user_id", user.ID)
	c.Set("user_info_id", user.UserInfoID)
	c.Set("email", user.Email)
	c.Set("nickname", user.Nickname)
	c.Set("avatar", user.Avatar)
	c.Set("login_type", user.LoginType)
	c.Set("role", stringsJoin(user.Roles, ","))
	c.Set("is_disable", user.IsDisable)
	c.Set("permissions", user.PermissionsList())
}

// GetUserID 从Gin Context获取当前用户ID (0=未登录)
func GetUserID(c *gin.Context) uint {
	if uid, exists := c.Get("user_id"); exists {
		switch v := uid.(type) {
		case float64:
			return uint(v)
		case int64:
			return uint(v)
		case int:
			return uint(v)
		case uint:
			return v
		}
	}
	return 0
}

// GetUserRole 从Gin Context获取当前用户角色字符串
func GetUserRole(c *gin.Context) string {
	if role, exists := c.Get("role"); exists {
		return role.(string)
	}
	return ""
}

// GetUserDetail 从Gin Context获取完整用户详情
func GetUserDetail(c *gin.Context) *dto.UserDetailsDTO {
	if ud, exists := c.Get("user_detail"); exists {
		if u, ok := ud.(*dto.UserDetailsDTO); ok {
			return u
		}
	}
	return nil
}

// RequireLogin 检查是否已登录的辅助函数
func RequireLogin(c *gin.Context) bool {
	return GetUserID(c) != 0
}

// RequireAdmin 检查是否为管理员的辅助函数
func RequireAdmin(c *gin.Context) bool {
	return IsAdmin(c) && RequireLogin(c)
}

// ===== 辅助函数 =====

func stringsJoin(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
