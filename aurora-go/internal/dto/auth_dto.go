package dto

import (
	"time"
)

// ===== JWT/认证相关DTO (对标Java UserDetailsDTO + SocialTokenDTO) =====

// UserDetailsDTO 用户详情(存入Redis Session, 对标Java UserDetailsDTO)
// 包含: 基本信息、角色列表、权限列表、Session过期时间
type UserDetailsDTO struct {
	ID           uint      `json:"id"`
	UserInfoID    uint      `json:"userInfoId"`
	Email        string    `json:"email"`
	LoginType    int       `json:"loginType"`     // 1邮箱 3QQ
	Username     string    `json:"username"`
	Password     string    `json:"-"`
	Roles        []string  `json:"roles"`         // 角色名列表 [admin,user]
	Nickname     string    `json:"nickname"`
	Avatar       string    `json:"avatar"`
	Intro        string    `json:"intro"`
	Website      string    `json:"website"`
	IsSubscribe  int8      `json:"isSubscribe"`   // 是否订阅
	IPAddress    string    `json:"ipAddress"`
	IPSource     string    `json:"ipSource"`
	IsDisable    int8      `json:"isDisable"`     // 0正常 1禁用
	Browser      string    `json:"browser"`
	OS           string    `json:"os"`
	ExpireTime   time.Time `json:"expireTime"`    // Session过期时间
	LastLoginTime time.Time `json:"lastLoginTime"`
}

// PermissionsList 返回角色+权限的合并列表 (对标Java getAuthorities)
func (u *UserDetailsDTO) PermissionsList() []string {
	if len(u.Roles) == 0 {
		return []string{}
	}
	return u.Roles // 简化版: 角色即权限 (后续可扩展为 role:permission)
}

// ToUserDetails 返回自身 (适配接口转换)
func (u *UserDetailsDTO) ToUserDetails() *UserDetailsDTO {
	return u
}

// IsAdmin 检查是否为管理员
func (u *UserDetailsDTO) IsAdmin() bool {
	for _, r := range u.Roles {
		if r == "admin" {
			return true
		}
	}
	return false
}

// ===== OAuth DTO =====

// QQLoginVO QQ登录请求VO (前端传accessToken+openID)
type QQLoginVO struct {
	AccessToken string `json:"access_token" binding:"required"` // QQ OAuth access_token
	OpenID      string `json:"open_id" binding:"required"`      // QQ用户唯一标识
}

// QQUserInfoDTO QQ用户信息响应 (对标Java QQUserInfoDTO)
type QQUserInfoDTO struct {
	Nickname     string `json:"nickname"`              // QQ昵称
	Figureurl    string `json:"figureurl"`             // 头像(30x30)
	Figureurl_1  string `json:"figureurl_1"`           // 头像(100x100)
	Figureurl_2  string `json:"figureurl_2"`           // 头像(40x40)
	FigureurlQQ1 string `json:"figureurl_qq_1"`        // QQ空间头像(100x100)
	Gender       string `json:"gender"`                // 性别
	City         string `json:"city"`                  // 城市
	Province     string `json:"province"`               // 省份
	Year         string `json:"year"`                  // 出生年份
}

// SocialTokenDTO 社交登录Token信息 (对标Java SocialTokenDTO)
type SocialTokenDTO struct {
	OpenID    string `json:"openId"`    // 第三方平台用户标识
	AccessToken string `json:"accessToken"` // 平台访问令牌
	LoginType int    `json:"loginType"`  // 登录类型: 3=QQ
}

// SocialUserInfoDTO 社交用户信息 (对标Java SocialUserInfoDTO)
type SocialUserInfoDTO struct {
	Nickname string `json:"nickname"` // 昵称
	Avatar   string `json:"avatar"`   // 头像URL
}

// UserInfoDTO 用户信息DTO (用于登录/社交注册返回, 对标Java UserInfoDTO)
type UserInfoDTO struct {
	ID       uint   `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Email    string `json:"email,omitempty"`
	Token    string `json:"token,omitempty"`
}
