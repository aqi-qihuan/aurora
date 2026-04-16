package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/model"
	"github.com/aurora-go/aurora/internal/scheduler"
	"github.com/aurora-go/aurora/internal/vo"
	"gorm.io/gorm"
)

// JobService 定时任务业务逻辑 (对标 Java QuartzServiceImpl)
type JobService struct {
	db        *gorm.DB
	scheduler *scheduler.Scheduler // 调度器引用（对标Java Scheduler）
}

func NewJobService(db *gorm.DB, scheduler *scheduler.Scheduler) *JobService {
	return &JobService{db: db, scheduler: scheduler}
}

// CreateJob 创建定时任务
func (s *JobService) CreateJob(ctx context.Context, jobVO vo.JobVO) (*model.Job, error) {
	job := model.Job{
		JobName:        jobVO.JobName,
		JobGroup:       jobVO.JobGroup,
		InvokeTarget:   jobVO.InvokeTarget, // 全限定类名或函数名
		CronExpression: jobVO.CronExpression,
		Status:         0, // 默认暂停
	}

	if err := s.db.WithContext(ctx).Create(&job).Error; err != nil {
		return nil, fmt.Errorf("创建任务失败: %w", err)
	}

	slog.Info("创建定时任务", "id", job.ID, "name", job.JobName, "cron", job.CronExpression)

	// 注册到调度器（对标Java ScheduleUtil.createScheduleJob）
	if err := s.scheduler.AddJob(job); err != nil {
		slog.Error("注册任务到调度器失败", "id", job.ID, "error", err)
		// 不返回错误，任务已保存到数据库，调度器注册失败不影响数据
	}

	return &job, nil
}

// UpdateJob 更新定时任务
func (s *JobService) UpdateJob(ctx context.Context, id uint, jobVO vo.JobVO) error {
	var job model.Job
	if err := s.db.WithContext(ctx).First(&job, id).Error; err != nil {
		return errors.ErrJobNotFound
	}

	updates := map[string]interface{}{
		"job_name":        jobVO.JobName,
		"job_group":       jobVO.JobGroup,
		"invoke_target":   jobVO.InvokeTarget,
		"cron_expression": jobVO.CronExpression,
	}

	if err := s.db.WithContext(ctx).Model(&job).Updates(updates).Error; err != nil {
		return err
	}

	// 更新调度器（对标Java updateSchedulerJob）
	job.JobName = jobVO.JobName
	job.JobGroup = jobVO.JobGroup
	job.InvokeTarget = jobVO.InvokeTarget
	job.CronExpression = jobVO.CronExpression
	if err := s.scheduler.UpdateJob(job); err != nil {
		slog.Error("更新调度器任务失败", "id", id, "error", err)
	}

	return nil
}

// DeleteJob 删除定时任务
func (s *JobService) DeleteJob(ctx context.Context, id uint) error {
	var job model.Job

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&job, id).Error; err != nil {
			return errors.ErrJobNotFound
		}

		// 清除调度日志
		tx.Delete(&model.JobLog{}, "job_id = ?", id)

		if err := tx.Delete(&job).Error; err != nil {
			return fmt.Errorf("删除任务失败: %w", err)
		}

		// 从调度器移除（对标Java scheduler.deleteJob）
		if err := s.scheduler.RemoveJob(job.JobName); err != nil {
			slog.Error("从调度器移除任务失败", "id", id, "error", err)
		}

		return nil
	})
}

