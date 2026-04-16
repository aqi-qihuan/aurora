package strategy

import (
	"context"
)

// SearchStrategy 搜索策略接口 (对标Java com.aurora.strategy.SearchStrategy)
// 定义文章搜索的统一行为契约，支持多种后端实现（ES/MySQL）
type SearchStrategy interface {
	// SearchArticles 搜索文章（支持高亮、分页）
	// keywords: 搜索关键词（支持中文分词）
	// current: 当前页码
	// size: 每页大小
	SearchArticles(ctx context.Context, keywords string, current, size int) ([]map[string]interface{}, int, error)
}
