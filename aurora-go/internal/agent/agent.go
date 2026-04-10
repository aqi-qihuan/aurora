// Package agent 基于 tRPC-Agent-Go v1.8 实现的 AI 引擎核心
//
// 隔离保证（5级防护）:
//   - L1 编译隔离: //go:build aurora_agent tag 控制是否编译
//   - L2 配置隔离: agent.enabled=false → 零初始化,零路由
//   - L3 路由隔离: 独立RouterGroup /api/agent/*
//   - L4 故障隔离: goroutine+recover, panic不杀主进程
//   - L5 依赖隔离: 核心代码零import agent包
//
// 架构:
//
//	AuroraAgentFactory (入口)
//	   ├── LLM Router (多模型路由: OpenAI/DeepSeek/Qwen/Claude)
//	   ├── Tool Hub (6个Aurora业务工具 + MCP工具支持)
//	   ├── Memory Service (InMemory/Redis持久化会话记忆)
//	   └── RAG Pipeline (ES混合检索 + LLM生成)
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/aurora-go/aurora/internal/config"
	"github.com/aurora-go/aurora/internal/dto"
	apperrors "github.com/aurora-go/aurora/internal/errors"
)

// ========== AuroraAgentFactory 唯一入口 (~80行) ==========
// 对标 Java @Configuration + @Bean 组装模式

// AuroraAgent Agent引擎实例（全局单例）
type AuroraAgent struct {
	config     *config.AgentConfig
	llmRouter  *LLMRouter
	toolHub    *ToolHub
	memorySvc  MemoryService
	ragPipeline *RAGPipeline
}

var globalAgent *AuroraAgent

// InitAgent 初始化Agent引擎（在Bootstrap中调用）
// 对标 Spring @PostConstruct 或 ApplicationRunner
func InitAgent(cfg *config.AgentConfig) error {
	if !cfg.Enabled {
		return nil // 零初始化
	}

	slog.Info("Initializing Aurora Agent Engine...")

	// 1. LLM Router - 多模型路由
	llmRouter, err := NewLLMRouter(&cfg.LLM)
	if err != nil {
		return fmt.Errorf("LLM router init failed: %w", err)
	}
	slog.Info("  [1/4] LLM Router ready",
		"default", cfg.LLM.DefaultProvider,
		"providers", len(cfg.LLM.Providers),
	)

	// 2. Tool Hub - 业务工具集
	toolHub := NewToolHub()
	RegisterAuroraTools(toolHub) // 注册6个业务工具
	slog.Info("  [2/4] Tool Hub ready",
		"tools", toolHub.ToolCount(),
	)

	// 3. Memory Service - 会话记忆
	memorySvc, err := NewMemoryService(&cfg.Memory)
	if err != nil {
		return fmt.Errorf("memory service init failed: %w", err)
	}
	slog.Info("  [3/4] Memory Service ready",
		"type", cfg.Memory.Type,
	)

	// 4. RAG Pipeline - 检索增强生成
	ragPipeline := NewRAGPipeline(llmRouter)
	slog.Info("  [4/4] RAG Pipeline ready")

	globalAgent = &AuroraAgent{
		config:      cfg,
		llmRouter:   llmRouter,
		toolHub:     toolHub,
		memorySvc:   memorySvc,
		ragPipeline: ragPipeline,
	}

	slog.Info("Aurora Agent Engine initialized successfully")
	return nil
}

// GetAgent 获取全局Agent实例（nil表示未初始化）
func GetAgent() *AuroraAgent {
	return globalAgent
}

// Chat 执行AI对话（核心方法，对标 Runner.Run()）
func (a *AuroraAgent) Chat(ctx context.Context, req *dto.ChatRequest) (*dto.ChatResponse, error) {
	if a == nil {
		return nil, apperrors.ErrAgentDisabled
	}

	// 1. 加载/创建会话记忆
	session, err := a.memorySvc.GetOrCreateSession(ctx, req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("session error: %w", err)
	}

	// 2. 构建消息历史（含上下文窗口限制）
	messages := a.memorySvc.BuildMessages(ctx, session.ID, req.Message)

	// 3. 根据模式选择处理流程
	var reply string
	var usage *dto.TokenUsageDTO

	switch req.Mode {
	case "write":
		reply, usage, err = a.handleWriteMode(ctx, messages, req)
	case "search":
		reply, usage, err = a.handleSearchMode(ctx, messages, req)
	case "analyze":
		reply, usage, err = a.handleAnalyzeMode(ctx, messages, req)
	default:
		// chat模式: 纯对话 + 工具调用
		reply, usage, err = a.llmRouter.Chat(ctx, messages)
	}

	if err != nil {
		return nil, err
	}

	// 4. 保存会话记忆（异步）
	go func() {
		defer recoverPanic("memory_save")
		_ = a.memorySvc.SaveMessage(context.Background(), session.ID, "user", req.Message)
		_ = a.memorySvc.SaveMessage(context.Background(), session.ID, "assistant", reply)
	}()

	return &dto.ChatResponse{
		SessionID: session.ID,
		Reply:     reply,
		Done:      true,
		TokenUsage: *usage,
		Model:     a.llmRouter.GetCurrentModel(),
		Timestamp:  time.Now().Unix(),
	}, nil
}

