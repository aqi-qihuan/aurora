package agent

import (
	"context"
	"log/slog"
	"math"
	"regexp"
	"strings"

	"github.com/aurora-go/aurora/internal/dto"
)

// ========== AI语义搜索增强 ==========
// 在传统ES关键词搜索基础上增加AI理解层
// 能力: 查询扩展 / 意图识别 / 结果重排序 / 摘要生成 / 相关推荐
//
// 架构:
//   User Query → IntentRecognizer → QueryExpander → ES Search → ResultReranker → SummaryGenerator
//                                                          ↓
//                                              RelatedQuery Suggester

// SemanticSearchEngine 语义搜索引擎
type SemanticSearchEngine struct {
	ragPipeline *RAGPipeline
	llmRouter   *LLMRouter
}

// NewSemanticSearchEngine 创建语义搜索引擎
func NewSemanticSearchEngine(pipeline *RAGPipeline, router *LLMRouter) *SemanticSearchEngine {
	return &SemanticSearchEngine{ragPipeline: pipeline, llmRouter: router}
}

// SearchIntent 搜索意图类型
type SearchIntent string

const (
	IntentInformational SearchIntent = "informational" // 信息查找: "什么是goroutine"
	IntentNavigational  SearchIntent = "navigational"  // 导航定位: "Go官方文档"
	IntentTransactional SearchIntent = "transactional" // 操作执行: "如何安装Go"
	IntentComparison    SearchIntent = "comparison"    // 对比选择: "Gin vs Echo哪个好"
	IntentTroubleshoot  SearchIntent = "troubleshoot"  // 问题排查: "连接MySQL超时怎么办"
)

// IntentRecognitionResult 意图识别结果
type IntentRecognitionResult struct {
	Intent      SearchIntent `json:"intent"`
	Confidence  float64      `json:"confidence"`
	Entities    []QueryEntity `json:"entities"`
	RawQuery    string        `json:"rawQuery"`
	NormalizedQuery string    `json:"normalizedQuery"`
	ExpandedQueries []string  `json:"expandedQueries"`
}

// QueryEntity 查询实体
type QueryEntity struct {
	Text   string `json:"text"`
	Type   string `json:"type"`   // tech/framework/concept/language/version
	Offset int    `json:"offset"` // 原文中的位置
}

// EnhancedSearchResult 增强搜索结果
type EnhancedSearchResult struct {
	dto.SearchResponse
	QueryUnderstanding *IntentRecognitionResult `json:"queryUnderstanding,omitempty"`
	AnswerType         string                   `json:"answerType"` // direct_answer/list/snippet/fallback
	RelatedQueries     []string                 `json:"relatedQueries"`
	Facets             map[string][]FacetItem    `json:"facets,omitempty"`
}

// FacetItem 聚合分面项
type FacetItem struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

// ========== 公开方法 ==========

// EnhancedSearch 执行增强语义搜索
func (se *SemanticSearchEngine) EnhancedSearch(ctx context.Context, req *dto.SearchRequest) (*EnhancedSearchResult, error) {
	// Step 1: 意图识别 + 查询扩展
	intent := se.RecognizeIntent(req.Query)

	// Step 2: 用扩展查询执行搜索
	searchReq := &dto.SearchRequest{
		Query:    intent.NormalizedQuery,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		UseRAG:   req.UseRAG,
		TopK:     req.TopK,
	}

	var searchResp *dto.SearchResponse
	var err error

	if req.UseRAG {
		searchResp, err = se.ragPipeline.Retrieve(ctx, searchReq)
	} else {
		searchResp, err = se.ragPipeline.SemanticSearch(ctx, searchReq)
	}

	if err != nil {
		slog.Warn("Enhanced search base search failed", "error", err)
		searchResp = &dto.SearchResponse{Results: []dto.AgentSearchResult{}, Total: 0}
	}

	// Step 3: 结果重排序（基于意图相关性）
	se.rerankByIntent(searchResp.Results, intent)

	// Step 4: 生成答案类型判断
	answerType := se.determineAnswerType(intent, searchResp)

	// Step 5: 生成相关问题推荐
	relatedQueries := se.suggestRelatedQueries(ctx, intent, req.Query)

	result := &EnhancedSearchResult{
		SearchResponse:     *searchResp,
		QueryUnderstanding: intent,
		AnswerType:         answerType,
		RelatedQueries:     relatedQueries,
	}

	return result, nil
}

// RecognizeIntent 识别搜索意图
func (se *SemanticSearchEngine) RecognizeIntent(query string) *IntentRecognitionResult {
	result := &IntentRecognitionResult{
		RawQuery:         query,
		NormalizedQuery:  normalizeQuery(query),
		Confidence:       0.7,
	}

	// 1. 实体提取
	result.Entities = extractEntities(query)

	// 2. 基于规则的意图识别
	intent, conf := classifyByRules(query)
	result.Intent = intent
	if conf > result.Confidence {
		result.Confidence = conf
	}

	// 3. 查询扩展
	result.ExpandedQueries = expandQuery(query, intent, result.Entities)

	return result
}

