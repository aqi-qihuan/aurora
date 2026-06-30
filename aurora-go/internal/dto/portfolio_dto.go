package dto

import "time"

// PortfolioDTO 前台作品集展示 DTO（首页/列表页）
type PortfolioDTO struct {
	ID              uint       `json:"id"`
	Name            string     `json:"name"`
	FullName        string     `json:"fullName,omitempty"`
	Description     string     `json:"description,omitempty"`
	HtmlURL         string     `json:"htmlUrl"`
	Homepage        string     `json:"homepage,omitempty"`
	Language        string     `json:"language,omitempty"`
	StargazersCount int        `json:"stargazersCount"`
	ForksCount      int        `json:"forksCount"`
	Topics          []string   `json:"topics,omitempty"`
	RepoUpdatedAt   *time.Time `json:"repoUpdatedAt,omitempty"`
	Cover           string     `json:"cover,omitempty"`
	Sort            int        `json:"sort,omitempty"`
	IsFeatured      int8       `json:"isFeatured,omitempty"`
}

// PortfolioAdminDTO 后台作品集管理 DTO（含隐藏项与可见性）
type PortfolioAdminDTO struct {
	ID              uint       `json:"id"`
	RepoID          int64      `json:"repoId"`
	Name            string     `json:"name"`
	FullName        string     `json:"fullName,omitempty"`
	Description     string     `json:"description,omitempty"`
	HtmlURL         string     `json:"htmlUrl"`
	Homepage        string     `json:"homepage,omitempty"`
	Language        string     `json:"language,omitempty"`
	StargazersCount int        `json:"stargazersCount"`
	ForksCount      int        `json:"forksCount"`
	Topics          []string   `json:"topics,omitempty"`
	RepoCreatedAt   *time.Time `json:"repoCreatedAt,omitempty"`
	RepoUpdatedAt   *time.Time `json:"repoUpdatedAt,omitempty"`
	Cover           string     `json:"cover,omitempty"`
	Sort            int        `json:"sort"`
	IsFeatured      int8       `json:"isFeatured"`
	IsVisible       int8       `json:"isVisible"`
	CreateTime      time.Time  `json:"createTime"`
}
