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

// @Summary SSE流式AI对话
// @Tags AI Agent
// @Description 通过SSE流式输出或同步模式进行AI对话
// @Accept json
// @Produce json
// @Param message query string false "对话消息"
// @Param sessionId query string false "会话ID"
// @Param mode query string false "对话模式"
// @Param stream query bool false "是否流式输出"
// @Success 200 {object} object "成功返回AI响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 500 {object} object "服务器内部错误"
// @Router /api/agent/chat [get]
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
		fmt.Fprintf(c.Writer, "data: %s\n\n", data)
		flusher.Flush()

		if chunk.Done {
			break
		}
	}
}

// @Summary AI写作助手
// @Tags AI Agent
// @Description 使用AI生成文章/内容
// @Accept json
// @Produce json
// @Param request body dto.WriteRequest true "写作请求"
// @Success 200 {object} object "成功返回写作结果"
// @Failure 400 {object} object "请求参数错误"
// @Failure 500 {object} object "服务器内部错误"
// @Router /api/agent/write [post]
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

// @Summary AI语义搜索
// @Tags AI Agent
// @Description 使用AI进行语义搜索
// @Accept json
// @Produce json
// @Param request body dto.SearchRequest true "搜索请求"
// @Success 200 {object} object "成功返回搜索结果"
// @Failure 400 {object} object "请求参数错误"
// @Failure 500 {object} object "服务器内部错误"
// @Router /api/agent/search [post]
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

// @Summary 数据分析+AI洞察
// @Tags AI Agent
// @Description 使用AI进行数据分析并获取洞察结果
// @Accept json
// @Produce json
// @Param request body dto.AnalyzeRequest true "分析请求"
// @Success 200 {object} object "成功返回分析结果"
// @Failure 400 {object} object "请求参数错误"
// @Failure 500 {object} object "服务器内部错误"
// @Router /api/agent/analyze [post]
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// @Summary AI对话会话列表
// @Tags AI Agent
// @Description 获取用户的AI对话会话列表
// @Accept json
// @Produce json
// @Success 200 {object} object "成功返回会话列表"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Router /api/agent/sessions [get]
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
