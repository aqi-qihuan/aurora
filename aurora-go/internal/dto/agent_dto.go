package dto

import "time"

// ========== Agent 请求/响应 DTO ==========

// ChatRequest AI对话请求
type ChatRequest struct {
	Message   string            `json:"message" binding:"required" example:"写一篇关于Go语言的文章"`     // 用户消息
	SessionID string            `json:"sessionId" example:"sess_abc123"`                          // 会话ID（为空则新建）
	Mode      string            `json:"mode" enum:"chat,write,search,analyze" example:"chat"`       // 对话模式
	Stream    bool              `json:"stream" example:"true"`                                       // 是否SSE流式输出
	Context   map[string]interface{} `json:"context,omitempty"`                                    // 上下文信息（文章ID、搜索关键词等）
}

// ChatResponse AI对话响应
type ChatResponse struct {
	SessionID  string          `json:"sessionId"`
	Reply      string          `json:"reply"`                // 完整回复文本
	Content    string          `json:"content,omitempty"`     // 增量内容(流式)
	Done       bool            `json:"done"`                 // 是否结束
	TokenUsage TokenUsageDTO   `json:"tokenUsage,omitempty"`
	Model      string          `json:"model"`                // 使用的模型
	Timestamp  int64           `json:"timestamp"`
}

// WriteRequest AI写作请求
type WriteRequest struct {
	Action    string `json:"action" binding:"required" enum:"generate,continue,polish,summarize,keywords,seo" example:"generate"`
	Topic     string `json:"topic,omitempty" example:"Go语言并发编程实战"`
	Content   string `json:"content,omitempty" example:"现有文章内容..."`
	ArticleID uint   `json:"articleId,omitempty"`
	Style     string `json:"style,omitempty" enum:"professional,casual,technical" example:"professional"`
	Length    string `json:"length,omitempty" enum:"short,medium,long" example:"medium"`
}

// WriteResponse AI写作响应
type WriteResponse struct {
	Title     string `json:"title,omitempty"`
	Content   string `json:"content"`
	Summary   string `json:"summary,omitempty"`
	Keywords  []string `json:"keywords,omitempty"`
	WordCount int    `json:"wordCount"`
	Action    string `json:"action"`
}

// SearchRequest AI语义搜索请求
type SearchRequest struct {
	Query      string `json:"query" binding:"required" example:"Go语言如何实现高并发"`
	PageNum    int    `json:"pageNum" example:"1"`
	PageSize   int    `json:"pageSize" example:"10"`
	UseRAG     bool   `json:"useRag" example:"true"`        // 是否启用RAG增强
	TopK       int    `json:"topK" example:"5"`            // RAG检索数量
}

// SearchResponse AI语义搜索响应
type SearchResponse struct {
	Results []AgentSearchResult `json:"results"`
	Answer  string             `json:"answer,omitempty"` // AI生成的直接答案
	Total   int64              `json:"total"`
	Query   string             `json:"query"`
	Sources []string           `json:"sources,omitempty"` // 引用来源
}

// AgentSearchResult 单条AI搜索结果
type AgentSearchResult struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	Relevance   float64 `json:"relevance"`   // 相关度分数 0-1
	Highlight   string `json:"highlight"`    // 高亮片段
	URL         string `json:"url"`
	PublishedAt string `json:"publishedAt"`
}

// AnalyzeRequest 数据分析请求
type AnalyzeRequest struct {
	Type      string `json:"type" binding:"required" enum:"traffic,content,user,seo,trend" example:"traffic"`
	DateRange string `json:"dateRange,omitempty" example:"7d"`
	Metric    string `json:"metric,omitempty" example:"pv,uv,bounce_rate"`
}

// AnalyzeResponse 数据分析响应
type AnalyzeResponse struct {
	Type        string                 `json:"type"`
	Data        map[string]interface{} `json:"data"`
	Insights    []string               `json:"insights"`     // AI生成的洞察建议
	Charts      []ChartConfig          `json:"charts,omitempty"` // 图表配置
	Summary     string                 `json:"summary"`      // 自然语言总结
}

// ChartConfig 图表配置（用于前端渲染）
type ChartConfig struct {
	Type   string                 `json:"type"`   // line/bar/pie/radar
	Title  string                 `json:"title"`
	Data   map[string]interface{} `json:"data"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// TokenUsageDTO Token使用统计
type TokenUsageDTO struct {
	PromptTokens     int `json:"promptTokens"`
	CompletionTokens int `json:"completionTokens"`
	TotalTokens      int `json:"totalTokens"`
}

// SessionInfo 会话信息
type SessionInfo struct {
	ID           string    `json:"id"`
	UserID       uint      `json:"userId"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	MessageCount int       `json:"messageCount"`
}

// ========== 审核相关 DTO (P0-11) ==========

// ModerationDTO 内容审核响应
type ModerationDTO struct {
	Passed      bool           `json:"passed"`
	Action      string         `json:"action"`
	Score       int            `json:"score"`
	Violations  []ViolationDTO `json:"violations"`
	DurationMs  int64          `json:"durationMs"`
}

// ViolationDTO 违规记录DTO
type ViolationDTO struct {
	RuleID      string `json:"ruleId"`
	RuleName    string `json:"ruleName"`
	Category    string `json:"category"`
	Severity    int    `json:"severity"`
	MatchedText string `json:"matchedText"`
	Action      string `json:"action"`
	Message     string `json:"message"`
}

// CommentReviewDTO 评论审核请求/响应(Handler用)
type CommentReviewDTO struct {
	Passed      bool              `json:"passed"`
	RiskLevel   string            `json:"riskLevel"`
	Score       float64           `json:"score"`
	Category    string            `json:"category"`
	Reason      string            `json:"reason"`
	Suggestions []string          `json:"suggestions"`
	AutoReply   *string           `json:"autoReply,omitempty"`
	Sentiment   *SentimentDTO     `json:"sentiment,omitempty"`
}

// SentimentDTO 情感分析DTO
type SentimentDTO struct {
	Label      string   `json:"label"`
	Confidence float64  `json:"confidence"`
	Emotions   []string `json:"emotions"`
	Summary    string   `json:"summary"`
}

// WorkflowExecutionDTO 工作流执行响应DTO
type WorkflowExecutionDTO struct {
	ContextID  string                 `json:"contextId"`
	WorkflowID string                 `json:"workflowId"`
	Status     string                 `json:"status"`
	Input      map[string]interface{} `json:"input,omitempty"`
	Output     map[string]interface{} `json:"output,omitempty"`
	NodeResults map[string]interface{} `json:"nodeResults,omitempty"`
	Errors     []string               `json:"errors,omitempty"`
	DurationMs int64                  `json:"durationMs"`
}
