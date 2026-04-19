package mq

import (
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/aurora-go/aurora/internal/config"
	"github.com/aurora-go/aurora/internal/constant"
)

var Conn *amqp.Connection
var Channel *amqp.Channel

// InitRabbitMQ 初始化 RabbitMQ 连接
func InitRabbitMQ(cfg *config.RabbitMQConfig) error {
	var err error

	Conn, err = amqp.Dial(cfg.URL())
	if err != nil {
		slog.Error("Failed to connect to RabbitMQ", "error", err)
		return err
	}

	Channel, err = Conn.Channel()
	if err != nil {
		slog.Error("Failed to open channel", "error", err)
		return err
	}

	// 声明交换机
	err = Channel.ExchangeDeclare(
		constant.ExchangeDirect, "direct",
		true, false, false, false, nil,
	)
	if err != nil {
		slog.Error("Failed to declare direct exchange", "error", err)
		return err
	}

	err = Channel.ExchangeDeclare(
		constant.ExchangeTopic, "topic",
		true, false, false, false, nil,
	)
	if err != nil {
		slog.Error("Failed to declare topic exchange", "error", err)
		return err
	}

	// 声明订阅广播交换机 (对标Java FanoutExchange: subscribe_exchange)
	err = Channel.ExchangeDeclare(
		constant.ExchangeSubscribe, "fanout",
		true, false, false, false, nil,
	)
	if err != nil {
		slog.Error("Failed to declare subscribe exchange", "error", err)
		return err
	}

	// 设置QoS
	err = Channel.Qos(cfg.PrefetchCount, 0, false)
	if err != nil {
		slog.Error("Failed to set QoS", "error", err)
	}

	slog.Info("RabbitMQ connected successfully",
		"host", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		"prefetch_count", cfg.PrefetchCount,
	)
	return nil
}

// CloseRabbitMQ 关闭 RabbitMQ 连接
func CloseRabbitMQ() error {
	if Channel != nil {
		Channel.Close()
	}
	if Conn != nil {
		return Conn.Close()
	}
	return nil
}

// GetChannel 获取 RabbitMQ Channel
func GetChannel() *amqp.Channel {
	return Channel
}

// DeclareQueue 声明队列并绑定到交换机
func DeclareQueue(queueName, routingKey, exchange string) (amqp.Queue, error) {
	queue, err := Channel.QueueDeclare(
		queueName,
		true, false, false, false, nil,
	)
	if err != nil {
		return queue, err
	}

	err = Channel.QueueBind(queueName, routingKey, exchange, false, nil)
	return queue, err
}

// PublishSubscribe 发布订阅通知消息 (对标Java rabbitTemplate.convertAndSend)
// 对标Java: ArticleServiceImpl.saveOrUpdateArticle() → rabbitTemplate.convertAndSend(SUBSCRIBE_EXCHANGE, "*", message)
func PublishSubscribe(articleID uint) error {
	if Channel == nil {
		return fmt.Errorf("RabbitMQ channel not initialized")
	}

	// 构建消息体 (对标Java: JSON.toJSONBytes(article.getId()))
	msgBody := fmt.Sprintf(`{"articleId":%d}`, articleID)

	err := Channel.Publish(
		constant.ExchangeSubscribe, // 订阅广播交换机
		"",                          // FanoutExchange 忽略 routing key
		false, false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(msgBody),
			MessageId:   fmt.Sprintf("sub-%d", articleID),
		},
	)
	if err != nil {
		return fmt.Errorf("publish subscribe notification failed: %w", err)
	}

	slog.Info("📢 已发送文章订阅通知", "article_id", articleID)
	return nil
}
