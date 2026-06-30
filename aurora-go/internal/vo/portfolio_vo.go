package vo

// PortfolioVO 后台编辑作品集（仅覆盖人工配置项，同步时不触碰这些字段）
type PortfolioVO struct {
	ID         uint   `json:"id" binding:"required"`
	Cover      string `json:"cover,omitempty" binding:"omitempty,max=500"`
	Sort       int    `json:"sort,omitempty"`
	IsFeatured *int8  `json:"isFeatured,omitempty" binding:"omitempty,oneof=0 1"`
	IsVisible  *int8  `json:"isVisible,omitempty" binding:"omitempty,oneof=0 1"`
}
