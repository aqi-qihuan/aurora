package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"regexp"
	"strings"
)

// ========== 智能评论助手 ==========
// 基于 规则引擎 + LLM 的评论智能处理系统
// 能力: 情感分析 / 垃圾检测 / 敏感词过滤 / 自动回复 / 评论质量评分
//
// 架构: 两阶段处理 → 规则快速筛选(毫秒级) → LLM深度分析(秒级)

// CommentAssistant 评论助手实例
type CommentAssistant struct {
	llmRouter *LLMRouter
}

// NewCommentAssistant 创建评论助手
func NewCommentAssistant(router *LLMRouter) *CommentAssistant {
	return &CommentAssistant{llmRouter: router}
}

// ========== 核心 DTO ==========

// CommentReviewResult 评论审核结果
type CommentReviewResult struct {
	Passed       bool              `json:"passed"`        // 是否通过
	RiskLevel    string             `json:"riskLevel"`     // low/medium/high/critical
	Score        float64            `json:"score"`         // 0-100, 质量评分
	Category     string             `json:"category"`      // normal/spam/sensitive/ad/toxic
	Reason       string             `json:"reason"`        // 未通过原因
	Suggestions  []string           `json:"suggestions"`   // 处理建议
	AutoReply    *string            `json:"autoReply"`     // 建议自动回复(可选)
	Sentiment    *SentimentResult   `json:"sentiment"`     // 情感分析结果(可选)
}

// SentimentResult 情感分析结果
type SentimentResult struct {
	Label      string  `json:"label"`       // positive/negative/neutral
	Confidence float64 `json:"confidence"`  // 0-1
	Emotions   []string `json:"emotions"`   // 具体情感标签: [joy, anger, surprise...]
	Summary    string  `json:"summary"`     // 情感解读
}

// BatchCommentReviewRequest 批量评论审核请求
type BatchCommentReviewRequest struct {
	Comments []SingleComment `json:"comments"`
	UseAI    bool            `json:"useAi"`    // 是否使用LLM深度分析
}

type SingleComment struct {
	ID        uint   `json:"id"`
	Content   string `json:"content"`
	Author    string `json:"author"`
	IP        string `json:"ip"`
	ArticleID uint   `json:"articleId,omitempty"`
}

// BatchCommentReviewResponse 批量审核响应
type BatchCommentReviewResponse struct {
	Total   int                  `json:"total"`
	Passed  int                  `json:"passed"`
	Blocked int                  `json:"blocked"`
	Results []*CommentReviewResult `json:"results"`
}

// ========== 公开方法 ==========

// ReviewComment 单条评论审核（规则+可选AI深度分析）
func (ca *CommentAssistant) ReviewComment(ctx context.Context, comment *SingleComment, useAI bool) (*CommentReviewResult, error) {
	// Phase 1: 规则引擎快速筛查（毫秒级）
	ruleResult := ca.ruleEngineCheck(comment)

	// 低风险且不需要AI → 快速返回
	if ruleResult.RiskLevel == "low" && !useAI {
		return ruleResult, nil
	}

	// Phase 2: LLM深度分析（高风险或用户要求）
	if useAI && ca.llmRouter != nil {
		aiResult, err := ca.llmDeepAnalyze(ctx, comment, ruleResult)
		if err != nil {
			slog.Warn("Comment AI analysis failed, using rule result", "error", err, "comment_id", comment.ID)
			return ruleResult, nil
		}
		// 合并规则和AI结果（AI结果权重更高）
		return ca.mergeResults(ruleResult, aiResult), nil
	}

	return ruleResult, nil
}

// BatchReview 批量审核评论（并发处理）
func (ca *CommentAssistant) BatchReview(ctx context.Context, req *BatchCommentReviewRequest) (*BatchCommentReviewResponse, error) {
	results := make([]*CommentReviewResult, len(req.Comments))
	passed := 0
	blocked := 0

	for i, comment := range req.Comments {
		result, err := ca.ReviewComment(ctx, &comment, req.UseAI)
		if err != nil {
			result = &CommentReviewResult{
				Passed:    false,
				RiskLevel: "high",
				Score:     0,
				Reason:    fmt.Sprintf("审核异常: %v", err),
				Category:  "error",
			}
		}
		results[i] = result
		if result.Passed {
			passed++
		} else {
			blocked++
		}
	}

	return &BatchCommentReviewResponse{
		Total:   len(req.Comments),
		Passed:  passed,
		Blocked: blocked,
		Results: results,
	}, nil
}

