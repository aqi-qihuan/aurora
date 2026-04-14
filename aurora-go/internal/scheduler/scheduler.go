// Package scheduler 基于 robfig/cron v3 实现定时任务调度系统
// 对标 Java Quartz 框架 (AuroraQuartz.java + ScheduleUtil + JobInvokeUtil)
//
// 架构对比:
//
//	Java Quartz:  t_job表 → @PostConstruct init() → Scheduler.createScheduleJob() → CronTrigger → JobExecute()
//	Go cron:    Registry注册 → Cron.AddFunc(cronExpr) → goroutine触发 → JobHandler()
//
// 特性:
//   - 支持标准Cron表达式 (秒 分 时 日 月 周)
//   - 从数据库动态加载任务（对标Java @PostConstruct init）
//   - 自动记录执行日志到t_job_log表
//   - panic恢复机制(单个任务崩溃不影响其他任务)
//   - 优雅关闭(signal-based)
package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/redis/go-redis/v9"

	"github.com/aurora-go/aurora/internal/model"
	"gorm.io/gorm"
)

// Scheduler 调度器核心管理器 (对标Java Quartz Scheduler)
type Scheduler struct {
	cron      *cron.Cron
	db        *gorm.DB
	rdb       *redis.Client // Redis客户端（任务需要）
	jobMap    map[string]JobHandler   // invokeTarget → handler
	entryIDs  map[string]cron.EntryID // jobName → cron.EntryID
	mu        sync.RWMutex
	running   bool
	siteURL   string // 网站访问URL (用于百度SEO等需要生成URL的任务)
}

// JobHandler 定时任务处理函数类型
type JobHandler func(ctx context.Context) error

// NewScheduler 创建调度器实例
func NewScheduler(db *gorm.DB, rdb *redis.Client, siteURL string) *Scheduler {
	return &Scheduler{
		cron:     cron.New(cron.WithSeconds(), cron.WithChain(cron.Recover(cron.DefaultLogger))),
		db:       db,
		rdb:      rdb,
		jobMap:   make(map[string]JobHandler),
		entryIDs: make(map[string]cron.EntryID),
		siteURL:  siteURL,
	}
}

// Start 启动调度器 (对标Java Scheduler.start())
func (s *Scheduler) Start() {
	s.mu.Lock()
	s.running = true
	s.mu.Unlock()

	s.cron.Start()
	slog.Info("定时任务调度器已启动", "taskCount", len(s.entryIDs))

	// 监听系统信号实现优雅关闭
	go s.waitForShutdown()
}

// Stop 停止调度器 (优雅关闭, 对标Java Scheduler.shutdown())
func (s *Scheduler) Stop() context.Context {
	s.mu.Lock()
	s.running = false
	s.mu.Unlock()

	slog.Info("正在停止定时任务调度器...")
	ctx := s.cron.Stop()
	slog.Info("定时任务调度器已停止")
	return ctx
}

