package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/aurora-go/aurora/internal/constant"
	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/infrastructure/email"
	"github.com/aurora-go/aurora/internal/infrastructure/mq"
	"github.com/aurora-go/aurora/internal/model"
	"gorm.io/gorm"
)

// SubscribeConsumer 文章订阅通知消费者
// 对标Java: SubscribeConsumer (@RabbitListener + @RabbitHandler)
//
// 消息流:
//   ArticleService.saveOrUpdateArticle() → rabbitTemplate.convertAndSend(SUBSCRIBE_EXCHANGE, articleId)
//     → subscribe_queue → SubscribeConsumer.process()
//     → 查询文章详情 + 查询所有订阅用户 → 逐个发送邮件
type SubscribeConsumer struct {
	ch      *amqp.Channel
	db      *gorm.DB
	logger  *slog.Logger
	siteURL string // 网站URL (用于生成文章链接)
}

// NewSubscribeConsumer 创建文章订阅通知消费者
func NewSubscribeConsumer(ch *amqp.Channel, db *gorm.DB, siteURL string, logger *slog.Logger) *SubscribeConsumer {
	return &SubscribeConsumer{
		ch:      ch,
		db:      db,
		logger:  logger,
		siteURL: siteURL,
	}
}

// Start 启动消费循环 (阻塞运行在goroutine中)
func (c *SubscribeConsumer) Start() error {
	queue, err := mq.DeclareQueue(constant.QueueSubscribeNotify, constant.RoutingKeySubscribe, constant.ExchangeDirect)
	if err != nil {
		return fmt.Errorf("declare subscribe notify queue failed: %w", err)
	}

	msgs, err := c.ch.Consume(
		queue.Name,
		"subscribe-consumer",
		false, // 手动确认
		false, false, false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("register subscribe consumer failed: %w", err)
	}

	c.logger.Info("✅ 文章订阅通知消费者已启动",
		"queue", constant.QueueSubscribeNotify,
		"exchange", constant.ExchangeDirect,
	)

	go func() {
		for msg := range msgs {
			if err := c.processMessage(msg); err != nil {
				c.logger.Error("❌ 处理订阅通知失败",
					"error", err,
					"message_id", msg.MessageId,
				)
				_ = msg.Nack(false, false) // 订阅邮件不重入队
			} else {
				_ = msg.Ack(false)
			}
		}
		c.logger.Warn("⚠️ 订阅消费者通道关闭")
	}()

	return nil
}

// processMessage 处理单条订阅消息
// 对标Java: SubscribeConsumer.process(byte[] data)
//
// Java流程:
//   Integer articleId = JSON.parseObject(new String(data), Integer.class);
//   Article article = articleService.getById(articleId);
//   List<UserInfo> users = userInfoService.list(isSubscribe=TRUE);
//   for each user → build EmailDTO → emailUtil.sendHtmlMail(emailDTO);
func (c *SubscribeConsumer) processMessage(msg amqp.Delivery) error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	var articleID uint

	// 兼容两种格式: 整数ID 或 DTO结构体
	var subMsg dto.SubscribeMessageDTO
	if err := json.Unmarshal(msg.Body, &subMsg); err == nil && subMsg.ArticleID > 0 {
		articleID = subMsg.ArticleID
	} else {
		// 回退: 尝试解析纯整数 (对标Java版直接传Integer)
		var id int
		if json.Unmarshal(msg.Body, &id); err == nil && id > 0 {
			articleID = uint(id)
		} else {
			return fmt.Errorf("无法解析文章ID")
		}
	}

	// Step 1: 查询文章详情
	var article model.Article
	err := c.db.WithContext(ctx).First(&article, articleID).Error
	if err != nil {
		// 文章可能已被删除, 不算错误
		c.logger.Warn("查询文章失败(可能已删除)", "article_id", articleID, "error", err)
		return nil
	}

	// Step 2: 查询所有已订阅用户 (对标Java isSubscribe=TRUE)
	var subscribers []model.UserInfo
	err = c.db.WithContext(ctx).
		Where("is_subscribe = ? AND email != '' AND is_delete = ?", 1, 0).
		Find(&subscribers).Error
	if err != nil {
		return fmt.Errorf("查询订阅用户失败: %w", err)
	}

	if len(subscribers) == 0 {
		c.logger.Debug("无订阅用户", "article_id", articleID)
		return nil
	}

	// Step 3: 构建文章链接和邮件内容
	articleURL := c.siteURL + "/articles/" + fmt.Sprintf("%d", articleID)

	// Step 4: 批量发送邮件 (逐个发送, 避免群发被拦截)
	successCount := 0
	failCount := 0

	for _, user := range subscribers {
		subscribeData := email.SubscribeNotifyData{
			ArticleTitle:   article.ArticleTitle,
			ArticleSummary: truncateContent(article.ArticleContent, 150),
			ArticleURL:     articleURL,
			SiteName:       c.siteName(),
			UnsubscribeURL: c.siteURL + "/unsubscribe?email=" + user.Email,
			CurrentYear:    time.Now().Year(),
		}

		subject := "📝 文章订阅"
		if !article.UpdateTime.IsZero() && !article.UpdateTime.Equal(article.CreateTime) {
			subject = "🔄 文章更新通知"
		}

		if err := email.SendSubscribeNotification(user.Email, subject, subscribeData); err != nil {
			c.logger.Warn("订阅邮件发送失败",
				"email", user.Email,
				"article_id", articleID,
				"error", err,
			)
			failCount++
		} else {
			successCount++
		}

		// 发送间隔100ms, 避免SMTP频率限制
		time.Sleep(100 * time.Millisecond)
	}

	c.logger.Info("📬 文章订阅通知完成",
		"article_id", articleID,
		"title", article.ArticleTitle,
		"total_subscribers", len(subscribers),
		"success", successCount,
		"failed", failCount,
	)

	return nil
}

// ==================== 辅助函数 ====================

// siteName 获取站点名称 (TODO: 从WebsiteConfig读取, 当前硬编码)
func (c *SubscribeConsumer) siteName() string {
	return "Aurora Blog" // TODO: c.websiteConfig.GetSiteName()
}

// truncateContent 截断文章内容作为摘要
func truncateContent(content string, maxLen int) string {
	if len(content) <= maxLen {
		return stripHTML(content)
	}
	runes := []rune(stripHTML(content))
	if len(runes) <= maxLen {
		return string(runes)
	}
	return string(runes[:maxLen]) + "..."
}

// stripHTML 去除HTML标签 (简单正则实现)
func stripHTML(html string) string {
	result := make([]byte, 0, len(html))
	inTag := false
	for i := 0; i < len(html); i++ {
		switch html[i] {
		case '<':
			inTag = true
		case '>':
			inTag = false
		default:
			if !inTag {
				result = append(result, html[i])
			}
		}
	}
	return string(result)
}
