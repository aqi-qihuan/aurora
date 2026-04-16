package scheduler

import (
	"context"
	"fmt"
	"log/slog"

	"gorm.io/gorm"
)

// CleanLogJob 定时任务日志清理任务
// 对标 Java AuroraQuartz.clearJobLogs() (Cron: 0 0 0 * * ?, ID=85)
//
// 业务逻辑:
// 1. DELETE FROM t_job_log WHERE create_time < DATE_SUB(NOW(), INTERVAL 30 DAY)
// 2. 默认保留最近30天的日志
// 3. 可配置保留天数
type CleanLogJob struct {
	db *gorm.DB
}

// NewCleanLogJob 创建日志清理任务实例
func NewCleanLogJob(db *gorm.DB) *CleanLogJob {
	return &CleanLogJob{db: db}
}

const defaultRetentionDays = 30 // 默认保留天数(对标Java版无参数调用cleanJobLogs)

// Run 执行日志清理
func (j *CleanLogJob) Run(ctx context.Context, params ...interface{}) error {
	// 执行删除操作 (对标Java jobLogMapper.delete(null) - 全表清理 或 带条件的部分清理)
	// Go改进: 只清理超过保留期的日志，而非全表删除
	result := j.db.WithContext(ctx).
		Exec("DELETE FROM t_job_log WHERE create_time < DATE_SUB(NOW(), INTERVAL ? DAY)", defaultRetentionDays)

	if result.Error != nil {
		return fmt.Errorf("failed to clean job logs: %w", result.Error)
	}

	rowsAffected := result.RowsAffected
	if rowsAffected > 0 {
		slog.Info("调度日志清理完成", "deleted_rows", rowsAffected, "retention_days", defaultRetentionDays)
	} else {
		slog.Debug("无需清理的过期日志", "retention_days", defaultRetentionDays)
	}

	return nil
}
