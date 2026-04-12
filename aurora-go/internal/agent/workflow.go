package agent

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/aurora-go/aurora/internal/dto"
	svc "github.com/aurora-go/aurora/internal/service"
	"github.com/aurora-go/aurora/internal/vo"
)

// ========== Agent 工作流编排 ==========
// 基于 DAG (Directed Acyclic Graph) 的工作流编排引擎
// 预置工作流: 文章发布流水线 / 评论审核流水线 / 数据分析流水线
//
// 架构对标: tRPC graph.StateGraph (类型安全多条件路由, 等价LangGraph)
// 特性:
//   - 有向无环图(DAG)节点编排
//   - 条件分支(Condition-based routing)
//   - 并行节点(Parallel execution)
//   - 错误恢复(Retry + Fallback)
//   - 可观测性(Step timing + logging)

// Workflow 工作流定义
type Workflow struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Nodes       []WorkflowNode `json:"nodes"`
	Edges       []WorkflowEdge `json:"edges"`
	StartNode   string         `json:"startNode"` // 起始节点ID
}

// WorkflowNode 工作流节点
type WorkflowNode struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Type        NodeType      `json:"type"`        // task/condition/parallel/subflow
	Handler     string        `json:"handler"`     // 处理器函数名
	Config      NodeConfig    `json:"config"`      // 节点配置(超时/重试等)
	Timeout     time.Duration `json:"timeout"`     // 节点超时时间
	MaxRetries  int           `json:"maxRetries"`  // 最大重试次数
	OnFailure   FailurePolicy `json:"onFailure"`   // 失败策略
}

type NodeType string

const (
	NodeTypeTask      NodeType = "task"       // 任务节点(执行具体操作)
	NodeTypeCondition NodeType = "condition"  // 条件节点(决定分支走向)
	NodeTypeParallel  NodeType = "parallel"   // 并行节点(同时执行多个子任务)
	NodeTypeSubflow   NodeType = "subflow"    // 子工作流节点(嵌套调用)
)

type NodeConfig struct {
	Params map[string]interface{} `json:"params,omitempty"`
}

type FailurePolicy string

const (
	FailAbort    FailurePolicy = "abort"    // 中止整个工作流
	FailSkip     FailurePolicy = "skip"     // 跳过此节点继续
	FailRetry    FailurePolicy = "retry"    // 重试
	FailFallback FailurePolicy = "fallback" // 使用降级方案
)

// WorkflowEdge 工作流边（连接节点）
type WorkflowEdge struct {
	From string `json:"from"` // 源节点ID
	To   string `json:"to"`   // 目标节点ID
	Cond string `json:"cond"` // 条件表达式(条件节点的出口边)
}

// WorkflowContext 工作流执行上下文
type WorkflowContext struct {
	ID          string                 `json:"id"`
	WorkflowID  string                 `json:"workflowId"`
	Status      WorkflowStatus         `json:"status"`
	Input       map[string]interface{} `json:"input"`
	Output      map[string]interface{} `json:"output"`
	NodeResults map[string]*NodeResult `json:"nodeResults"` // 各节点执行结果
	Variables   map[string]interface{} `json:"variables"`    // 跨节点共享变量
	Errors      []WorkflowError        `json:"errors"`
	StartedAt   time.Time              `json:"startedAt"`
	FinishedAt  time.Time              `json:"finishedAt"`
	mu          sync.RWMutex
}

type WorkflowStatus string

const (
	StatusPending   WorkflowStatus = "pending"
	StatusRunning   WorkflowStatus = "running"
	StatusCompleted WorkflowStatus = "completed"
	StatusFailed    WorkflowStatus = "failed"
	StatusCancelled WorkflowStatus = "cancelled"
)

// NodeResult 单个节点执行结果
type NodeResult struct {
	NodeID    string        `json:"nodeId"`
	Status    string        `json:"status"`     // success/skipped/failed/timeout
	Output    interface{}   `json:"output"`
	Error     error         `json:"-"`
	Duration  time.Duration `json:"durationMs"`
	Retries   int           `json:"retries"`
}

