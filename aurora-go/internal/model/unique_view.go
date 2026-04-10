package model

import (
	"time"
)

// UniqueView 每日独立访客统计记录 (对应 t_unique_view 表)
// 对标 Java 版: viewsCount 字段存储当日独立访客数(从Redis Set的SCARD获取)
type UniqueView struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	CreateTime time.Time `gorm:"index;not null;uniqueIndex" json:"createTime"` // 统计日期(精确到天)
	ViewsCount int       `gorm:"default:0;not null" json:"viewsCount"`          // 当日独立访客数量
}

func (UniqueView) TableName() string { return "t_unique_view" }
