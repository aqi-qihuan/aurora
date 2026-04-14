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
		UserID:        log.UserID,
		OptModule:     log.Module,
		OptType:       log.Operation,
		OptMethod:     log.Method,
		OptUri:        log.URL,
		IpAddress:     log.IP,
		RequestMethod: log.RequestMethod,
		RequestParam:  log.RequestParam,
		ResponseData:  log.ResponseData,
		OptDesc:       log.OptDesc,
		Nickname:      log.Nickname,
		IpSource:      log.IpSource,
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

// ListOperationLogs 分页查询操作日志（对标Java OperationLogServiceImpl.listOperationLogs）
// Java: .like(StringUtils.isNotBlank(conditionVO.getKeywords()), OperationLog::getOptModule, conditionVO.getKeywords())
//       .or().like(StringUtils.isNotBlank(conditionVO.getKeywords()), OperationLog::getOptDesc, conditionVO.getKeywords())
func (s *OperationLogService) ListOperationLogs(ctx context.Context, cond dto.ConditionVO, page dto.PageVO) (*dto.PageResultDTO, error) {
	var logs []model.OperationLog
	var count int64

	baseQuery := s.db.WithContext(ctx).Model(&model.OperationLog{})

	// 关键词搜索：搜索 opt_module 或 opt_desc（对标Java）
	if cond.Keywords != "" {
		baseQuery = baseQuery.Where("opt_module LIKE ? OR opt_desc LIKE ?", "%"+cond.Keywords+"%", "%"+cond.Keywords+"%")
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
		return nil, fmt.Errorf("查询操作日志失败: %w", err)
	}

	// 转换为DTO（字段完全对标Java OperationLogDTO）
	list := make([]dto.OperationLogDTO, len(logs))
	for i, l := range logs {
		list[i] = dto.OperationLogDTO{
			ID:             l.ID,
			OptModule:      l.OptModule,
			OptUri:         l.OptUri,
			OptType:        l.OptType,
			OptMethod:      l.OptMethod,
			OptDesc:        l.OptDesc,
			RequestMethod:  l.RequestMethod,
			RequestParam:   l.RequestParam,
			ResponseData:   l.ResponseData,
			Nickname:       l.Nickname,
			IpAddress:      l.IpAddress,
			IpSource:       l.IpSource,
			CreateTime:     l.CreateTime,
		}
	}

	return &dto.PageResultDTO{
		List:     list,
		Count:    count,
		PageNum:  page.PageNum,
		PageSize: page.PageSize,
	}, nil
}

// DeleteOperationLogs 批量删除操作日志（对标Java: removeByIds）
func (s *OperationLogService) DeleteOperationLogs(ctx context.Context, ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	result := s.db.WithContext(ctx).Where("id IN ?", ids).Delete(&model.OperationLog{})
	if result.Error != nil {
		return fmt.Errorf("批量删除操作日志失败: %w", result.Error)
	}
	slog.Info("批量删除操作日志", "count", result.RowsAffected, "ids", ids)
	return nil
}

// DeleteOperationLog 删除单个操作日志
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