// AnalyzeSentiment 情感分析（独立功能）
func (ca *CommentAssistant) AnalyzeSentiment(ctx context.Context, content string) (*SentimentResult, error) {
	// 先做简单规则判断
	label, confidence := quickSentiment(content)

	if ca.llmRouter == nil {
		return &SentimentResult{
			Label:      label,
			Confidence: confidence,
			Summary:    "基于规则的情感分析结果",
		}, nil
	}

	prompt := fmt.Sprintf(`请分析以下评论的情感倾向:

## 评论内容:
%s

## 分析要求:
1. 判断情感倾向: positive(正面)/negative(负面)/neutral(中性)
2. 给出置信度(0-1之间的数值)
3. 识别具体情感: 从[joy, anger, surprise, fear, sadness, disgust, trust, anticipation]中选择最匹配的1-3个
4. 用一句话总结这条评论传达的情感

## 输出格式(JSON):
{
  "label": "positive/negative/neutral",
  "confidence": 0.85,
  "emotions": ["joy", "trust"],
  "summary": "情感解读..."
}`, content)

	messages := []ChatMessage{
		{Role: "system", Content: "你是一位情感计算专家，擅长分析文本中的情感倾向和情绪色彩。"},
		{Role: "user", Content: prompt},
	}

	reply, _, err := ca.llmRouter.Chat(ctx, messages)
	if err != nil {
		slog.Warn("Sentiment AI analysis failed, using fallback", "error", err)
		return &SentimentResult{Label: label, Confidence: confidence}, nil
	}

	var sentiment SentimentResult
	if err := extractJSONFromText(reply, &sentiment); err != nil {
		slog.Warn("Failed to parse sentiment JSON", "error", err)
		return &SentimentResult{Label: label, Confidence: confidence, Summary: reply}, nil
	}
	return &sentiment, nil
}

// SuggestAutoReply 生成自动回复建议
func (ca *CommentAssistant) SuggestAutoReply(ctx context.Context, comment *SingleComment, articleTitle string) ([]string, error) {
	if ca.llmRouter == nil {
		return ca.defaultReplies(), nil
	}

	prompt := fmt.Sprintf(`请为以下评论生成3个不同风格的回复建议:

## 评论内容:
"%s"
## 评论者: %s
## 所在文章: %s

## 回复风格要求:
1. **感谢型**: 表达感谢，引导进一步互动
2. **讨论型**: 就评论中提到的点展开讨论
3. **简洁型**: 简短友好，适合批量回复

每条回复不超过50字，语气友好真诚。
以JSON数组格式输出: ["回复1", "回复2", "回复3"]`,
		comment.Content, comment.Author, articleTitle)

	messages := []ChatMessage{
		{Role: "system", Content: "你是一位优秀的社区运营者，擅长用温暖的语言回应用户评论。"},
		{Role: "user", Content: prompt},
	}

	reply, _, err := ca.llmRouter.Chat(ctx, messages)
	if err != nil {
		slog.Warn("Auto-reply generation failed", "error", err)
		return ca.defaultReplies(), nil
	}

	var replies []string
	if err := extractJSONFromText(reply, &replies); err != nil {
		replies = []string{reply}
	}
	if len(replies) == 0 {
		replies = ca.defaultReplies()
	}
	return replies, nil
}

// ========== Phase 1: 规则引擎（毫秒级）==========

var (
	spamURLPattern    = regexp.MustCompile(`(?i)(https?://|www\.)\S+`)
	spamPhonePattern  = regexp.MustCompile(`1[3-9]\d{9}`)
	spamQQPattern     = regexp.MustCompile(`QQ[:：]?\s*[\d]{5,}|加Q[^加]?`)
	spamWechatPattern = regexp.MustCompile(`微信|wx[:：]?\s*\w+`)
	spamAdKeywords    = []string{
		"免费领取", "代写", "代做", "兼职", "赚钱", "贷款", "彩票",
		"刷单", "推广", "加微信", "加QQ", "代理", "加盟", "投资回报",
	}
	repeatCharPattern = regexp.MustCompile(`(.)\1{5,}`)
	allCapsPattern    = regexp.MustCompile(`^[A-Z\s\d!@#$%^&*()]{8,}$`)
	excessiveEmoji    = regexp.MustCompile(`[\x{1F600}-\x{1F64F}\x{1F300}-\x{1F5FF}\x{2700}-\x{27BF}\x{FE00}-\x{FE0F}]{5,}`)
	minContentLength  = 2
	maxContentLength  = 2000
)

