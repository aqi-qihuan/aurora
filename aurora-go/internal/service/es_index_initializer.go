package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// ESIndexInitializer Elasticsearch 索引初始化器（对标 Java ElasticsearchIndexInitializer）
type ESIndexInitializer struct {
	esService *ESService
	db        *gorm.DB
}

// NewESIndexInitializer 创建索引初始化器
func NewESIndexInitializer(esService *ESService, db *gorm.DB) *ESIndexInitializer {
	return &ESIndexInitializer{
		esService: esService,
		db:        db,
	}
}

// Initialize 初始化索引并同步数据
func (init *ESIndexInitializer) Initialize(ctx context.Context) error {
	log.Println("🚀 开始初始化 Elasticsearch 索引...")

	// 1. 检查并创建 article 索引
	if err := init.esService.InitArticleIndex(ctx); err != nil {
		return fmt.Errorf("初始化 article 索引失败: %w", err)
	}

	// 2. 同步已有文章数据
	if err := init.SyncExistingArticles(ctx); err != nil {
		log.Printf("⚠️  同步已有文章数据失败: %v", err)
		// 不阻断启动，允许后续手动同步
	}

	log.Println("✅ Elasticsearch 索引初始化完成")
	return nil
}

// SyncExistingArticles 同步已有文章数据到 ES（对标 Java syncExistingArticles）
func (init *ESIndexInitializer) SyncExistingArticles(ctx context.Context) error {
	log.Println("📊 开始同步已有文章数据到 ES...")

	// 查询所有已发布且未删除的文章
	type ArticleData struct {
		ID             uint   `gorm:"column:id"`
		ArticleTitle   string `gorm:"column:article_title"`
		ArticleContent string `gorm:"column:article_content"`
		IsDelete       int8   `gorm:"column:is_delete"`
		Status         int8   `gorm:"column:status"`
	}

	var articles []ArticleData
	result := init.db.Table("t_article").
		Select("id, article_title, article_content, is_delete, status").
		Where("is_delete = ? AND status = ?", 0, 1). // 未删除且公开
		Find(&articles)

	if result.Error != nil {
		return fmt.Errorf("查询文章失败: %w", result.Error)
	}

	if len(articles) == 0 {
		log.Println("ℹ️  没有需要同步的文章")
		return nil
	}

	log.Printf("📝 找到 %d 篇文章待同步", len(articles))

	// 批量索引（每批 100 篇）
	batchSize := 100
	totalSynced := 0

	for i := 0; i < len(articles); i += batchSize {
		end := i + batchSize
		if end > len(articles) {
			end = len(articles)
		}

		batch := articles[i:end]
		docs := make([]map[string]interface{}, 0, len(batch))

		for _, article := range batch {
			doc := map[string]interface{}{
				"id":             article.ID,
				"articleTitle":   article.ArticleTitle,
				"articleContent": article.ArticleContent,
				"isDelete":       article.IsDelete,
				"status":         article.Status,
			}
			docs = append(docs, doc)
		}

		if err := init.esService.BulkIndexArticles(ctx, docs); err != nil {
			return fmt.Errorf("批量索引第 %d-%d 篇失败: %w", i+1, end, err)
		}

		totalSynced += len(batch)
		log.Printf("  ✓ 已同步 %d/%d 篇", totalSynced, len(articles))

		// 避免过快请求
		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("✅ 成功同步 %d 篇文章到 ES", totalSynced)
	return nil
}
