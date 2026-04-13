package model

import (
	"time"

	"gorm.io/gorm"
)

// Resource 资源实体 (对应 t_resource 表)
type Resource struct {
	ID            uint      `gorm:"primarykey;column:id" json:"id"`
	ResourceName  string    `gorm:"size:50;not null;column:resource_name" json:"resourceName"`
	URL           string    `gorm:"size:255;column:url" json:"url"`
	RequestMethod string    `gorm:"size:10;column:request_method" json:"requestMethod"` // GET/POST/PUT/DELETE
	ParentID      *uint     `gorm:"column:parent_id" json:"parentId"`                    // 父模块id（模块的parentId为null）
	IsAnonymous   int8      `gorm:"column:is_anonymous;default:0" json:"isAnonymous"`   // 是否匿名访问 0否 1是
	CreateTime    time.Time `gorm:"column:create_time" json:"createTime"`
	UpdateTime    *time.Time `gorm:"column:update_time" json:"updateTime,omitempty"`
}

func (Resource) TableName() string { return "t_resource" }

// RoleResource 角色-资源权限关联表
type RoleResource struct {
	RoleID      uint `gorm:"primaryKey" json:"roleId"`
	ResourceID uint `gorm:"primaryKey" json:"resourceId"`
	IsForbidden int8 `gorm:"default:0" json:"isForbidden"` // 0允许 -1禁止
}
func (RoleResource) TableName() string { return "t_role_resource" }

func (r *Resource) BeforeCreate(tx *gorm.DB) error { r.CreateTime = time.Now(); return nil }
