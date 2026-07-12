package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/aurora-go/aurora/internal/config"
	"github.com/aurora-go/aurora/internal/dto"

	"trpc.group/trpc-go/trpc-agent-go/model"
	"trpc.group/trpc-go/trpc-agent-go/model/openai"
)

// ========== LLM 多模型路由器 ==========
// 基于 tRPC-Agent-Go model/openai (v1.10.0) 实现
// 支持: OpenAI GPT / DeepSeek / 阿里通义千问(Qwen) / 腾讯混元(Hunyuan) / Anthropic Claude
//
// OpenAI/DeepSeek/Qwen/Hunyuan → tRPC model/openai (WithVariant)
// Claude → 保留手搓 Anthropic API 适配 (tRPC 无 VariantClaude)

// ChatMessage 对话消息（对标 OpenAI ChatCompletion message format）
type ChatMessage struct {
	Role    string `json:"role"`    // system/user/assistant/tool
	Content string `json:"content"` // 文本内容
	Name    string `json:"name,omitempty"`
}

// streamChunk SSE流式数据块（handler 层依赖此结构）
type streamChunk struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Delta struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

// newStreamChunk 构造 streamChunk 的辅助函数（简化匿名结构体构造）
func newStreamChunk(id, object, modelName string, created int64, content string, finishReason *string) streamChunk {
	return streamChunk{
		ID:      id,
		Object:  object,
		Created: created,
		Model:   modelName,
		Choices: []struct {
			Delta        struct {
				Role    string `json:"role,omitempty"`
				Content string `json:"content,omitempty"`
			} `json:"delta"`
			FinishReason *string `json:"finish_reason"`
		}{
			{
				Delta: struct {
					Role    string `json:"role,omitempty"`
					Content string `json:"content,omitempty"`
				}{Role: "assistant", Content: content},
				FinishReason: finishReason,
			},
		},
	}
}

// ========== Anthropic Claude 专用类型（ClaudeClient 依赖） ==========

type anthropicRequest struct {
	Model     string         `json:"model"`
	MaxTokens int            `json:"max_messages"`
	System    string         `json:"system,omitempty"`
	Messages  []anthropicMsg `json:"messages"`
}
type anthropicMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// ========== LLM 客户端接口 ==========

// LLMClient 单个LLM提供商客户端接口
type LLMClient interface {
	Chat(ctx context.Context, messages []ChatMessage, opts *config.LLMProvider) (string, *dto.TokenUsageDTO, error)
	ChatStream(ctx context.Context, messages []ChatMessage, opts *config.LLMProvider) (<-chan streamChunk, error)
	Close()
}

// ========== LLMRouter 多模型路由器 ==========

// LLMRouter 多模型路由器（全局单例）
type LLMRouter struct {
	mu              sync.RWMutex
	defaultProvider string
	providers       map[string]*config.LLMProvider
	clients         map[string]LLMClient
	httpClient      *http.Client // 仅 ClaudeClient 使用
}

// NewLLMRouter 创建多模型路由器
func NewLLMRouter(llmCfg *config.AgentLLMConfig) (*LLMRouter, error) {
	router := &LLMRouter{
		defaultProvider: llmCfg.DefaultProvider,
		providers:       make(map[string]*config.LLMProvider),
		clients:         make(map[string]LLMClient),
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}

	for name, provider := range llmCfg.Providers {
		if provider.APIKey == "" {
			slog.Warn("Skipping LLM provider (no API key)", "provider", name)
			continue
		}

		router.providers[name] = &provider

		// 根据provider名称选择客户端实现
		var client LLMClient
		switch name {
		case "claude":
			// Claude 保留手搓 Anthropic API 适配（tRPC 无 VariantClaude）
			client = &ClaudeClient{httpClient: router.httpClient}
		default:
			// OpenAI/DeepSeek/Qwen/Hunyuan 使用 tRPC model/openai
			trpcModel := newTRPCModel(name, &provider)
			client = &OpenAICompatibleClient{model: trpcModel}
		}

		router.clients[name] = client
		slog.Info("  LLM Provider registered", "name", name, "model", provider.Model, "base_url", provider.BaseURL)
	}

	if len(router.clients) == 0 {
		return nil, fmt.Errorf("no valid LLM providers configured (need at least one API key)")
	}

	return router, nil
}

// newTRPCModel 根据provider名称创建 tRPC model/openai 实例
func newTRPCModel(providerName string, opts *config.LLMProvider) model.Model {
	variant := inferVariant(providerName, opts.BaseURL)

	m := openai.New(opts.Model,
		openai.WithAPIKey(opts.APIKey),
		openai.WithBaseURL(opts.BaseURL),
		openai.WithVariant(variant),
	)

	slog.Info("tRPC model created",
		"provider", providerName,
		"model", opts.Model,
		"variant", variant,
	)
	return m
}

