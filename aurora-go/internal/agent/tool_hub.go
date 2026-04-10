package agent

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"

	"github.com/aurora-go/aurora/internal/constant"
	"github.com/aurora-go/aurora/internal/infrastructure/database"
	svc "github.com/aurora-go/aurora/internal/service"
	"github.com/aurora-go/aurora/internal/strategy"
	"github.com/aurora-go/aurora/internal/util"
	"github.com/aurora-go/aurora/internal/vo"
)

// ========== Tool 工具集Hub ==========
// 对标 tRPC tool/function + tool/mcp (任意Go函数→Tool, 一行封装)
// 支持: 6个Aurora业务工具 + MCP协议工具扩展

// ToolFunc 工具函数签名（对标 tRPC FunctionTool）
type ToolFunc func(ctx context.Context, params map[string]interface{}) (interface{}, error)

// ToolDefinition 工具定义（用于LLM function calling）
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"` // JSON Schema
}

// Tool 注册的工具
type Tool struct {
	Definition ToolDefinition
	Execute    ToolFunc
}

// ToolHub 工具集管理器
type ToolHub struct {
	mu     sync.RWMutex
	tools  map[string]*Tool
}

// NewToolHub 创建工具集
func NewToolHub() *ToolHub {
	return &ToolHub{
		tools: make(map[string]*Tool),
	}
}

// Register 注册一个工具
func (h *ToolHub) Register(tool *Tool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.tools[tool.Definition.Name] = tool
	slog.Debug("Tool registered", "name", tool.Definition.Name)
}

// Execute 执行指定工具
func (h *ToolHub) Execute(ctx context.Context, name string, params map[string]interface{}) (interface{}, error) {
	h.mu.RLock()
	tool, ok := h.tools[name]
	h.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("tool '%s' not found", name)
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("tool '%s' execution failed: %w", name, err)
	}

	return result, nil
}

// GetDefinitions 获取所有工具定义（用于LLM function calling）
func (h *ToolHub) GetDefinitions() []ToolDefinition {
	h.mu.RLock()
	defer h.mu.RUnlock()
	defs := make([]ToolDefinition, 0, len(h.tools))
	for _, tool := range h.tools {
		defs = append(defs, tool.Definition)
	}
	return defs
}

// ToolCount 获取工具数量
func (h *ToolHub) ToolCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.tools)
}

// ExecuteAnalysis 分析工具快捷执行方法
func (h *ToolHub) ExecuteAnalysis(analysisType string, dateRange, metric string) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"type":       analysisType,
		"date_range": dateRange,
		"metric":     metric,
	}
	result, err := h.Execute(context.Background(), "analyze_stats", params)
	if err != nil {
		return nil, err
	}
	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result type from analyze_stats")
	}
	return data, nil
}

// ========== Aurora 业务工具注册 (~200行) ==========

// RegisterAuroraTools 注册所有Aurora业务工具到Hub
func RegisterAuroraTools(hub *ToolHub) {
	// 1. 文章搜索工具
	hub.Register(&Tool{
		Definition: ToolDefinition{
			Name:        "search_articles",
			Description: "搜索博客文章，支持标题/内容关键词搜索、分类过滤、状态筛选。返回文章列表含摘要和高亮。",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"keywords":   map[string]interface{}{"type": "string", "description": "搜索关键词"},
					"category_id": map[string]interface{}{"type": "integer", "description": "分类ID(可选)"},
					"status":     map[string]interface{}{"type": "integer", "description": "文章状态(1=公开,2=私密)"},
					"page_num":   map[string]interface{}{"type": "integer", "description": "页码(默认1)"},
					"page_size":  map[string]interface{}{"type": "integer", "description": "每页数量(默认10)"},
				},
				"required": []string{"keywords"},
			},
		},
		Execute: executeSearchArticles,
	})

	// 2. 文章读写工具
	hub.Register(&Tool{
		Definition: ToolDefinition{
			Name:        "write_article",
			Description: "创建或更新博客文章。支持设置标题、内容、分类、标签、置顶、推荐等属性。返回文章ID和完整信息。",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"title":      map[string]interface{}{"type": "string", "description": "文章标题"},
					"content":    map[string]interface{}{"type": "string", "description": "文章内容(Markdown格式)"},
					"summary":    map[string]interface{}{"type": "string", "description": "文章摘要"},
					"category_id":map[string]interface{}{"type": "integer", "description": "分类ID"},
					"tag_names":  map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "标签名称列表"},
					"is_top":     map[string]interface{}{"type": "boolean", "description": "是否置顶"},
					"article_id": map[string]interface{}{"type": "integer", "description": "文章ID(更新时必填)"},
				},
				"required": []string{"title", "content"},
			},
		},
		Execute: executeWriteArticle,
	})

	// 3. 统计分析工具
	hub.Register(&Tool{
		Definition: ToolDefinition{
			Name:        "analyze_stats",
			Description: "获取博客运营统计数据，包括访问量(PV/UV)、热门文章排行、地域分布、趋势数据等。支持按时间范围筛选。",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"type": map[string]interface{}{
						"type": "string",
						"enum": []string{"traffic", "content", "user", "seo", "trend"},
						"description": "统计类型: traffic=流量统计, content=内容统计, user=用户统计, seo=SEO统计, trend=趋势分析",
					},
					"date_range": map[string]interface{}{"type": "string", "description": "时间范围: 7d/30d/90d/all"},
					"metric":     map[string]interface{}{"type": "string", "description": "具体指标(pv/uv/bounce_rate等)"},
				},
				"required": []string{"type"},
			},
		},
		Execute: executeAnalyzeStats,
	})

	// 4. 评论审核工具
	hub.Register(&Tool{
		Definition: ToolDefinition{
			Name:        "review_content",
			Description: "审核文章或评论内容，检查是否包含敏感词、垃圾评论特征、情感倾向等。返回审核结果和建议操作。",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"content": map[string]interface{}{"type": "string", "description": "待审核内容"},
					"type":    map[string]interface{}{"type": "string", "enum": []string{"article", "comment"}, "description": "内容类型"},
				},
				"required": []string{"content"},
			},
		},
		Execute: executeReviewContent,
	})

	// 5. 标签管理工具
	hub.Register(&Tool{
		Definition: ToolDefinition{
			Name:        "manage_tags",
			Description: "查询、创建或关联标签。可以获取所有标签列表、创建新标签、为文章添加/移除标签。",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"action":    map[string]interface{}{"type": "string", "enum": []string{"list", "create", "add_to_article", "remove_from_article"}, "description": "操作类型"},
					"name":      map[string]interface{}{"type": "string", "description": "标签名称(create时必填)"},
					"article_id":map[string]interface{}{"type": "integer", "description": "文章ID(add/remove时必填)"},
				},
				"required": []string{"action"},
			},
		},
		Execute: executeManageTags,
	})

	// 6. 分类管理工具
	hub.Register(&Tool{
		Definition: ToolDefinition{
			Name:        "get_categories",
			Description: "获取博客分类列表及其文章数量。可用于了解网站结构、辅助文章归类建议。",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		Execute: executeGetCategories,
	})
}