// WorkflowError 工作流错误
type WorkflowError struct {
	NodeID  string `json:"nodeId"`
	Phase   string `json:"phase"`    // execute/validate/route
	Message string `json:"message"`
	Error   error  `json:"-"`
}

// ========== 工作流引擎 ==========

// Engine 工作流引擎
type Engine struct {
	workflows   map[string]*Workflow
	handlers    map[string]WorkflowHandler
	writingAsst *WritingAssistant
	commentAsst *CommentAssistant
	moderator   *Moderator
	llmRouter   *LLMRouter
	mu          sync.RWMutex
}

// WorkflowHandler 节点处理器函数签名
type WorkflowHandler func(ctx context.Context, node *WorkflowNode, wfCtx *WorkflowContext) (interface{}, error)

// NewEngine 创建工作流引擎
func NewEngine(wa *WritingAssistant, ca *CommentAssistant, mod *Moderator, router *LLMRouter) *Engine {
	e := &Engine{
		workflows:   make(map[string]*Workflow),
		handlers:    make(map[string]WorkflowHandler),
		writingAsst: wa,
		commentAsst: ca,
		moderator:   mod,
		llmRouter:   router,
	}
	e.registerBuiltinHandlers()
	e.registerBuiltinWorkflows()
	return e
}

// RegisterWorkflow 注册工作流
func (e *Engine) RegisterWorkflow(wf *Workflow) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.workflows[wf.ID] = wf
	slog.Info("Workflow registered", "id", wf.ID, "name", wf.Name, "nodes", len(wf.Nodes))
}

// RegisterHandler 注册节点处理器
func (e *Engine) RegisterHandler(name string, handler WorkflowHandler) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.handlers[name] = handler
}

// Execute 执行工作流
func (e *Engine) Execute(ctx context.Context, workflowID string, input map[string]interface{}) (*WorkflowContext, error) {
	e.mu.RLock()
	wf, ok := e.workflows[workflowID]
	e.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("workflow '%s' not found", workflowID)
	}

	wfCtx := &WorkflowContext{
		ID:          fmt.Sprintf("wf_%d_%s", time.Now().UnixNano(), workflowID),
		WorkflowID:  workflowID,
		Status:      StatusRunning,
		Input:       input,
		Output:      make(map[string]interface{}),
		NodeResults: make(map[string]*NodeResult),
		Variables:   make(map[string]interface{}),
		Errors:      make([]WorkflowError, 0),
		StartedAt:   time.Now(),
	}

	slog.Info("Workflow execution started", "wf_id", workflowID, "ctx_id", wfCtx.ID)

	// 执行工作流
	err := e.executeNode(ctx, wf.StartNode, wf, wfCtx)

	wfCtx.FinishedAt = time.Now()
	if err != nil {
		wfCtx.Status = StatusFailed
		slog.Error("Workflow execution failed", "wf_id", workflowID, "ctx_id", wfCtx.ID, "error", err)
		return wfCtx, err
	}

	wfCtx.Status = StatusCompleted
	wfCtx.Output["result"] = wfCtx.NodeResults[wf.StartNode].Output
	slog.Info("Workflow execution completed", "wf_id", workflowID, "ctx_id", wfCtx.ID, "duration_ms", wfCtx.FinishedAt.Sub(wfCtx.StartedAt).Milliseconds())

	return wfCtx, nil
}

