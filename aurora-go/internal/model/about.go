package model

import (
	"time"

	"gorm.io/gorm"
)

// About 关于页面实体 (对应 t_about 表)
type About struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Content     string    `gorm:"type:longtext" json:"content"`
	HTMLContent string    `gorm:"type:longtext" json:"htmlContent"`
	CreateTime  time.Time `json:"createTime"`
	UpdateTime  time.Time `json:"updateTime"`
}

func (About) TableName() string { return "t_about" }

func (a *About) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	a.CreateTime = now
	a.UpdateTime = now
	return nil
}

func (a *About) BeforeUpdate(tx *gorm.DB) error {
	a.UpdateTime = time.Now()
	return nil
}