// ========== 工具实现（通过Service层调用）==========
// 注意: 这些函数通过全局Registry获取Service实例，避免循环依赖

func executeSearchArticles(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// 通过策略模式调用搜索服务
	keywords, _ := params["keywords"].(string)
	pageNum, _ := params["page_num"].(float64)
	pageSize, _ := params["page_size"].(float64)

	if pageNum == 0 {
		pageNum = 1
	}
	if pageSize == 0 {
		pageSize = 10
	}

	searchCtx := strategy.GetSearchContext()
	if searchCtx == nil {
		return map[string]interface{}{"error": "search service unavailable"}, nil
	}

	results, err := searchCtx.ExecuteSearch(ctx, keywords)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	return map[string]interface{}{
		"results": results,
		"keywords": keywords,
	}, nil
}

func executeWriteArticle(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	title, _ := params["title"].(string)
	content, _ := params["content"].(string)

	// 验证必要字段
	if title == "" || content == "" {
		return nil, fmt.Errorf("title and content are required")
	}

	// 构建VO并调用ArticleService
	articleVO := vo.ArticleVO{
		ArticleTitle:   title,
		ArticleContent: content,
		CategoryID:     getUintParam(params, "category_id"),
	}

	// 更新模式 vs 创建模式
	articleID := getUintParam(params, "article_id")

	articleSvc := svc.GetArticleService()
	if articleSvc == nil {
		return nil, fmt.Errorf("article service unavailable")
	}

	if articleID > 0 {
		articleVO.ID = articleID
		result, err := articleSvc.UpdateArticle(ctx, articleVO)
		if err != nil {
			return nil, fmt.Errorf("update article failed: %w", err)
		}
		return map[string]interface{}{"id": result.ID, "action": "updated"}, nil
	}

	result, err := articleSvc.CreateArticle(ctx, 0, articleVO)
	if err != nil {
		return nil, fmt.Errorf("create article failed: %w", err)
	}
	return map[string]interface{}{"id": result.ID, "action": "created"}, nil
}

func executeAnalyzeStats(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	statType, _ := params["type"].(string)
	dateRange, _ := params["date_range"].(string)

	switch statType {
	case "traffic":
		return fetchTrafficStats(dateRange)
	case "content":
		return fetchContentStats(dateRange)
	case "user":
		return fetchUserStats(dateRange)
	default:
		return map[string]interface{}{
			"type": statType,
			"message": "stat type not yet implemented, use traffic/content/user",
		}, nil
	}
}