// ========== 意图识别 ==========

func classifyByRules(query string) (SearchIntent, float64) {
	q := strings.ToLower(strings.TrimSpace(query))

	infoPatterns := []struct {
		pattern string
		conf    float64
	}{
		{"什么", 0.8}, {"如何", 0.75}, {"怎么", 0.75}, {"为什么", 0.85},
		{"是什么", 0.9}, {"原理", 0.85}, {"区别", 0.7}, {"对比", 0.7},
		{"教程", 0.8}, {"入门", 0.78}, {"指南", 0.78}, {"详解", 0.82},
		{"vs ", 0.85}, {"和.*的区别", 0.88}, {"哪个好", 0.9},
		{"报错", 0.88}, {"错误", 0.8}, {"失败", 0.78}, {"解决", 0.8},
		{"配置", 0.72}, {"部署", 0.72}, {"安装", 0.7},
	}

	maxConf := 0.5
	bestIntent := IntentInformational

	for _, p := range infoPatterns {
		if matched, _ := regexp.MatchString(p.pattern, q); matched {
			if p.conf > maxConf {
				maxConf = p.conf
				switch {
				case strings.Contains(p.pattern, "区别"), strings.Contains(p.pattern, "vs "), p.pattern == "哪个好":
					bestIntent = IntentComparison
				case strings.Contains(p.pattern, "报错"), strings.Contains(p.pattern, "错误"), strings.Contains(p.pattern, "解决"):
					bestIntent = IntentTroubleshoot
				case strings.Contains(p.pattern, "配置"), strings.Contains(p.pattern, "部署"), strings.Contains(p.pattern, "安装"):
					bestIntent = IntentTransactional
				default:
					bestIntent = IntentInformational
				}
			}
		}
	}

	// URL/导航类
	if regexp.MustCompile(`^(http|www\.|/)`).MatchString(q) {
		return IntentNavigational, 0.95
	}

	return bestIntent, maxConf
}

// ========== 实体提取 ==========

var entityPatterns = map[string]*regexp.Regexp{
	"language":    regexp.MustCompile(`(?i)\b(Go|Java|Python|JavaScript|TypeScript|Rust|Kotlin|Swift|C\+\+|C#|PHP|Ruby)\b`),
	"framework":   regexp.MustCompile(`(?i)\b(Gin|Echo|Fiber|SpringBoot|Django|Flask|FastAPI|Express|Nextjs|Nuxt|Vue|React|Angular)\b`),
	"concept":     regexp.MustCompile(`(?i)\b(goroutine|channel|middleware|ORM|RESTful|GraphQL|gRPC|microservice|container|kubernetes)\b`),
	"database":    regexp.MustCompile(`(?i)\b(MySQL|PostgreSQL|Redis|MongoDB|Elasticsearch|SQLite|MariaDB)\b`),
	"cloud":       regexp.MustCompile(`(?i)\b(AWS|Azure|GCP|Docker|Kubernetes|Nginx|Linux|Ubuntu|CentOS)\b`),
	"version":      regexp.MustCompile(`\b(\d+\.\d+(\.\d+)?(?:-[a-zA-Z]+)?)\b`),
}

func extractEntities(query string) []QueryEntity {
	var entities []QueryEntity
	for entityType, pattern := range entityPatterns {
		matches := pattern.FindAllStringSubmatchIndex(query, -1)
		for _, m := range matches {
			entities = append(entities, QueryEntity{
				Text:   query[m[0]:m[1]],
				Type:   entityType,
				Offset: m[0],
			})
		}
	}
	return entities
}

// ========== 查询规范化 + 扩展 ==========

func normalizeQuery(query string) string {
	q := strings.TrimSpace(query)
	q = regexp.MustCompile(`\s+`).ReplaceAllString(q, " ") // 合并多余空格
	q = regexp.MustCompile(`[?？!！。，]+$`).ReplaceAllString(q, "") // 去除尾部标点
	return q
}

func expandQuery(query string, intent SearchIntent, entities []QueryEntity) []string {
	queries := make(map[string]bool)
	queries[query] = true

	// 基于实体的同义词扩展
	for _, e := range entities {
		switch e.Type {
		case "language":
			synonyms := getLanguageSynonyms(e.Text)
			for _, s := range synonyms {
				newQ := strings.Replace(query, e.Text, s, 1)
				queries[newQ] = true
			}
		case "framework":
			synonyms := getFrameworkSynonyms(e.Text)
			for _, s := range synonyms {
				newQ := strings.Replace(query, e.Text, s, 1)
				queries[newQ] = true
			}
		}
	}

	// 基于意图的前缀扩展
	switch intent {
	case IntentInformational:
		queries[query+" 入门"] = true
		queries[query+" 教程"] = true
		queries[query+" 是什么"] = true
	case IntentTroubleshoot:
		queries[query+" 解决方案"] = true
		queries[query+" 怎么办"] = true
	case IntentTransactional:
		queries[query+" 步骤"] = true
		queries[query+" 方法"] = true
	}

	result := make([]string, 0, len(queries))
	for q := range queries {
		if q != query {
			result = append(result, q)
		}
	}
	if len(result) > 5 {
		result = result[:5]
	}
	return result
}

