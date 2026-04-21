package dto

import "time"

// UserInfoInCard 文章卡片中的用户信息精简版（用于ArticleDTO和ArticleCardDTO的author字段）
type UserInfoInCard struct {
	ID       uint   `json:"id"`
	Nickname string `json:"nickname,omitempty"`
	Website  string `json:"website,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
}

// ArticleDTO 文章数据传输对象 (用于API响应)
type ArticleDTO struct {
	ID             uint           `json:"id"`
	UserID         uint           `json:"userId"`
	ArticleCover   string        `json:"articleCover"`
	ArticleTitle   string        `json:"articleTitle"`
	ArticleContent string        `json:"articleContent"`
	IsTop          int8          `json:"isTop"`
	IsFeatured     int8          `json:"isFeatured"`
	Status         int8          `json:"status"`
	Type           int8          `json:"type"`
	ViewCount      uint64        `json:"viewsCount"`
	LikeCount      int64         `json:"likeCount"`
	CategoryID     uint           `json:"categoryId"`
	CategoryName   string        `json:"categoryName,omitempty"`
	Nickname       string        `json:"nickname,omitempty"`         // 兼容旧前端
	Avatar         string        `json:"avatar,omitempty"`             // 兼容旧前端
	Author         *UserInfoInCard `json:"author,omitempty"`        // 前端详情页使用 article.author.nickname/avatar
	Tags           []TagDTO      `json:"tags,omitempty"`
	CreateTime     time.Time     `json:"createTime"`
}

// ArticleCardDTO 文章卡片(列表页精简版) - 完全对标Java版 ArticleCardDTO
type ArticleCardDTO struct {
	ID             uint           `json:"id"`
	ArticleCover   string         `json:"articleCover"`
	ArticleTitle   string         `json:"articleTitle"`
	ArticleContent string         `json:"articleContent"`
	IsTop          int8           `json:"isTop"`
	IsFeatured     int8           `json:"isFeatured"`
	Author         *UserInfoInCard `json:"author,omitempty"`      // 对标Java: UserInfo author（嵌套对象）
	CategoryName   string         `json:"categoryName,omitempty"`
	Tags           []TagInCard    `json:"tags,omitempty"`      // 对标Java: List<Tag> tags（不是tagDTOs）
	Status         int8           `json:"status"`
	CreateTime     time.Time      `json:"createTime"`
	UpdateTime     time.Time      `json:"updateTime,omitempty"`
}

// TagInCard 标签精简版（对标Java Tag实体，用于ArticleCardDTO）
type TagInCard struct {
	ID      uint   `json:"id"`
	TagName string `json:"tagName,omitempty"`
}

// ArticleAdminDTO 后台文章管理DTO（完全对标Java ArticleAdminDTO）
// 用于: GET /api/admin/articles 后台文章列表
type ArticleAdminDTO struct {
	ID             uint       `json:"id"`
	ArticleCover   string     `json:"articleCover"`
	ArticleTitle   string     `json:"articleTitle"`
	CreateTime     time.Time `json:"createTime"`
	ViewsCount     int        `json:"viewsCount"`
	CategoryName   string     `json:"categoryName"`
	TagDTOs        []TagDTO   `json:"tagDTOs"`
	IsTop          int8       `json:"isTop"`
	IsFeatured     int8       `json:"isFeatured"`
	IsDelete       int8       `json:"isDelete"`
	Status         int8       `json:"status"`
	Type           int8       `json:"type"`
}

// ArticleSearchDTO ES搜索结果DTO
type ArticleSearchDTO struct {
	ID             uint     `json:"id"`
	ArticleTitle   string   `json:"articleTitle"`
	ArticleContent string   `json:"articleContent"`
	Highlight      []string `json:"highlight,omitempty"`
	Score          float64  `json:"score"`
}

// TagDTO 标签DTO
type TagDTO struct {
	ID           uint      `json:"id"`
	TagName      string    `json:"tagName"`
	ArticleCount int       `json:"articleCount,omitempty"`
	CreateTime   time.Time `json:"createTime,omitempty"`
}

// ArticleAdminViewDTO 后台文章编辑详情DTO（对标Java ArticleAdminViewDTO）
// 用于: GET /api/admin/articles/:id 编辑文章回显
type ArticleAdminViewDTO struct {
	ID              uint     `json:"id"`
	ArticleTitle    string   `json:"articleTitle"`
	ArticleAbstract string   `json:"articleAbstract,omitempty"`
	ArticleContent  string   `json:"articleContent"`
	ArticleCover    string   `json:"articleCover,omitempty"`
	CategoryName    string   `json:"categoryName,omitempty"`
	TagNames        []string `json:"tagNames"` // 对标Java: List<String> tagNames
	IsTop           int8     `json:"isTop"`
	IsFeatured      int8     `json:"isFeatured"`
	Status          int8     `json:"status"`
	Type            int8     `json:"type"`
	OriginalURL     string   `json:"originalUrl,omitempty"`
	Password        string   `json:"password,omitempty"`
}
