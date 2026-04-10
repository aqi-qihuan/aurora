package model

import (
	"time"

	"gorm.io/gorm"
)

// FriendLink 友链实体 (对应 t_friend_link 表)
type FriendLink struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	UserID      *uint     `json:"userId,omitempty"`                     // 申请者用户ID
	LinkName    string    `gorm:"size:50;not null" json:"linkName"`
	LinkAvatar  string    `gorm:"size:1024" json:"linkAvatar"`
	LinkAddress string    `gorm:"size:500;not null" json:"linkAddress"`
	LinkIntro   string    `gorm:"size:500" json:"linkIntro"`
	Status      int8      `gorm:"default:0;index" json:"status"`         // 0审核中 1通过 -1拒绝
	CreateTime  time.Time `json:"createTime"`

	// 关联
	UserInfo *UserInfo `gorm:"foreignKey:UserID" json:"userInfo,omitempty"` // 申请者
}

func (FriendLink) TableName() string { return "t_friend_link" }

func (f *FriendLink) BeforeCreate(tx *gorm.DB) error { f.CreateTime = time.Now(); return nil }
