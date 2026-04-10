package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/aurora-go/aurora/internal/constant"
	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/model"
	"github.com/aurora-go/aurora/internal/service"
	"github.com/gin-gonic/gin"
)

// RBAC 基于角色的动态权限控制中间件 (对标Java FilterInvocationSecurityMetadataSourceImpl + AccessDecisionManagerImpl)
//
// 权限检查流程:
//  1. 从Context获取当前用户角色列表
//  2. 查询该路径+方法对应的资源权限配置
//  3. 判断用户角色是否在允许的角色列表中
//  4. 超级管理员(admin角色)自动放行
func RBAC(registry *service.Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "未登录",
			})
			return
		}

		// 获取用户角色(从Redis Session已存入Context)
		userRoles := GetUserRole(c)
		if userRoles == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "无法获取用户角色信息",
			})
			return
		}

		method := c.Request.Method
		path := c.Request.URL.Path

		// 查询资源-角色映射
		resourceRoles, err := registry.Resource.ListResourceRoles(c.Request.Context())
		if err != nil {
			// 查询失败时默认放行(安全策略: 宁可放过不可误杀)
			c.Next()
			return
		}

		// 匹配当前请求路径和方法的资源权限
		requiredRoles := matchResourceRoles(resourceRoles, method, path)
		if requiredRoles == nil {
			// 未配置权限的资源，默认放行
			c.Next()
			return
		}

		// 超级管理员自动放行
		if userRoles == constant.RoleAdmin {
			c.Next()
			return
		}

		// 检查用户角色是否在允许的角色列表中
		if !containsRole(userRoles, requiredRoles) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": fmt.Sprintf("权限不足，需要以下角色之一: %s", strings.Join(requiredRoles, ",")),
			})
			return
		}

		c.Next()
	}
}

// matchResourceRoles 匹配当前请求的URL和方法对应的角色列表
// 使用Ant Path风格通配符匹配 (对标Java AntPathMatcher)
func matchResourceRoles(resourceRoles []dto.ResourceRoleDTO, method string, urlPath string) []string {
	for _, rr := range resourceRoles {
		if rr.RequestMethod != method {
			continue
		}
		if matchAntPath(rr.URL, urlPath) && len(rr.RoleList) > 0 {
			return rr.RoleList
		}
	}
	return nil
}

// matchAntPath Ant风格路径匹配
// 支持: /api/articles/* , /api/articles/{id} , 精确匹配
func matchAntPath(pattern, path string) bool {
	pattern = strings.TrimSuffix(pattern, "/")
	path = strings.TrimSuffix(path, "/")

	// 精确匹配
	if pattern == path {
		return true
	}

	// 通配符 * 匹配一级目录
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		return strings.HasPrefix(path, prefix+"/") || path == prefix
	}

	// 通配符 ** 匹配多级目录
	if strings.Contains(pattern, "**") {
		prefix := strings.Split(pattern, "**")[0]
		return strings.HasPrefix(path, strings.TrimRight(prefix, "/"))
	}

	return false
}

// containsRole 检查用户的角色字符串是否包含任一所需角色
// userRoles 格式: "admin,user" 或 "user"
func containsRole(userRoles string, requiredRoles []string) bool {
	userRoleSet := make(map[string]bool)
	for _, r := range strings.Split(userRoles, ",") {
		userRoleSet[strings.TrimSpace(r)] = true
	}
	for _, req := range requiredRoles {
		if userRoleSet[req] {
			return true
		}
	}
	return false
}

// ===== RBAC辅助函数 =====

// HasPermission 在Handler内部进行细粒度权限检查
// 用法: if !middleware.HasPermission(c, "article:delete") { return }
func HasPermission(c *gin.Context, permission string) bool {
	rolesRaw, exists := c.Get("permissions")
	if !exists {
		return false
	}
	switch v := rolesRaw.(type) {
	case []string:
		for _, p := range v {
			if p == permission || p == constant.RoleAdmin {
				return true
			}
		}
	case []model.Role:
		for _, r := range v {
			if r.RoleName == permission || r.RoleName == constant.RoleAdmin {
				return true
			}
		}
	}
	return false
}

// IsAdmin 检查当前用户是否为超级管理员
func IsAdmin(c *gin.Context) bool {
	return GetUserRole(c) == constant.RoleAdmin
}
