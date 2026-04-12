package model

import (
	"time"

	"gorm.io/gorm"
)

// Talk 说说/微语实体 (对应 t_talk 表)
type Talk struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	UserID     uint      `gorm:"index" json:"userId"`
	Content    string    `gorm:"type:longtext;not null" json:"content"`
	IsTop      int8      `gorm:"default:0" json:"isTop"`
	Status     int8      `gorm:"default:1" json:"status"`       // 1公开 2私密
	CreateTime time.Time `json:"createTime"`

	// 关联
	UserInfo *UserInfo `gorm:"foreignKey:UserID" json:"userInfo,omitempty"`
}

func (Talk) TableName() string { return "t_talk" }

func (t *Talk) BeforeCreate(tx *gorm.DB) error { t.CreateTime = time.Now(); return nil }
