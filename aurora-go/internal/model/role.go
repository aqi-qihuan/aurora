package model

import (
	"time"

	"gorm.io/gorm"
)

// Role 角色实体 (对应 t_role 表)
type Role struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	RoleName    string    `gorm:"size:30;not null;uniqueIndex" json:"roleName"`
	RoleLabel   string    `gorm:"size:50;not null" json:"roleLabel"`   // 角色标签(如:超级管理员)
	Description  string    `gorm:"size:200" json:"description"`        // 角色描述
	IsDisable   int8      `gorm:"default:0" json:"isDisable"`
	IsDefault   int8      `gorm:"default:0" json:"isDefault"`         // 是否默认角色
	CreateTime  time.Time `json:"createTime"`

	// 关联
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
