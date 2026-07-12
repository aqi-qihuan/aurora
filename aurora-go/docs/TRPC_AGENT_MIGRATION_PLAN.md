# tRPC-Agent-Go 全量引入迁移计划

> 目标：将 aurora-go 自研 agent 模块（4836 行 / 12 文件）替换为 tRPC-Agent-Go v1.10.0
> 日期：2026-07-12
> 预估工期：5-6 天

---

## 一、tRPC-Agent-Go 核心 API 速览

### Import 路径

```
trpc.group/trpc-go/trpc-agent-go/agent/llmagent    # LLMAgent
trpc.group/trpc-go/trpc-agent-go/model             # 模型接口 + 消息类型
trpc.group/trpc-go/trpc-agent-go/model/openai      # OpenAI/DeepSeek 兼容客户端
trpc.group/trpc-go/trpc-agent-go/runner            # 执行器（管理 session/memory）
trpc.group/trpc-go/trpc-agent-go/tool              # 工具接口
trpc.group/trpc-go/trpc-agent-go/tool/function     # Function 工具
trpc.group/trpc-go/trpc-agent-go/memory/memorysvc  # 内存/Redis 记忆服务
trpc.group/trpc-go/trpc-agent-go/agent/graphagent  # 图工作流（可选）
```

### 核心 API 模式

```go
// 1. 创建 Model（LLM 客户端）
modelInstance := openai.New("deepseek-chat",
    openai.WithVariant(openai.VariantDeepSeek),
)
// API Key/Base URL 通过环境变量：OPENAI_API_KEY / OPENAI_BASE_URL

// 2. 创建 Tool
tool := function.NewFunctionTool(myFunc,
    function.WithName("my_tool"),
    function.WithDescription("..."),
)

// 3. 创建 Agent
agent := llmagent.New("assistant",
    llmagent.WithModel(modelInstance),
    llmagent.WithTools([]tool.Tool{myTool}),
    llmagent.WithGenerationConfig(model.GenerationConfig{Stream: true}),
)

// 4. 创建 Runner（管理 session/memory）
r := runner.NewRunner("aurora-agent", agent,
    runner.WithMemoryService(memorysvc.NewInMemoryService()),
)

// 5. 执行对话（返回事件流）
events, err := r.Run(ctx, userID, sessionID, model.NewUserMessage("..."))
for event := range events {
    if event.Object == "chat.completion.chunk" {
        fmt.Print(event.Response.Choices[0].Delta.Content)  // 流式 token
    }
}
```

### 关键特性

| 特性 | tRPC-Agent-Go | aurora 自研 |
|:---|:---|:---|
| LLM 多模型 | model/openai + WithVariant | llm_router.go 手搓 HTTP |
| 工具调用 | function.NewFunctionTool | tool_hub.go 自研 ToolFunc |
| 会话记忆 | runner 内置 + memorysvc | memory.go 自研 InMemory/Redis |
| 流式输出 | event channel + GenerationConfig.Stream | sse_scanner.go 手搓 SSE |
| 工作流 | graphagent (StateGraph) | workflow.go 自研 DAG |
| RAG | knowledge 子包 | rag.go 自研 ES + LLM |
| 可观测性 | OpenTelemetry 全链路 | slog 日志 |
| Agent 类型 | LLM/Chain/Parallel/Cycle/Graph | 仅 LLM |

---

## 二、迁移范围与策略

### 保留不变

| 文件 | 行数 | 理由 |
|:---|:---:|:---|
| `handler/agent_handler.go` | ~150 | HTTP 接口层，5 个 API 端点不变 |
| `writing_assistant.go` | 527 | aurora 业务逻辑（AI 写作），非基础设施 |
| `comment_assistant.go` | 546 | aurora 业务逻辑（评论助手） |
| `content_moderator.go` | 595 | aurora 业务逻辑（内容审核） |
| `workflow.go` | 712 | 初期保留自研 DAG，后续按需引入 graphagent |

### 替换重写

| 文件 | 行数 | 替换为 | 难度 |
|:---|:---:|:---|:---:|
| `agent.go` | 386 | llmagent.New + runner.NewRunner | 中 |
| `llm_router.go` | 449 | model/openai.New + WithVariant | 低 |
| `memory.go` | 308 | memorysvc (InMemory + Redis) | 中 |
| `tool_hub.go` | 524 | tool/function.NewFunctionTool | 低 |
| `sse_scanner.go` | 72 | tRPC event channel | 低（可删除） |

### 评估后决定

| 文件 | 行数 | 评估点 |
|:---|:---:|:---|
| `rag.go` | 235 | tRPC knowledge 是否支持 ES 后端？不支持则保留 |
| `semantic_search.go` | 417 | 同上 |

### 5 级隔离保证（保持不变）

| 级别 | 机制 | 迁移后 |
|:---|:---|:---|
| L1 编译隔离 | `//go:build aurora_agent` tag | ✅ 保持 |
| L2 配置隔离 | `agent.enabled=false` | ✅ 保持 |
| L3 路由隔离 | 独立 `/api/agent/*` RouterGroup | ✅ 保持 |
| L4 故障隔离 | goroutine + recover | ✅ 保持（tRPC runner 内部也有 recover） |
| L5 依赖隔离 | 核心 Service 零 import agent 包 | ✅ 保持 |

