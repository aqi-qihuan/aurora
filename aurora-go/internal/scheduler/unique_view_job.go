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
// 1. Redis SCARD unique_visitor → 获取Set元素个数(独立访客数)
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
	// Step 1: 从Redis Set获取独立访客数量 (对标Java redisService.sSize(UNIQUE_VISITOR))
	count, err := j.rdb.SCard(ctx, constant.UniqueVisitorPrefix).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("failed to get unique visitor count: %w", err)
	}

	// Step 2: 构建UniqueView实体 (时间设为前一天, 对标Java LocalDateTimeUtil.offset(now, -1, DAYS))
	yesterday := time.Now().AddDate(0, 0, -1).Truncate(24 * time.Hour)
	uniqueView := model.UniqueView{
		CreateTime: yesterday,
		ViewsCount: int(count),
	}

	// Step 3: 写入MySQL (对标Java uniqueViewMapper.insert(uniqueView))
	if err := j.db.Create(&uniqueView).Error; err != nil {
		return fmt.Errorf("failed to insert unique view: %w", err)
	}

	slog.Info("UniqueView统计保存成功",
		"date", yesterday.Format("2006-01-02"),
		"views_count", count,
	)
	return nil
}
