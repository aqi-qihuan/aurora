package model

import "time"

// Portfolio 作品集实体 (对应 t_portfolio 表)
// 通过定时任务从 GitHub API 同步仓库快照，后台可覆盖封面/排序/置顶/可见性
type Portfolio struct {
	ID              uint       `gorm:"primarykey;column:id" json:"id"`
	RepoID          int64      `gorm:"column:repo_id;uniqueIndex" json:"repoId"`
	Name            string     `gorm:"column:name;size:128;not null" json:"name"`
	FullName        string     `gorm:"column:full_name;size:255" json:"fullName"`
	Description     string     `gorm:"column:description;size:500" json:"description"`
	HtmlURL         string     `gorm:"column:html_url;size:500;not null" json:"htmlUrl"`
	Homepage        string     `gorm:"column:homepage;size:500" json:"homepage"`
	Language        string     `gorm:"column:language;size:64" json:"language"`
	StargazersCount int        `gorm:"column:stargazers_count;default:0" json:"stargazersCount"`
	ForksCount      int        `gorm:"column:forks_count;default:0" json:"forksCount"`
	Topics          string     `gorm:"column:topics;type:text" json:"topics"` // JSON 数组字符串
	RepoCreatedAt   *time.Time `gorm:"column:repo_created_at" json:"repoCreatedAt,omitempty"`
	RepoUpdatedAt   *time.Time `gorm:"column:repo_updated_at" json:"repoUpdatedAt,omitempty"`
	Cover           string     `gorm:"column:cover;size:500" json:"cover"`           // 后台自定义封面，覆盖默认
	Sort            int        `gorm:"column:sort;default:0" json:"sort"`            // 排序权重，越大越靠前
	IsFeatured      int8       `gorm:"column:is_featured;default:0" json:"isFeatured"` // 是否首页置顶
	IsVisible       int8       `gorm:"column:is_visible;default:1" json:"isVisible"`   // 是否展示 0隐藏 1展示
	CreateTime      time.Time  `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateTime      time.Time  `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
}

func (Portfolio) TableName() string { return "t_portfolio" }