// ListJobs 分页查询定时任务列表
func (s *JobService) ListJobs(ctx context.Context, cond dto.ConditionVO, page dto.PageVO) (*dto.PageResultDTO, error) {
	var jobs []model.Job
	var count int64

	baseQuery := s.db.WithContext(ctx).Model(&model.Job{})

	if cond.Keywords != "" {
		baseQuery = baseQuery.Where("job_name LIKE ?", "%"+cond.Keywords+"%")
	}
	if cond.Status != nil && *cond.Status >= 0 {
		statusInt := int(*cond.Status)
		baseQuery = baseQuery.Where("status = ?", statusInt)
	}

	baseQuery.Count(&count)

	offset := page.GetOffset()
	err := baseQuery.
		Order("create_time DESC").
		Limit(page.PageSize).
		Offset(offset).
		Find(&jobs).Error

	if err != nil {
		return nil, fmt.Errorf("查询任务列表失败: %w", err)
	}

	list := make([]dto.JobDTO, len(jobs))
	for i, j := range jobs {
		list[i] = dto.JobDTO{
			ID:             j.ID,
			JobName:        j.JobName,
			JobGroup:       j.JobGroup,
			InvokeTarget:   j.InvokeTarget,
			CronExpression: j.CronExpression,
			MisfirePolicy:  j.MisfirePolicy,
			Concurrent:     j.Concurrent,
			Status:         j.Status,
			Remark:         j.Remark,
			CreateTime:     j.CreateTime,
			UpdateTime:     j.UpdateTime,
		}
	}

	return &dto.PageResultDTO{
		List:     list,
		Count:    count,
		PageNum:  page.PageNum,
		PageSize: page.PageSize,
	}, nil
}

// ChangeJobStatus 切换任务状态(暂停/恢复)（对标Java JobServiceImpl.updateJobStatus）
func (s *JobService) ChangeJobStatus(ctx context.Context, id uint, status int8) error {
	var job model.Job
	if err := s.db.WithContext(ctx).First(&job, id).Error; err != nil {
		return errors.ErrJobNotFound
	}

	// 如果状态没有变化，直接返回
	if job.Status == int(status) {
		return nil
	}

	// 更新数据库状态
	result := s.db.WithContext(ctx).
		Model(&model.Job{}).
		Where("id = ?", id).
		Update("status", status)

	if result.Error != nil {
		return fmt.Errorf("切换任务状态失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.ErrJobNotFound
	}

	// 更新调度器状态（对标Java scheduler.pauseJob/resumeJob）
	var err error
	if status == 0 { // 暂停
		err = s.scheduler.PauseJob(job.JobName)
	} else if status == 1 { // 恢复
		err = s.scheduler.ResumeJob(job)
	}

	if err != nil {
		slog.Error("更新调度器任务状态失败", "id", id, "error", err)
		// 不返回错误，数据库已更新
	}

	action := "暂停"
	if status == 1 {
		action = "恢复"
	}
	slog.Info("任务状态变更", "id", id, "name", job.JobName, "action", action)
	return nil
}

// RunJobNow 手动触发执行一次任务 (对标Java JobServiceImpl.runJob → scheduler.triggerJob())
func (s *JobService) RunJobNow(ctx context.Context, id uint) (*model.JobLog, error) {
	var job model.Job
	if err := s.db.WithContext(ctx).First(&job, id).Error; err != nil {
		return nil, errors.ErrJobNotFound
	}

	// 调用调度器执行（对标Java scheduler.triggerJob）
	if err := s.scheduler.RunJobNow(ctx, job.JobName, job.InvokeTarget); err != nil {
		return nil, err
	}

	slog.Info("手动触发任务执行", "id", id, "name", job.JobName)
	return &model.JobLog{
		JobID:        job.ID,
		JobName:      job.JobName,
		JobGroup:     job.JobGroup,
		InvokeTarget: job.InvokeTarget,
		Status:       0,
		JobMessage:   "手动触发执行成功",
	}, nil
}

// GetJobByID 根据ID获取任务详情
func (s *JobService) GetJobByID(ctx context.Context, id uint) (*dto.JobDetailDTO, error) {
	var job model.Job

	err := s.db.WithContext(ctx).First(&job, id).Error
	if err != nil {
		if errors.IsStd(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrJobNotFound
		}
		return nil, fmt.Errorf("查询任务详情失败: %w", err)
	}

	dto := &dto.JobDetailDTO{
		ID:             job.ID,
		JobName:        job.JobName,
		JobGroup:       job.JobGroup,
		InvokeTarget:   job.InvokeTarget,
		CronExpression: job.CronExpression,
		MisfirePolicy:  job.MisfirePolicy,
		Concurrent:     job.Concurrent,
		Status:         job.Status,
		Remark:         job.Remark,
		CreateTime:     job.CreateTime,
		UpdateTime:     job.UpdateTime,
	}
	return dto, nil
}
