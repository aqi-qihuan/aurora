package vo

// ArticleVO 文章发布/编辑请求（对齐Java版ArticleVO）
type ArticleVO struct {
	ID              uint     `json:"id,omitempty"`                   // 编辑时传
	ArticleTitle    string   `json:"articleTitle" binding:"required,max=200"`
	ArticleContent  string   `json:"articleContent" binding:"required"`
	ArticleAbstract string   `json:"articleAbstract,omitempty"`      // 文章摘要
	ArticleCover    string   `json:"articleCover,omitempty"`         // 文章封面
	CategoryName    string   `json:"categoryName,omitempty"`         // 分类名称（对齐Java）
	TagNames        []string `json:"tagNames,omitempty"`             // 标签名称列表（对齐Java）
	Type            int8     `json:"type" binding:"omitempty,oneof=1 2 3"`
	OriginalURL     string   `json:"originalUrl,omitempty" binding:"omitempty,max=500"`
	IsTop           int8     `json:"isTop" binding:"omitempty,oneof=0 1"`
	IsFeatured      int8     `json:"isFeatured" binding:"omitempty,oneof=0 1"`
	Status          *int8    `json:"status" binding:"omitempty,oneof=0 1 2 3"`
	Password        string   `json:"password,omitempty" binding:"omitempty,max=64"` // 密码保护文章
}

// ArticleTopFeaturedVO 置顶/推荐操作（对标Java ArticleTopFeaturedVO）
// 注意: 使用指针类型 *int8，对标Java Integer（对象类型），避免 required 验证将0视为空值
type ArticleTopFeaturedVO struct {
	ID         uint   `json:"id" binding:"required"`
	IsTop      *int8  `json:"isTop" binding:"required,oneof=0 1"`       // 指针类型，对标Java Integer
	IsFeatured *int8  `json:"isFeatured" binding:"required,oneof=0 1"` // 指针类型，对标Java Integer
}
