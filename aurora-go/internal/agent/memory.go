package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/aurora-go/aurora/internal/config"
	"github.com/aurora-go/aurora/internal/infrastructure/database"
	"github.com/redis/go-redis/v9"
)

// ========== Memory 会话记忆服务 ==========
// 对标 tRPC memory/memorysvc + Redis持久化适配器 (~60行)
// 支持: InMemory(开发) / Redis持久化(生产) 两种模式

// SessionMessage 会话中的单条消息
type SessionMessage struct {
	Role      string    `json:"role"`      // user/assistant/system/tool
	Content   string    `json:"content"`   // 消息内容
	Timestamp time.Time `json:"timestamp"` // 时间戳
}

// Session 会话（包含完整消息历史）
type Session struct {
	ID        string           `json:"id"`
	UserID    uint             `json:"user_id"`
	Messages  []SessionMessage `json:"messages"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
	mu        sync.RWMutex     // 保护Messages并发访问
}

// MemoryService 记忆服务接口
type MemoryService interface {
	GetOrCreateSession(ctx context.Context, sessionID string) (*Session, error)
	SaveMessage(ctx context.Context, sessionID, role, content string) error
	BuildMessages(ctx context.Context, sessionID, newMessage string) []ChatMessage
	ListSessions(ctx context.Context, userID uint) ([]*Session, error)
	DeleteSession(ctx context.Context, sessionID string) error
	Close()
}

// ========== InMemory 实现 (开发环境) ==========

type inMemoryMemory struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	maxTurns int // 最大保留轮数（防止上下文无限增长）
}

func newInMemoryMemory(maxTurns int) *inMemoryMemory {
	return &inMemoryMemory{
		sessions: make(map[string]*Session),
		maxTurns: maxTurns,
	}
}

func (m *inMemoryMemory) GetOrCreateSession(ctx context.Context, sessionID string) (*Session, error) {
	if sessionID == "" {
		sessionID = generateSessionID()
	}

	m.mu.RLock()
	sess, ok := m.sessions[sessionID]
	m.mu.RUnlock()

	if ok {
		return sess, nil
	}

	sess = &Session{
		ID:        sessionID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	m.mu.Lock()
	m.sessions[sessionID] = sess
	m.mu.Unlock()

	return sess, nil
}

func (m *inMemoryMemory) SaveMessage(_ context.Context, sessionID, role, content string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess, ok := m.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	sess.mu.Lock()
	sess.Messages = append(sess.Messages, SessionMessage{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})
	sess.UpdatedAt = time.Now()

	// 截断超出最大轮数的消息（保留最近的maxTurns轮）
	if m.maxTurns > 0 && len(sess.Messages) > m.maxTurns*2 {
		sess.Messages = sess.Messages[len(sess.Messages)-m.maxTurns*2:]
	}
	sess.mu.Unlock()

	return nil
}

func (m *inMemoryMemory) BuildMessages(_ context.Context, sessionID, newMessage string) []ChatMessage {
	m.mu.RLock()
	sess, ok := m.sessions[sessionID]
	m.mu.RUnlock()

	var result []ChatMessage

	if ok && len(sess.Messages) > 0 {
		sess.mu.RLock()
		for _, msg := range sess.Messages {
			result = append(result, ChatMessage{Role: msg.Role, Content: msg.Content})
		}
		sess.mu.RUnlock()
	}

	// 追加当前用户消息
	result = append(result, ChatMessage{Role: "user", Content: newMessage})

	return result
}

func (m *inMemoryMemory) ListSessions(_ context.Context, _ uint) ([]*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	sessions := make([]*Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func (m *inMemoryMemory) DeleteSession(_ context.Context, sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, sessionID)
	return nil
}

func (m *inMemoryMemory) Close() {
	m.mu.Lock()
	m.sessions = make(map[string]*Session)
	m.mu.Unlock()
}

// ========== Redis 持久化实现 (生产环境) ==========

const (
	sessionPrefix   = "agent:session:"
	sessionTTL      = 7 * 24 * time.Hour // 会话过期时间7天
	maxRedisMsgLen  = 50                 // Redis中最多保存的消息数
)

type redisMemory struct {
	client   *redis.Client
	maxTurns int
}

func newRedisMemory(client *redis.Client, maxTurns int) *redisMemory {
	return &redisMemory{
		client:   client,
		maxTurns: maxTurns,
	}
}

func (m *redisMemory) GetOrCreateSession(ctx context.Context, sessionID string) (*Session, error) {
	key := sessionPrefix + sessionID

	data, err := m.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		// 新会话
		if sessionID == "" {
			sessionID = generateSessionID()
		}
		sess := &Session{
			ID:        sessionID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		return sess, nil
	} else if err != nil {
		return nil, fmt.Errorf("get session failed: %w", err)
	}

	var sess Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, fmt.Errorf("unmarshal session failed: %w", err)
	}

	return &sess, nil
}

func (m *redisMemory) SaveMessage(ctx context.Context, sessionID, role, content string) error {
	key := sessionPrefix + sessionID

	// 获取现有会话数据
	sess, err := m.GetOrCreateSession(ctx, sessionID)
	if err != nil {
		return err
	}

	sess.mu.Lock()
	sess.Messages = append(sess.Messages, SessionMessage{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})
	sess.UpdatedAt = time.Now()

	// 限制消息数量
	if m.maxTurns > 0 && len(sess.Messages) > m.maxTurns*2 {
		sess.Messages = sess.Messages[len(sess.Messages)-m.maxTurns*2:]
	}
	sess.mu.Unlock()

	// 序列化并保存到Redis
	data, err := json.Marshal(sess)
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}

	return m.client.Set(ctx, key, data, sessionTTL).Err()
}

func (m *redisMemory) BuildMessages(ctx context.Context, sessionID, newMessage string) []ChatMessage {
	sess, err := m.GetOrCreateSession(ctx, sessionID)
	if err != nil || sess == nil {
		// 无历史记录，返回当前消息
		return []ChatMessage{{Role: "user", Content: newMessage}}
	}

	var result []ChatMessage
	sess.mu.RLock()
	for _, msg := range sess.Messages {
		result = append(result, ChatMessage{Role: msg.Role, Content: msg.Content})
	}
	sess.mu.RUnlock()

	result = append(result, ChatMessage{Role: "user", Content: newMessage})
	return result
}

func (m *redisMemory) ListSessions(ctx context.Context, userID uint) ([]*Session, error) {
	iter := m.client.Scan(ctx, 0, sessionPrefix+"*", 0).Iterator()
	var sessions []*Session

	for iter.Next(ctx) {
		data, err := m.client.Get(ctx, iter.Val()).Bytes()
		if err != nil {
			continue
		}
		var sess Session
		if err := json.Unmarshal(data, &sess); err != nil {
			continue
		}
		if userID == 0 || sess.UserID == userID {
			sessions = append(sessions, &sess)
		}
	}

	return sessions, iter.Err()
}

func (m *redisMemory) DeleteSession(ctx context.Context, sessionID string) error {
	return m.client.Del(ctx, sessionPrefix+sessionID).Err()
}

func (m *redisMemory) Close() {} // Redis连接由基础设施层管理

// ========== 工厂方法 + 辅助函数 ==========

// NewMemoryService 根据配置创建记忆服务
func NewMemoryService(memCfg *config.AgentMemoryConfig) (MemoryService, error) {
	switch memCfg.Type {
	case "redis":
		rdb := database.GetRedis()
		if rdb == nil {
			slog.Warn("Redis unavailable for agent memory, falling back to InMemory")
			return newInMemoryMemory(memCfg.MaxTurns), nil
		}
		slog.Info("Using Redis-backed Agent memory")
		return newRedisMemory(rdb, memCfg.MaxTurns), nil

	default:
		slog.Info("Using InMemory Agent memory (development mode)")
		return newInMemoryMemory(memCfg.MaxTurns), nil
	}
}

var sessionCounter uint64

func generateSessionID() string {
	sessionCounter++
	ts := time.Now().UnixNano()
	return fmt.Sprintf("sess_%d_%x", ts, sessionCounter%0xFFFF)
}
