package model

import (
	"time"

	"gorm.io/gorm"
)

// Menu 菜单实体 (对应 t_menu 表)
// 数据库实际字段: id, name, path, component, icon, create_time, update_time, order_num, parent_id, is_hidden
type Menu struct {
	ID         uint      `gorm:"primarykey;column:id" json:"id"`
	Name       string    `gorm:"column:name;size:20;not null" json:"name"`
	Path       string    `gorm:"column:path;size:50;not null" json:"path"`
	Component  string    `gorm:"column:component;size:50;not null" json:"component"`
	Icon       string    `gorm:"column:icon;size:50;not null" json:"icon"`
	OrderNum   int       `gorm:"column:order_num;not null" json:"orderNum"`
	ParentID   *uint     `gorm:"column:parent_id" json:"parentId"`
	IsHidden   int8      `gorm:"column:is_hidden;default:0;not null" json:"isHidden"`
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateTime *time.Time `gorm:"column:update_time;autoUpdateTime" json:"updateTime,omitempty"`

	// 关联
	Children []Menu `gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE" json:"children,omitempty"`
	Roles    []Role `gorm:"many2many:t_role_menu;" json:"-"`
}

func (Menu) TableName() string { return "t_menu" }

func (m *Menu) BeforeCreate(tx *gorm.DB) error { m.CreateTime = time.Now(); return nil }
