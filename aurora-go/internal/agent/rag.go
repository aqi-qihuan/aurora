package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/strategy"
)

// ========== RAG 检索增强生成管道 ==========
// 对标 tRPC knowledge 包 (文档检索增强生成内置支持)
// 流程: 用户Query → ES语义检索 → 上下文构建 → LLM生成答案

// RAGPipeline RAG管道实例
type RAGPipeline struct {
	llmRouter *LLMRouter
}

// NewRAGPipeline 创建RAG管道
func NewRAGPipeline(router *LLMRouter) *RAGPipeline {
	return &RAGPipeline{llmRouter: router}
}

// Retrieve 完整RAG检索增强生成
func (p *RAGPipeline) Retrieve(ctx context.Context, req *dto.SearchRequest) (*dto.SearchResponse, error) {
	// Step 1: ES混合检索（关键词 + 语义）
	searchResults, err := p.esHybridSearch(ctx, req)
	if err != nil {
		slog.Warn("RAG ES search failed", "error", err)
		searchResults = []dto.AgentSearchResult{}
	}

	// Step 2: 构建上下文（搜索结果 → 提示词模板）
	contextStr := p.buildRagContext(searchResults)

	// Step 3: LLM基于上下文生成答案
	systemPrompt := fmt.Sprintf(ragSystemPromptTemplate(), len(searchResults), contextStr)

	messages := []ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: req.Query},
	}

	answer, _, err := p.llmRouter.Chat(ctx, messages)
	if err != nil {
		slog.Warn("RAG LLM generation failed, returning raw results", "error", err)
		return p.formatRawResults(req.Query, searchResults), nil
	}

	// Step 4: 收集引用来源
	sources := make([]string, 0, len(searchResults))
	for _, r := range searchResults {
		if !contains(sources, r.URL) {
			sources = append(sources, r.URL)
		}
	}

	return &dto.SearchResponse{
		Results: searchResults,
		Answer:  answer,
		Total:   int64(len(searchResults)),
		Query:   req.Query,
		Sources: sources,
	}, nil
}

// SemanticSearch 纯语义搜索（降级方案，不使用RAG增强）
func (p *RAGPipeline) SemanticSearch(ctx context.Context, req *dto.SearchRequest) (*dto.SearchResponse, error) {
	results, err := p.esHybridSearch(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	return p.formatRawResults(req.Query, results), nil
}

// ========== 内部方法 ==========

func (p *RAGPipeline) esHybridSearch(ctx context.Context, req *dto.SearchRequest) ([]dto.AgentSearchResult, error) {
	// 通过策略模式调用ES/MySQL搜索
	searchCtx := strategy.GetSearchContext()
	if searchCtx == nil {
		return nil, fmt.Errorf("search service unavailable")
	}

	rawResults, total, err := searchCtx.ExecuteSearch(ctx, req.Query, 1, req.TopK)
	if err != nil {
		return nil, err
	}

	// 转换为AgentSearchResult格式
	results := make([]dto.AgentSearchResult, 0, len(rawResults))
	for _, item := range rawResults {
		result := dto.AgentSearchResult{
			ID:      uint(item["id"].(float64)),
			Title:   item["articleTitle"].(string),
			Summary: truncateText(item["articleContent"].(string), 150),
			URL:     fmt.Sprintf("/articles/%d", uint(item["id"].(float64))),
		}
		results = append(results, result)
	}

	_ = total // 保留总数供后续使用

	topK := req.TopK
	if topK <= 0 {
		topK = 10
	}
	if len(results) > topK {
		return results[:topK], nil
	}
	return results, nil
}

func (p *RAGPipeline) buildRagContext(results []dto.AgentSearchResult) string {
	var sb strings.Builder
	for i, r := range results {
		sb.WriteString(fmt.Sprintf("\n[%d] 标题: %s\n", i+1, r.Title))
		sb.WriteString(fmt.Sprintf("    摘要: %s\n", r.Summary))
		if r.Highlight != "" {
			sb.WriteString(fmt.Sprintf("    高亮: %s\n", r.Highlight))
		}
		sb.WriteString(fmt.Sprintf("    来源: %s\n", r.URL))
	}
	return sb.String()
}

func (p *RAGPipeline) formatRawResults(query string, results []dto.AgentSearchResult) *dto.SearchResponse {
	return &dto.SearchResponse{
		Results: results,
		Total:   int64(len(results)),
		Query:   query,
	}
}

// ========== Prompt模板 ==========

func ragSystemPromptTemplate() string {
	return `你是一个专业的博客助手。请根据以下【参考信息】回答用户的问题。

## 要求：
1. 基于提供的参考信息进行回答，不要编造内容
2. 如果参考信息不足，明确说明"根据现有资料无法确定"
3. 回答要自然流畅，不要机械地罗列来源
4. 引用具体信息时标注 [序号]
5. 使用中文回答

## 参考信息 (%d条):
%s`
}

func systemPromptForWriting() string {
	return `你是一位资深的博客写作专家。你的任务是根据用户的请求创作高质量博客文章。

## 写作要求:
1. 内容原创、有价值、有深度
2. 结构清晰，包含引言、正文、总结
3. Markdown格式排版，适当使用标题、列表、代码块
4. 语言专业但不晦涩，适合技术读者
5. 字数控制在合理范围内(通常1000-3000字)

## 输出格式:
- title: 文章标题(吸引人但不过度标题党)
- content: 正文(Markdown格式)
- summary: 一句话摘要(50字以内)
- keywords: 关键标签(3-5个,用逗号分隔)`
}

func systemPromptForAnalysis() string {
	return `你是一位数据分析师。请分析提供的数据并给出有价值的洞察。

## 分析要求:
1. 识别数据中的关键趋势和异常点
2. 给出可操作的建议(而非空泛的描述)
3. 用简洁的语言解释复杂的数据关系
4. 区分事实陈述和推测性建议
5. 输出结构化的JSON格式结果

## 输出格式:
{
  "summary": "一句话总体评价",
  "insights": ["洞察1", "洞察2", ...],
  "recommendations": ["建议1", "建议2", ...],
  "alerts": ["需要关注的风险点"]
}`
}

func analysisSystemPrompt(rawData map[string]interface{}) string {
	dataJSON, _ := json.MarshalIndent(rawData, "", "  ")
	return fmt.Sprintf(`你是一位博客运营数据分析专家。以下是最新的运营统计数据:

## 原始数据 (JSON):
%s

请基于以上数据进行深度分析，识别趋势、发现机会、指出风险。
输出中文自然语言分析结果。`, string(dataJSON))
}

func searchSystemPrompt(contextStr string) string {
	return `你是一位博客搜索助手。用户正在搜索文章。

## 已找到的相关文章:
%s

请基于以上搜索结果，帮助用户找到最相关的信息或直接回答问题。如果搜索结果不足以回答用户问题，请诚实说明。`
}

// ========== 辅助函数 ==========

func calculateRelevance(score float64, highlight string) float64 {
	relevance := score
	if score < 1.0 && score > 0 {
		relevance = min(score*10, 1.0) // 归一化到0-1
	}
	if relevance < 0.3 {
		relevance = 0.3 // 最小相关性
	}
	// 高亮匹配加分
	if strings.Contains(highlight, "<mark>") {
		relevance = min(relevance+0.15, 1.0)
	}
	return relevance
}

func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
