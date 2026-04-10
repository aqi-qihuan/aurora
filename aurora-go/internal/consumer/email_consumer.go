package consumer

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/aurora-go/aurora/internal/constant"
	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/infrastructure/email"
	"github.com/aurora-go/aurora/internal/infrastructure/mq"
)

// EmailConsumer 评论邮件通知消费者
// 对标Java: CommentNoticeConsumer (@RabbitListener + @RabbitHandler)
//
// 消息流:
//   CommentService.saveComment() → rabbitTemplate.convertAndSend(EMAIL_EXCHANGE, EmailDTO)
//     → email_queue → EmailConsumer.process() → email.SendCommentNotification()
//
// 支持的邮件类型 (对标Java CommentServiceImpl 中的3种通知):
//  1. COMMENT_REMIND: 新评论通知 (博主收到新评论)
//  2. MENTION_REMIND: @提醒通知 (被@的用户)
//  3. REVIEW_REMIND: 审核提醒 (评论审核通过/拒绝 - 可选扩展)
type EmailConsumer struct {
	ch      *amqp.Channel
	logger  *slog.Logger
}

// NewEmailConsumer 创建评论邮件通知消费者
func NewEmailConsumer(ch *amqp.Channel, logger *slog.Logger) *EmailConsumer {
	return &EmailConsumer{
		ch:     ch,
		logger: logger,
	}
}

// Start 启动消费循环 (阻塞运行在goroutine中)
func (c *EmailConsumer) Start() error {
	queue, err := mq.DeclareQueue(constant.QueueEmailNotify, constant.RoutingKeyEmail, constant.ExchangeDirect)
	if err != nil {
		return fmt.Errorf("declare email notify queue failed: %w", err)
	}

	msgs, err := c.ch.Consume(
		queue.Name,
		"email-consumer",
		false, // 手动确认
		false, false, false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("register email consumer failed: %w", err)
	}

	c.logger.Info("✅ 评论邮件通知消费者已启动",
		"queue", constant.QueueEmailNotify,
		"exchange", constant.ExchangeDirect,
	)

	go func() {
		for msg := range msgs {
			if err := c.processMessage(msg); err != nil {
				c.logger.Error("❌ 处理邮件通知失败",
					"error", err,
					"message_id", msg.MessageId,
				)
				// 邮件失败通常不重入队 (避免无限重试轰炸), 直接ACK丢弃或发到死信队列
				_ = msg.Nack(false, false) // requeue=false, 死信
			} else {
				_ = msg.Ack(false)
			}
		}
		c.logger.Warn("⚠️ 邮件消费者通道关闭")
	}()

	return nil
}

// processMessage 处理单条邮件通知消息
// 对标Java: CommentNoticeConsumer.process(byte[] data)
func (c *EmailConsumer) processMessage(msg amqp.Delivery) error {
	var emailDTO dto.EmailDTO
	if err := json.Unmarshal(msg.Body, &emailDTO); err != nil {
		return fmt.Errorf("反序列化EmailDTO失败: %w", err)
	}

	if emailDTO.Email == "" {
		return fmt.Errorf("收件人邮箱为空")
	}

	// 从commentMap构建模板数据 (对齐Java版Thymeleaf模板变量)
	commentData := buildCommentNotifyData(emailDTO)

	// 发送HTML邮件 (调用email基础设施层)
	if err := email.SendCommentNotification(
		emailDTO.Email,
		emailDTO.Subject,
		commentData,
	); err != nil {
		return fmt.Errorf("发送邮件失败(to=%s): %w", emailDTO.Email, err)
	}

	c.logger.Info("📧 评论邮件通知已发送",
		"to", emailDTO.Email,
		"subject", emailDTO.Subject,
	)

	return nil
}

// ==================== 辅助函数 ====================

// buildCommentNotifyData 将Java版EmailDTO.commentMap转换为Go版CommentNotifyData结构体
// 对标Java Thymeleaf模板变量: content / url / title / nickname / time 等
func buildCommentNotifyData(dto dto.EmailDTO) email.CommentNotifyData {
	data := email.CommentNotifyData{
		SiteName:    "Aurora Blog", // TODO: 从WebsiteConfig读取
		CurrentYear: time.Now().Year(),
	}

	if dto.CommentMap != nil {
		// 安全提取模板变量 (兼容Java版Map<String,Object>)
		if v, ok := dto.CommentMap["content"]; ok {
			data.Content = toString(v)
		}
		if v, ok := dto.CommentMap["url"]; ok {
			data.ArticleURL = toString(v)
		}
		if v, ok := dto.CommentMap["title"]; ok {
			data.ArticleTitle = toString(v)
		}
		if v, ok := dto.CommentMap["nickname"]; ok {
			data.Nickname = toString(v)
		}
	}

	return data
}

// toString 安全转换interface{}为string
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	default:
		return fmt.Sprintf("%v", val)
	}
}