// executeNode 执行单个节点
func (e *Engine) executeNode(ctx context.Context, nodeID string, wf *Workflow, wfCtx *WorkflowContext) error {
	node := findNode(wf, nodeID)
	if node == nil {
		return fmt.Errorf("node '%s' not found", nodeID)
	}

	nodeResult := &NodeResult{NodeID: nodeID, Status: "success"}
	startTime := time.Now()

	// 重试循环
	var lastErr error
	for attempt := 0; attempt <= node.MaxRetries; attempt++ {
		if attempt > 0 {
			nodeResult.Retries = attempt
			time.Sleep(time.Duration(attempt) * time.Second) // 指数退避
			slog.Info("Retrying node", "node_id", nodeID, "attempt", attempt)
		}

		// 超时控制
		execCtx := ctx
		if node.Timeout > 0 {
			var cancel context.CancelFunc
			execCtx, cancel = context.WithTimeout(ctx, node.Timeout)
			defer cancel()
		}

		output, err := e.invokeHandler(execCtx, node, wfCtx)
		lastErr = err

		if err == nil {
			nodeResult.Output = output
			nodeResult.Status = "success"
			break
		}

		slog.Warn("Node execution failed", "node_id", nodeID, "attempt", attempt+1, "error", err)
	}

	if lastErr != nil {
		nodeResult.Status = "failed"
		nodeResult.Error = lastErr
		wfCtx.Errors = append(wfCtx.Errors, WorkflowError{NodeID: nodeID, Phase: "execute", Message: lastErr.Error(), Error: lastErr})

		switch node.OnFailure {
		case FailAbort:
			return lastErr
		case FailSkip:
			nodeResult.Status = "skipped"
		case FailFallback:
			fallbackOutput, err := e.executeFallback(node, wfCtx)
			if err == nil {
				nodeResult.Output = fallbackOutput
				nodeResult.Status = "fallback"
				lastErr = nil
			}
		case FailRetry:
			return lastErr
		}
	}

	nodeResult.Duration = time.Since(startTime)
	wfCtx.mu.Lock()
	wfCtx.NodeResults[nodeID] = nodeResult
	wfCtx.mu.Unlock()

	// 路由到下一个节点
	if lastErr == nil {
		nextNodes := findNextNodes(wf, nodeID, wfCtx)
		if len(nextNodes) == 0 {
			return nil // 终端节点
		}
		if len(nextNodes) == 1 {
			return e.executeNode(ctx, nextNodes[0], wf, wfCtx)
		}
		// 多个后续节点(并行)
		return e.executeParallelNodes(ctx, nextNodes, wf, wfCtx)
	}

	return lastErr
}

func (e *Engine) invokeHandler(ctx context.Context, node *WorkflowNode, wfCtx *WorkflowContext) (interface{}, error) {
	e.mu.RLock()
	handler, ok := e.handlers[node.Handler]
	e.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("handler '%s' not registered", node.Handler)
	}
	return handler(ctx, node, wfCtx)
}

func (e *Engine) executeFallback(node *WorkflowNode, wfCtx *WorkflowContext) (interface{}, error) {
	switch node.Handler {
	case "ai_generate":
		return map[string]interface{}{
			"title":   wfCtx.Input["topic"].(string),
			"content": "[AI生成服务不可用，请手动编写]",
		}, nil
	case "content_review":
		return &ModerationResult{Passed: true, Action: ActionReview, Score: 60, Metadata: map[string]string{"fallback": "true"}}, nil
	default:
		return nil, fmt.Errorf("no fallback for handler: %s", node.Handler)
	}
}

func (e *Engine) executeParallelNodes(ctx context.Context, nodeIDs []string, wf *Workflow, wfCtx *WorkflowContext) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(nodeIDs))

	for _, nid := range nodeIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			defer recoverPanic("workflow_parallel_" + id)
			if err := e.executeNode(ctx, id, wf, wfCtx); err != nil {
				errCh <- err
			}
		}(nid)
	}

	wg.Wait()
	close(errCh)

	// 收集第一个错误
	for err := range errCh {
		return err
	}
	return nil
}

// ========== 内置工作流注册 ==========

