# Aurora Agent 开发指南

> 基于 **tRPC-Agent-Go v1.8** (腾讯开源) 构建的 AI Agent 智能体模块

---

## 一、架构概览

### 1.1 Agent模块位置

```
aurora-go/
├── internal/
│   ├── service/
│   │   └── agent/                    # Agent业务层 (可选插件)
│   │       ├── aurora_agent.go       # Agent工厂 (~80行)
│   │       ├── aurora_tools.go       # 业务工具注册 (~200行)
│   │       ├── aurora_handler.go     # Gin→tRPC桥接 (~100行)
│   │       ├── aurora_workflow.go    # 预置工作流 (~150行)
│   │       └── memory_adapter.go     # Redis持久化适配器 (~60行)
│   ├── agent/                        # Agent引擎封装 (隔离层)
│       ├── factory.go                # 引擎初始化入口
│       └── types.go                  # 公开类型定义
│   └── handler/
│       └── agent_handler.go          # HTTP处理器
```

### 1.2 组件关系图

```
HTTP Request
    │
    ▼
┌──────────────────┐
│  AgentHandler     │ ← Gin HTTP处理器
│  (agent_handler)  │
└────────┬─────────┘
         │ 调用
         ▼
┌──────────────────┐
│  AuroraAgent      │ ← 工厂组装
│  (service/agent)  │
└────────┬─────────┘
    ┌────┴────┬──────────┐
    ▼         ▼          ▼
┌────────┐┌────────┐┌──────────┐
│ LLM    │ Tool   │ Memory   │
│ Router │ Hub    │ Service  │
│(多模型) │(函数集) │(会话记忆)│
└────────┘└────────┘└──────────┘
    │         │          │
    ▼         ▼          ▼
 OpenAI     Search    Redis
 DeepSeek   Write     (持久化)
 Claude     Review
 Qwen       Analyze
```

---

## 二、快速开始

### 2.1 启用Agent

```yaml
# configs/config.yaml
agent:
  enabled: true                          # 总开关
  llm:
    default_model: deepseek-chat         # 默认模型
    providers:                           # 多模型配置
      openai:
        api_key: ${OPENAI_API_KEY}
        base_url: https://api.openai.com/v1
      deepseek:
        api_key: ${DEEPSEEK_API_KEY}
        base_url: https://api.deepseek.com
      anthropic:
        api_key: ${ANTHROPIC_API_KEY}
        base_url: https://api.anthropic.com
  rag:
    enabled: true                        # RAG增强检索
    es_index: aurora_article             # ES索引名
    top_k: 5                            # 检索数量
  moderation:
    auto_enabled: true                  # 自动内容审核
    sensitive_words: true               # 敏感词过滤
  stream:
    default: true                       # 默认流式输出
    buffer_size: 2048                   # 缓冲区大小
```

### 2.2 调用示例

#### AI对话

```bash
curl -N -X POST http://localhost:8080/api/agent/chat \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: text/event-stream" \
  -d '{
    "message": "帮我写一篇关于Go语言Gin框架的博客文章",
    "model": "deepseek-chat",
    "stream": true
  }'
```

#### AI写作助手

```bash
curl -X POST http://localhost:8080/api/agent/write \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "operation": "generate",
    "topic": "微服务架构设计最佳实践",
    "style": "professional",
    "length": "long"
  }'
```

#### AI语义搜索

```bash
curl -X POST http://localhost:8080/api/agent/search \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "如何在生产环境中优化Go程序的内存使用？",
    "useRag": true,
    "topK": 5
  }'
```

#### 数据分析洞察

```bash
curl -X POST http://localhost:8080/api/agent/analyze \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "analysisType": "overview",
    "dateRange": {
      "start": "2026-03-01",
      "end": "2026-04-11"
    }
  }'
```

---

## 三、核心组件详解

### 3.1 LLM多模型路由 (LLM Router)

支持同时接入多个LLM Provider, 通过配置动态切换:

| Provider | 模型ID | 最大上下文 | 适用场景 |
|----------|--------|-----------|---------|
| OpenAI | `gpt-4o` | 128K | 复杂推理/长文分析 |
| Anthropic | `claude-3.5-sonnet` | 200K | 长文档/代码审查 |
| DeepSeek | `deepseek-chat` | 65K | 中文写作/性价比首选 |
| Alibaba | `qwen-max` | 32K | 国内合规场景 |

**添加新的Provider只需在config.yaml中添加配置, 零代码改动:**

```yaml
llm:
  providers:
    custom_provider:                      # 新Provider
      api_key: ${CUSTOM_API_KEY}
      base_url: https://api.custom.com/v1
      models:
        - id: custom-model-1
          name: Custom Model
          max_tokens: 8192
```

### 3.2 Tool工具集 (Tool Hub)

Aurora注册的业务工具, Agent可自主决定是否调用:

| 工具名 | 功能 | 输入 | 输出 |
|--------|------|------|------|
| `SearchArticles` | 搜索文章 | 关键词/分类/标签 | 文章列表+摘要 |
| `WriteArticle` | 创建/编辑文章 | 标题/内容/分类 | 文章ID |
| `GetArticleDetail` | 获取文章详情 | 文章ID | 完整文章 |
| `AnalyzeStats` | 获取统计数据 | 时间范围/类型 | 统计结果 |
| `ManageTags` | 管理标签 | 操作类型/标签名 | 结果 |
| `ReviewContent` | 内容审核 | 文本内容 | 审核+评分 |
| `GetWebsiteConfig` | 获取网站配置 | - | 配置信息 |

