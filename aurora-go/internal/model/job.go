package model

import (
	"time"

	"gorm.io/gorm"
)

// Job 定时任务实体 (对应 t_job 表)
// 对标Java: id, job_name, job_group, invoke_target, cron_expression, misfire_policy, concurrent, status, create_time, update_time, remark
type Job struct {
	ID             uint       `gorm:"primarykey;column:id" json:"id"`
	JobName        string     `gorm:"column:job_name;size:64;not null" json:"jobName"`
	JobGroup       string     `gorm:"column:job_group;size:64;not null;default:DEFAULT" json:"jobGroup"`
	InvokeTarget   string     `gorm:"column:invoke_target;size:500;not null" json:"invokeTarget"`
	CronExpression string     `gorm:"column:cron_expression;size:255" json:"cronExpression"`
	MisfirePolicy  int        `gorm:"column:misfire_policy;default:3" json:"misfirePolicy"`
	Concurrent     int        `gorm:"column:concurrent;default:1" json:"concurrent"`
	Status         int        `gorm:"column:status;default:0" json:"status"` // 0暂停 1正常（Java: 0暂停 1正常）
	Remark         string     `gorm:"column:remark;size:500" json:"remark"`
	CreateTime     time.Time  `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateTime     time.Time  `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
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
