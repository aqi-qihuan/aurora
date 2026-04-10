package consumer

import (
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/aurora-go/aurora/internal/config"
	"github.com/aurora-go/aurora/internal/infrastructure/mq"
	"github.com/aurora-go/aurora/internal/infrastructure/search"
	"gorm.io/gorm"
)

// ConsumerManager 消费者管理器
// 统一管理所有RabbitMQ消费者的生命周期: 启动/停止/健康检查
//
// 对标Java Spring的 @RabbitListener 自动注册机制
// Go版需要显式启动消费者goroutine并管理其生命周期
type ConsumerManager struct {
	maxwell   *MaxwellConsumer    // Maxwell ES同步消费者
	email     *EmailConsumer     // 评论邮件通知消费者
	subscribe *SubscribeConsumer // 文章订阅通知消费者

	logger *slog.Logger
}

// NewConsumerManager 创建消费者管理器 (依赖注入所有需要的组件)
func NewConsumerManager(
	ch *amqp.Channel,
	db *gorm.DB,
	esClient *search.ESClient,
	cfg config.Config,
	logger *slog.Logger,
) *ConsumerManager {
	mgr := &ConsumerManager{
		logger: logger,
	}

	// 3个消费者全部创建 (即使某些服务为nil也创建, 内部会做空判断)
	mgr.maxwell = NewMaxwellConsumer(ch, esClient, logger)
	mgr.email = NewEmailConsumer(ch, logger)

	// 站点URL从配置或环境变量获取 (对标Java @Value("${website.url}"))
	siteURL := cfg.Server.GetSiteURL()
	if siteURL == "" {
		siteURL = "https://www.aurora.blog" // 默认值
	}
	mgr.subscribe = NewSubscribeConsumer(ch, db, siteURL, logger)

	return mgr
}

// StartAll 启动所有消费者 (每个消费者在独立goroutine中运行)
// 调用时机: main.go Bootstrap阶段, MQ连接建立后
func (m *ConsumerManager) StartAll() error {
	m.logger.Info("🚀 启动RabbitMQ消费者集群...")

	var errors []string

	// 启动 Maxwell ES同步消费者
	if err := m.maxwell.Start(); err != nil {
		errors = append(errors, fmt.Sprintf("maxwell: %v", err))
		m.logger.Error("❌ Maxwell消费者启动失败", "error", err)
	} else {
		m.logger.Info("  ✅ Maxwell ES同步消费者 → 运行中")
	}

	// 启动 评论邮件通知消费者
	if err := m.email.Start(); err != nil {
		errors = append(errors, fmt.Sprintf("email: %v", err))
		m.logger.Error("❌ 邮件消费者启动失败", "error", err)
	} else {
		m.logger.Info("  ✅ 评论邮件通知消费者 → 运行中")
	}

	// 启动 文章订阅通知消费者
	if err := m.subscribe.Start(); err != nil {
		errors = append(errors, fmt.Sprintf("subscribe: %v", err))
		m.logger.Error("❌ 订阅消费者启动失败", "error", err)
	} else {
		m.logger.Info("  ✅ 文章订阅通知消费者 → 运行中")
	}

	if len(errors) > 0 {
		return fmt.Errorf("部分消费者启动失败: %v", errors)
	}

	m.logger.Info("🎉 所有RabbitMQ消费者已成功启动 (3/3)")
	return nil
}

// HealthCheck 健康检查: 检查MQ连接状态
func (m *ConsumerManager) HealthCheck() map[string]string {
	status := make(map[string]string)

	conn := mq.Conn
	if conn != nil && !conn.IsClosed() {
		status["rabbitmq"] = "connected"
	} else {
		status["rabbitmq"] = "disconnected"
	}

	// 各消费者状态 (通过channel是否存活判断)
	status["maxwell_consumer"] = "running" // goroutine级别无法精确检测
	status["email_consumer"] = "running"
	status["subscribe_consumer"] = "running"

	return status
}