// ========== 结果重排序 ==========

func (se *SemanticSearchEngine) rerankByIntent(results []dto.AgentSearchResult, intent *IntentRecognitionResult) {
	if len(results) == 0 {
		return
	}

	for i := range results {
		scoreBoost := 0.0
		titleLower := strings.ToLower(results[i].Title)
		summaryLower := strings.ToLower(results[i].Summary)

		// 实体匹配加分
		for _, e := range intent.Entities {
			if strings.Contains(titleLower, strings.ToLower(e.Text)) {
				scoreBoost += 0.1
			}
			if strings.Contains(summaryLower, strings.ToLower(e.Text)) {
				scoreBoost += 0.05
			}
		}

		// 标题完全匹配加分
		if strings.Contains(titleLower, intent.NormalizedQuery) {
			scoreBoost += 0.15
		}

		results[i].Relevance = math.Min(1.0, results[i].Relevance+scoreBoost)
	}

	// 冒泡排序（小数据量够用）
	for i := 0; i < len(results)-1; i++ {
		for j := 0; j < len(results)-1-i; j++ {
			if results[j].Relevance < results[j+1].Relevance {
				results[j], results[j+1] = results[j+1], results[j]
			}
		}
	}
}

// ========== 答案类型判断 ==========

func (se *SemanticSearchEngine) determineAnswerType(intent *IntentRecognitionResult, searchResp *dto.SearchResponse) string {
	if searchResp.Answer != "" && len(searchResp.Answer) > 20 {
		return "direct_answer"
	}
	if intent.Intent == IntentTroubleshoot && len(searchResp.Results) > 0 {
		return "snippet"
	}
	if intent.Intent == IntentComparison && len(searchResp.Results) > 1 {
		return "list"
	}
	if len(searchResp.Results) == 1 {
		return "direct_answer"
	}
	if len(searchResp.Results) == 0 {
		return "fallback"
	}
	return "list"
}

// ========== 相关问题推荐 ==========

func (se *SemanticSearchEngine) suggestRelatedQueries(_ context.Context, intent *IntentRecognitionResult, originalQuery string) []string {
	related := make([]string, 0, 5)

	// 基于意图的固定模板
	switch intent.Intent {
	case IntentInformational:
		related = append(related, originalQuery+" 实战案例")
		related = append(related, originalQuery+" 最佳实践")
		related = append(related, originalQuery+" 优缺点")
	case IntentTroubleshoot:
		related = append(related, originalQuery+" 常见原因")
		related = append(related, originalQuery+" 排查步骤")
		related = append(related, originalQuery+" 替代方案")
	case IntentComparison:
		related = append(related, originalQuery+" 性能对比")
		related = append(related, originalQuery+" 选型建议")
	case IntentTransactional:
		related = append(related, originalQuery+" 注意事项")
		related = append(related, originalQuery+" 配置详解")
	default:
		related = append(related, originalQuery+" 简介")
		related = append(related, originalQuery+" 示例")
	}

	// 基于实体的推荐
	for _, e := range intent.Entities {
		switch e.Type {
		case "framework":
			related = append(related, e.Text+" 入门教程")
			related = append(related, e.Text+" 配置指南")
		case "database":
			related = append(related, e.Text+" 优化技巧")
		case "concept":
			related = append(related, e.Text+" 应用场景")
		}
	}

	if len(related) > 6 {
		related = related[:6]
	}
	return related
}

// ========== 同义词库 ==========

func getLanguageSynonyms(lang string) []string {
	synonyms := map[string][]string{
		"Golang": {"go语言", "golang", "Go"}, "Go": {"golang", "go语言", "Golang"},
		"JavaScript": {"JS", "javascript", "ECMAScript"}, "JS": {"JavaScript", "javascript", "ECMAScript"},
		"Python": {"python", "Py", "py"},
		"Java": {"java", "JDK", "JVM"},
	}
	if s, ok := synonyms[lang]; ok {
		return s
	}
	if s, ok := synonyms[strings.Title(lang)]; ok {
		return s
	}
	return nil
}

func getFrameworkSynonyms(fw string) []string {
	synonyms := map[string][]string{
		"Gin": {"gin框架", "gin-gonic", "Gin Web Framework"},
		"Vue": {"vue.js", "vue3", "Vue.js", "Vue2"},
		"React": {"react.js", "React.js", "Next.js"},
		"Docker": {"docker容器", "容器化", "Docker Compose"},
		"SpringBoot": {"spring boot", "Spring-Boot", "springboot"},
	}
	if s, ok := synonyms[fw]; ok {
		return s
	}
	if s, ok := synonyms[strings.Title(fw)]; ok {
		return s
	}
	return nil
}
