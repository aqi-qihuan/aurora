package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/aurora-go/aurora/internal/dto"
)

// ========== 智能写作助手 ==========
// 基于 LLM Router 实现的博客文章写作AI助手
// 支持: 生成/续写/润色/摘要提取/关键词推荐/SEO优化 6种模式
//
// 对标: tRPC workflow.Execute(ArticlePublishFlow) 的写作环节
// 架构: AuroraAgent.Write() → WritingAssistant.Execute() → LLMRouter.Chat()

// WritingAssistant 写作助手实例（无状态，方法级并发安全）
type WritingAssistant struct {
	llmRouter *LLMRouter
}

// NewWritingAssistant 创建写作助手
func NewWritingAssistant(router *LLMRouter) *WritingAssistant {
	return &WritingAssistant{llmRouter: router}
}

// Execute 执行写作操作（核心入口）
func (w *WritingAssistant) Execute(ctx context.Context, req *dto.WriteRequest) (*dto.WriteResponse, error) {
	switch req.Action {
	case "generate":
		return w.generateArticle(ctx, req)
	case "continue":
		return w.continueArticle(ctx, req)
	case "polish":
		return w.polishArticle(ctx, req)
	case "summarize":
		return w.summarizeArticle(ctx, req)
	case "keywords":
		return w.extractKeywords(ctx, req)
	case "seo":
		return w.optimizeSEO(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported write action: %s", req.Action)
	}
}

// ========== 1. 文章生成 (generate) ==========
// 根据主题/关键词从零生成一篇完整文章

func (w *WritingAssistant) generateArticle(ctx context.Context, req *dto.WriteRequest) (*dto.WriteResponse, error) {
	prompt := fmt.Sprintf(`请根据以下要求创作一篇完整的博客文章:

## 主题/话题:
%s

## 风格要求: %s
## 长度要求: %s

## 输出格式(JSON):
{
  "title": "吸引人的标题",
  "content": "Markdown格式的正文内容(包含适当的标题层级、代码块、列表)",
  "summary": "一句话摘要(50字以内)",
  "keywords": ["标签1", "标签2", "标签3"]
}

## 写作规范:
1. 内容原创、有深度、有实用价值
2. 技术文章需包含代码示例和解释
3. 使用Markdown格式排版
4. 语言专业但不晦涩
5. 字数控制在%s`,
		req.Topic,
		w.resolveStyle(req.Style),
		w.resolveLength(req.Length),
		w.wordCountHint(req.Length),
	)

	messages := []ChatMessage{
		{Role: "system", Content: systemPromptForWriting()},
		{Role: "user", Content: prompt},
	}

	reply, _, err := w.llmRouter.Chat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("article generation failed: %w", err)
	}

	resp := parseAIGeneratedContent(reply, req.Action)
	return resp, nil
}

// ========== 2. 续写 (continue) ==========
// 基于已有文章内容续写下一章节或段落

func (w *WritingAssistant) continueArticle(ctx context.Context, req *dto.WriteRequest) (*dto.WriteResponse, error) {
	if req.Content == "" {
		return nil, fmt.Errorf("content is required for continue action")
	}

	prompt := fmt.Sprintf(`以下是当前已写的内容，请自然地续写接下来的内容。

## 已有内容(%d字):
---
%s
---

## 续写要求:
1. 保持与上文一致的写作风格和语调
2. 自然衔接上一段结尾
3. 深入展开论述，不要重复已有内容
4. 如果是技术文章，增加更多细节和示例
5. 续写字数约%s

## 输出格式:
仅输出续写部分的内容(Markdown格式)，不需要输出标题。`,
		utf8.RuneCountInString(req.Content),
		req.Content,
		w.wordCountHint(req.Length),
	)

	messages := []ChatMessage{
		{Role: "system", Content: systemPromptForWriting() + "\n\n你现在的工作是续写已有内容。"},
		{Role: "user", Content: prompt},
	}

	reply, _, err := w.llmRouter.Chat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("article continuation failed: %w", err)
	}

	return &dto.WriteResponse{
		Title:    "",
		Content:  cleanMarkdown(reply),
		Summary:  "",
		Keywords: nil,
		WordCount: utf8.RuneCountInString(reply),
		Action:   "continue",
	}, nil
}