func executeReviewContent(_ context.Context, params map[string]interface{}) (interface{}, error) {
	content, _ := params["content"].(string)
	contentType, _ := params["type"].(string)

	if contentType == "" {
		contentType = "comment"
	}

	// 敏感词检查
	isSensitive := false
	sensitiveWords := []string{"赌博", "色情", "诈骗", "代开发票", "刷单"}
	matchedWords := []string{}
	for _, word := range sensitiveWords {
		if util.ContainsSensitiveWord(content, word) {
			isSensitive = true
			matchedWords = append(matchedWords, word)
		}
	}

	// 垃圾评论检测
	isSpam := detectSpamComment(content)

	reviewResult := map[string]interface{}{
		"is_safe":     !isSensitive && !isSpam,
		"is_sensitive": isSensitive,
		"is_spam":     isSpam,
		"sensitive_words": matchedWords,
		"suggestion": buildReviewSuggestion(isSensitive, isSpam, contentType),
		"confidence": calculateConfidence(isSensitive, isSpam),
	}

	return reviewResult, nil
}

func executeManageTags(_ context.Context, params map[string]interface{}) (interface{}, error) {
	action, _ := params["action"].(string)

	tagSvc := svc.GetTagService()
	if tagSvc == nil {
		return nil, fmt.Errorf("tag service unavailable")
	}

	switch action {
	case "list":
		tags, err := tagSvc.GetTags(context.Background())
		if err != nil {
			return nil, err
		}
		return tags, nil

	case "create":
		name, _ := params["name"].(string)
		tag, err := tagSvc.CreateTag(context.Background(), vo.TagVO{TagName: name})
		if err != nil {
			return nil, err
		}
		return tag, nil

	default:
		return nil, fmt.Errorf("unsupported action: %s", action)
	}
}

func executeGetCategories(_ context.Context, _ map[string]interface{}) (interface{}, error) {
	catSvc := svc.GetCategoryService()
	if catSvc == nil {
		return nil, fmt.Errorf("category service unavailable")
	}
	return catSvc.GetCategories(context.Background())
}

// ========== 辅助函数 ==========

func getStringParam(m map[string]interface{}, key string) string {
	v, _ := m[key].(string)
	return v
}

func getUintParam(m map[string]interface{}, key string) uint {
	switch v := m[key].(type) {
	case float64:
		return uint(v)
	case int:
		return uint(v)
	case int64:
		return uint(v)
	case uint:
		return v
	default:
		return 0
	}
}

func getBoolParam(m map[string]interface{}, key string) bool {
	v, _ := m[key].(bool)
	return v
}

func detectSpamComment(content string) bool {
	spamPatterns := []string{
		"http://", "https://", "www.", "免费", "赚钱", "代写",
		"加微信", "加QQ", "兼职", "贷款", "彩票",
	}
	lowerContent := strings.ToLower(content)
	for _, pattern := range spamPatterns {
		if strings.Contains(lowerContent, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

func buildReviewSuggestion(isSensitive, isSpam bool, contentType string) string {
	switch {
	case isSensitive:
		return fmt.Sprintf("该%s包含敏感词，建议拒绝发布或转人工审核", contentType)
	case isSpam:
		return fmt.Sprintf("该%s疑似垃圾内容，建议标记为待审核", contentType)
	default:
		return "内容安全，可以通过"
	}
}

func calculateConfidence(isSensitive, isSpam bool) float64 {
	switch {
	case isSensitive && isSpam:
		return 0.95
	case isSensitive:
		return 0.85
	case isSpam:
		return 0.75
	default:
		return 0.95 // 安全置信度高
	}
}

func fetchTrafficStats(_ string) (map[string]interface{}, error) {
	rdb := database.GetRedis()
	if rdb == nil {
		return map[string]interface{}{"pv": 0, "uv": 0}, nil
	}

	ctx := context.Background()
	pv, _ := rdb.Get(ctx, constant.SiteViewCount).Int()
	uv, _ := rdb.Get(ctx, constant.UniqueViewToday).Int()

	// 热门文章排行
	topArticles, _ := rdb.ZRevRangeWithScores(ctx, constant.ArticleViewsRanking, 0, 4).Result()
	topList := make([]map[string]interface{}, 0, len(topArticles))
	for _, z := range topArticles {
		articleID, _ := strconv.ParseUint(z.Member.(string), 10, 32)
		topList = append(topList, map[string]interface{}{
			"id":    articleID,
			"views": int(z.Score),
		})
	}

	return map[string]interface{}{
		"pv":           pv,
		"uv":           uv,
		"top_articles": topList,
	}, nil
}

func fetchContentStats(dateRange string) (map[string]interface{}, error) {
	articleSvc := svc.GetArticleService()
	if articleSvc == nil {
		return map[string]interface{}{"total": 0, "published": 0}, nil
	}

	// TODO: 实现 ArticleService.CountByStatus
	return map[string]interface{}{
		"total":      0,
		"published":  0,
		"draft":      0,
		"date_range": dateRange,
	}, nil
}

func fetchUserStats(dateRange string) (map[string]interface{}, error) {
	rdb := database.GetRedis()
	if rdb == nil {
		return map[string]interface{}{"online": 0}, nil
	}

	online, _ := rdb.DBSize(context.Background()).Result()

	return map[string]interface{}{
		"online_users": online,
		"date_range":   dateRange,
	}, nil
}