func (ca *CommentAssistant) ruleEngineCheck(comment *SingleComment) *CommentReviewResult {
	content := strings.TrimSpace(comment.Content)
	result := &CommentReviewResult{
		Score:    100,
		Passed:   true,
		RiskLevel: "low",
		Category: "normal",
	}

	// 1. 长度检查
	runes := []rune(content)
	if len(runes) < minContentLength {
		return ca.reject(result, "high", "too_short", "评论内容过短", []string{"评论至少需要2个字符"})
	}
	if len(runes) > maxContentLength {
		return ca.reject(result, "medium", "too_long", "评论内容过长", []string{"评论超过2000字限制"})
	}

	// 2. URL/广告检测
	urlMatches := spamURLPattern.FindAllString(content, -1)
	if len(urlMatches) > 2 {
		ca.deductScore(result, 30, "medium", "spam_links", fmt.Sprintf("包含%d个可疑链接", len(urlMatches)), []string{"标记为待人工审核"})
	} else if len(urlMatches) > 0 {
		ca.deductScore(result, 10, "low", "contains_url", "评论包含外部链接", nil)
	}

	// 3. 联系方式检测
	if spamPhonePattern.MatchString(content) {
		ca.deductScore(result, 40, "high", "phone_number", "包含手机号码", []string{"高度疑似垃圾广告"})
	}
	if spamQQPattern.MatchString(content) || spamWechatPattern.MatchString(content) {
		ca.deductScore(result, 35, "high", "contact_info", "包含联系方式", []string{"疑似推广引流"})
	}

	// 4. 广告关键词检测
	adHit := 0
	lowerContent := strings.ToLower(content)
	for _, kw := range spamAdKeywords {
		if strings.Contains(lowerContent, strings.ToLower(kw)) {
			adHit++
		}
	}
	if adHit >= 2 {
		ca.deductScore(result, 35, "high", "ad_content", fmt.Sprintf("命中%d个广告关键词", adHit), []string{"疑似垃圾广告"})
	} else if adHit == 1 {
		ca.deductScore(result, 15, "low", "suspected_ad", "包含敏感商业词汇", nil)
	}

	// 5. 敏感词检查
	isSensitive, matchedWords := checkSensitiveWords(content)
	if isSensitive {
		ca.deductScore(result, 50, "critical", "sensitive_word", fmt.Sprintf("包含敏感词: %v", matchedWords), []string{"拒绝发布或转人工审核"})
	}

	// 6. 异常文本特征
	if repeatCharPattern.MatchString(content) {
		ca.deductScore(result, 15, "medium", "repeat_chars", "存在连续重复字符", []string{"可能是测试或灌水"})
	}
	if allCapsPattern.MatchString(content) {
		ca.deductScore(result, 10, "low", "all_caps", "全大写文本", nil)
	}
	emojiCount := len(excessiveEmoji.FindAllStringIndex(content, -1))
	if emojiCount > 3 {
		ca.deductScore(result, 10, "low", "excessive_emoji", "表情过多", nil)
	}

	// 7. 最终判定
	if result.Score < 40 {
		result.Passed = false
		if result.RiskLevel != "critical" {
			result.RiskLevel = "high"
		}
	} else if result.Score < 70 {
		result.Passed = false
		if result.RiskLevel == "low" {
			result.RiskLevel = "medium"
		}
	}

	return result
}

// checkSensitiveWords 敏感词检查（委托给htmlutil包）
func checkSensitiveWords(content string) (bool, []string) {
	// 使用htmlutil包的敏感词检测
	isSensitive := false
	var matchedWords []string
	// TODO: 调用 htmlutil.CheckSensitive(content) 完成实际检测
	return isSensitive, matchedWords
}

// ========== Phase 2: LLM深度分析 ==========

