package model

import (
	"time"

	"gorm.io/gorm"
)

// Resource 资源实体 (对应 t_resource 表)
type Resource struct {
	ID            uint   `gorm:"primarykey" json:"id"`
	ResourceName  string `gorm:"size:100;not null;uniqueIndex" json:"resourceName"`
	URL           string `gorm:"size:500;not null" json:"url"`
	RequestMethod string `gorm:"size:10;not null;index" json:"requestMethod"` // GET/POST/PUT/DELETE
	CategoryID    uint   `gorm:"index" json:"categoryId"`
	Description   string `gorm:"size:500" json:"description"`
	Size          int64  `json:"size"`                                    // 文件大小(bytes)
	Nickname      string `gorm:"size:50" json:"nickname"`
	IsDelete      int8   `gorm:"default:0;index" json:"isDelete"`
	CreateTime    time.Time `json:"createTime"`

	// 关联
	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
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