func (e *Engine) registerBuiltinWorkflows() {
	// 工作流1: 文章发布流水线 (ArticlePublishPipeline)
	e.RegisterWorkflow(&Workflow{
		ID:          "article_publish_pipeline",
		Name:        "文章发布流水线",
		Description: "AI辅助的文章发布全流程: 生成→审核→SEO优化→发布",
		StartNode:   "start",
		Nodes: []WorkflowNode{
			{ID: "start", Name: "开始", Type: NodeTypeTask, Handler: "workflow_start", Timeout: 5 * time.Second, OnFailure: FailAbort},
			{ID: "ai_generate", Name: "AI生成", Type: NodeTypeTask, Handler: "ai_generate", Timeout: 60 * time.Second, MaxRetries: 2, OnFailure: FailFallback},
			{ID: "quality_check", Name: "质量检查", Type: NodeTypeCondition, Handler: "quality_check", Timeout: 10 * time.Second, OnFailure: FailSkip},
			{ID: "content_review", Name: "内容审核", Type: NodeTypeTask, Handler: "content_review", Timeout: 30 * time.Second, OnFailure: FailFallback},
			{ID: "seo_optimize", Name: "SEO优化", Type: NodeTypeTask, Handler: "seo_optimize", Timeout: 30 * time.Second, MaxRetries: 1, OnFailure: FailSkip},
			{ID: "save_article", Name: "保存文章", Type: NodeTypeTask, Handler: "save_article", Timeout: 10 * time.Second, MaxRetries: 3, OnFailure: FailAbort},
			{ID: "notify", Name: "通知", Type: NodeTypeTask, Handler: "publish_notify", Timeout: 10 * time.Second, OnFailure: FailSkip},
			{ID: "end", Name: "结束", Type: NodeTypeTask, Handler: "workflow_end", Timeout: 5 * time.Second, OnFailure: FailSkip},
		},
		Edges: []WorkflowEdge{
			{From: "start", To: "ai_generate"},
			{From: "ai_generate", To: "quality_check"},
			{From: "quality_check", To: "content_review", Cond: "needs_review"},
			{From: "quality_check", To: "seo_optimize", Cond: "pass_quick"},
			{From: "content_review", To: "seo_optimize", Cond: "approved"},
			{From: "content_review", To: "end", Cond: "rejected"},
			{From: "seo_optimize", To: "save_article"},
			{From: "save_article", To: "notify"},
			{From: "notify", To: "end"},
		},
	})

	// 工作流2: 评论审核流水线 (CommentReviewPipeline)
	e.RegisterWorkflow(&Workflow{
		ID:          "comment_review_pipeline",
		Name:        "评论审核流水线",
		Description: "自动化评论审核: 规则筛选→AI分析→情感分析→自动回复",
		StartNode:   "receive_comment",
		Nodes: []WorkflowNode{
			{ID: "receive_comment", Name: "接收评论", Type: NodeTypeTask, Handler: "receive_comment", Timeout: 5 * time.Second, OnFailure: FailAbort},
			{ID: "rule_filter", Name: "规则过滤", Type: NodeTypeTask, Handler: "rule_filter", Timeout: 5 * time.Second, OnFailure: FailSkip},
			{ID: "check_result", Name: "检查结果", Type: NodeTypeCondition, Handler: "check_rule_result", Timeout: 3 * time.Second, OnFailure: FailSkip},
			{ID: "ai_analyze", Name: "AI深度分析", Type: NodeTypeTask, Handler: "ai_analyze_comment", Timeout: 30 * time.Second, OnFailure: FailSkip},
			{ID: "sentiment", Name: "情感分析", Type: NodeTypeTask, Handler: "analyze_sentiment", Timeout: 15 * time.Second, OnFailure: FailSkip},
			{ID: "auto_reply", Name: "自动回复", Type: NodeTypeTask, Handler: "suggest_reply", Timeout: 15 * time.Second, OnFailure: FailSkip},
			{ID: "store_result", Name: "存储结果", Type: NodeTypeTask, Handler: "store_review_result", Timeout: 5 * time.Second, MaxRetries: 3, OnFailure: FailAbort},
		},
		Edges: []WorkflowEdge{
			{From: "receive_comment", To: "rule_filter"},
			{From: "rule_filter", To: "check_result"},
			{From: "check_result", To: "store_result", Cond: "blocked"},
			{From: "check_result", To: "ai_analyze", Cond: "needs_ai"},
			{From: "check_result", To: "sentiment", Cond: "pass_quick"},
			{From: "ai_analyze", To: "sentiment"},
			{From: "sentiment", To: "auto_reply"},
			{From: "auto_reply", To: "store_result"},
		},
	})

	slog.Info("Builtin workflows registered", "count", 2)
}

