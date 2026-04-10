package dto

import "time"

// ArticleDTO 文章数据传输对象 (用于API响应)
type ArticleDTO struct {
	ID             uint       `json:"id"`
	UserID         uint       `json:"userId"`
	ArticleCover   string     `json:"articleCover"`
	ArticleTitle   string     `json:"articleTitle"`
	ArticleContent string     `json:"articleContent"`
	IsTop          int8       `json:"isTop"`
	IsFeatured     int8       `json:"isFeatured"`
	Status         int8       `json:"status"`
	Type           int8       `json:"type"`
	ViewCount      uint64     `json:"viewCount"`
	LikeCount      int64      `json:"likeCount"`
	CategoryID     uint       `json:"categoryId"`
	CategoryName   string     `json:"categoryName,omitempty"`
	Nickname       string     `json:"nickname,omitempty"`
	Avatar         string     `json:"avatar,omitempty"`
	Tags           []TagDTO   `json:"tags,omitempty"`
	CreateTime     time.Time  `json:"createTime"`
}

// ArticleCardDTO 文章卡片(列表页精简版)
type ArticleCardDTO struct {
	ID           uint     `json:"id"`
	ArticleTitle string   `json:"articleTitle"`
	ArticleCover string   `json:"articleCover"`
	IsTop        int8     `json:"isTop"`
	IsFeatured   int8     `json:"isFeatured"`
	Status       int8     `json:"status"`
	ViewCount    uint64   `json:"viewCount"`
	CategoryName string   `json:"categoryName,omitempty"`
	Nickname     string   `json:"nickname,omitempty"`
	Tags         []TagDTO `json:"tags,omitempty"`
	CreateTime   time.Time `json:"createTime"`
}

// ArticleSearchDTO ES搜索结果DTO
type ArticleSearchDTO struct {
	ID             uint     `json:"id"`
	ArticleTitle   string   `json:"articleTitle"`
	ArticleContent string   `json:"articleContent"` // 高亮片段
	Highlight      []string `json:"highlight,omitempty"`
	Score          float64  `json:"score"` // 相关性评分
}

// TagDTO 标签DTO
type TagDTO struct {
	ID       uint   `json:"id"`
	TagName  string `json:"tagName"`
}
