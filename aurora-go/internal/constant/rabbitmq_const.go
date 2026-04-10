package constant

// ==================== RabbitMQ 常量定义 ====================
// 对标 Java 版 RabbitMQConfig 中的交换机/队列/路由键配置

const (
	// Exchange 交换机
	ExchangeDirect = "aurora.direct"   // 直连交换机: 邮件通知/订阅通知
	ExchangeTopic  = "aurora.topic"    // Topic交换机: Maxwell ES数据同步

	// Queue 队列名
	QueueEmailNotify    = "aurora.email.notify"     // 评论邮件通知队列
	QueueSubscribeNotify = "aurora.subscribe.notify" // 文章订阅通知队列
	QueueMaxwellSync     = "aurora.es.sync"          // Maxwell ES数据同步队列

	// Routing Key 路由键
	RoutingKeyEmail     = "email.notify"
	RoutingKeySubscribe = "subscribe.notify"
	RoutingKeyMaxwellDB = "maxwell.db" // Maxwell 数据库变更事件

	// Maxwell 相关常量（对标 Java 版 Maxwell 配置）
	MaxwellDatabase    = "aurora"       // 监听的数据库名
	MaxwellTablePrefix = "tb_"          // 监听的表前缀（Aurora表都以 tb_ 开头）
)
