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
//   - 内置6个预定义任务(对标Java版t_job表的5条记录+ES同步)
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

	"github.com/aurora-go/aurora/internal/infrastructure/database"
	"github.com/aurora-go/aurora/internal/model"
	"gorm.io/gorm"
)

// Scheduler 调度器核心管理器 (对标Java Quartz Scheduler)
type Scheduler struct {
	cron      *cron.Cron
	db        *gorm.DB
	jobMap    map[string]JobHandler       // invokeTarget → handler
	entryIDs  map[string]cron.EntryID     // jobName → cron.EntryID
	mu        sync.RWMutex
	running   bool
	siteURL   string // 网站访问URL (用于百度SEO等需要生成URL的任务)
}

// JobHandler 定时任务处理函数类型
type JobHandler func(ctx context.Context) error

// JobResult 任务执行结果
type JobResult struct {
	JobID   uint
	JobName string
	Status  int8    // 0=成功 1=失败
	Duration int64  // 执行耗时(ms)
	ErrorMsg string
}

// NewScheduler 创建调度器实例
func NewScheduler(db *gorm.DB, siteURL string) *Scheduler {
	return &Scheduler{
		cron:     cron.New(cron.WithSeconds(), cron.WithChain(cron.Recover(cron.DefaultLogger))),
		db:       db,
		jobMap:   make(map[string]JobHandler),
		entryIDs: make(map[string]cron.EntryID),
		siteURL:  siteURL,
	}
}

// Register 注册一个定时任务 (程序化注册, 对标Java t_job表动态注册)
func (s *Scheduler) Register(jobName, invokeTarget, cronExpr string, handler JobHandler) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 验证Cron表达式
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	if _, err := parser.Parse(cronExpr); err != nil {
		return err
	}

	s.jobMap[invokeTarget] = handler

	// 添加到Cron调度器
	entryID, err := s.cron.AddFunc(cronExpr, func() {
		s.executeJob(context.Background(), jobName, invokeTarget, handler)
	})
	if err != nil {
		return err
	}

	s.entryIDs[jobName] = entryID
	slog.Info("定时任务已注册", "job", jobName, "cron", cronExpr, "target", invokeTarget)
	return nil
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

// GetEntryIDs 获取所有已注册任务的 EntryID
func (s *Scheduler) GetEntryIDs() map[string]cron.EntryID {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]cron.EntryID, len(s.entryIDs))
	for k, v := range s.entryIDs {
		result[k] = v
	}
	return result
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
	go s.recordJobLog(jobName, status, duration, errorMsg)
}

// recordJobLog 记录调度日志到数据库 (对标Java JobLogMapper.insert)
func (s *Scheduler) recordJobLog(jobName string, status int8, duration int64, errorMsg string) {
	log := model.JobLog{
		JobName:  jobName,
		Status:   status,
		Duration: &duration,
		ErrorMsg: errorMsg,
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

// InitDefaultTasks 初始化所有预定义定时任务 (对标Java @PostConstruct init() 从t_job加载)
// Java版数据来自 aurora.sql 中 INSERT INTO `t_job` 的5条记录:
//   ID=81 统计用户地域分布  '0 0,30 * * * ?'
//   ID=82 统计访问量        '0 0 0 * * ?'
//   ID=83 清空redis访客记录  '0 0 1 * * ?'
//   ID=84 百度SEO          '0 0/10 * * * ?'
//   ID=85 清理定时任务日志   '0 0 0 * * ?'
// + ES全量同步 (Go增强, Java版通过手动触发 importDataIntoES)
func (s *Scheduler) InitDefaultTasks() error {
	rdb := database.GetRedis()

	// 注册所有内置任务
	tasks := []struct {
		name         string
		invokeTarget string
		cronExpr     string
		handler      JobHandler
	}{
		{
			name:         "统计访问量",
			invokeTarget: "auroraQuartz.saveUniqueView",
			cronExpr:     "0 0 0 * * *",           // 每天00:00:00 (对标Java ID=82)
			handler:      NewUniqueViewJob(s.db, rdb).Run,
		},
		{
			name:         "清空redis访客记录",
			invokeTarget: "auroraQuartz.clear",
			cronExpr:     "0 0 1 * * *",           // 每天01:00:00 (对标Java ID=83)
			handler:      NewClearCacheJob(rdb).Run,
		},
		{
			name:         "统计用户地域分布",
			invokeTarget: "auroraQuartz.statisticalUserArea",
			cronExpr:     "0 0,30 * * * *",         // 每30分钟一次 (对标Java ID=81)
			handler:      NewUserAreaJob(s.db, rdb).Run,
		},
		{
			name:         "百度SEO推送",
			invokeTarget: "auroraQuartz.baiduSeo",
			cronExpr:     "0 0/10 * * * *",          // 每10分钟一次 (对标Java ID=84)
			handler:      NewBaiduSeoJob(s.db, s.siteURL).Run,
		},
		{
			name:         "清理定时任务日志",
			invokeTarget: "auroraQuartz.clearJobLogs",
			cronExpr:     "0 0 3 * * *",             // 每天03:00:00 (对标Java ID=85)
			handler:      NewCleanLogJob(s.db).Run,
		},
		{
			name:         "ES全量数据同步",
			invokeTarget: "auroraQuartz.importDataIntoES",
			cronExpr:     "0 0 4 * * *",             // 每天04:00:00 (Go增强: 自动全量同步)
			handler:      NewESSyncJob(s.db).Run,
		},
	}

	for _, task := range tasks {
		if err := s.Register(task.name, task.invokeTarget, task.cronExpr, task.handler); err != nil {
			slog.Error("注册定时任务失败", "job", task.name, "error", err)
			continue
		}
	}

	slog.Info("默认定时任务初始化完成", "total", len(tasks))
	return nil
}