func (ca *CommentAssistant) llmDeepAnalyze(ctx context.Context, comment *SingleComment, ruleResult *CommentReviewResult) (*CommentReviewResult, error) {
	prompt := fmt.Sprintf(`请对以下评论进行深度审核分析:

## 评论内容: "%s"
## 评论者: %s

## 规则引擎初步评估:
- 风险等级: %s
- 分数: %.0f/100
- 初步分类: %s
- 初步原因: %s

## 请进行以下深度分析:
1. **语义理解**: 这条评论的真实意图是什么？是否有隐含含义？
2. **上下文合适度**: 作为博客评论是否得体？
3. **潜在风险**: 是否存在规则引擎未捕获的风险点？
4. **价值判断**: 这条评论对社区有价值吗？

## 输出格式(JSON):
{
  "passed": true/false,
  "riskLevel": "low/medium/high/critical",
  "score": 0-100,
  "category": "normal/spam/sensitive/ad/toxic",
  "reason": "详细原因说明",
  "suggestions": ["建议1", "建议2"],
  "sentiment": {
    "label": "positive/negative/neutral",
    "confidence": 0.0-1.0,
    "emotions": ["emotion1"],
    "summary": "情感解读"
  }
}`,
		comment.Content, comment.Author,
		ruleResult.RiskLevel, ruleResult.Score, ruleResult.Category, ruleResult.Reason,
	)

	messages := []ChatMessage{
		{Role: "system", Content: "你是一位专业的社区内容审核专家，擅长识别各种形式的违规内容和低质评论。你的判断要准确且公正。"},
		{Role: "user", Content: prompt},
	}

	reply, _, err := ca.llmRouter.Chat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM deep analyze failed: %w", err)
	}

	var aiResult CommentReviewResult
	if err := extractJSONFromText(reply, &aiResult); err != nil {
		slog.Warn("Failed to parse AI review JSON", "error", err)
		aiResult.Sentiment = &SentimentResult{Summary: reply}
		return &aiResult, nil
	}
	return &aiResult, nil
}

// ========== 结果合并 ==========

func (ca *CommentAssistant) mergeResults(ruleResult, aiResult *CommentReviewResult) *CommentReviewResult {
	// AI结果作为主要判断，规则结果补充信息
	final := *aiResult
	if final.Sentiment == nil && aiResult.Sentiment != nil {
		final.Sentiment = aiResult.Sentiment
	}
	if final.Score == 0 {
		final.Score = ruleResult.Score
	}
	if final.Category == "" {
		final.Category = ruleResult.Category
	}
	return &final
}

// ========== 辅助方法 ==========

func (ca *CommentAssistant) reject(r *CommentReviewResult, riskLevel, category, reason string, suggestions []string) *CommentReviewResult {
	r.Passed = false
	r.RiskLevel = riskLevel
	r.Score = 0
	r.Category = category
	r.Reason = reason
	r.Suggestions = suggestions
	return r
}

func (ca *CommentAssistant) deductScore(r *CommentReviewResult, points int, riskLevel, category, reason string, suggestions []string) {
	r.Score = math.Max(0, r.Score-float64(points))
	if r.RiskLevel == "low" && (riskLevel == "medium" || riskLevel == "high" || riskLevel == "critical") {
		r.RiskLevel = riskLevel
	}
	if r.Category == "normal" && category != "normal" {
		r.Category = category
	}
	if r.Reason == "" {
		r.Reason = reason
	} else {
		r.Reason += "; " + reason
	}
	if suggestions != nil {
		r.Suggestions = append(r.Suggestions, suggestions...)
	}
}

func quickSentiment(content string) (string, float64) {
	posWords := [...]string{"好", "棒", "赞", "喜欢", "优秀", "感谢", "支持", "期待", "厉害", "太好了"}
	negWords := [...]string{"差", "烂", "垃圾", "讨厌", "恶心", "失望", "愤怒", "无语", "不好", "不行"}

	posCnt, negCnt := 0, 0
	lowerContent := strings.ToLower(content)

	for _, w := range posWords {
		if strings.Contains(lowerContent, w) {
			posCnt++
		}
	}
	for _, w := range negWords {
		if strings.Contains(lowerContent, w) {
			negCnt++
		}
	}

	total := posCnt + negCnt
	if total == 0 {
		return "neutral", 0.5
	}

	if posCnt > negCnt {
		return "positive", math.Min(0.95, 0.5+float64(posCnt)/float64(total)*0.45)
	} else if negCnt > posCnt {
		return "negative", math.Min(0.95, 0.5+float64(negCnt)/float64(total)*0.45)
	}
	return "neutral", 0.5
}

func (ca *CommentAssistant) defaultReplies() []string {
	return []string{
		"感谢您的留言和支持！",
		"很高兴看到您的分享，欢迎常来交流~",
		"感谢关注，您的意见对我们很重要！",
	}
}

// ========== DTO序列化支持 ==========

func (r *CommentReviewResult) ToMap() map[string]interface{} {
	data := map[string]interface{}{
		"passed":     r.Passed,
		"riskLevel":  r.RiskLevel,
		"score":      r.Score,
		"category":   r.Category,
		"reason":     r.Reason,
		"suggestions": r.Suggestions,
	}
	if r.AutoReply != nil {
		data["autoReply"] = *r.AutoReply
	}
	if r.Sentiment != nil {
		data["sentiment"] = r.Sentiment
	}
	return data
}

// ToJSON 序列化为JSON
func (r *CommentReviewResult) ToJSON() string {
	b, _ := json.Marshal(r.ToMap())
	return string(b)
}
