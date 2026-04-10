package strategy

import (
	"context"

	"github.com/aurora-go/aurora/internal/dto"
)

// SearchStrategy 搜索策略接口 (对标Java com.aurora.strategy.SearchStrategy)
// 定义文章搜索的统一行为契约，支持多种后端实现（ES/MySQL）
type SearchStrategy interface {
	// SearchArticle 搜索文章（支持高亮）
	// keywords: 搜索关键词（支持中文分词）
	SearchArticle(ctx context.Context, keywords string) ([]dto.ArticleSearchDTO, error)
}