// inferVariant 根据provider名称或BaseURL推断模型变体
func inferVariant(providerName, baseURL string) openai.Variant {
	switch strings.ToLower(providerName) {
	case "deepseek":
		return openai.VariantDeepSeek
	case "qwen":
		return openai.VariantQwen
	case "hunyuan":
		return openai.VariantHunyuan
	default:
		// 也通过 BaseURL 推断（tRPC openai 包内部也有此逻辑，这里提前设置）
		return openai.VariantOpenAI
	}
}

// Chat 同步对话（自动路由到默认模型）
func (r *LLMRouter) Chat(ctx context.Context, messages []ChatMessage) (string, *dto.TokenUsageDTO, error) {
	return r.ChatWithProvider(ctx, messages, r.defaultProvider)
}

// ChatWithProvider 指定Provider进行对话
func (r *LLMRouter) ChatWithProvider(ctx context.Context, messages []ChatMessage, providerName string) (string, *dto.TokenUsageDTO, error) {
	r.mu.RLock()
	provider, ok := r.providers[providerName]
	client, hasClient := r.clients[providerName]
	r.mu.RUnlock()

	if !ok || !hasClient {
		return "", nil, fmt.Errorf("LLM provider '%s' not found", providerName)
	}

	reply, usage, err := client.Chat(ctx, messages, provider)
	if err != nil {
		return "", nil, fmt.Errorf("%s chat failed: %w", providerName, err)
	}

	return reply, usage, nil
}

// ChatStream 流式对话（SSE输出，对标 event.StreamingEvents）
func (r *LLMRouter) ChatStream(ctx context.Context, messages []ChatMessage) (<-chan streamChunk, error) {
	return r.ChatStreamWithProvider(ctx, messages, r.defaultProvider)
}

func (r *LLMRouter) ChatStreamWithProvider(ctx context.Context, messages []ChatMessage, providerName string) (<-chan streamChunk, error) {
	r.mu.RLock()
	provider, ok := r.providers[providerName]
	client, hasClient := r.clients[providerName]
	r.mu.RUnlock()

	if !ok || !hasClient {
		return nil, fmt.Errorf("LLM provider '%s' not found", providerName)
	}

	ch, err := client.ChatStream(ctx, messages, provider)
	if err != nil {
		return nil, fmt.Errorf("%s stream failed: %w", providerName, err)
	}

	return ch, nil
}

// GetCurrentModel 获取当前默认模型名
func (r *LLMRouter) GetCurrentModel() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if p, ok := r.providers[r.defaultProvider]; ok {
		return p.Model
	}
	return "unknown"
}

// GetDefaultTRPCModel 获取默认 provider 的 tRPC model.Model 实例
// 用于创建 tRPC llmagent + runner
func (r *LLMRouter) GetDefaultTRPCModel() model.Model {
	r.mu.RLock()
	defer r.mu.RUnlock()
	client, ok := r.clients[r.defaultProvider]
	if !ok {
		return nil
	}
	if oc, ok := client.(*OpenAICompatibleClient); ok {
		return oc.model
	}
	return nil // ClaudeClient 无 tRPC model
}

// GetAvailableProviders 获取所有可用Provider列表
func (r *LLMRouter) GetAvailableProviders() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// Close 关闭所有连接
func (r *LLMRouter) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, client := range r.clients {
		client.Close()
	}
}

// ========== OpenAI 兼容客户端（基于 tRPC-Agent-Go model/openai） ==========
// 替换原手搓 HTTP 实现，使用 tRPC model/openai 的 GenerateContent API
// 支持: OpenAI GPT / DeepSeek / Qwen / Hunyuan

type OpenAICompatibleClient struct {
	model model.Model // tRPC model/openai 实例
}

// toModelMessages 将 ChatMessage 转换为 tRPC model.Message
func toModelMessages(messages []ChatMessage) []model.Message {
	msgs := make([]model.Message, len(messages))
	for i, msg := range messages {
		msgs[i] = model.Message{
			Role:    model.Role(msg.Role),
			Content: msg.Content,
		}
	}
	return msgs
}

// buildGenerationConfig 从 LLMProvider 配置构建 GenerationConfig
func buildGenerationConfig(opts *config.LLMProvider, stream bool) model.GenerationConfig {
	cfg := model.GenerationConfig{
		Stream: stream,
	}
	if opts.Temperature > 0 {
		cfg.Temperature = &opts.Temperature
	}
	if opts.MaxTokens > 0 {
		cfg.MaxTokens = &opts.MaxTokens
	}
	return cfg
}

