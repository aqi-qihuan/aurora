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

// ArticleCardDTO 文章卡片(列表页精简版) - 完全对标Java版 ArticleCardDTO
type ArticleCardDTO struct {
	ID             uint       `json:"id"`
	ArticleCover   string     `json:"articleCover"`
	ArticleTitle   string     `json:"articleTitle"`
	ArticleContent string     `json:"articleContent"`
	IsTop          int8       `json:"isTop"`
	IsFeatured     int8       `json:"isFeatured"`
	Author         *UserInfoInCard `json:"author,omitempty"`      // 对标Java: UserInfo author（嵌套对象）
	CategoryName   string     `json:"categoryName,omitempty"`
	Tags           []TagInCard `json:"tags,omitempty"`      // 对标Java: List<Tag> tags（不是tagDTOs）
	Status         int8       `json:"status"`
	CreateTime     time.Time  `json:"createTime"`
	UpdateTime     time.Time  `json:"updateTime,omitempty"`
}

// UserInfoInCard 文章卡片中的用户信息精简版（用于ArticleCardDTO的author字段）
type UserInfoInCard struct {
	ID       uint   `json:"id"`
	Nickname string `json:"nickname,omitempty"`
	Website  string `json:"website,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
}

// ArticleAdminDTO 后台文章管理DTO（完全对标Java ArticleAdminDTO）
// 用于: GET /api/admin/articles 后台文章列表
// 注意: 字段名必须与Java完全一致，否则前端无法渲染
type ArticleAdminDTO struct {
	ID             uint       `json:"id"`
	ArticleCover   string     `json:"articleCover"`
	ArticleTitle   string     `json:"articleTitle"`
	CreateTime     time.Time  `json:"createTime"`
	ViewsCount     int        `json:"viewsCount"` // 浏览量
	CategoryName   string     `json:"categoryName"`
	TagDTOs        []TagDTO   `json:"tagDTOs"`    // 标签列表（注意是tagDTOs不是tags）
	IsTop          int8       `json:"isTop"`      // 置顶
	IsFeatured     int8       `json:"isFeatured"` // 推荐
	IsDelete       int8       `json:"isDelete"`   // 删除状态
	Status         int8       `json:"status"`     // 状态
	Type           int8       `json:"type"`       // 类型 1原创 2转载 3翻译
}

// TagInCard 标签精简版（对标Java Tag实体，用于ArticleCardDTO）
type TagInCard struct {
	ID      uint   `json:"id"`
	TagName string `json:"tagName,omitempty"`
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