// ========== 3. 润色 (polish) ==========
// 对已有文章进行语言润色、结构调整、可读性优化

func (w *WritingAssistant) polishArticle(ctx context.Context, req *dto.WriteRequest) (*dto.WriteResponse, error) {
	if req.Content == "" {
		return nil, fmt.Errorf("content is required for polish action")
	}

	prompt := fmt.Sprintf(`请对以下文章进行润色优化:

## 原文(%d字):
---
%s
---

## 润色方向:
1. 修正语法错误和错别字
2. 优化句子结构，提升流畅度
3. 调整段落逻辑顺序
4. 替换过于口语化或啰嗦的表达
5. 保持原文核心观点不变
6. 确保专业术语使用准确

## 输出格式:
直接输出润色后的完整Markdown正文。`,
		utf8.RuneCountInString(req.Content),
		req.Content,
	)

	messages := []ChatMessage{
		{Role: "system", Content: "你是一位专业的中文编辑，擅长技术文章的润色和优化。保持原文风格的同时提升表达质量。"},
		{Role: "user", Content: prompt},
	}

	reply, _, err := w.llmRouter.Chat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("article polish failed: %w", err)
	}

	return &dto.WriteResponse{
		Title:    "",
		Content:  cleanMarkdown(reply),
		Summary:  "已完成文章润色优化",
		Keywords: nil,
		WordCount: utf8.RuneCountInString(reply),
		Action:   "polish",
	}, nil
}

// ========== 4. 摘要提取 (summarize) ==========
// 自动提取文章摘要（支持多长度版本）

func (w *WritingAssistant) summarizeArticle(ctx context.Context, req *dto.WriteRequest) (*dto.WriteResponse, error) {
	if req.Content == "" {
		return nil, fmt.Errorf("content is required for summarize action")
	}

	prompt := fmt.Sprintf(`请为以下文章撰写摘要:

## 文章内容:
%s

## 要求:
1. 提炼核心观点和价值主张
2. 控制在50-100字之间
3. 吸引读者点击阅读全文
4. 不使用"本文介绍了"等模板化开头

## 同时提供:
- short_summary: 一句话摘要(30字以内，用于列表展示)
- long_summary: 详细摘要(80-120字，用于详情页)
- key_points: 3-5个关键要点(数组)

## 输出格式(JSON):
{
  "short_summary": "...",
  "long_summary": "...",
  "key_points": ["要点1", "要点2", ...]
}`,
		req.Content,
	)

	messages := []ChatMessage{
		{Role: "system", Content: "你是一位资深编辑，擅长用精炼的语言概括文章精髓。"},
		{Role: "user", Content: prompt},
	}

	reply, _, err := w.llmRouter.Chat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("summarization failed: %w", err)
	}

	var summaryData struct {
		ShortSummary string   `json:"short_summary"`
		LongSummary  string   `json:"long_summary"`
		KeyPoints    []string `json:"key_points"`
	}

	if err := extractJSONFromText(reply, &summaryData); err != nil {
		slog.Warn("Failed to parse summary JSON, using raw text", "error", err)
		summaryData.LongSummary = reply
		summaryData.ShortSummary = truncateText(reply, 30)
	}

	return &dto.WriteResponse{
		Content:  summaryData.LongSummary,
		Summary:  summaryData.ShortSummary,
		Keywords: summaryData.KeyPoints,
		WordCount: utf8.RuneCountInString(summaryData.LongSummary),
		Action:   "summarize",
	}, nil
}

// ========== 5. 关键词推荐 (keywords) ==========
// 基于内容自动推荐SEO友好的标签/关键词

