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
	"github.com/aurora-go/aurora/internal/infrastructure/mq"
	"github.com/aurora-go/aurora/internal/infrastructure/search"
)

// MaxwellConsumer Maxwell ES同步消费者
// 对标Java: MaxWellConsumer (@RabbitListener + @RabbitHandler)
// 功能: 监听 MySQL Binlog 变更 → 同步数据到 Elasticsearch (insert/update/delete)
//
// 消息流:
//   Maxwell(监听binlog) → maxwell_queue → MaxwellConsumer.process() → ES IndexDocument/DeleteDocument
type MaxwellConsumer struct {
	ch       *amqp.Channel
	esClient *search.ESClient
	logger   *slog.Logger
}

// NewMaxwellConsumer 创建Maxwell ES同步消费者
func NewMaxwellConsumer(ch *amqp.Channel, esClient *search.ESClient, logger *slog.Logger) *MaxwellConsumer {
	return &MaxwellConsumer{
		ch:       ch,
		esClient: esClient,
		logger:   logger,
	}
}

// Start 启动消费循环 (阻塞运行在goroutine中)
// 对标Java: @RabbitListener(queues = MAXWELL_QUEUE) + @RabbitHandler process(byte[] data)
func (c *MaxwellConsumer) Start() error {
	// Step 1: 声明队列并绑定到Topic交换机
	queue, err := mq.DeclareQueue(constant.QueueMaxwellSync, constant.RoutingKeyMaxwellDB, constant.ExchangeTopic)
	if err != nil {
		return fmt.Errorf("declare maxwell queue failed: %w", err)
	}

	// Step 2: 注册消费者
	msgs, err := c.ch.Consume(
		queue.Name, // queue
		"maxwell-consumer",    // consumer tag
		false,                 // auto-ack: false → 手动确认
		false,                 // exclusive
		false,                 // no-local
		false,                 // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("register maxwell consumer failed: %w", err)
	}

	c.logger.Info("✅ Maxwell ES同步消费者已启动",
		"queue", constant.QueueMaxwellSync,
		"exchange", constant.ExchangeTopic,
		"routing_key", constant.RoutingKeyMaxwellDB,
	)

	// Step 3: 消费消息循环
	go func() {
		for msg := range msgs {
			if err := c.processMessage(msg); err != nil {
				// 处理失败: Nack并重入队(requeue=true), 让其他消费者重试或延迟处理
				c.logger.Error("❌ 处理Maxwell消息失败",
					"error", err,
					"message_id", msg.MessageId,
					"requeue", true,
				)
				_ = msg.Nack(false, true) // requeue=true 重入队
			} else {
				// 处理成功: ACK确认
				_ = msg.Ack(false)
			}
		}
		c.logger.Warn("⚠️ Maxwell消费者通道关闭")
	}()

	return nil
}

// processMessage 处理单条Binlog消息 (对标Java MaxWellConsumer.process())
func (c *MaxwellConsumer) processMessage(msg amqp.Delivery) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 安全检查: ES未启用时跳过 (对标Java elasticsearchClient == null)
	if c.esClient == nil {
		c.logger.Debug("Elasticsearch未启用, 跳过ES同步", "body", string(msg.Body))
		return nil
	}

	// Step 1: 反序列化 MaxwellDataDTO
	var maxwellDTO dto.MaxwellDataDTO
	if err := json.Unmarshal(msg.Body, &maxwellDTO); err != nil {
		return fmt.Errorf("反序列化MaxwellData失败: %w", err)
	}

	// 仅处理 aurora 数据库的变更
	if maxwellDTO.Database != constant.MaxwellDatabase {
		return nil
	}

	// 仅处理文章表 (tb_article)
	if !isArticleTable(maxwellDTO.Table) {
		return nil
	}

	// Step 2: 根据操作类型分发处理 (对标Java switch(type))
	switch maxwellDTO.Type {
	case "insert", "update":
		return c.syncToES(ctx, maxwellDTO.Data)

	case "delete":
		return c.deleteFromES(ctx, maxwellDTO.Data)

	default:
		c.logger.Debug("忽略非CRUD操作类型", "type", maxwellDTO.Type)
		return nil
	}
}

// syncToES 将文章数据索引到 Elasticsearch (对标Java insert/update分支)
func (c *MaxwellConsumer) syncToES(ctx context.Context, data map[string]interface{}) error {
	articleID := extractArticleID(data)
	if articleID == "" {
		return fmt.Errorf("缺少文章ID字段")
	}

	// 构建ES文档 (对标Java ArticleSearchDTO)
	searchDoc := map[string]interface{}{
		"id":              articleID,
		"articleTitle":     getStringField(data, "article_title"),
		"articleContent":   getStringField(data, "article_content"),
		"isDelete":         getIntField(data, "is_delete", 0),
		"status":           getIntField(data, "status", 1),
	}

	indexName := "article"
	if err := c.esClient.IndexDocument(ctx, indexName, articleID, searchDoc); err != nil {
		return fmt.Errorf("ES索引文档失败(id=%s): %w", articleID, err)
	}

	c.logger.Info("📄 文章已同步到ES",
		"article_id", articleID,
		"action", "index",
		"title", searchDoc["articleTitle"],
	)

	return nil
}

// deleteFromES 从Elasticsearch删除文档 (对标Java delete分支)
func (c *MaxwellConsumer) deleteFromES(ctx context.Context, data map[string]interface{}) error {
	articleID := extractArticleID(data)
	if articleID == "" {
		return fmt.Errorf("缺少文章ID字段")
	}

	if err := c.esClient.DeleteDocument(ctx, "article", articleID); err != nil {
		return fmt.Errorf("ES删除文档失败(id=%s): %w", articleID, err)
	}

	c.logger.Info("🗑️ 文章已从ES删除",
		"article_id", articleID,
		"action", "delete",
	)

	return nil
}

// ==================== 辅助函数 ====================

// isArticleTable 判断是否为文章相关表 (支持 tb_article / t_article 等命名变体)
func isArticleTable(table string) bool {
	return table == "t_article" || table == "tb_article" ||
		table == "article" || table == constant.MaxwellTablePrefix+"article"
}

// extractArticleID 从data中提取文章ID
func extractArticleID(data map[string]interface{}) string {
	// 支持多种字段名格式: id, article_id, articleId (GORM默认)
	for _, key := range []string{"id", "article_id", "articleId"} {
		if v, ok := data[key]; ok {
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
}

// getStringField 安全获取string字段
func getStringField(data map[string]interface{}, key string) string {
	v, ok := data[key]
	if !ok {
		return ""
	}
	// 处理JSON number类型 (MySQL数字字段可能被解析为float64)
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return fmt.Sprintf("%.0f", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// getIntField 安全获取int字段
func getIntField(data map[string]interface{}, key string, defaultVal int) int {
	v, ok := data[key]
	if !ok {
		return defaultVal
	}
	switch val := v.(type) {
	case float64:
		return int(val)
	case int:
		return val
	case int8:
		return int(val)
	default:
		return defaultVal
	}
}
