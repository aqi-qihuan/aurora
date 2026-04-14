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
		OptUri:        log.URL,
		OptMethod:     log.Method,
		RequestMethod: log.RequestMethod,
		RequestParam:  log.RequestParam,
		OptDesc:       log.OptDesc,
		ExceptionInfo: log.ExceptionInfo,
		IpAddress:     log.IP,
		IpSource:      log.IpSource,
	}

	if err := s.db.WithContext(ctx).Create(&elog).Error; err != nil {
		return fmt.Errorf("保存异常日志失败: %w", err)
	}
	errMsg := elog.ExceptionInfo
	if len(errMsg) > 200 {
		errMsg = errMsg[:200]
	}
	slog.Error("异常日志记录", "id", elog.ID, "url", elog.OptUri, "error", errMsg)
	return nil
}

// ListExceptionLogs 分页查询异常日志（对标Java ExceptionLogServiceImpl.listExceptionLogs）
// Java: .like(StringUtils.isNotBlank(conditionVO.getKeywords()), ExceptionLog::getOptDesc, conditionVO.getKeywords())
func (s *ExceptionLogService) ListExceptionLogs(ctx context.Context, cond dto.ConditionVO, page dto.PageVO) (*dto.PageResultDTO, error) {
	var logs []model.ExceptionLog
	var count int64

	baseQuery := s.db.WithContext(ctx).Model(&model.ExceptionLog{})

	// 关键词搜索：搜索 opt_desc（对标Java）
	if cond.Keywords != "" {
		baseQuery = baseQuery.Where("opt_desc LIKE ?", "%"+cond.Keywords+"%")
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

	// 转换为DTO（字段完全对标Java ExceptionLogDTO）
	list := make([]dto.ExceptionLogDTO, len(logs))
	for i, l := range logs {
		list[i] = dto.ExceptionLogDTO{
			ID:            l.ID,
			OptUri:        l.OptUri,
			OptMethod:     l.OptMethod,
			RequestMethod: l.RequestMethod,
			RequestParam:  l.RequestParam,
			OptDesc:       l.OptDesc,
			ExceptionInfo: l.ExceptionInfo,
			IpAddress:     l.IpAddress,
			IpSource:      l.IpSource,
			CreateTime:    l.CreateTime,
		}
	}

	return &dto.PageResultDTO{
		List:     list,
		Count:    count,
		PageNum:  page.PageNum,
		PageSize: page.PageSize,
	}, nil
}

// DeleteExceptionLogs 批量删除异常日志（对标Java: removeByIds）
func (s *ExceptionLogService) DeleteExceptionLogs(ctx context.Context, ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	result := s.db.WithContext(ctx).Where("id IN ?", ids).Delete(&model.ExceptionLog{})
	if result.Error != nil {
		return fmt.Errorf("批量删除异常日志失败: %w", result.Error)
	}
	slog.Info("批量删除异常日志", "count", result.RowsAffected, "ids", ids)
	return nil
}
