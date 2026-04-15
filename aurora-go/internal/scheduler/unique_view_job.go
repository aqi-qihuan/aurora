package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/aurora-go/aurora/internal/constant"
	"github.com/aurora-go/aurora/internal/model"
	"gorm.io/gorm"
)

// UniqueViewJob UniqueView每日独立访客统计
// 对标 Java AuroraQuartz.saveUniqueView() (Cron: 0 0 0 * * ?, ID=82)
//
// 业务逻辑:
// 1. Redis SCARD unique_visitor:YYYY-MM-DD → 获取Set元素个数(独立访客数)
// 2. 构建 UniqueView 实体 (createTime=前一天, viewsCount=count)
// 3. MySQL INSERT INTO t_unique_view
type UniqueViewJob struct {
	db  *gorm.DB
	rdb *redis.Client
}

// NewUniqueViewJob 创建UniqueView统计任务实例
func NewUniqueViewJob(db *gorm.DB, rdb *redis.Client) *UniqueViewJob {
	return &UniqueViewJob{db: db, rdb: rdb}
}

// Run 执行统计任务 (实现 JobHandler 接口)
func (j *UniqueViewJob) Run(ctx context.Context) error {
	// Step 1: 构建昨天的日期 key（按天拆分，避免累积）
	yesterday := time.Now().AddDate(0, 0, -1)
	yesterdayStr := yesterday.Format("2006-01-02")
	key := fmt.Sprintf("%s:%s", constant.UniqueVisitor, yesterdayStr)
	
	// Step 2: 从Redis Set获取前一天的独立访客数量
	count, err := j.rdb.SCard(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("failed to get unique visitor count: %w", err)
	}

	// Step 3: 构建UniqueView实体 (createTime=前一天, viewsCount=count)
	uniqueView := model.UniqueView{
		CreateTime: yesterday.Truncate(24 * time.Hour),
		ViewsCount: int(count),
	}

	// Step 4: 写入MySQL (先删除旧记录，再插入新记录，避免重复)
	// 使用 Upsert 逻辑：如果当天已存在记录，先删除
	j.db.Where("DATE(create_time) = ?", yesterdayStr).Delete(&model.UniqueView{})
	
	if err := j.db.Create(&uniqueView).Error; err != nil {
		return fmt.Errorf("failed to insert unique view: %w", err)
	}

	// Step 5: 清空 Redis（对标Java redisService.del(UNIQUE_VISITOR)）
	j.rdb.Del(ctx, key)

	slog.Info("UniqueView统计保存成功",
		"date", yesterdayStr,
		"views_count", count,
	)
	return nil
}
