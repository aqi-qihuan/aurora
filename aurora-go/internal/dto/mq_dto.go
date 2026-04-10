package dto

// ==================== RabbitMQ 消息 DTO ====================
// 对标 Java 版: MaxwellDataDTO / EmailDTO (消息队列传输对象)

// MaxwellDataDTO Maxwell Binlog 数据变更消息
// 来源: MySQL → Maxwell(监听binlog) → RabbitMQ(maxwell_queue) → Go Consumer
// 对标Java: com.aurora.model.dto.MaxwellDataDTO
type MaxwellDataDTO struct {
	Database string                 `json:"database"` // 数据库名
	XID      int                    `json:"xid"`      // 事务ID
	Data     map[string]interface{} `json:"data"`     // 变更的行数据 (JSON Object)
	Commit   bool                   `json:"commit"`   // 是否提交
	Type     string                 `json:"type"`     // 操作类型: insert/update/delete
	Table    string                 `json:"table"`    // 表名 (如 tb_article)
	TS       int64                  `json:"ts"`       // Unix时间戳(秒)
}

// EmailDTO 邮件通知消息 (评论/@提醒/订阅)
// 对标Java: com.aurora.model.dto.EmailDTO
type EmailDTO struct {
	Email      string                 `json:"email"`                // 收件人邮箱
	Subject    string                 `json:"subject"`              // 邮件主题
	CommentMap map[string]interface{} `json:"commentMap,omitempty"` // 模板变量 (Thymeleaf context)
	Template   string                 `json:"template,omitempty"`   // 模板文件名 (如 common.html / owner.html)
}

// SubscribeMessageDTO 文章订阅通知消息
// 对标Java: SubscribeConsumer.process() 中从队列读取的是 Integer(articleId), 但Go用结构体更好
type SubscribeMessageDTO struct {
	ArticleID uint `json:"articleId"` // 新发布/更新的文章ID
}