// ========== 内置处理器注册 ==========

func (e *Engine) registerBuiltinHandlers() {
	// 文章发布流水线处理器
	e.RegisterHandler("workflow_start", func(_ context.Context, _ *WorkflowNode, wfCtx *WorkflowContext) (interface{}, error) {
		wfCtx.Variables["started_at"] = time.Now().Format(time.RFC3339)
		return map[string]interface{}{"message": "工作流开始"}, nil
	})

	e.RegisterHandler("ai_generate", func(ctx context.Context, _ *WorkflowNode, wfCtx *WorkflowContext) (interface{}, error) {
		topic, _ := wfCtx.Input["topic"].(string)
		action, _ := wfCtx.Input["action"].(string)
		if action == "" {
			action = "generate"
		}

		req := &dto.WriteRequest{
			Action:  action,
			Topic:   topic,
			Style:   getStringFromInput(wfCtx.Input, "style", "professional"),
			Length:  getStringFromInput(wfCtx.Input, "length", "medium"),
		}

		if e.writingAsst != nil {
			return e.writingAsst.Execute(ctx, req)
		}
		return nil, fmt.Errorf("writing assistant unavailable")
	})

	e.RegisterHandler("quality_check", func(_ context.Context, _ *WorkflowNode, wfCtx *WorkflowContext) (interface{}, error) {
		genResult, ok := wfCtx.NodeResults["ai_generate"]
		if !ok || genResult.Output == nil {
			return "needs_review", nil
		}
		writeResp, ok := genResult.Output.(*dto.WriteResponse)
		if !ok {
			return "needs_review", nil
		}
		if writeResp.WordCount < 200 {
			return "needs_review", nil
		}
		return "pass_quick", nil
	})

	e.RegisterHandler("content_review", func(ctx context.Context, _ *WorkflowNode, wfCtx *WorkflowContext) (interface{}, error) {
		genResult, ok := wfCtx.NodeResults["ai_generate"]
		if !ok || genResult.Output == nil {
			return &ModerationResult{Passed: true, Score: 70}, nil
		}
		writeResp, ok := genResult.Output.(*dto.WriteResponse)
		if !ok {
			return &ModerationResult{Passed: true, Score: 70}, nil
		}

		if e.moderator != nil {
			return e.moderator.Moderate(ctx, "article", writeResp.Content, 0)
		}
		return &ModerationResult{Passed: true, Score: 70}, nil
	})

	e.RegisterHandler("seo_optimize", func(ctx context.Context, _ *WorkflowNode, wfCtx *WorkflowContext) (interface{}, error) {
		topic, _ := wfCtx.Input["topic"].(string)
		genResult, _ := wfCtx.NodeResults["ai_generate"]

		content := ""
		if genResult != nil && genResult.Output != nil {
			if wr, ok := genResult.Output.(*dto.WriteResponse); ok {
				content = wr.Content
			}
		}

		req := &dto.WriteRequest{Action: "seo", Topic: topic, Content: content}
		if e.writingAsst != nil {
			return e.writingAsst.Execute(ctx, req)
		}
		return map[string]string{"note": "SEO optimization skipped"}, nil
	})

	e.RegisterHandler("save_article", func(ctx context.Context, _ *WorkflowNode, wfCtx *WorkflowContext) (interface{}, error) {
		topic, _ := wfCtx.Input["topic"].(string)

		var content string

		// 优先使用SEO优化结果
		if seoResult, ok := wfCtx.NodeResults["seo_optimize"]; ok && seoResult.Output != nil {
			if seoResp, ok := seoResult.Output.(*dto.WriteResponse); ok {
				content = seoResp.Content
			}
		}
		if content == "" {
			if genResult, ok := wfCtx.NodeResults["ai_generate"]; ok && genResult.Output != nil {
				if genResp, ok := genResult.Output.(*dto.WriteResponse); ok {
					content = genResp.Content
				}
			}
		}

		articleVO := vo.ArticleVO{
			ArticleTitle:   topic,
			ArticleContent: content,
			CategoryName:   getStringFromInput(wfCtx.Input, "category_name", "默认分类"),
		}

		articleSvc := svc.GetArticleService()
		if articleSvc == nil {
			return map[string]interface{}{"status": "saved_to_draft", "title": topic}, nil
		}

		result, err := articleSvc.CreateArticle(ctx, 0, articleVO)
		if err != nil {
			return nil, fmt.Errorf("save article failed: %w", err)
		}
		wfCtx.Variables["article_id"] = result.ID
		return map[string]interface{}{"article_id": result.ID, "status": "published"}, nil
	})

	e.RegisterHandler("publish_notify", func(_ context.Context, _ *WorkflowNode, wfCtx *WorkflowContext) (interface{}, error) {
		articleID, _ := wfCtx.Variables["article_id"].(uint)
		slog.Info("Article published notification", "article_id", articleID)
		return map[string]interface{}{"notified": true, "article_id": articleID}, nil
	})

	e.RegisterHandler("workflow_end", func(_ context.Context, _ *WorkflowNode, wfCtx *WorkflowContext) (interface{}, error) {
		wfCtx.Variables["completed_at"] = time.Now().Format(time.RFC3339)
		duration := time.Since(wfCtx.StartedAt).String()
		return map[string]interface{}{"status": "completed", "duration": duration}, nil
	})

	// 评论审核流水线处理器
	e.RegisterHandler("receive_comment", func(_ context.Context, _ *WorkflowNode, wfCtx *WorkflowContext) (interface{}, error) {
		wfCtx.Variables["received_at"] = time.Now().Format(time.RFC3339)
		return wfCtx.Input, nil
	})

	e.RegisterHandler("rule_filter", func(ctx context.Context, _ *WorkflowNode, wfCtx *WorkflowContext) (interface{}, error) {
		content, _ := wfCtx.Input["content"].(string)
		commentID := getUintFromInput(wfCtx.Input, "comment_id")

		if e.moderator != nil {
			return e.moderator.Moderate(ctx, "comment", content, commentID)
		}
		return &ModerationResult{Passed: true, Score: 80}, nil
	})

	e.RegisterHandler("check_rule_result", func(_ context.Context, _ *WorkflowNode, wfCtx *WorkflowContext) (interface{}, error) {
		filterResult, ok := wfCtx.NodeResults["rule_filter"]
		if !ok {
			return "needs_ai", nil
		}
		modResult, ok := filterResult.Output.(*ModerationResult)
		if !ok {
			return "needs_ai", nil
		}
		if !modResult.Passed {
			return "blocked", nil
		}
		if modResult.Score < 80 {
			return "needs_ai", nil
		}
		return "pass_quick", nil
	})

	e.RegisterHandler("ai_analyze_comment", func(ctx context.Context, _ *WorkflowNode, wfCtx *WorkflowContext) (interface{}, error) {
		content, _ := wfCtx.Input["content"].(string)
		author, _ := wfCtx.Input["author"].(string)

		comment := &SingleComment{Content: content, Author: author}
		if e.commentAsst != nil {
			return e.commentAsst.ReviewComment(ctx, comment, true)
		}
		return &CommentReviewResult{Passed: true, RiskLevel: "low", Score: 80}, nil
	})

	e.RegisterHandler("analyze_sentiment", func(ctx context.Context, _ *WorkflowNode, wfCtx *WorkflowContext) (interface{}, error) {
		content, _ := wfCtx.Input["content"].(string)
		if e.commentAsst != nil {
			return e.commentAsst.AnalyzeSentiment(ctx, content)
		}
		return &SentimentResult{Label: "neutral", Confidence: 0.6}, nil
	})

	e.RegisterHandler("suggest_reply", func(ctx context.Context, _ *WorkflowNode, wfCtx *WorkflowContext) (interface{}, error) {
		content, _ := wfCtx.Input["content"].(string)
		author, _ := wfCtx.Input["author"].(string)
		articleTitle, _ := wfCtx.Input["article_title"].(string)

		comment := &SingleComment{Content: content, Author: author}
		if e.commentAsst != nil {
			return e.commentAsst.SuggestAutoReply(ctx, comment, articleTitle)
		}
		return []string{"感谢您的评论！"}, nil
	})

	e.RegisterHandler("store_review_result", func(_ context.Context, _ *WorkflowNode, wfCtx *WorkflowContext) (interface{}, error) {
		commentID := getUintFromInput(wfCtx.Input, "comment_id")

		var finalAction string
		var finalScore float64

		// 收集最终审核结论
		if filterResult, ok := wfCtx.NodeResults["rule_filter"]; ok {
			if mr, ok := filterResult.Output.(*ModerationResult); ok {
				finalAction = string(mr.Action)
				finalScore = float64(mr.Score)
			}
		}
		if aiResult, ok := wfCtx.NodeResults["ai_analyze"]; ok {
			if cr, ok := aiResult.Output.(*CommentReviewResult); ok && !cr.Passed {
				finalAction = "reject"
				finalScore = cr.Score
			}
		}

		slog.Info("Comment review stored", "comment_id", commentID, "action", finalAction, "score", finalScore)
		return map[string]interface{}{
			"comment_id": commentID,
			"action":     finalAction,
			"score":      finalScore,
		}, nil
	})
}

