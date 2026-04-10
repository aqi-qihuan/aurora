package scheduler

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"

	"github.com/aurora-go/aurora/internal/constant"
)

// ClearCacheJob Redis缓存清理任务
// 对标 Java AuroraQuartz.clear() (Cron: 0 0 1 * * ?, ID=83)
//
// 业务逻辑:
// 1. DEL unique_visitor (清空独立访客Set)
// 2. DEL visitor_area   (清空访客地域缓存)
// 在UniqueView统计完成后1小时执行(避免数据丢失)
type ClearCacheJob struct {
	rdb *redis.Client
}

// NewClearCacheJob 创建缓存清理任务实例
func NewClearCacheJob(rdb *redis.Client) *ClearCacheJob {
	return &ClearCacheJob{rdb: rdb}
}

// Run 执行缓存清理任务
func (j *ClearCacheJob) Run(ctx context.Context) error {
	// Step 1: 删除独立访客 Set (对标Java redisService.del(UNIQUE_VISITOR))
	result1, err := j.rdb.Del(ctx,
		constant.UniqueVisitorPrefix,
		constant.VisitorAreaPrefix,
	).Result()
	if err != nil {
		return fmt.Errorf("failed to clear cache keys: %w", err)
	}

	slog.Info("Redis缓存清理成功",
		"deleted_keys", result1,
		"keys", []string{constant.UniqueVisitorPrefix, constant.VisitorAreaPrefix},
	)
	return nil
}
