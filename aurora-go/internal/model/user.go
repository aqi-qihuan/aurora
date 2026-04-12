package model

import (
	"time"

	"gorm.io/gorm"
)

// UserInfo 用户信息实体 (对应 t_user_info 表)
type UserInfo struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	Email         string    `gorm:"size:50" json:"email"`
	Nickname      string    `gorm:"size:50;not null" json:"nickname"`
	Avatar        string    `gorm:"size:1024;not null" json:"avatar"`
	Intro         string    `gorm:"size:255" json:"intro"`          // 个人简介
	Website       string    `gorm:"size:255" json:"website"`        // 个人网站
	IsSubscribe   int8      `gorm:"default:0" json:"isSubscribe"`   // 是否订阅
	IsDisable     int8      `gorm:"default:0;not null" json:"isDisable"`
	CreateTime    time.Time `json:"createTime"`
	UpdateTime    time.Time `json:"updateTime,omitempty"`

	// 关联
	Roles []Role `gorm:"many2many:t_user_role;foreignKey:ID;joinForeignKey:user_id;References:ID;joinReferences:role_id" json:"roles,omitempty"`
}

func (UserInfo) TableName() string {
	return "t_user_info"
}

// UserAuth 用户认证信息 (对应 t_user_auth 表，支持多登录方式)
type UserAuth struct {
	ID            uint       `gorm:"primarykey" json:"id"`
	UserID        uint       `gorm:"column:user_info_id;index;not null" json:"userId"` // 对应数据库 user_info_id
	Username      string     `gorm:"size:50;uniqueIndex" json:"username"`              // 登录用户名
	Password      string     `gorm:"size:100;not null" json:"password"`                // BCrypt 加密
	LoginType     int8       `gorm:"not null" json:"loginType"`                        // 1 邮箱 2QQ 3 手机
	LastLoginTime *time.Time `json:"lastLoginTime,omitempty"`                          // 最后登录时间
	IPAddress     string     `gorm:"size:255" json:"ipAddress,omitempty"`              // 登录 IP
	IPSource      string     `gorm:"size:255" json:"ipSource,omitempty"`               // IP 归属地
	CreateTime    time.Time  `json:"createTime"`
	UpdateTime    time.Time  `json:"updateTime,omitempty"`

	// 关联
	UserInfo *UserInfo `gorm:"foreignKey:UserID" json:"userInfo,omitempty"`
}

func (UserAuth) TableName() string {
	return "t_user_auth"
}

// UserRole 用户-角色关联表 (多对多)
type UserRole struct {
	UserID uint `gorm:"primaryKey" json:"userId"`
	RoleID uint `gorm:"primaryKey" json:"roleId"`
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