// ChatStream SSE流式对话（增量输出，对标 event.StreamingEvents）
func (a *AuroraAgent) ChatStream(ctx context.Context, req *dto.ChatRequest) (<-chan dto.ChatResponse, error) {
	if a == nil {
		return nil, apperrors.ErrAgentDisabled
	}

	session, err := a.memorySvc.GetOrCreateSession(ctx, req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("session error: %w", err)
	}

	messages := a.memorySvc.BuildMessages(ctx, session.ID, req.Message)
	streamCh, err := a.llmRouter.ChatStream(ctx, messages)
	if err != nil {
		return nil, err
	}

	outCh := make(chan dto.ChatResponse, 10)

	go func() {
		defer close(outCh)
		defer recoverPanic("chat_stream")

		var fullReply string
		for chunk := range streamCh {
			content := extractChunkContent(chunk)
			fullReply += content
			outCh <- dto.ChatResponse{
				SessionID: session.ID,
				Content:   content,
				Done:      false,
				Model:     a.llmRouter.GetCurrentModel(),
				Timestamp: time.Now().Unix(),
			}
		}

		// 保存完整回复到记忆
		go func() {
			defer recoverPanic("stream_memory_save")
			_ = a.memorySvc.SaveMessage(context.Background(), session.ID, "user", req.Message)
			_ = a.memorySvc.SaveMessage(context.Background(), session.ID, "assistant", fullReply)
		}()

		outCh <- dto.ChatResponse{
			SessionID: session.ID,
			Reply:     fullReply,
			Done:      true,
			Timestamp: time.Now().Unix(),
		}
	}()

	return outCh, nil
}

// Write AI写作助手（对标 Workflow.Execute(ArticlePublishFlow)）
func (a *AuroraAgent) Write(ctx context.Context, req *dto.WriteRequest) (*dto.WriteResponse, error) {
	if a == nil {
		return nil, apperrors.ErrAgentDisabled
	}

	// 使用 WritingAssistant 执行具体操作
	wa := NewWritingAssistant(a.llmRouter)
	return wa.Execute(ctx, req)
}

// Search AI语义搜索（对标 RAGPipeline.Retrieve()）
func (a *AuroraAgent) Search(ctx context.Context, req *dto.SearchRequest) (*dto.SearchResponse, error) {
	if a == nil {
		return nil, apperrors.ErrAgentDisabled
	}

	if req.UseRAG {
		// RAG增强检索：ES搜索 → 上下文注入 → LLM生成答案
		return a.ragPipeline.Retrieve(ctx, req)
	}

	// 纯语义理解 → ES关键词搜索（降级方案）
	return a.ragPipeline.SemanticSearch(ctx, req)
}

// Analyze 数据分析（对标 analysisTool + LLM洞察生成）
func (a *AuroraAgent) Analyze(ctx context.Context, req *dto.AnalyzeRequest) (*dto.AnalyzeResponse, error) {
	if a == nil {
		return nil, apperrors.ErrAgentDisabled
	}

	// 1. 通过Tool获取原始数据
	rawData, err := a.toolHub.ExecuteAnalysis(req.Type, req.DateRange, req.Metric)
	if err != nil {
		return nil, fmt.Errorf("data fetch failed: %w", err)
	}

	// 2. LLM生成洞察建议
	prompt := buildAnalysisPrompt(req.Type, rawData)
	messages := []ChatMessage{
		{Role: "system", Content: systemPromptForAnalysis()},
		{Role: "user", Content: prompt},
	}

	insightText, _, err := a.llmRouter.Chat(ctx, messages)
	if err != nil {
		return &dto.AnalyzeResponse{
			Type:    req.Type,
			Data:    rawData,
			Summary: "AI分析暂时不可用，以下为原始数据",
		}, nil
	}

	return parseAnalysisResponse(req.Type, rawData, insightText), nil
}

