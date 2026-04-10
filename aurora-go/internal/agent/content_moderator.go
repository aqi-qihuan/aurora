package agent

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/aurora-go/aurora/internal/dto"
)

// ========== 内容审核系统 ==========
// 多层次审核管道: 关键词过滤 → 正则模式匹配 → LLM深度审核
// 支持异步批量审核 + 审计日志 + 规则热更新

// Moderator 审核器实例
type Moderator struct {
	llmRouter    *LLMRouter
	ruleSets     map[string]*RuleSet // 按类别组织的规则集
	mu           sync.RWMutex
	auditLog     []ModerationEvent  // 内存审计日志(生产环境应写入DB)
	enableAsync  bool               // 是否启用异步LLM审核
}

// RuleSet 规则集合
type RuleSet struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Version     string        `json:"version"`
	Rules       []Rule        `json:"rules"`
	Action      ModerationAction `json:"action"` // 触发时的默认动作
	Enabled     bool          `json:"enabled"`
}

// Rule 单条规则
type Rule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`        // keyword/regex/pattern/length/custom
	Pattern     string `json:"pattern"`     // 匹配模式
	Severity    int    `json:"severity"`     // 1-10严重程度
	Action      ModerationAction `json:"action"`
	Enabled     bool   `json:"enabled"`
	MatchCount  int64  `json:"-"`            // 运行时统计
}

// ModerationAction 审核动作
type ModerationAction string

const (
	ActionAllow    ModerationAction = "allow"     // 通过
	ActionReject   ModerationAction = "reject"    // 拒绝
	ActionReview   ModerationAction = "review"    // 人工审核
	ActionQuarantine ModerationAction = "quarantine" // 隔离
)

// ModerationResult 审核结果
type ModerationResult struct {
	Passed      bool                `json:"passed"`
	Action      ModerationAction     `json:"action"`
	Score       int                 `json:"score"`        // 0-100 安全分数
	Violations  []Violation         `json:"violations"`   // 违规详情
	ProcessedAt time.Time           `json:"processedAt"`
	Duration    time.Duration       `json:"durationMs"`
	Metadata    map[string]string   `json:"metadata"`
}

// Violation 违规记录
type Violation struct {
	RuleID      string             `json:"ruleId"`
	RuleName    string             `json:"ruleName"`
	Category    string             `json:"category"`    // spam/sensitive/ad/toxic/inappropriate
	Severity    int                `json:"severity"`
	MatchedText string             `json:"matchedText"`
	Action      ModerationAction    `json:"action"`
	Message     string             `json:"message"`
}

// ModerationEvent 审计事件
type ModerationEvent struct {
	Timestamp   time.Time          `json:"timestamp"`
	ContentType string             `json:"contentType"`  // article/comment/talk/message
	ContentID   uint               `json:"contentId"`
	Result      *ModerationResult  `json:"result"`
	Operator    string             `json:"operator"`     // system/manual/ai
}

// ReviewTask 异步审核任务
type ReviewTask struct {
	ID        string
	Content   string
	Type      string // article/comment
	ContentID uint
	Callback  func(*ModerationResult)
	CreatedAt time.Time
}

// ========== 工厂方法 ==========

// NewModerator 创建审核器（带内置规则）
func NewModerator(router *LLMRouter) *Moderator {
	m := &Moderator{
		llmRouter:   router,
		ruleSets:    make(map[string]*RuleSet),
		auditLog:    make([]ModerationEvent, 0, 100),
		enableAsync: true,
	}

	// 加载内置规则集
	m.LoadBuiltinRuleSets()
	return m
}

// LoadBuiltinRuleSets 加载内置规则集
func (m *Moderator) LoadBuiltinRuleSets() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ruleSets["sensitive_words"] = &RuleSet{
		Name:        "敏感词过滤",
		Description: "政治敏感词、违法内容检测",
		Version:     "1.0.0",
		Action:      ActionQuarantine,
		Enabled:     true,
		Rules: []Rule{
			{ID: "SENS-001", Name: "政治敏感", Type: "keyword", Pattern: "涉政关键词", Severity: 10, Action: ActionQuarantine, Enabled: true},
		},
	}

	m.ruleSets["spam_detection"] = &RuleSet{
		Name:        "垃圾内容检测",
		Description: "广告、引流、刷屏等垃圾内容",
		Version:     "1.0.0",
		Action:      ActionReject,
		Enabled:     true,
		Rules: []Rule{
			{ID: "SPAM-001", Name: "外部链接过多", Type: "regex", Pattern: `(https?:\/\/\S+){3,}`, Severity: 7, Action: ActionReview, Enabled: true},
			{ID: "SPAM-002", Name: "联系方式", Type: "regex", Pattern: `(微信|wx|QQ|电话)[^\s]{2,10}`, Severity: 8, Action: ActionReject, Enabled: true},
			{ID: "SPAM-003", Name: "广告关键词", Type: "keyword", Pattern: "代写|代做|兼职|贷款|彩票|刷单|推广|加盟|投资回报", Severity: 7, Action: ActionReview, Enabled: true},
			{ID: "SPAM-004", Name: "连续重复字符", Type: "regex", Pattern: `(.)\1{6,}`, Severity: 5, Action: ActionReview, Enabled: true},
		},
	}

	m.ruleSets["content_quality"] = &RuleSet{
		Name:        "内容质量检测",
		Description: "内容质量基本检查",
		Version:     "1.0.0",
		Action:      ActionReview,
		Enabled:     true,
		Rules: []Rule{
			{ID: "QUAL-001", Name: "内容过短", Type: "pattern", Pattern: "min_length:5", Severity: 3, Action: ActionReview, Enabled: true},
			{ID: "QUAL-002", Name: "内容过长", Type: "pattern", Pattern: "max_length:10000", Severity: 2, Action: ActionReview, Enabled: true},
			{ID: "QUAL-003", Name: "全大写", Type: "regex", Pattern: `^[A-Z\s\d\W]{10,}$`, Severity: 3, Action: ActionReview, Enabled: true},
		},
	}

	m.ruleSets["toxic_content"] = &RuleSet{
		Name:        "有毒内容检测",
		Description: "辱骂、人身攻击、仇恨言论",
		Version:     "1.0.0",
		Action:      ActionReject,
		Enabled:     true,
		Rules: []Rule{
			{ID: "TOXIC-001", Name: "辱骂性词汇", Type: "regex", Pattern: `(蠢|傻|笨蛋|白痴|脑残|垃圾人渣|滚蛋|闭嘴|去死)`, Severity: 9, Action: ActionReject, Enabled: true},
			{ID: "TOXIC-002", Name: "人身攻击", Type: "regex", Pattern: `(你是|你就是|真是个).*(傻|笨|垃圾|废物|菜鸟)`, Severity: 8, Action: ActionReject, Enabled: true},
		},
	}

	slog.Info("Moderator loaded builtin rule sets", "count", len(m.ruleSets))
}

// ========== 核心审核方法 ==========

// Moderate 审核内容（同步，完整流程）
func (m *Moderator) Moderate(ctx context.Context, contentType string, content string, contentID uint) (*ModerationResult, error) {
	startTime := time.Now()
	result := &ModerationResult{
		Passed:      true,
		Action:      ActionAllow,
		Score:       100,
		Violations:  make([]Violation, 0),
		ProcessedAt: time.Now(),
		Metadata:    make(map[string]string),
	}

	m.mu.RLock()
	ruleSets := m.ruleSets
	m.mu.RUnlock()

	// 第一遍: 快速规则扫描
	for setName, rs := range ruleSets {
		if !rs.Enabled {
			continue
		}
		violations := m.evaluateRuleSet(rs, content)
		if len(violations) > 0 {
			result.Violations = append(result.Violations, violations...)
			for _, v := range violations {
				result.Score -= v.Severity * 3
			}
			// 高严重度规则立即触发动作
			if rs.Action == ActionReject || rs.Action == ActionQuarantine {
				result.Passed = false
				result.Action = rs.Action
				result.Duration = time.Since(startTime)
				m.logAudit(contentType, contentID, result, "system")
				return result, nil
			}
		}
		result.Metadata["ruleset_checked_"+setName] = "ok"
	}

	// 分数下限保护
	if result.Score < 0 {
		result.Score = 0
	}

	// 第二遍: LLM深度审核（如果启用且有LLM）
	if m.enableAsync && m.llmRouter != nil && result.Score < 80 {
		aiViolations, err := m.aiDeepReview(ctx, contentType, content)
		if err != nil {
			slog.Warn("Moderator AI deep review failed", "error", err)
			// LLM失败不影响规则审核结果
		} else if len(aiViolations) > 0 {
			result.Violations = append(result.Violations, aiViolations...)
			result.Score -= 20
			if result.Score < 0 {
				result.Score = 0
			}
		}
	}

	// 最终判定
	if result.Score < 40 {
		result.Passed = false
		result.Action = ActionReject
	} else if result.Score < 70 {
		result.Passed = false
		result.Action = ActionReview
	} else if len(result.Violations) > 0 {
		result.Action = ActionReview
	}

	result.Duration = time.Since(startTime)
	m.logAudit(contentType, contentID, result, "system")
	return result, nil
}

// ModerateAsync 异步审核（适合非实时场景）
func (m *Moderator) ModerateAsync(contentType string, content string, contentID uint, callback func(*ModerationResult)) {
	task := &ReviewTask{
		ID:        fmt.Sprintf("review_%d_%d", contentID, time.Now().UnixNano()),
		Content:   content,
		Type:      contentType,
		ContentID: contentID,
		Callback:  callback,
		CreatedAt: time.Now(),
	}

	go func() {
		defer recoverPanic("moderate_async")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := m.Moderate(ctx, task.Content, task.Type, task.ContentID)
		if err != nil {
			result = &ModerationResult{
				Passed:      false,
				Action:      ActionReview,
				Score:       0,
				Violations:  []Violation{{Message: fmt.Sprintf("审核异常: %v", err)}},
				ProcessedAt: time.Now(),
			}
		}

		if task.Callback != nil {
			task.Callback(result)
		}
	}()

	slog.Debug("Async moderation task submitted", "task_id", task.ID)
}

// ========== 规则评估 ==========

func (m *Moderator) evaluateRuleSet(rs *RuleSet, content string) []Violation {
	var violations []Violation
	for i := range rs.Rules {
		rule := &rs.Rules[i]
		if !rule.Enabled {
			continue
		}

		matched := m.matchRule(rule, content)
		if matched != "" {
			rule.MatchCount++ // 运行时统计
			violations = append(violations, Violation{
				RuleID:      rule.ID,
				RuleName:    rule.Name,
				Category:    rs.Name,
				Severity:    rule.Severity,
				MatchedText: truncateText(matched, 100),
				Action:      rule.Action,
				Message:     fmt.Sprintf("触犯规则 [%s]: %s", rule.Name, truncateText(matched, 50)),
			})
		}
	}
	return violations
}

func (m *Moderator) matchRule(rule *Rule, content string) string {
	switch rule.Type {
	case "keyword":
		// 关键词匹配（支持 | 分隔的多关键词）
		keywords := strings.Split(rule.Pattern, "|")
		for _, kw := range keywords {
			kw = strings.TrimSpace(kw)
			if kw == "" {
				continue
			}
			if strings.Contains(content, kw) {
				return kw
			}
		}
		return ""

	case "regex":
		re, err := regexp.Compile(rule.Pattern)
		if err != nil {
			slog.Warn("Invalid regex rule", "rule_id", rule.ID, "pattern", rule.Pattern, "error", err)
			return ""
		}
		loc := re.FindStringIndex(content)
		if loc != nil {
			start := loc[0] - 20
			end := loc[1] + 20
			if start < 0 {
				start = 0
			}
			if end > len(content) {
				end = len(content)
			}
			return content[start:end]
		}
		return ""

	case "pattern":
		switch {
		case strings.HasPrefix(rule.Pattern, "min_length:"):
			var minLen int
			fmt.Sscanf(rule.Pattern, "min_length:%d", &minLen)
			if len([]rune(content)) < minLen {
				return fmt.Sprintf("内容长度%d < 最低要求%d", len([]rune(content)), minLen)
			}
		case strings.HasPrefix(rule.Pattern, "max_length:"):
			var maxLen int
			fmt.Sscanf(rule.Pattern, "max_length:%d", &maxLen)
			if len([]rune(content)) > maxLen {
				return fmt.Sprintf("内容长度%d > 最大限制%d", len([]rune(content)), maxLen)
			}
		}
		return ""

	default:
		return ""
	}
}

// ========== LLM深度审核 ==========

func (m *Moderator) aiDeepReview(ctx context.Context, contentType string, content string) ([]Violation, error) {
	prompt := fmt.Sprintf(`请对以下内容进行深度合规审核:

