package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/model"
	"gorm.io/gorm"
)

// ExceptionLogService 异常日志业务逻辑 (对标 Java ExceptionLogServiceImpl)
type ExceptionLogService struct {
	db *gorm.DB
}

func NewExceptionLogService(db *gorm.DB) *ExceptionLogService {
	return &ExceptionLogService{db: db}
}

// SaveExceptionLog 保存异常日志 (由Recovery中间件调用, 对标 ExceptionLogAspect)
func (s *ExceptionLogService) SaveExceptionLog(ctx context.Context, log dto.ExceptionLogVO) error {
	elog := model.ExceptionLog{
		UserID:    log.UserID,
		URL:       log.URL,
		Method:    log.Method,
		IP:        log.IP,
		ErrorMsg:  log.ErrorMsg,
		Stacktrace: log.Stacktrace,
		Status:    1, // 默认未处理
	}

	if err := s.db.WithContext(ctx).Create(&elog).Error; err != nil {
		return fmt.Errorf("保存异常日志失败: %w", err)
	}
	errMsg := elog.ErrorMsg
	if len(errMsg) > 200 {
		errMsg = errMsg[:200]
	}
	slog.Error("异常日志记录", "id", elog.ID, "url", elog.URL, "error", errMsg)
	return nil
}

// ListExceptionLogs 分页查询异常日志
func (s *ExceptionLogService) ListExceptionLogs(ctx context.Context, cond dto.ConditionVO, page dto.PageVO) (*dto.PageResultDTO, error) {
	var logs []model.ExceptionLog
	var count int64

	baseQuery := s.db.WithContext(ctx).Model(&model.ExceptionLog{})

	if cond.Keywords != "" {
		baseQuery = baseQuery.Where("error_msg LIKE ?", "%"+cond.Keywords+"%")
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
		return nil, fmt.Errorf("查询异常日志失败: %w", err)
	}

	list := make([]dto.ExceptionLogDTO, len(logs))
	for i, l := range logs {
		list[i] = dto.ExceptionLogDTO{
			ID:         l.ID,
			UserID:     l.UserID,
			URL:        l.URL,
			Method:     l.Method,
			IP:         l.IP,
			ErrorMsg:   l.ErrorMsg,
			Stacktrace: l.Stacktrace,
			Status:     l.Status,
			CreateTime:  l.CreateTime,
		}
	}

	return &dto.PageResultDTO{
		List:     list,
		Count:    count,
		PageNum:  page.PageNum,
		PageSize: page.PageSize,
	}, nil
}

// DeleteExceptionLog 删除异常日志
func (s *ExceptionLogService) DeleteExceptionLog(ctx context.Context, id uint) error {
	result := s.db.WithContext(ctx).Delete(&model.ExceptionLog{}, id)
	if result.Error != nil {
		return fmt.Errorf("删除异常日志失败: %w", result.Error)
	}
	return nil
}