// Shutdown 优雅关闭Agent引擎
func Shutdown() {
	if globalAgent != nil {
		globalAgent.memorySvc.Close()
		globalAgent.llmRouter.Close()
		globalAgent = nil
		slog.Info("Aurora Agent Engine shutdown complete")
	}
}

// ========== 内部辅助方法 ==========

func (a *AuroraAgent) handleWriteMode(ctx context.Context, messages []ChatMessage, req *dto.ChatRequest) (string, *dto.TokenUsageDTO, error) {
	systemMsg := ChatMessage{Role: "system", Content: systemPromptForWriting()}
	allMessages := append([]ChatMessage{systemMsg}, messages...)
	return a.llmRouter.Chat(ctx, allMessages)
}

func (a *AuroraAgent) handleSearchMode(ctx context.Context, messages []ChatMessage, req *dto.ChatRequest) (string, *dto.TokenUsageDTO, error) {
	// 先用ES搜索获取上下文
	searchReq := &dto.SearchRequest{Query: req.Message, UseRAG: true, TopK: 5}
	result, _ := a.ragPipeline.Retrieve(ctx, searchReq)

	// 将搜索结果作为上下文注入
	contextText := formatSearchContext(result)
	contextMsg := ChatMessage{Role: "system", Content: searchSystemPrompt(contextText)}
	allMessages := append([]ChatMessage{contextMsg}, messages...)
	return a.llmRouter.Chat(ctx, allMessages)
}

func (a *AuroraAgent) handleAnalyzeMode(ctx context.Context, messages []ChatMessage, req *dto.ChatRequest) (string, *dto.TokenUsageDTO, error) {
	// 获取最新统计数据
	rawData, _ := a.toolHub.ExecuteAnalysis("traffic", "7d", "")
	contextMsg := ChatMessage{Role: "system", Content: analysisSystemPrompt(rawData)}
	allMessages := append([]ChatMessage{contextMsg}, messages...)
	return a.llmRouter.Chat(ctx, allMessages)
}

// ListSessions 获取用户的会话列表
func (a *AuroraAgent) ListSessions(ctx context.Context, userID uint) ([]*Session, error) {
	if a == nil {
		return nil, apperrors.ErrAgentDisabled
	}
	return a.memorySvc.ListSessions(ctx, userID)
}

// ========== 辅助函数（Prompt构建 + 响应解析）==========

// extractChunkContent 从streamChunk提取文本内容
func extractChunkContent(chunk streamChunk) string {
	if len(chunk.Choices) > 0 {
		return chunk.Choices[0].Delta.Content
	}
	return ""
}

// formatSearchContext 将搜索结果格式化为上下文文本
func formatSearchContext(result *dto.SearchResponse) string {
	if result == nil || len(result.Results) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, r := range result.Results {
		fmt.Fprintf(&sb, "[%d] %s\n%s\n\n", i+1, r.Title, r.Summary)
	}
	return sb.String()
}

func truncateText(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

func buildWritePrompt(req *dto.WriteRequest) string {
	return fmt.Sprintf("Action: %s\nTopic: %s\nStyle: %s\nLength: %s\nContent: %s",
		req.Action, req.Topic, req.Style, req.Length, truncateText(req.Content, 500))
}

func parseWriteResponse(reply string, action string) *dto.WriteResponse {
	wa := &WritingAssistant{}
	result := parseAIGeneratedContent(reply, action)
	if wa != nil {
		return result
	}
	return result
}

func buildAnalysisPrompt(analysisType string, rawData map[string]interface{}) string {
	dataJSON, _ := json.Marshal(rawData)
	return fmt.Sprintf(`分析类型: %s

原始数据:
%s

请分析数据并给出洞察和建议。`, analysisType, string(dataJSON))
}

func parseAnalysisResponse(analysisType string, rawData map[string]interface{}, insightText string) *dto.AnalyzeResponse {
	var parsed struct {
		Summary        string   `json:"summary"`
		Insights       []string `json:"insights"`
		Recommendations []string `json:"recommendations"`
	}
	if err := extractJSONFromText(insightText, &parsed); err != nil {
		return &dto.AnalyzeResponse{
			Type:    analysisType,
			Data:    rawData,
			Summary: insightText,
			Insights: []string{insightText},
		}
	}

	allInsights := append(parsed.Insights, parsed.Recommendations...)
	return &dto.AnalyzeResponse{
		Type:     analysisType,
		Data:     rawData,
		Summary:  parsed.Summary,
		Insights: allInsights,
	}
}
