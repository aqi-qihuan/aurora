package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/aurora-go/aurora/internal/agent"
	"github.com/aurora-go/aurora/internal/dto"
	"github.com/gin-gonic/gin"
)

// ========== AgentHandler AI Agent HTTP端点 ==========
// 对标 tRPC Agent → Gin Bridge (~100行)
// 隔离保证: 独立文件, 仅在 /api/agent/* 路由组注册

type AgentHandler struct{}

func NewAgentHandler() *AgentHandler {
	return &AgentHandler{}
}

// Chat SSE流式AI对话
// GET  /api/agent/chat?message=...&sessionId=...
// POST /api/agent/chat
func (h *AgentHandler) Chat(c *gin.Context) {
	var req dto.ChatRequest

	if c.Request.Method == http.MethodGet {
		// GET请求: query参数模式(适合SSE)
		req.Message = c.Query("message")
		req.SessionID = c.Query("sessionId")
		req.Mode = c.DefaultQuery("mode", "chat")
		req.Stream = true // GET默认流式输出
	} else {
		// POST请求: JSON body模式
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	if req.Message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message is required"})
		return
	}

	a := agent.GetAgent()
	if a == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI Agent is not enabled or not initialized"})
		return
	}

	if req.Stream {
		h.handleStreamChat(c, a, &req)
	} else {
		h.handleSyncChat(c, a, &req)
	}
}

func (h *AgentHandler) handleSyncChat(c *gin.Context, a *agent.AuroraAgent, req *dto.ChatRequest) {
	ctx := context.Background()
	resp, err := a.Chat(ctx, req)
	if err != nil {
		slog.Error("Agent chat failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"message": "AI对话处理失败，请稍后重试",
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AgentHandler) handleStreamChat(c *gin.Context, a *agent.AuroraAgent, req *dto.ChatRequest) {
	ctx := context.Background()

	ch, err := a.ChatStream(ctx, req)
	if err != nil {
		slog.Error("Agent stream chat failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// SSE响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		slog.Error("SSE flusher not supported")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server-Sent Events not supported"})
		return
	}

	for chunk := range ch {
		data, _ := json.Marshal(chunk)
		fmt.Fprintf(c.Writer(), "data: %s\n\n", data)
		flusher.Flush()

		if chunk.Done {
			break
		}
	}
}

// Write AI写作助手
// POST /api/agent/write
func (h *AgentHandler) Write(c *gin.Context) {
	var req dto.WriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	a := agent.GetAgent()
	if a == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI Agent is not enabled"})
		return
	}

	ctx := context.Background()
	resp, err := a.Write(ctx, &req)
	if err != nil {
		slog.Error("Agent write failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Search AI语义搜索
// POST /api/agent/search
func (h *AgentHandler) Search(c *gin.Context) {
	var req dto.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	a := agent.GetAgent()
	if a == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI Agent is not enabled"})
		return
	}

	ctx := context.Background()
	resp, err := a.Search(ctx, &req)
	if err != nil {
		slog.Error("Agent search failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Analyze 数据分析 + AI洞察
// POST /api/agent/analyze
func (h *AgentHandler) Analyze(c *gin.Context) {
	var req dto.AnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	a := agent.GetAgent()
	if a == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI Agent is not enabled"})
		return
	}

	ctx := context.Background()
	resp, err := a.Analyze(ctx, &req)
	if err != nil {
		slog.Error("Agent analyze failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Sessions 会话列表
// GET /api/agent/sessions
func (h *AgentHandler) Sessions(c *gin.Context) {
	a := agent.GetAgent()
	if a == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI Agent is not enabled"})
		return
	}

	// TODO: 从JWT中提取UserID
	userID := uint(0)

	sessions, err := a.ListSessions(context.Background(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sessions)
}
