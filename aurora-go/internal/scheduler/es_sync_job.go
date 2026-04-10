package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/infrastructure/search"
	"github.com/aurora-go/aurora/internal/model"
	"gorm.io/gorm"
)

// ESSyncJob Elasticsearch全量数据同步任务
// 对标 Java AuroraQuartz.importDataIntoES() (手动触发, Go版改为自动定时)
//
// 业务逻辑:
// 1. 删除现有ES索引 (删除后重建, 保证数据一致性)
// 2. 查询所有已发布文章
// 3. 逐条索引到ES (Article → ArticleSearchDTO → ES.index)
// 4. 错误容错: 单条失败不影响其他文章
//
// Go增强点:
//   - 使用 BulkIndex 批量索引提升性能 (替代逐条index)
//   - 支持分批处理 (每批100条, 避免内存溢出)
//   - ES未启用时自动跳过
type ESSyncJob struct {
	db *gorm.DB
}

// NewESSyncJob 创建ES全量同步任务实例
func NewESSyncJob(db *gorm.DB) *ESSyncJob {
	return &ESSyncJob{db: db}
}

const (
	esSyncBatchSize = 100 // 每批处理的文档数量
	esIndexName     = "article"
)

// Run 执行ES全量同步
func (j *ESSyncJob) Run(ctx context.Context) error {
	// Step 0: 检查ES是否可用 (对标Java elasticsearchClient==null)
	if search.Client == nil {
		slog.Warn("Elasticsearch客户端未启用, 跳过ES全量同步")
		return nil
	}

	startTime := time.Now()
	slog.Info("开始ES全量数据同步...")

	// Step 1: 删除现有索引 (对标Java elasticsearchClient.indices().delete(d -> d.index("article")))
	fullIndexName := search.Client.GetFullIndexName(esIndexName)
	if err := search.Client.DeleteIndex(ctx, fullIndexName); err != nil {
		// 索引不存在不视为错误
		slog.Debug("删除旧索引结果(不存在时忽略)", "index", fullIndexName, "error", err)
	}

	// Step 2: 查询所有已发布文章 (对标Java articleService.list())
	var total int64
	offset := 0
	batchNum := 0

	for {
		var articles []model.Article
		
		query := j.db.WithContext(ctx).
			Where("is_delete = 0 AND status = 1"). // 只同步未删除且公开的文章
			Order("id ASC").
			Limit(esSyncBatchSize).
			Offset(offset)
		
		result := query.Find(&articles)
		if result.Error != nil {
			return fmt.Errorf("failed to query articles: %w", result.Error)
		}

		if len(articles) == 0 {
			break // 没有更多数据了
		}

		batchNum++
		
		// Step 3: 批量索引到ES (对标Java for(article : articles) { es.index(...) })
		if err := search.Client.BulkIndex(ctx, fullIndexName, articles); err != nil {
			// 批次失败记录警告但继续处理下一批
			slog.Error("批量索引失败", "batch", batchNum, "count", len(articles), "error", err)
		} else {
			slog.Debug("批次索引成功", "batch", batchNum, "count", len(articles))
		}

		total += int64(len(articles))
		offset += len(articles)

		// 避免过于频繁的请求
		if len(articles) == esSyncBatchSize {
			time.Sleep(100 * time.Millisecond)
		}
	}

	duration := time.Since(startTime)
	slog.Info("ES全量数据同步完成",
		"total_articles", total,
		"batches", batchNum,
		"duration", duration.String(),
		"index", fullIndexName,
	)

	return nil
}

// ArticleToSearchDTO 将Article模型转换为ES搜索DTO (对标Java BeanCopyUtil.copyObject(article, ArticleSearchDTO.class))
func ArticleToSearchDTO(a model.Article) dto.ArticleSearchDTO {
	return dto.ArticleSearchDTO{
		ID:             a.ID,
		ArticleTitle:   a.ArticleTitle,
		ArticleContent: a.ArticleContent,
	}
}
