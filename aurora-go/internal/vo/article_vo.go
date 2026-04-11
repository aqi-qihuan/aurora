package vo

// ArticleVO 文章发布/编辑请求
type ArticleVO struct {
	ID             uint   `json:"id,omitempty"`              // 编辑时传
	ArticleTitle   string `json:"articleTitle" binding:"required,max=200"`
	ArticleContent string `json:"articleContent" binding:"required"`
	CategoryID     uint   `json:"categoryId" binding:"required"`
	TagIDs         []uint `json:"tagIds"`                     // 标签ID列表
	Type           int8   `json:"type" binding:"omitempty,oneof=1 2 3"`
	OriginalURL    string `json:"originalUrl,omitempty" binding:"omitempty,max=500"`
	IsTop          int8   `json:"isTop" binding:"omitempty,oneof=0 1"`
	IsFeatured     int8   `json:"isFeatured" binding:"omitempty,oneof=0 1"`
	Status         *int8  `json:"status" binding:"omitempty,oneof=0 1 2 3"`
	Password       string `json:"password,omitempty" binding:"omitempty,max=64"` // 密码保护文章
}

// ArticleTopFeaturedVO 置顶/推荐操作
type ArticleTopFeaturedVO struct {
	ID       uint `json:"id" binding:"required"`
	IsTop    int8 `json:"isTop" binding:"required,oneof=0 1"`
	IsFeatured int8 `json:"isFeatured" binding:"required,oneof=0 1"`
}
