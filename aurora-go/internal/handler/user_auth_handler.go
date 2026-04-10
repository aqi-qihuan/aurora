package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/util"
)

// UserAuthHandler 用户认证处理器（对标 Java UserController + UserAuthController）
// 端点: 8个 (注册/登录/登出/OAuth/密码/邮箱验证码)
type UserAuthHandler struct {
	// userAuthService service.UserAuthService // P0-5 注入
}

// NewUserAuthHandler 创建用户认证Handler
func NewUserAuthHandler() *UserAuthHandler {
	return &UserAuthHandler{}
}

// Register 用户注册
// POST /api/auth/register
// 对标 UserController.register()
func (h *UserAuthHandler) Register(c *gin.Context) {
	var registerVO dto.RegisterVO
	if err := c.ShouldBindJSON(&registerVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParam.WithMsg(err.Error()))
		return
	}
	_ = registerVO // TODO: P0-5 调用Service注册

	util.ResponseSuccess(c, "注册成功")
}

// Login 用户登录
// POST /api/auth/login
// 对标 UserController.login() - 返回JWT Token + UserInfo
func (h *UserAuthHandler) Login(c *gin.Context) {
	var loginVO dto.LoginVO
	if err := c.ShouldBindJSON(&loginVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParam.WithMsg(err.Error()))
		return
	}
	_ = loginVO // TODO: P0-5 验证用户名密码 → 生成JWT → 返回

	util.ResponseSuccess(c, map[string]interface{}{
		"token":     "jwt_token_placeholder",
		"userInfo":  nil,
		"expiresIn": 86400,
	})
}

// Logout 用户登出（将Token加入Redis黑名单）
// POST /api/auth/logout
// 对标 UserController.logout() - Redis黑名单机制
func (h *UserAuthHandler) Logout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	_ = token // TODO: P0-5 将Token加入Redis黑名单

	util.ResponseSuccess(c, "登出成功")
}

// QQLogin QQ OAuth 登录回调
// POST /api/auth/qq/callback
// 对标 QQLoginStrategy.login()
func (h *UserAuthHandler) QQLogin(c *gin.Context) {
	var qqVO dto.QQLoginVO
	if err := c.ShouldBindJSON(&qqVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParam.WithMsg(err.Error()))
		return
	}
	_ = qqVO // TODO: P0-6 OAuth流程

	util.ResponseSuccess(c, map[string]interface{}{
		"token":    "qq_jwt_token",
		"isNewUser": false,
	})
}

// SendVerificationCode 发送邮箱验证码
// POST /api/auth/code
// 用于: 注册绑定邮箱 / 修改邮箱 / 找回密码
func (h *UserAuthHandler) SendVerificationCode(c *gin.Context) {
	var emailVO dto.EmailVO
	if err := c.ShouldBindJSON(&emailVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParam.WithMsg(err.Error()))
		return
	}
	_ = emailVO // TODO: P0-5 生成验证码 → 存入Redis(5分钟过期) → 发送邮件

	util.ResponseSuccess(c, "验证码已发送，请查收邮件")
}

// UpdatePassword 修改密码
// PUT /api/user/password
// 对标 UserController.updatePassword()
func (h *UserAuthHandler) UpdatePassword(c *gin.Context) {
	var passwordVO dto.PasswordVO
	if err := c.ShouldBindJSON(&passwordVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParam.WithMsg(err.Error()))
		return
	}
	_ = passwordVO // TODO: P0-5 校验旧密码 → BCrypt新密码存储

	util.ResponseSuccess(c, "密码修改成功")
}

// ResetPassword 重置密码（通过邮箱验证码）
// PUT /api/auth/password/reset
func (h *UserAuthHandler) ResetPassword(c *gin.Context) {
	var resetVO dto.ResetPasswordVO
	if err := c.ShouldBindJSON(&resetVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParam.WithMsg(err.Error()))
		return
	}
	_ = resetVO // TODO: P0-5 校验验证码 → 更新密码

	util.ResponseSuccess(c, "密码重置成功")
}

// GetUserInfo 获取当前登录用户信息
// GET /api/user/info
// 对标 UserController.getUserInfo()
func (h *UserAuthHandler) GetUserInfo(c *gin.Context) {
	// 从 JWT 中提取 userId
	userID, exists := c.Get("userId")
	if !exists {
		util.ResponseError(c, errors.ErrUnauthorized)
		return
	}
	_ = userID // TODO: P0-5 根据ID查询完整UserInfo

	util.ResponseSuccess(c, map[string]interface{}{
		"id":       userID,
		"nickname": "",
		"avatar":   "",
	})
}
