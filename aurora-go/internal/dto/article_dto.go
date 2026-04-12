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
	ViewCount      uint64     `json:"viewsCount"`
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
	IsDelete     int8     `json:"isDelete"` // 新增：前端操作列依赖此字段
	Status       int8     `json:"status"`
	Type         int8     `json:"type"`
	ViewCount    uint64   `json:"viewsCount"`
	CategoryName string   `json:"categoryName,omitempty"`
	Nickname     string   `json:"nickname,omitempty"`
	TagDTOs      []TagDTO `json:"tagDTOs,omitempty"` // 修改：对齐 Java 字段名
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
	ID           uint      `json:"id"`
	TagName      string    `json:"tagName"`
	ArticleCount int       `json:"articleCount,omitempty"` // 新增：文章数量
	CreateTime   time.Time `json:"createTime,omitempty"`
}