func (w *WritingAssistant) extractKeywords(ctx context.Context, req *dto.WriteRequest) (*dto.WriteResponse, error) {
	content := req.Content
	if content == "" && req.Topic != "" {
		content = fmt.Sprintf("主题: %s\n请基于此主题推荐相关关键词", req.Topic)
	}
	if content == "" {
		return nil, fmt.Errorf("content or topic is required")
	}

	prompt := fmt.Sprintf(`请为以下内容分析并推荐标签/关键词:

## 内容:
%s

## 分析维度:
1. 核心主题词(2-3个): 直接描述内容的词
2. 技术栈/工具词(2-3个): 涉及的具体技术和工具
3. 长尾关键词(2-3个): 用户可能搜索的相关词
4. 分类建议: 最适合归入的分类名称

## 输出格式(JSON):
{
  "primary": ["核心词1", "核心词2"],
  "technical": ["技术词1", "技术词2"],
  "long_tail": ["长尾词1"],
  "category_suggestion": "推荐分类名",
  "all_keywords": ["合并去重后的所有关键词"]
}`,
		content,
	)

	messages := []ChatMessage{
		{Role: "system", Content: "你是一位SEO专家，擅长从内容中提取高价值的关键词和标签。"},
		{Role: "user", Content: prompt},
	}

	reply, _, err := w.llmRouter.Chat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("keyword extraction failed: %w", err)
	}

	var kwData struct {
		Primary           []string `json:"primary"`
		Technical         []string `json:"technical"`
		LongTail          []string `json:"long_tail"`
		CategorySuggestion string  `json:"category_suggestion"`
		AllKeywords       []string `json:"all_keywords"`
	}

	if err := extractJSONFromText(reply, &kwData); err != nil {
		slog.Warn("Failed to parse keywords JSON", "error", err)
		kwData.AllKeywords = extractKeywordsFallback(reply)
	}

	allKw := kwData.AllKeywords
	if len(allKw) == 0 {
		allKw = append(kwData.Primary, kwData.Technical...)
		allKw = append(allKw, kwData.LongTail...)
	}

	return &dto.WriteResponse{
		Content:  kwData.CategorySuggestion,
		Summary:  fmt.Sprintf("共推荐%d个关键词", len(allKw)),
		Keywords: allKw,
		WordCount: len(allKw),
		Action:   "keywords",
	}, nil
}

// ========== 6. SEO优化 (seo) ==========
// 对文章进行SEO分析和优化建议

func (w *WritingAssistant) optimizeSEO(ctx context.Context, req *dto.WriteRequest) (*dto.WriteResponse, error) {
	if req.Content == "" && req.Topic == "" {
		return nil, fmt.Errorf("content or topic is required for seo action")
	}

	title := req.Topic
	content := req.Content

	prompt := fmt.Sprintf(`请对以下文章进行SEO优化分析:

## 文章标题: %s
## 文章内容(%d字):
%s

## 分析项目:
1. **标题优化**: 当前标题是否足够吸引搜索点击?给出改进建议
2. **Meta描述**: 生成155字符以内的meta description
3. **关键词密度**: 主要关键词是否合理分布?
4. **结构优化**: H1/H2/H3标题结构是否清晰?
5. **内部链接建议**: 建议链接到哪些相关页面?
6. **外链机会**: 是否需要引用外部权威资源?

## 输出格式(JSON):
{
  "title_score": 85,
  "title_suggestions": ["建议1"],
  "meta_description": "优化后的meta描述",
  "keyword_density": {"主要词": "密度百分比"},
  "structure_feedback": "H标签结构反馈",
  "internal_link_suggestions": ["建议链接的页面"],
  "overall_score": 82,
  "optimization_tips": ["具体优化建议1", "建议2"]
}`,
		title, utf8.RuneCountInString(content), content,
	)

	messages := []ChatMessage{
		{Role: "system", Content: "你是一位SEO专家，精通Google和Baidu搜索引擎优化算法。给出具体可执行的建议。"},
		{Role: "user", Content: prompt},
	}

	reply, _, err := w.llmRouter.Chat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("SEO optimization failed: %w", err)
	}

	var seoData struct {
		TitleScore            int      `json:"title_score"`
		TitleSuggestions     []string `json:"title_suggestions"`
		MetaDescription      string   `json:"meta_description"`
		KeywordDensity       map[string]string `json:"keyword_density"`
		StructureFeedback    string   `json:"structure_feedback"`
		InternalLinkSuggestions []string `json:"internal_link_suggestions"`
		OverallScore         int      `json:"overall_score"`
		OptimizationTips     []string `json:"optimization_tips"`
	}

	if err := extractJSONFromText(reply, &seoData); err != nil {
		slog.Warn("Failed to parse SEO JSON", "error", err)
		seoData.OptimizationTips = []string{"请查看原始分析结果: " + reply}
	}

	seoJSON, _ := json.Marshal(seoData)
	return &dto.WriteResponse{
		Title:    seoData.MetaDescription,
		Content:  string(seoJSON),
		Summary:  fmt.Sprintf("SEO评分: %d/100", seoData.OverallScore),
		Keywords: seoData.OptimizationTips,
		WordCount: utf8.RuneCountInString(reply),
		Action:   "seo",
	}, nil
}

