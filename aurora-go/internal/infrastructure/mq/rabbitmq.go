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