---

## 三、分阶段计划

### 阶段 0：引入依赖 + 配置迁移（0.5 天）

**任务**：
1. `go get trpc.group/trpc-go/trpc-agent-go@v1.10.0`
2. 验证 go.mod 依赖不冲突
3. 配置结构迁移：
   - 现有 `AgentConfig.LLM.Providers` → tRPC 环境变量模式
   - 保留 `agent.enabled` 开关
   - 新增 `OPENAI_API_KEY` / `OPENAI_BASE_URL` 环境变量映射
4. 编译验证（agent 模块未启用时不影响主流程）

**交付**：go.mod 引入依赖，编译通过，配置兼容

**风险**：tRPC-Agent-Go 间接依赖可能与 aurora 现有依赖冲突（go.sum 冲突）

---

### 阶段 1：LLM Router 替换（1 天）

**现状**：`llm_router.go` 449 行，手搓 HTTP 请求调 OpenAI/DeepSeek/Qwen/Claude

**目标**：用 `model/openai` 替换，保留 `LLMRouter` 接口

**改造**：
```go
// 保留接口
type LLMRouter struct {
    models map[string]model.Model  // tRPC model 实例
    defaultProvider string
}

func NewLLMRouter(cfg *config.LLMConfig) (*LLMRouter, error) {
    // 用 openai.New + WithVariant 创建各 provider 的 model
    // OpenAI:    openai.New("gpt-4o-mini")
    // DeepSeek:  openai.New("deepseek-chat", openai.WithVariant(openai.VariantDeepSeek))
    // Qwen:      openai.New("qwen-plus", openai.WithBaseURL(...))  // 需验证
    // Claude:    需验证是否支持（Anthropic API 格式不同）
}
```

**待验证**：
- [ ] Qwen 是否通过 OpenAI 兼容 API 支持？（阿里 DashScope 有 OpenAI 兼容模式）
- [ ] Claude 是否有 VariantAnthropic？（若无，需保留 Claude 手搓适配）

**交付**：llm_router.go 重写，单元测试通过

---

### 阶段 2：Tool Hub 替换（0.5 天）

**现状**：`tool_hub.go` 524 行，自研 ToolFunc/ToolDefinition/ToolHub

**目标**：用 `tool/function.NewFunctionTool` 替换底层，保留注册接口

**改造**：
```go
// 6 个业务工具迁移
func RegisterAuroraTools(hub *ToolHub) {
    // 每个工具用 function.NewFunctionTool 包装
    searchTool := function.NewFunctionTool(searchArticles,
        function.WithName("search_articles"),
        function.WithDescription("搜索博客文章"),
    )
    hub.Register(searchTool)
}
```

**交付**：tool_hub.go 重写，6 个业务工具迁移完成

---

### 阶段 3：Memory/Session 替换（1 天）

**现状**：`memory.go` 308 行，自研 InMemory + Redis 会话记忆

**目标**：用 `memorysvc` 替换，session 管理交给 runner

**改造**：
```go
// 创建 memory service
var mem memorysvc.Service
if cfg.Memory.Type == "redis" {
    mem = memorysvc.NewRedisService(redisAddr)  // 需验证 API
} else {
    mem = memorysvc.NewInMemoryService()
}

// 传入 runner
r := runner.NewRunner("aurora-agent", agent,
    runner.WithMemoryService(mem),
)
```

**待验证**：
- [ ] memorysvc 的 Redis 配置 API
- [ ] session 持久化格式是否兼容现有数据
- [ ] ListSessions/DeleteSession 接口是否可用

**交付**：memory.go 重写或删除（由 runner 接管），MemoryService 接口适配

---

### 阶段 4：Agent 入口重写（1 天）

**现状**：`agent.go` 386 行，AuroraAgent 单例 + InitAgent

**目标**：用 `llmagent.New` + `runner.NewRunner` 重写

**改造**：
```go
var globalRunner runner.Runner

func InitAgent(cfg *config.AgentConfig) error {
    if !cfg.Enabled {
        return nil
    }
    // 1. 创建 model
    // 2. 创建 tools
    // 3. 创建 agent (llmagent.New)
    // 4. 创建 runner (runner.NewRunner)
    // 5. 赋值 globalRunner
}

// Chat 对话（SSE 流式）
func Chat(ctx context.Context, userID, sessionID, message string) (<-chan *model.Event, error) {
    return globalRunner.Run(ctx, userID, sessionID, model.NewUserMessage(message))
}
```

**handler 层适配**：
```go
// agent_handler.go 的 Chat 方法
events, err := agent.Chat(ctx, userID, sessionID, message)
for event := range events {
    // event.Response.Choices[0].Delta.Content → SSE 写入
    sseWrite(c, event)
}
```

**交付**：agent.go 重写，handler 层 SSE 适配

---

### 阶段 5：RAG/Search 评估（0.5 天）