// IsRunning 调度器运行状态
func (s *Scheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// InitFromDatabase 从数据库加载所有任务并注册到调度器（对标Java @PostConstruct init()）
// Java版数据来自 aurora.sql 中 INSERT INTO `t_job` 的记录
func (s *Scheduler) InitFromDatabase() error {
	slog.Info("正在从数据库加载定时任务...")

	// 清空调度器（对标Java scheduler.clear()）
	s.cron.Stop()
	s.cron = cron.New(cron.WithSeconds(), cron.WithChain(cron.Recover(cron.DefaultLogger)))
	s.jobMap = make(map[string]JobHandler)
	s.entryIDs = make(map[string]cron.EntryID)

	// 查询所有任务
	var jobs []model.Job
	if err := s.db.Find(&jobs).Error; err != nil {
		return fmt.Errorf("查询定时任务失败: %w", err)
	}

	if len(jobs) == 0 {
		slog.Info("数据库中没有定时任务")
		return nil
	}

	// 注册每个任务
	for _, job := range jobs {
		if err := s.addJobToScheduler(job); err != nil {
			slog.Error("注册定时任务失败", "job", job.JobName, "id", job.ID, "error", err)
			continue
		}
		slog.Info("定时任务已注册", "id", job.ID, "name", job.JobName, "cron", job.CronExpression, "target", job.InvokeTarget)
	}

	slog.Info("从数据库加载定时任务完成", "total", len(jobs), "success", len(s.entryIDs))
	return nil
}

// addJobToScheduler 将单个任务添加到调度器（对标Java ScheduleUtil.createScheduleJob）
func (s *Scheduler) addJobToScheduler(job model.Job) error {
	// 保存 invokeTarget 到 jobMap
	s.jobMap[job.InvokeTarget] = func(ctx context.Context) error {
		// 调用任务函数（对标Java JobInvokeUtil.invokeMethod）
		return InvokeMethod(ctx, job.InvokeTarget)
	}

	// 添加到 Cron 调度器
	invokeTarget := job.InvokeTarget
	jobName := job.JobName
	entryID, err := s.cron.AddFunc(job.CronExpression, func() {
		handler := s.jobMap[invokeTarget]
		if handler != nil {
			s.executeJob(context.Background(), jobName, invokeTarget, handler)
		}
	})
	if err != nil {
		return fmt.Errorf("添加任务到Cron失败: %w", err)
	}

	s.entryIDs[jobName] = entryID
	return nil
}

// AddJob 动态添加定时任务（对标Java JobServiceImpl.saveJob → ScheduleUtil.createScheduleJob）
func (s *Scheduler) AddJob(job model.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.addJobToScheduler(job); err != nil {
		return err
	}

	slog.Info("动态添加定时任务成功", "id", job.ID, "name", job.JobName)
	return nil
}

// UpdateJob 更新定时任务（对标Java JobServiceImpl.updateJob）
func (s *Scheduler) UpdateJob(job model.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 先删除旧任务
	if entryID, ok := s.entryIDs[job.JobName]; ok {
		s.cron.Remove(entryID)
		delete(s.entryIDs, job.JobName)
		delete(s.jobMap, job.InvokeTarget)
	}

	// 再添加新任务
	if err := s.addJobToScheduler(job); err != nil {
		return err
	}

	slog.Info("更新定时任务成功", "id", job.ID, "name", job.JobName)
	return nil
}

// RemoveJob 删除定时任务（对标Java JobServiceImpl.deleteJobs）
func (s *Scheduler) RemoveJob(jobName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entryID, ok := s.entryIDs[jobName]; ok {
		s.cron.Remove(entryID)
		delete(s.entryIDs, jobName)
		slog.Info("删除定时任务成功", "name", jobName)
	}

	return nil
}

// RunJobNow 立即执行一次任务（对标Java JobServiceImpl.runJob → scheduler.triggerJob）
func (s *Scheduler) RunJobNow(ctx context.Context, jobName string, invokeTarget string) error {
	s.mu.RLock()
	handler, ok := s.jobMap[invokeTarget]
	s.mu.RUnlock()

	if !ok {
		return fmt.Errorf("任务未注册: %s", invokeTarget)
	}

	// 异步执行（对标Java triggerJob）
	go func() {
		s.executeJob(ctx, jobName, invokeTarget, handler)
	}()

	slog.Info("手动触发任务执行", "name", jobName, "target", invokeTarget)
	return nil
}

// executeJob 执行单个任务 (对标Java AbstractQuartzJob.execute() + JobLog记录)
func (s *Scheduler) executeJob(ctx context.Context, jobName, invokeTarget string, handler JobHandler) {
	startTime := time.Now()

	slog.Debug("开始执行定时任务", "job", jobName)

	// 执行任务并捕获异常(对标Java try-catch)
	var execErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				execErr = fmt.Errorf("panic recovered: %v", r)
				slog.Error("定时任务发生panic", "job", jobName, "error", r)
			}
		}()
		execErr = handler(ctx)
	}()

	duration := time.Since(startTime).Milliseconds()

	// 记录执行日志 (对标Java AbstractQuartzJob.after() → JobLogMapper.insert)
	status := int8(0) // 成功
	errorMsg := ""
	if execErr != nil {
		status = 1 // 失败
		errorMsg = execErr.Error()
		if len(errorMsg) > 2000 {
			errorMsg = errorMsg[:2000] // 截断超长错误信息(对齐Java ExceptionUtil.getTrace截断)
		}
		slog.Error("定时任务执行失败", "job", jobName, "error", errorMsg, "duration_ms", duration)
	} else {
		slog.Info("定时任务执行完成", "job", jobName, "duration_ms", duration)
	}

	// 异步写入日志(避免影响性能)
	go s.recordJobLog(jobName, invokeTarget, status, duration, errorMsg)
}

// recordJobLog 记录调度日志到数据库 (对标Java JobLogMapper.insert)
func (s *Scheduler) recordJobLog(jobName, invokeTarget string, status int8, duration int64, errorMsg string) {
	now := time.Now()
	startTime := now.Add(-time.Duration(duration) * time.Millisecond)

	log := model.JobLog{
		JobName:       jobName,
		InvokeTarget:  invokeTarget,
		Status:        status,
		StartTime:     &startTime,
		EndTime:       &now,
		ExceptionInfo: errorMsg,
		JobMessage:    "定时任务执行完成",
	}

	if err := s.db.Create(&log).Error; err != nil {
		slog.Error("记录调度日志失败", "error", err, "job", jobName)
	}
}

// waitForShutdown 监听系统信号实现优雅关闭
func (s *Scheduler) waitForShutdown() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh
	slog.Info("收到终止信号, 正在关闭调度器...", "signal", sig.String())
	s.Stop()
}
