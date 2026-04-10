package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/model"
	"gorm.io/gorm"
)

// OperationLogService 操作日志业务逻辑 (对标 Java OperationLogServiceImpl)
type OperationLogService struct {
	db *gorm.DB
}

func NewOperationLogService(db *gorm.DB) *OperationLogService {
	return &OperationLogService{db: db}
}

// SaveOperationLog 保存操作日志 (由中间件调用, 对标 @OptLog AOP)
func (s *OperationLogService) SaveOperationLog(ctx context.Context, log dto.OperationLogVO) error {
	oplog := model.OperationLog{
		UserID:    log.UserID,
		Module:    log.Module,
		Operation: log.Operation,
		Method:    log.Method,
		URL:       log.URL,
		IP:        log.IP,
		Duration:  log.Duration,
		Status:    log.Status,
		ErrorMsg:  log.ErrorMsg,
	}

	if err := s.db.WithContext(ctx).Create(&oplog).Error; err != nil {
		return fmt.Errorf("保存操作日志失败: %w", err)
	}
	return nil
}

// Save 便捷方法: 直接接收DTO并保存 (AccessLog中间件调用, 异步安全)
func (s *OperationLogService) Save(log *dto.OperationLogVO) error {
	ctx := context.Background()
	return s.SaveOperationLog(ctx, *log)
}

// ListOperationLogs 分页查询操作日志
func (s *OperationLogService) ListOperationLogs(ctx context.Context, cond dto.ConditionVO, page dto.PageVO) (*dto.PageResultDTO, error) {
	var logs []model.OperationLog
	var count int64

	baseQuery := s.db.WithContext(ctx).Model(&model.OperationLog{})

	if cond.Keywords != "" {
		baseQuery = baseQuery.Where("operation LIKE ? OR module LIKE ?", "%"+cond.Keywords+"%", "%"+cond.Keywords+"%")
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
		Preload("UserInfo", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, nickname")
		}).
		Order("create_time DESC").
		Limit(page.PageSize).
		Offset(offset).
		Find(&logs).Error

	if err != nil {
		return nil, fmt.Errorf("查询操作日志失败: %w", err)
	}

	list := make([]dto.OperationLogDTO, len(logs))
	for i, l := range logs {
		list[i] = dto.OperationLogDTO{
			ID:        l.ID,
			UserID:    l.UserID,
			Module:    l.Module,
			Operation: l.Operation,
			Method:    l.Method,
			URL:       l.URL,
			IP:        l.IP,
			Duration:  l.Duration,
			Status:    l.Status,
			ErrorMsg:  l.ErrorMsg,
			CreateTIme: l.CreateTime,
		}
		if l.UserInfo != nil {
			list[i].Nickname = l.UserInfo.Nickname
		}
	}

	return &dto.PageResultDTO{
		List:     list,
		Count:    count,
		PageNum:  page.PageNum,
		PageSize: page.PageSize,
	}, nil
}

// DeleteOperationLog 删除操作日志
func (s *OperationLogService) DeleteOperationLog(ctx context.Context, id uint) error {
	result := s.db.WithContext(ctx).Delete(&model.OperationLog{}, id)
	if result.Error != nil {
		return fmt.Errorf("删除操作日志失败: %w", result.Error)
	}
	return nil
}

// ClearOldLogs 清理过期操作日志(保留最近N天)
func (s *OperationLogService) ClearOldLogs(ctx context.Context, days int) error {
	cutoff := time.Now().AddDate(0, 0, -days)
	result := s.db.WithContext(ctx).
		Where("create_time < ?", cutoff).
		Delete(&model.OperationLog{})

	if result.Error != nil {
		return fmt.Errorf("清理操作日志失败: %w", result.Error)
	}
	slog.Info("操作日志清理完成", "rows", result.RowsAffected, "days", days)
	return nil
}