**评估点**：tRPC `knowledge` 子包是否支持 ES 后端

- **支持 ES** → 用 knowledge 替换 rag.go + semantic_search.go
- **不支持 ES** → 保留自研，仅替换 LLM 调用部分（用 LLMRouter 替换）

**交付**：评估报告 + 决策

---

### 阶段 6：SSE 适配 + 清理（0.5 天）

**任务**：
1. sse_scanner.go → 删除（tRPC event channel 替代）
2. handler 层 SSE 输出适配（event → SSE 格式转换）
3. 清理废弃代码

**交付**：SSE 流式输出正常工作

---

### 阶段 7：集成测试（1 天）

**任务**：
1. 5 个 API 端点全量测试（chat/write/search/analyze/sessions）
2. 5 级隔离验证（编译 tag / 配置开关 / 路由 / 故障 / 依赖）
3. 多模型测试（DeepSeek / OpenAI / Qwen / Claude）
4. 流式输出测试（SSE）
5. 性能基准测试

**交付**：测试报告，全绿

---

## 四、风险矩阵

| 风险 | 概率 | 影响 | 缓解措施 |
|:---|:---:|:---:|:---|
| go.mod 依赖冲突 | 中 | 高 | 阶段 0 优先验证，冲突则手动 resolve |
| Qwen/Claude 不支持 | 中 | 中 | 保留 Claude 手搓适配，Qwen 用 OpenAI 兼容模式 |
| Redis session API 不明确 | 高 | 中 | 阶段 3 前查源码确认，必要时用 InMemory 过渡 |
| knowledge 不支持 ES | 高 | 低 | 保留自研 RAG，仅替换 LLM 调用 |
| 事件流 → SSE 适配复杂 | 中 | 中 | 阶段 6 预留缓冲，参考 tRPC examples |
| 依赖膨胀影响二进制体积 | 低 | 低 | agent 模块可选编译，不影响核心博客 |
| 回归风险 | 中 | 高 | 每阶段独立提交 + 测试，可逐阶段回滚 |

---

## 五、配置迁移对照

### 现有配置（config.yaml）

```yaml
agent:
  enabled: false
  llm:
    default_provider: deepseek
    providers:
      openai:
        api_key: ""
        base_url: https://api.openai.com/v1
        model: gpt-4o-mini
      deepseek:
        api_key: ""
        base_url: https://api.deepseek.com/v1
        model: deepseek-chat
      qwen:
        api_key: ""
        base_url: https://dashscope.aliyuncs.com/compatible-mode/v1
        model: qwen-plus
      claude:
        api_key: ""
        base_url: https://api.anthropic.com/v1
        model: claude-sonnet-4-20250514
  memory:
    type: redis
    max_turns: 20
```

### 迁移后配置

```yaml
agent:
  enabled: false
  # tRPC-Agent-Go 通过环境变量配置 API Key
  # OPENAI_API_KEY / OPENAI_BASE_URL
  default_provider: deepseek
  model:
    name: deepseek-chat
    variant: deepseek          # openai/deepseek（qwen 用 openai 兼容模式）
  memory:
    type: redis                # inmemory/redis
    redis_addr: ""             # 由 AURORA_REDIS_HOST 覆盖
    max_turns: 20
```

### 环境变量映射

| 环境变量 | 用途 |
|:---|:---|
| `OPENAI_API_KEY` | LLM API Key（tRPC 原生读取） |
| `OPENAI_BASE_URL` | LLM Base URL（tRPC 原生读取） |
| `AURORA_AGENT_ENABLED` | Agent 开关（保持） |
| `AURORA_AGENT_MODEL_NAME` | 默认模型名 |
| `AURORA_AGENT_MODEL_VARIANT` | 模型变体（openai/deepseek） |

---

## 六、时间线

| 阶段 | 工期 | 累计 | 交付物 |
|:---:|:---:|:---:|:---|
| 0. 引入依赖 + 配置 | 0.5d | 0.5d | go.mod + 配置兼容 |
| 1. LLM Router 替换 | 1d | 1.5d | llm_router.go 重写 |
| 2. Tool Hub 替换 | 0.5d | 2d | tool_hub.go 重写 |
| 3. Memory/Session 替换 | 1d | 3d | memory.go 重写 |
| 4. Agent 入口重写 | 1d | 4d | agent.go 重写 + SSE 适配 |
| 5. RAG/Search 评估 | 0.5d | 4.5d | 评估报告 + 决策 |
| 6. SSE 适配 + 清理 | 0.5d | 5d | sse_scanner.go 删除 |
| 7. 集成测试 | 1d | 6d | 测试报告 |

---

## 七、回滚策略

每阶段独立提交，可逐阶段回滚：
- 阶段 0 失败 → 回滚 go.mod，不影响现有代码
- 阶段 1-4 失败 → 回滚对应文件，保留之前阶段成果
- 阶段 7 失败 → 整体回滚到迁移前快照（git revert）

**安全网**：agent 模块有 `//go:build aurora_agent` 编译 tag，迁移期间默认不编译，不影响核心博客系统。
