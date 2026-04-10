package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/model"
	"gorm.io/gorm"
)

// JobLogService 调度日志业务逻辑 (对标 Java JobLogServiceImpl)
type JobLogService struct {
	db *gorm.DB
}

func NewJobLogService(db *gorm.DB) *JobLogService {
	return &JobLogService{db: db}
}

// ListJobLogs 分页查询调度日志
func (s *JobLogService) ListJobLogs(ctx context.Context, cond dto.ConditionVO, page dto.PageVO) (*dto.PageResultDTO, error) {
	var logs []model.JobLog
	var count int64

	baseQuery := s.db.WithContext(ctx).Model(&model.JobLog{})

	if cond.Keywords != "" {
		baseQuery = baseQuery.Where("job_name LIKE ?", "%"+cond.Keywords+"%")
	}
	if cond.Status != nil && *cond.Status >= 0 {
		baseQuery = baseQuery.Where("status = ?", *cond.Status)
	}
	if cond.DateStart != "" {
		baseQuery = baseQuery.Where("create_time >= ?", cond.DateStart)
	}
	if cond.DateEnd != "" {
		baseQuery = baseQuery.Where("create_time <= ?", cond.DateEnd+" 23:59:59")
	}

	baseQuery.Count(&count)

	offset := page.GetOffset()
	err := baseQuery.
		Order("create_time DESC").
		Limit(page.PageSize).
		Offset(offset).
		Find(&logs).Error

	if err != nil {
		return nil, fmt.Errorf("查询日志列表失败: %w", err)
	}

	list := make([]dto.JobLogDTO, len(logs))
	for i, l := range logs {
		list[i] = dto.JobLogDTO{
			ID:        l.ID,
			JobID:     l.JobID,
			JobName:   l.JobName,
			JobGroup:  l.JobGroup,
			Status:    l.Status,
			Duration:  l.Duration,
			ErrorMsg:  l.ErrorMsg,
			CreateTime: l.CreateTime,
		}
	}

	return &dto.PageResultDTO{
		List:     list,
		Count:    count,
		PageNum:  page.PageNum,
		PageSize: page.PageSize,
	}, nil
}

// ClearJobLogs 清理过期日志(保留最近N天)
func (s *JobLogService) ClearJobLogs(ctx context.Context, days int) error {
	result := s.db.WithContext(ctx).
		Exec("DELETE FROM t_job_log WHERE create_time < DATE_SUB(NOW(), INTERVAL ? DAY)", days)

	if result.Error != nil {
		return fmt.Errorf("清理日志失败: %w", result.Error)
	}

	slog.Info("调度日志清理完成", "rows", result.RowsAffected, "days", days)
	return nil
}
