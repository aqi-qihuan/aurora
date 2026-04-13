package model

import (
	"time"

	"gorm.io/gorm"
)

// Role 角色实体 (对应 t_role 表)
type Role struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	RoleName    string    `gorm:"column:role_name;size:20;not null;uniqueIndex" json:"roleName"`
	IsDisable   int8      `gorm:"column:is_disable;default:0" json:"isDisable"`
	CreateTime  time.Time `gorm:"column:create_time" json:"createTime"`
	UpdateTime  *time.Time `gorm:"column:update_time" json:"updateTime,omitempty"`

	// 关联（仅用于 GORM Preload/Association，不对应 t_role 表字段）
	Menus []Menu `gorm:"many2many:t_role_menu;" json:"menus,omitempty"`
}

func (Role) TableName() string { return "t_role" }

// RoleMenu 角色-菜单关联表
type RoleMenu struct {
	RoleID uint `gorm:"primaryKey" json:"roleId"`
	MenuID uint `gorm:"primaryKey" json:"menuId"`
}
func (RoleMenu) TableName() string { return "t_role_menu" }

func (r *Role) BeforeCreate(tx *gorm.DB) error { r.CreateTime = time.Now(); return nil }
func (r *Role) BeforeUpdate(tx *gorm.DB) error { now := time.Now(); r.UpdateTime = &now; return nil }
