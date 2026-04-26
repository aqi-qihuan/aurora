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
// 1. Redis SCARD unique_visitor → 获取全局Set元素个数(独立访客数)
// 2. 构建 UniqueView 实体 (createTime=前一天, viewsCount=count)
// 3. MySQL INSERT INTO t_unique_view
// 4. Redis DEL unique_visitor → 清空全局Set（每天凌晨0点重置）
type UniqueViewJob struct {
	db  *gorm.DB
	rdb *redis.Client
}

// NewUniqueViewJob 创建UniqueView统计任务实例
func NewUniqueViewJob(db *gorm.DB, rdb *redis.Client) *UniqueViewJob {
	return &UniqueViewJob{db: db, rdb: rdb}
}

// Run 执行统计任务 (实现 JobHandler 接口)
func (j *UniqueViewJob) Run(ctx context.Context, params ...interface{}) error {
	// Step 1: 构建昨天的日期
	yesterday := time.Now().AddDate(0, 0, -1)
	yesterdayStr := yesterday.Format("2006-01-02")

	// Step 2: 从全局 Redis Set 获取独立访客数量（对标Java）
	// Java 版使用全局 key: unique_visitor，每天凌晨0点读取后清空
	key := constant.UniqueVisitor
	count, err := j.rdb.SCard(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("failed to get unique visitor count: %w", err)
	}

	// Step 3: 防御性处理 — 当没有访客数据时，不写入0值记录
	// 避免 t_unique_view 表中出现 views_count=0 的脏数据导致图表异常
	if count == 0 {
		slog.Info("UniqueView无访客数据，跳过归档",
			"date", yesterdayStr,
		)
		// 仍然需要清空Redis Set（确保干净状态）
		j.rdb.Del(ctx, key)
		return nil
	}

	// Step 4: 构建UniqueView实体 (createTime=前一天, viewsCount=count)
	uniqueView := model.UniqueView{
		CreateTime: yesterday.Truncate(24 * time.Hour),
		ViewsCount: int(count),
	}

	// Step 5: 写入MySQL (先删除旧记录，再插入新记录，避免重复)
	j.db.Where("DATE(create_time) = ?", yesterdayStr).Delete(&model.UniqueView{})

	if err := j.db.Create(&uniqueView).Error; err != nil {
		return fmt.Errorf("failed to insert unique view: %w", err)
	}

	// Step 6: 清空全局 Redis Set（对标Java redisService.del(UNIQUE_VISITOR)）
	// 每天凌晨0点重置，开始统计新一天的访客
	j.rdb.Del(ctx, key)

	slog.Info("UniqueView统计保存成功",
		"date", yesterdayStr,
		"views_count", count,
	)
	return nil
}