// ========== 辅助函数 ==========

func (w *WritingAssistant) resolveStyle(style string) string {
	switch style {
	case "professional":
		return "专业学术风格"
	case "casual":
		return "轻松口语风格"
	case "technical":
		return "硬核技术深度风格"
	default:
		return "专业但易懂的风格(默认)"
	}
}

func (w *WritingAssistant) resolveLength(length string) string {
	switch length {
	case "short":
		return "短篇(800-1500字)"
	case "medium":
		return "中篇(1500-3000字)"
	case "long":
		return "长篇(3000-5000字)"
	default:
		return "中篇(1500-3000字)(默认)"
	}
}

func (w *WritingAssistant) wordCountHint(length string) string {
	switch length {
	case "short":
		return "800-1200字"
	case "medium":
		return "2000-3000字"
	case "long":
		return "4000-6000字"
	default:
		return "2000-3000字"
	}
}

// parseAIGeneratedContent 从LLM回复中解析结构化内容
func parseAIGeneratedContent(reply string, action string) *dto.WriteResponse {
	var data struct {
		Title    string   `json:"title"`
		Content  string   `json:"content"`
		Summary  string   `json:"summary"`
		Keywords []string `json:"keywords"`
	}

	if err := extractJSONFromText(reply, &data); err != nil {
		slog.Warn("Failed to parse AI generated content as JSON, using raw text", "error", err)
		return &dto.WriteResponse{
			Content:  cleanMarkdown(reply),
			WordCount: utf8.RuneCountInString(reply),
			Action:   action,
		}
	}

	if data.Content == "" {
		data.Content = data.Title + "\n\n" + reply
	}

	return &dto.WriteResponse{
		Title:    data.Title,
		Content:  cleanMarkdown(data.Content),
		Summary:  data.Summary,
		Keywords: data.Keywords,
		WordCount: utf8.RuneCountInString(data.Content),
		Action:   action,
	}
}

// cleanMarkdown 清理LLM生成的Markdown文本
func cleanMarkdown(text string) string {
	text = strings.TrimSpace(text)
	// 移除代码块开头的 ```lang 标记
	reOpen := regexp.MustCompile(`(?m)^` + "```" + `[a-zA-Z]*\n?`)
	text = reOpen.ReplaceAllString(text, "")
	// 移除代码块结尾的 ```
	reClose := regexp.MustCompile(`(?m)^` + "```" + `\n?$`)
	text = reClose.ReplaceAllString(text, "")
	for strings.HasPrefix(text, "```") || strings.HasSuffix(text, "```") {
		text = strings.TrimPrefix(text, "```")
		text = strings.TrimSuffix(text, "```")
		text = strings.TrimSpace(text)
	}
	return text
}

// extractJSONFromText 从可能包含markdown包裹的文本中提取JSON
func extractJSONFromText[T any](text string, target *T) error {
	text = strings.TrimSpace(text)
	jsonStart := strings.Index(text, "{")
	jsonEnd := strings.LastIndex(text, "}")
	if jsonStart >= 0 && jsonEnd > jsonStart {
		jsonStr := text[jsonStart : jsonEnd+1]
		return json.Unmarshal([]byte(jsonStr), target)
	}
	return fmt.Errorf("no JSON object found in text")
}

// extractKeywordsFallback 从纯文本中提取关键词(正则匹配大写/引号包裹的词)
func extractKeywordsFallback(text string) []string {
	re := regexp.MustCompile(`[""]([^""]+)[""]|([A-Za-z][A-Za-z_]+(?:[+-][A-Za-z][A-Za-z_]+)*)`)
	matches := re.FindAllStringSubmatch(text, -1)
	seen := make(map[string]bool)
	var result []string
	for _, m := range matches {
		word := m[1] + m[2]
		if !seen[word] && len(word) >= 2 && len(word) <= 20 {
			seen[word] = true
			result = append(result, word)
		}
	}
	if len(result) > 8 {
		result = result[:8]
	}
	return result
}
