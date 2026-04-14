package model

import (
	"time"

	"gorm.io/gorm"
)

// OperationLog 操作日志实体 (对标Java OperationLog，对应 t_operation_log 表)
// 数据库列名使用 opt_module, opt_type 等Java风格命名
type OperationLog struct {
	ID            uint      `gorm:"column:id;primarykey" json:"id"`
	OptModule     string    `gorm:"column:opt_module;size:20" json:"optModule"`       // 操作模块
	OptType       string    `gorm:"column:opt_type;size:20" json:"optType"`           // 操作类型
	OptUri        string    `gorm:"column:opt_uri;size:255" json:"optUri"`            // 操作url
	OptMethod     string    `gorm:"column:opt_method;size:255" json:"optMethod"`      // 操作方法
	OptDesc       string    `gorm:"column:opt_desc;size:255" json:"optDesc"`          // 操作描述
	RequestParam  string    `gorm:"column:request_param;type:longtext" json:"requestParam"`  // 请求参数
	RequestMethod string    `gorm:"column:request_method;size:20" json:"requestMethod"`    // 请求方式
	ResponseData  string    `gorm:"column:response_data;type:longtext" json:"responseData"` // 返回数据
	UserID        uint      `gorm:"column:user_id" json:"userId"`                     // 用户id
	Nickname      string    `gorm:"column:nickname;size:50" json:"nickname"`          // 用户昵称
	IpAddress     string    `gorm:"column:ip_address;size:255" json:"ipAddress"`      // 操作ip
	IpSource      string    `gorm:"column:ip_source;size:255" json:"ipSource"`        // 操作地址
	CreateTime    time.Time `gorm:"column:create_time" json:"createTime"`
	UpdateTime    *time.Time `gorm:"column:update_time" json:"updateTime"`
}

func (OperationLog) TableName() string { return "t_operation_log" }

func (o *OperationLog) BeforeCreate(tx *gorm.DB) error {
	o.CreateTime = time.Now()
	return nil
}

// ExceptionLog 异常日志实体 (对标Java ExceptionLog，对应 t_exception_log 表)
// 数据库列名使用 opt_uri, opt_method 等Java风格命名
type ExceptionLog struct {
	ID            uint      `gorm:"column:id;primarykey" json:"id"`
	OptUri        string    `gorm:"column:opt_uri;size:255" json:"optUri"`            // 请求接口
	OptMethod     string    `gorm:"column:opt_method;size:255" json:"optMethod"`      // 请求方法
	RequestMethod string    `gorm:"column:request_method;size:255" json:"requestMethod"` // 请求方式
	RequestParam  string    `gorm:"column:request_param;size:2000" json:"requestParam"`  // 请求参数
	OptDesc       string    `gorm:"column:opt_desc;size:255" json:"optDesc"`          // 操作描述
	ExceptionInfo string    `gorm:"column:exception_info;type:text" json:"exceptionInfo"` // 异常信息
	IpAddress     string    `gorm:"column:ip_address;size:255" json:"ipAddress"`      // ip
	IpSource      string    `gorm:"column:ip_source;size:255" json:"ipSource"`        // ip来源
	CreateTime    time.Time `gorm:"column:create_time" json:"createTime"`
}

func (ExceptionLog) TableName() string { return "t_exception_log" }

func (e *ExceptionLog) BeforeCreate(tx *gorm.DB) error {
	e.CreateTime = time.Now()
	return nil
}
