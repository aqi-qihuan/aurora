package model

import (
	"time"

	"gorm.io/gorm"
)

// Tag 标签实体 (对应 t_tag 表)
// 注意: 数据库表只有 id, tag_name, create_time, update_time
type Tag struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	TagName    string    `gorm:"size:20;not null" json:"tagName"`
	CreateTime time.Time `json:"createTime"`
	UpdateTime time.Time `json:"updateTime"`
}

func (Tag) TableName() string { return "t_tag" }

func (t *Tag) BeforeCreate(tx *gorm.DB) error { now := time.Now(); t.CreateTime = now; t.UpdateTime = now; return nil }
func (t *Tag) BeforeUpdate(tx *gorm.DB) error { t.UpdateTime = time.Now(); return nil }