**自定义工具步骤:**

```go
// 1. 定义工具函数
func MyCustomTool(ctx context.Context, input MyInput) (*MyOutput, error) {
    // 业务逻辑...
    return &MyOutput{Result: "success"}, nil
}

// 2. 在 aurora_tools.go 注册
func RegisterTools(registry *tool.Registry) error {
    // ...已有工具...
    
    registry.Register(tool.FunctionToolConfig{
        Name:        "my_custom_tool",
        Description: "自定义工具描述",
        InputSchema: tool.SchemaFromStruct(MyInput{}),
        Handler:     MyCustomTool,
    })
    
    return nil
}
```

### 3.3 Memory会话记忆 (Memory)

基于Redis的持久化会话记忆:

```
Session (会话)
├── ID: sess_uuid
├── Title: "Go博客写作讨论"
├── CreatedAt: timestamp
├── Messages (消息列表)
│   ├── [system] 系统Prompt
│   ├── [user] 用户消息1
│   ├── [assistant] AI回复1 (含tool_calls)
│   ├── [tool] 工具返回结果
│   ├── [user] 用户消息2
│   └── ...
├── Metadata (元数据)
│   ├── Model: deepseek-chat
│   ├── TotalTokens: 4500
│   └── MessageCount: 8
```

**Redis Key结构:**
```
aurora:agent:session:{session_id}    → Hash (会话元数据)
aurora:agent:msg:{session_id}        → List (消息历史)
aurora:agent:user:{user_id}:sessions → Set (用户的会话列表)
```

**Memory TTL策略:**
- 活跃会话: 7天不活跃自动清理
- 消息上限: 单会话最多保留200条
- Token预算: 超过8192 tokens自动摘要压缩

### 3.4 RAG管道 (知识检索增强)

5阶段处理流程:

```
用户问题 → [Query Understanding]
              ↓
         [Retrieval] ← ES混合检索(向量+关键词)
              ↓
         [ReRanking] ← 相关度重排序
              ↓
         [Context Assembly] ← 上下文拼接
              ↓
         [Generation] ← LLM生成答案
```

**RAG开关:**
- 全局开关: `agent.rag.enabled: false`
- 单次请求: `"useRag": false`
- 降级策略: ES不可用时自动fallback到MySQL全文搜索

### 3.5 工作流引擎 (Workflow DAG)

预置的工作流模板:

| 工作流名 | 触发方式 | 步骤 |
|----------|---------|------|
| `ArticlePublishFlow` | 发布文章 | AI生成→审核→SEO优化→定时发布 |
| `CommentReviewFlow` | 新评论 | 敏感词过滤→情感分析→垃圾判定→通知 |
| `ContentAnalysisFlow` | 定时 | 数据采集→趋势分析→报告生成→推送 |

**自定义工作流示例:**

```go
// 在 aurora_workflow.go 中定义
var CustomWorkflow = workflow.Define("custom_flow").
    Step("step1", aiTask, func(ctx context.Context, state State) (string, error) {
        return "step2", nil  // 下一步
    }).
    Step("step2", humanTask, func(ctx context.Context, state State) (string, error) {
        return workflow.END, nil  // 结束
    }).
    Build()
```

---

## 四、安全与限制

### 4.1 速率限制

| 端点 | 限制 | 窗口 |
|------|------|------|
| POST /agent/chat | 30次/分钟 | 滑动窗口 |
| POST /agent/write | 15次/分钟 | 滑动窗口 |
| POST /agent/search | 60次/分钟 | 滻动窗口 |
| POST /agent/analyze | 10次/分钟 | 滑动窗口 |

### 4.2 Token配额

```yaml
agent:
  limits:
    daily_tokens_per_user: 100000   # 每用户每日Token限额
    max_output_tokens: 4096         # 单次最大输出Tokens
    session_message_limit: 200      # 单会话消息上限
```

### 4.3 内容安全

- **敏感词过滤**: 内置敏感词库, 支持自定义扩展
- **LLM自带安全**: DeepSeek/GPT-4o/Claude均内置安全护栏
- **人工审核**: 评论可开启强制人工审核模式

---

## 五、监控与调试

### 5.1 使用统计

```bash
GET /api/agent/stats
```

返回: 对话数/写作数/搜索数/Token消耗/Daily趋势

### 5.2 日志关键字

```bash
# 查看Agent相关日志
grep "agent" logs/aurora.log

# 查看LLM调用日志
grep "llm\|model" logs/aurora.log

# 查看Tool调用日志
grep "tool_call" logs/aurora.log

# 查看错误日志
grep "error\|ERROR.*agent" logs/aurora.log
```

### 5.3 健康检查

```bash
GET /health
# 响应中包含 agentReady: true/false
```

---

## 六、故障排查

| 问题 | 可能原因 | 解决方案 |
|------|---------|---------|
| Agent端点404 | agent.enabled=false | 修改配置并重启 |
| LLM调用失败 | API Key无效/额度不足 | 检查.env中的API Key |
| SSE断连 | Nginx代理buffer | 添加proxy_buffering off |
| RAG无结果 | ES未启动/索引不存在 | 检查ES健康状态 |
| Redis连接超时 | 连接池耗尽 | 增大pool_size |
| Token超限 | 单次请求过长 | 减小max_tokens |

---

*文档版本: 1.0 | 最后更新: 2026-04-11*