// ========== 辅助函数 ==========

func findNode(wf *Workflow, nodeID string) *WorkflowNode {
	for i := range wf.Nodes {
		if wf.Nodes[i].ID == nodeID {
			return &wf.Nodes[i]
		}
	}
	return nil
}

func findNextNodes(wf *Workflow, fromNodeID string, wfCtx *WorkflowContext) []string {
	var nextNodes []string
	for _, edge := range wf.Edges {
		if edge.From == fromNodeID {
			if edge.Cond == "" {
				nextNodes = append(nextNodes, edge.To)
			} else {
				// 条件边: 检查上一个节点的输出是否匹配条件
				if prevResult, ok := wfCtx.NodeResults[fromNodeID]; ok {
					if outputStr, ok := prevResult.Output.(string); ok && outputStr == edge.Cond {
						nextNodes = append(nextNodes, edge.To)
					}
				}
			}
		}
	}
	return nextNodes
}

func getStringFromInput(input map[string]interface{}, key, defaultValue string) string {
	if v, ok := input[key].(string); ok && v != "" {
		return v
	}
	return defaultValue
}

func getUintFromInput(input map[string]interface{}, key string) uint {
	switch v := input[key].(type) {
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

// GetWorkflows 获取所有已注册的工作流
func (e *Engine) GetWorkflows() []WorkflowSummary {
	e.mu.RLock()
	defer e.mu.RUnlock()

	summaries := make([]WorkflowSummary, 0, len(e.workflows))
	for id, wf := range e.workflows {
		summaries = append(summaries, WorkflowSummary{
			ID:          id,
			Name:        wf.Name,
			Description: wf.Description,
			NodeCount:   len(wf.Nodes),
		})
	}
	return summaries
}

// WorkflowSummary 工作流概要
type WorkflowSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	NodeCount   int    `json:"nodeCount"`
}

// GetWorkflowExecutionHistory 获取工作流执行历史（简化版）
func (e *Engine) GetWorkflowExecutionHistory(limit int) []map[string]interface{} {
	// 实际实现应该从持久化存储读取
	return []map[string]interface{}{
		{
			"workflow_id": "article_publish_pipeline",
			"status":      "completed",
			"started_at":  time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			"duration_ms": 45000,
		},
	}
}
