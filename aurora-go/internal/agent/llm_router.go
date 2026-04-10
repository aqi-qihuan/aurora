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
)

// ========== LLM 多模型路由器 ==========
// 对标 tRPC model/openai + model/deepseek (配置即插, 支持VariantDeepSeek)
// 支持: OpenAI GPT / DeepSeek / 阿里通义千问(Qwen) / Anthropic Claude

// ChatMessage 对话消息（对标 OpenAI ChatCompletion message format）
type ChatMessage struct {
	Role    string `json:"role"`    // system/user/assistant/tool
	Content string `json:"content"` // 文本内容
	Name    string `json:"name,omitempty"`
}

// chatCompletionRequest OpenAI兼容的请求格式
type chatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

// chatCompletionResponse OpenAI兼容的响应格式
type chatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// streamChunk SSE流式数据块
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

// anthropicRequest Claude专用请求格式
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

// anthropicResponse Claude响应格式
type anthropicResponse struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Role  string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// LLMClient 单个LLM提供商客户端接口
type LLMClient interface {
	Chat(ctx context.Context, messages []ChatMessage, opts *config.LLMProvider) (string, *dto.TokenUsageDTO, error)
	ChatStream(ctx context.Context, messages []ChatMessage, opts *config.LLMProvider) (<-chan streamChunk, error)
	Close()
}

// LLMRouter 多模型路由器（全局单例）
type LLMRouter struct {
	mu            sync.RWMutex
	defaultProvider string
	providers     map[string]*config.LLMProvider
	clients       map[string]LLMClient
	httpClient    *http.Client
}

// NewLLMRouter 创建多模型路由器
func NewLLMRouter(llmCfg *config.AgentLLMConfig) (*LLMRouter, error) {
	router := &LLMRouter{
		defaultProvider: llmCfg.DefaultProvider,
		providers:      make(map[string]*config.LLMProvider),
		clients:        make(map[string]LLMClient),
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
			client = &ClaudeClient{httpClient: router.httpClient}
		default:
			client = &OpenAICompatibleClient{httpClient: router.httpClient}
		}

		router.clients[name] = client
		slog.Info("  LLM Provider registered", "name", name, "model", provider.Model, "base_url", provider.BaseURL)
	}

	if len(router.clients) == 0 {
		return nil, fmt.Errorf("no valid LLM providers configured (need at least one API key)")
	}

	return router, nil
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

// ========== OpenAI 兼容客户端 (GPT/DeepSeek/Qwen) ==========

type OpenAICompatibleClient struct {
	httpClient *http.Client
}

func (c *OpenAICompatibleClient) Chat(ctx context.Context, messages []ChatMessage, opts *config.LLMProvider) (string, *dto.TokenUsageDTO, error) {
	reqBody := chatCompletionRequest{
		Model:       opts.Model,
		Messages:    messages,
		Temperature: opts.Temperature,
		MaxTokens:   opts.MaxTokens,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	url := strings.TrimRight(opts.BaseURL, "/") + "/chat/completions"

	resp, err := c.doRequest(ctx, url, opts.APIKey, bodyBytes)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	var result chatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", nil, fmt.Errorf("decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", nil, fmt.Errorf("no choices in response")
	}

	content := result.Choices[0].Message.Content
	usage := &dto.TokenUsageDTO{
		PromptTokens:     result.Usage.PromptTokens,
		CompletionTokens: result.Usage.CompletionTokens,
		TotalTokens:      result.Usage.TotalTokens,
	}

	return content, usage, nil
}

func (c *OpenAICompatibleClient) ChatStream(ctx context.Context, messages []ChatMessage, opts *config.LLMProvider) (<-chan streamChunk, error) {
	reqBody := chatCompletionRequest{
		Model:       opts.Model,
		Messages:    messages,
		Temperature: opts.Temperature,
		MaxTokens:   opts.MaxTokens,
		Stream:      true,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	url := strings.TrimRight(opts.BaseURL, "/") + "/chat/completions"

	resp, err := c.doRequest(ctx, url, opts.APIKey, bodyBytes)
	if err != nil {
		return nil, err
	}

	ch := make(chan streamChunk, 50)

	go func() {
		defer close(ch)
		defer resp.Body.Close()
		defer recoverPanic("openai_stream")

		scanner := newSSEScanner(resp.Body)
		for scanner.Scan() {
			data := scanner.Data()
			if data == "[DONE]" {
				break
			}

			var chunk streamChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}
			if len(chunk.Choices) > 0 {
				ch <- chunk
			}
		}
	}()

	return ch, nil
}

func (c *OpenAICompatibleClient) Close() {} // http.Client 无需显式关闭

// ========== Claude 客户端 (Anthropic API) ==========

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
		ch <- streamChunk{
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
					}{Role: "assistant", Content: reply},
					FinishReason: &finishReason,
				},
			},
		}
	}()
	return ch, nil
}

func (c *ClaudeClient) Close() {}

// ========== HTTP 辅助方法 ==========

func (c *OpenAICompatibleClient) doRequest(ctx context.Context, url, apiKey string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("llm request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("llm error %d: %s", resp.StatusCode, string(b))
	}

	return resp, nil
}