## 类型: %s
## 内容:
%s

## 审核维度:
1. **法律合规**: 是否违反法律法规
2. **平台规范**: 是否违反社区准则
3. **价值观**: 是否传播不良价值观
4. **误导性**: 是否含有虚假或误导性信息

## 输出格式(JSON):
[
  {
    "category": "违规类别(sensitive/spam/ad/toxic/inappropriate/misleading)",
    "severity": 1-10,
    "matchedText": "涉及的具体内容片段",
    "message": "详细说明为什么违规"
  }
]

如果内容没有问题，返回空数组 []`,
		contentType, truncateText(content, 3000),
	)

	messages := []ChatMessage{
		{Role: "system", Content: "你是一位专业的内容审核专家，熟悉中国互联网内容管理法规和各平台社区规范。审核标准严格但公正。"},
		{Role: "user", Content: prompt},
	}

	reply, _, err := m.llmRouter.Chat(ctx, messages)
	if err != nil {
		return nil, err
	}

	var aiViolations []Violation
	if err := extractJSONFromText(reply, &aiViolations); err != nil {
		slog.Warn("Failed to parse moderation AI response", "error", err)
		// 尝试从文本中提取关键信息
		if strings.Contains(reply, "不合规") || strings.Contains(reply, "违规") || strings.Contains(reply, "有问题") {
			return []Violation{{
				Category: "ai_detected",
				Severity: 5,
				Message:  truncateText(reply, 200),
				Action:   ActionReview,
			}}, nil
		}
		return nil, nil
	}

	// 补充AI标记
	for i := range aiViolations {
		aiViolations[i].RuleID = "AI-" + fmt.Sprintf("%03d", i+1)
		aiViolations[i].Action = ActionReview
	}

	return aiViolations, nil
}

// ========== 审计日志 ==========

func (m *Moderator) logAudit(contentType string, contentID uint, result *ModerationResult, operator string) {
	event := ModerationEvent{
		Timestamp:   time.Now(),
		ContentType: contentType,
		ContentID:   contentID,
		Result:      result,
		Operator:    operator,
	}

	m.mu.Lock()
	m.auditLog = append(m.auditLog, event)
	// 保留最近1000条
	if len(m.auditLog) > 1000 {
		m.auditLog = m.auditLog[len(m.auditLog)-1000:]
	}
	m.mu.Unlock()
}

// GetAuditLog 获取审计日志
func (m *Moderator) GetAuditLog(limit int) []ModerationEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 || limit > len(m.auditLog) {
		limit = len(m.auditLog)
	}
	start := len(m.auditLog) - limit
	if start < 0 {
		start = 0
	}
	result := make([]ModerationEvent, limit)
	copy(result, m.auditLog[start:])
	return result
}

// GetStats 获取审核统计
func (m *Moderator) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	totalReviews := 0
	passedCount := 0
	rejectedCount := 0
	reviewCount := 0

	for _, event := range m.auditLog {
		totalReviews++
		switch event.Result.Action {
		case ActionAllow:
			passedCount++
		case ActionReject:
			rejectedCount++
		case ActionReview, ActionQuarantine:
			reviewCount++
		}
	}

	ruleStats := make(map[string]interface{})
	for name, rs := range m.ruleSets {
		totalMatches := int64(0)
		for _, r := range rs.Rules {
			totalMatches += r.MatchCount
		}
		ruleStats[name] = map[string]interface{}{
			"rules":     len(rs.Rules),
			"total_matches": totalMatches,
			"enabled":   rs.Enabled,
		}
	}

	return map[string]interface{}{
		"total_reviews": totalReviews,
		"passed":        passedCount,
		"rejected":      rejectedCount,
		"pending_review": reviewCount,
		"pass_rate":     float64(passedCount) / float64(maxInt(totalReviews, 1)),
		"rule_sets":     ruleStats,
		"log_size":      len(m.auditLog),
	}
}

// ========== 动态规则管理 ==========

// AddRule 动态添加规则
func (m *Moderator) AddRule(setName string, rule Rule) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	rs, ok := m.ruleSets[setName]
	if !ok {
		return fmt.Errorf("rule set '%s' not found", setName)
	}
	rs.Rules = append(rs.Rules, rule)
	slog.Info("Rule added", "set", setName, "rule_id", rule.ID, "name", rule.Name)
	return nil
}

// UpdateRuleStatus 更新规则状态
func (m *Moderator) UpdateRuleStatus(ruleID string, enabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, rs := range m.ruleSets {
		for i := range rs.Rules {
			if rs.Rules[i].ID == ruleID {
				rs.Rules[i].Enabled = enabled
				slog.Info("Rule status updated", "rule_id", ruleID, "enabled", enabled)
				return nil
			}
		}
	}
	return fmt.Errorf("rule '%s' not found", ruleID)
}

// EnableRuleSet 启用/禁用整个规则集
func (m *Moderator) EnableRuleSet(setName string, enabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	rs, ok := m.ruleSets[setName]
	if !ok {
		return fmt.Errorf("rule set '%s' not found", setName)
	}
	rs.Enabled = enabled
	slog.Info("Rule set status updated", "set", setName, "enabled", enabled)
	return nil
}

// ========== DTO转换 ==========

func (r *ModerationResult) ToDTO() *dto.ModerationDTO {
	violations := make([]dto.ViolationDTO, 0, len(r.Violations))
	for _, v := range r.Violations {
		violations = append(violations, dto.ViolationDTO{
			RuleID:      v.RuleID,
			RuleName:    v.RuleName,
			Category:    v.Category,
			Severity:    v.Severity,
			MatchedText: v.MatchedText,
			Action:      string(v.Action),
			Message:     v.Message,
		})
	}

	return &dto.ModerationDTO{
		Passed:      r.Passed,
		Action:      string(r.Action),
		Score:       r.Score,
		Violations:  violations,
		DurationMs:  r.Duration.Milliseconds(),
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
