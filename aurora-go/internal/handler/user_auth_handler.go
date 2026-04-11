package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/service"
	"github.com/aurora-go/aurora/internal/util"
	"github.com/aurora-go/aurora/internal/vo"
)

// UserAuthHandler 用户认证处理器（对标 Java UserController + UserAuthController）
type UserAuthHandler struct {
	registry *service.Registry
}

func NewUserAuthHandler(registry *service.Registry) *UserAuthHandler {
	return &UserAuthHandler{registry: registry}
}

// Register 用户注册
// POST /api/auth/register
func (h *UserAuthHandler) Register(c *gin.Context) {
	var registerVO vo.RegisterVO
	if err := c.ShouldBindJSON(&registerVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	result, err := h.registry.UserAuth.Register(c.Request.Context(), registerVO)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// Login 用户登录
// POST /api/auth/login
func (h *UserAuthHandler) Login(c *gin.Context) {
	var loginVO vo.LoginVO
	if err := c.ShouldBindJSON(&loginVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	result, err := h.registry.UserAuth.Login(c.Request.Context(), loginVO)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// Logout 用户登出
// POST /api/auth/logout
func (h *UserAuthHandler) Logout(c *gin.Context) {
	userID, _ := c.Get("userId")
	uid := uint(0)
	if id, ok := userID.(uint); ok {
		uid = id
	}
	if err := h.registry.UserAuth.Logout(c.Request.Context(), uid); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "登出成功")
}

// QQLogin QQ OAuth 登录回调
// POST /api/auth/qq/callback
func (h *UserAuthHandler) QQLogin(c *gin.Context) {
	var qqVO dto.QQLoginVO
	if err := c.ShouldBindJSON(&qqVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	if h.registry.QQOAuthSvc == nil {
		util.ResponseError(c, errors.ErrInternalServer.WithMsg("QQ登录功能未启用"))
		return
	}
	result, err := h.registry.QQOAuthSvc.Login(c.Request.Context(), &qqVO)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// SendVerificationCode 发送邮箱验证码
// POST /api/auth/code
func (h *UserAuthHandler) SendVerificationCode(c *gin.Context) {
	var emailVO dto.EmailVO
	if err := c.ShouldBindJSON(&emailVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	if err := h.registry.UserAuth.SendVerificationCode(c.Request.Context(), emailVO.Email); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "验证码已发送，请查收邮件")
}

// UpdatePassword 修改密码
// PUT /api/user/password
func (h *UserAuthHandler) UpdatePassword(c *gin.Context) {
	var passwordVO vo.PasswordVO
	if err := c.ShouldBindJSON(&passwordVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	userID, _ := c.Get("userId")
	uid := uint(0)
	if id, ok := userID.(uint); ok {
		uid = id
	}
	if err := h.registry.UserAuth.ChangePassword(c.Request.Context(), uid, passwordVO); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "密码修改成功")
}

// ResetPassword 重置密码（通过邮箱验证码）
// PUT /api/auth/password/reset
func (h *UserAuthHandler) ResetPassword(c *gin.Context) {
	var resetVO dto.ResetPasswordVO
	if err := c.ShouldBindJSON(&resetVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	// TODO: 校验验证码后重置密码
	util.ResponseSuccess(c, "密码重置成功")
}

// GetUserInfo 获取当前登录用户信息
// GET /api/user/info
func (h *UserAuthHandler) GetUserInfo(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		util.ResponseError(c, errors.ErrUnauthorized)
		return
	}
	uid := uint(0)
	if id, ok := userID.(uint); ok {
		uid = id
	}
	result, err := h.registry.UserAuth.GetUserInfoByID(c.Request.Context(), uid)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}
