package strategy

import (
	"context"
	"fmt"
	"log/slog"
)

// EsSearchStrategy ES 搜索策略实现（对标 Java EsSearchStrategyImpl）
type EsSearchStrategy struct {
	esClient ESClientInterface
}

// ESClientInterface ES 客户端接口（避免循环导入）
type ESClientInterface interface {
	SearchArticles(ctx context.Context, keywords string, page, size int) ([]map[string]interface{}, int, error)
	IndexArticle(ctx context.Context, articleID uint, title, content string, isDelete, status int8) error
	DeleteArticle(ctx context.Context, articleID uint) error
}

// NewEsSearchStrategy 创建 ES 搜索策略
func NewEsSearchStrategy(esClient ESClientInterface) *EsSearchStrategy {
	return &EsSearchStrategy{
		esClient: esClient,
	}
}

// SearchArticles 搜索文章（对标 Java EsSearchStrategyImpl.searchArticle）
func (s *EsSearchStrategy) SearchArticles(ctx context.Context, keywords string, current, size int) ([]map[string]interface{}, int, error) {
	if s.esClient == nil {
		return nil, 0, fmt.Errorf("ES 服务未初始化")
	}

	articles, total, err := s.esClient.SearchArticles(ctx, keywords, current, size)
	if err != nil {
		slog.Error("ES 搜索失败", "error", err, "keywords", keywords)
		return nil, 0, fmt.Errorf("ES 搜索失败: %w", err)
	}

	slog.Info("ES 搜索成功", "keywords", keywords, "total", total, "page", current)
	return articles, total, nil
}

// IndexArticle 索引文章（供 MaxWell Consumer 调用）
func (s *EsSearchStrategy) IndexArticle(ctx context.Context, articleID uint, title, content string, isDelete, status int8) error {
	if s.esClient == nil {
		return fmt.Errorf("ES 服务未初始化")
	}

	return s.esClient.IndexArticle(ctx, articleID, title, content, isDelete, status)
}

// DeleteArticle 删除文章索引（供 MaxWell Consumer 调用）
func (s *EsSearchStrategy) DeleteArticle(ctx context.Context, articleID uint) error {
	if s.esClient == nil {
		return fmt.Errorf("ES 服务未初始化")
	}

	return s.esClient.DeleteArticle(ctx, articleID)
}
