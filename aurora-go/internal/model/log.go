package model

import (
	"time"

	"gorm.io/gorm"
)

// OperationLog 操作日志实体 (对应 t_operation_log 表)
type OperationLog struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	Module     string    `gorm:"size:50;index" json:"module"`           // 操作模块
	Operation  string    `gorm:"size:50;index" json:"operation"`          // 操作类型
	Method     string    `gorm:"size:200" json:"method"`                 // 操作方法
	URL        string    `gorm:"size:500" json:"url"`                    // 请求路径
	IP         string    `gorm:"size:64" json:"ip"`
	Duration   *int64    `json:"duration,omitempty"`                     // 执行耗时(ms)
	Status     int8      `json:"status"`                                // 状态(0失败 1成功)
	ErrorMsg   string    `gorm:"type:text" json:"errorMsg"`              // 错误信息
	UserID     uint      `json:"userId"`

	// 关联 (预加载用, 不存DB)
	UserInfo *UserInfo `gorm:"foreignKey:UserID" json:"userInfo,omitempty"`
	CreateTime time.Time `json:"createTime"`
}

func (OperationLog) TableName() string { return "t_operation_log" }

func (o *OperationLog) BeforeCreate(tx *gorm.DB) error { o.CreateTime = time.Now(); return nil }

// ExceptionLog 异常日志实体 (对应 t_exception_log 表)
type ExceptionLog struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	URL        string    `gorm:"size:500" json:"url"`                   // 请求URL
	Method     string    `gorm:"size:200" json:"method"`                  // 请求方法
	IP         string    `gorm:"size:64" json:"ip"`
	ErrorMsg   string    `gorm:"type:longtext" json:"errorMsg"`            // 错误消息
	Stacktrace string    `gorm:"type:longtext" json:"stacktrace"`          // 堆栈信息
	Status     int8      `gorm:"default:1;index" json:"status"`           // 0已处理 1未处理
	UserID     uint      `json:"userId"`
	CreateTime time.Time `json:"createTime"`
}

func (ExceptionLog) TableName() string { return "t_exception_log" }

func (e *ExceptionLog) BeforeCreate(tx *gorm.DB) error { e.CreateTime = time.Now(); return nil }
