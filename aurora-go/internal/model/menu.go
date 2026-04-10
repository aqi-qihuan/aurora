package model

import (
	"time"

	"gorm.io/gorm"
)

// Menu 菜单实体 (对应 t_menu 表)
type Menu struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	ParentID   *uint     `gorm:"index;default:0" json:"parentId"`       // 父菜单ID, 0=顶级
	Name       string    `gorm:"size:30;not null" json:"name"`            // 菜单名称
	Path       string    `gorm:"size:100" json:"path"`                   // 前端路由路径
	Component  string    `gorm:"size:200" json:"component"`              // 组件路径
	Icon       string    `gorm:"size:100" json:"icon"`                  // 菜单图标
	Sort       int       `gorm:"default:0" json:"sort"`
	Type       int8      `gorm:"default:1;not null" json:"type"`         // 0目录 1菜单 2按钮
	Permission string    `gorm:"size:100" json:"permission"`             // 权限标识(如: article:list)
	Hidden     int8      `gorm:"default:0" json:"hidden"`               // 是否隐藏(0显示 1隐藏)
	OrderNum   *int      `json:"orderNum"`                              // 排序号
	CreateTime time.Time `json:"createTime"`

	// 关联
	Children []Menu `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Roles    []Role `gorm:"many2many:t_role_menu;" json:"-"`
}

func (Menu) TableName() string { return "t_menu" }

func (m *Menu) BeforeCreate(tx *gorm.DB) error { m.CreateTime = time.Now(); return nil }
