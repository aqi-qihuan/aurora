package model

import (
	"time"

	"gorm.io/gorm"
)

// Category 分类实体 (对应 t_category 表)
type Category struct {
	ID           uint       `gorm:"primarykey" json:"id"`
	CategoryName string     `gorm:"size:20;not null;uniqueIndex" json:"categoryName"`
	CreateTime   time.Time  `json:"createTime"`
	UpdateTime   time.Time  `json:"updateTime"`

	// 关联
	Articles []Article `gorm:"foreignKey:CategoryID" json:"articles,omitempty"`
}

func (Category) TableName() string { return "t_category" }

func (c *Category) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	c.CreateTime = now
	c.UpdateTime = now
	return nil
}

func (c *Category) BeforeUpdate(tx *gorm.DB) error {
	c.UpdateTime = time.Now()
	return nil
}