func (c *OpenAICompatibleClient) Chat(ctx context.Context, messages []ChatMessage, opts *config.LLMProvider) (string, *dto.TokenUsageDTO, error) {
	req := &model.Request{
		Messages:         toModelMessages(messages),
		GenerationConfig: buildGenerationConfig(opts, false),
	}

	respCh, err := c.model.GenerateContent(ctx, req)
	if err != nil {
		return "", nil, fmt.Errorf("tRPC generate content: %w", err)
	}

	// 收集完整响应（GenerateContent 返回 channel，非流式时只有一个响应）
	var content strings.Builder
	var usage *dto.TokenUsageDTO

	for resp := range respCh {
		if len(resp.Choices) > 0 {
			content.WriteString(resp.Choices[0].Delta.Content)
		}
		if resp.Usage != nil {
			usage = &dto.TokenUsageDTO{
				PromptTokens:     resp.Usage.PromptTokens,
				CompletionTokens: resp.Usage.CompletionTokens,
				TotalTokens:      resp.Usage.TotalTokens,
			}
		}
	}

	if content.Len() == 0 {
		return "", nil, fmt.Errorf("empty response from model")
	}

	return content.String(), usage, nil
}

func (c *OpenAICompatibleClient) ChatStream(ctx context.Context, messages []ChatMessage, opts *config.LLMProvider) (<-chan streamChunk, error) {
	req := &model.Request{
		Messages:         toModelMessages(messages),
		GenerationConfig: buildGenerationConfig(opts, true),
	}

	respCh, err := c.model.GenerateContent(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("tRPC generate content: %w", err)
	}

	// 转换 tRPC model.Response → aurora streamChunk
	ch := make(chan streamChunk, 50)

	go func() {
		defer close(ch)
		defer recoverPanic("trpc_openai_stream")

		for resp := range respCh {
			if len(resp.Choices) == 0 {
				continue
			}

			choice := resp.Choices[0]
			chunk := newStreamChunk(
				resp.ID,
				resp.Object,
				resp.Model,
				resp.Created,
				choice.Delta.Content,
				choice.FinishReason,
			)
			ch <- chunk
		}
	}()

	return ch, nil
}

func (c *OpenAICompatibleClient) Close() {} // tRPC model 无需显式关闭

// ========== Claude 客户端 (Anthropic API，保留手搓实现) ==========
// tRPC-Agent-Go 无 VariantClaude，Claude 保留手搓 Anthropic API 适配

type ClaudeClient struct {
	httpClient *http.Client
}

func (c *ClaudeClient) Chat(ctx context.Context, messages []ChatMessage, opts *config.LLMProvider) (string, *dto.TokenUsageDTO, error) {
	systemPrompt := ""
	anthropicMsgs := make([]anthropicMsg, 0, len(messages))

	for _, msg := range messages {
		switch msg.Role {
		case "system":
			systemPrompt = msg.Content
		case "user":
			anthropicMsgs = append(anthropicMsgs, anthropicMsg{Role: "user", Content: msg.Content})
		case "assistant":
			anthropicMsgs = append(anthropicMsgs, anthropicMsg{Role: "assistant", Content: msg.Content})
		}
	}

	reqBody := anthropicRequest{
		Model:     opts.Model,
		MaxTokens: opts.MaxTokens,
		System:    systemPrompt,
		Messages:  anthropicMsgs,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	url := strings.TrimRight(opts.BaseURL, "/") + "/messages"

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", opts.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", nil, fmt.Errorf("claude request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", nil, fmt.Errorf("claude error %d: %s", resp.StatusCode, string(body))
	}

	var result anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", nil, fmt.Errorf("decode claude response: %w", err)
	}

	content := ""
	for _, block := range result.Content {
		if block.Type == "text" {
			content += block.Text
		}
	}

	usage := &dto.TokenUsageDTO{
		PromptTokens:     result.Usage.InputTokens,
		CompletionTokens: result.Usage.OutputTokens,
		TotalTokens:      result.Usage.InputTokens + result.Usage.OutputTokens,
	}

	return content, usage, nil
}

func (c *ClaudeClient) ChatStream(ctx context.Context, messages []ChatMessage, opts *config.LLMProvider) (<-chan streamChunk, error) {
	// Claude流式API暂不实现，降级为同步+模拟chunk
	reply, _, err := c.Chat(ctx, messages, opts)
	if err != nil {
		return nil, err
	}

	ch := make(chan streamChunk, 1)
	go func() {
		defer close(ch)
		finishReason := "end_turn"
		ch <- newStreamChunk("", "chat.completion.chunk", opts.Model, time.Now().Unix(), reply, &finishReason)
	}()
	return ch, nil
}

func (c *ClaudeClient) Close() {}
