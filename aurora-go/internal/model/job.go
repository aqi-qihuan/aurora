package model

import (
	"time"

	"gorm.io/gorm"
)

// Job 定时任务实体 (对应 t_job 表)
type Job struct {
	ID             uint       `gorm:"primarykey" json:"id"`
	JobName        string     `gorm:"size:100;not null;uniqueIndex" json:"jobName"`
	JobGroup       string     `gorm:"size:50" json:"jobGroup"`              // 任务分组
	InvokeTarget   string     `gorm:"size:500" json:"invokeTarget"`          // 调用目标(全限定类名或函数名)
	CronExpression  string     `gorm:"size:100" json:"cronExpression"`       // Cron表达式
	MisfirePolicy  int8       `gorm:"default:1" json:"misfirePolicy"`        // 错失策略
	Status         int8       `gorm:"default:0;index" json:"status"`         // 0正常 1暂停
	Param          string     `gorm:"type:text" json:"param"`                // 参数(JSON)
	RetryCount     int        `gorm:"default:3" json:"retryCount"`           // 重试次数
	Interval       string     `gorm:"size:50" json:"interval"`               // 执行间隔(备用)
	LastRunTime    *time.Time `json:"lastRunTime,omitempty"`                // 上次执行时间
	CreateTime     time.Time  `json:"createTime"`
	UpdateTime     time.Time  `json:"updateTime"`

	// 关联
	Logs []JobLog `gorm:"foreignKey:JobID" json:"logs,omitempty"`
}

func (Job) TableName() string { return "t_job" }

// JobLog 调度日志实体 (对应 t_job_log 表)
// 数据库实际字段: id, job_id, job_name, job_group, invoke_target, job_message, status, exception_info, create_time, start_time, end_time
type JobLog struct {
	ID            uint       `gorm:"primarykey" json:"id"`
	JobID         uint       `gorm:"index;column:job_id" json:"jobId"`
	JobName       string     `gorm:"size:64;column:job_name" json:"jobName"`
	JobGroup      string     `gorm:"size:64;column:job_group" json:"jobGroup"`
	InvokeTarget  string     `gorm:"size:500;column:invoke_target" json:"invokeTarget"`
	JobMessage    string     `gorm:"size:500;column:job_message" json:"jobMessage"`
	Status        int8       `gorm:"index" json:"status"`                    // 0正常 1失败
	ExceptionInfo string     `gorm:"type:text;column:exception_info" json:"exceptionInfo"`
	StartTime     *time.Time `gorm:"column:start_time" json:"startTime,omitempty"`
	EndTime       *time.Time `gorm:"column:end_time" json:"endTime,omitempty"`
	CreateTime    time.Time  `gorm:"column:create_time" json:"createTime"`
}

func (JobLog) TableName() string { return "t_job_log" }

// Duration 计算执行耗时(毫秒)
func (jl *JobLog) Duration() *int64 {
	if jl.StartTime != nil && jl.EndTime != nil {
		duration := jl.EndTime.Sub(*jl.StartTime).Milliseconds()
		return &duration
	}
	return nil
}

func (j *Job) BeforeCreate(tx *gorm.DB) error { now := time.Now(); j.CreateTime = now; j.UpdateTime = now; return nil }
func (j *Job) BeforeUpdate(tx *gorm.DB) error { j.UpdateTime = time.Now(); return nil }
func (jl *JobLog) BeforeCreate(tx *gorm.DB) error { jl.CreateTime = time.Now(); return nil }
