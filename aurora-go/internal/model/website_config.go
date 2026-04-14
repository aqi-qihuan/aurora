package model

import (
	"time"

	"gorm.io/gorm"
)

// WebsiteConfig 网站配置实体 (对应 t_website_config 表)
// 对标Java: config字段存储JSON字符串，包含所有配置信息
type WebsiteConfig struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	Config     string    `gorm:"type:varchar(2000)" json:"config"` // 配置信息（JSON格式）
	CreateTime time.Time `json:"createTime"`
	UpdateTime time.Time `json:"updateTime"`
}

func (WebsiteConfig) TableName() string { return "t_website_config" }

func (w *WebsiteConfig) BeforeCreate(tx *gorm.DB) error { now := time.Now(); w.CreateTime = now; w.UpdateTime = now; return nil }
func (w *WebsiteConfig) BeforeUpdate(tx *gorm.DB) error { w.UpdateTime = time.Now(); return nil }
