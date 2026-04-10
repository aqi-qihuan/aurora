package model

import (
	"time"

	"gorm.io/gorm"
)

// UserInfo 用户信息实体 (对应 t_user_info 表)
type UserInfo struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	Email         string         `gorm:"size:50;uniqueIndex" json:"email"`
	Nickname      string         `gorm:"size:30;not null" json:"nickname"`
	Avatar        string         `gorm:"size:1024" json:"avatar"`
	Intro         string         `gorm:"size:500" json:"intro"`          // 个人简介
	WebSite       string         `gorm:"size:255" json:"webSite"`        // 个人网站
	IsDisable     int8           `gorm:"default:0;index" json:"isDisable"`
	BlogLinkCount int            `gorm:"default:0" json:"blogLinkCount"`  // 友链数
	ArticleCount  int            `gorm:"default:0" json:"articleCount"`
	CategoryCount int            `gorm:"default:0" json:"categoryCount"`
	TagCount      int            `gorm:"default:0" json:"tagCount"`
	IsSubscribe   int8           `gorm:"default:0;index" json:"isSubscribe"` // 是否订阅新文章通知
	IsDelete      int8           `gorm:"default:0;index" json:"isDelete"`     // 软删除标记
	CreateTime    time.Time      `json:"createTime"`
	UpdateTime    time.Time      `json:"updateTime"`

	// 关联
	Roles []Role `gorm:"many22:t_user_role;" json:"roles,omitempty"`
}

func (UserInfo) TableName() string {
	return "t_user_info"
}

// UserAuth 用户认证信息 (对应 t_user_auth 表, 支持多登录方式)
type UserAuth struct {
	ID           uint       `gorm:"primarykey" json:"id"`
	UserID       uint       `gorm:"index;not null" json:"userId"`
	Username     string     `gorm:"size:50;uniqueIndex" json:"username"`  // 登录用户名
	Password     string     `gorm:"size:100" json:"password"`            // BCrypt加密
	LoginType    int8       `gorm:"not null;index" json:"loginType"`     // 1邮箱 2QQ 3手机
	UUID         string     `gorm:"size:64;index" json:"uid"`           // 第三方登录UUID
	LastLoginTime *time.Time `json:"lastLoginTime,omitempty"`           // 最后登录时间
	IPAddress    string     `gorm:"size:64" json:"ipAddress,omitempty"`   // 登录IP
	IPSource     string     `gorm:"size:50" json:"ipSource,omitempty"`   // IP归属地

	// 关联
	UserInfo *UserInfo `gorm:"foreignKey:UserID" json:"userInfo,omitempty"`
}

func (UserAuth) TableName() string {
	return "t_user_auth"
}

// UserRole 用户-角色关联表 (多对多)
type UserRole struct {
 UserID uint `gorm:"primaryKey"json:"userId"`
 RoleID uint `gorm:"primaryKey"json:"roleId"`
}

func (UserRole) TableName() string { return "t_user_role" }

// BeforeCreate GORM钩子
func (u *UserInfo) BeforeCreate(tx *gorm.DB) error {
 now := time.Now()
 u.CreateTime = now
 u.UpdateTime = now
 return nil
}

func (u *UserInfo) BeforeUpdate(tx *gorm.DB) error {
 u.UpdateTime = time.Now()
 return nil
}
