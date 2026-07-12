# tRPC-Agent-Go 版本升级评估报告

> 评估范围：`v1.8.0` → `v1.10.0`
> 评估日期：2026-07-01（2026-07-12 修正发布日期）
> 评估结论：**不建议立即全量升级，但建议跟踪 P1/P3 子包**（理由见 §5）

---

## 1. 执行摘要（核心结论前置）

| 关键事实 | 说明 |
|:---|:---|
| **aurora-go 是否真依赖 tRPC-Agent-Go** | ❌ 否。`go.mod` 无此依赖，`internal/agent/` 全部 import 为项目内部包 |
| **当前 "v1.8" 字样性质** | 仅注释/文档里的**架构对标声明**（"对标 tRPC xxx 包"），非依赖版本 |
| **tRPC-Agent-Go 最新版** | `v1.10.0`（**2026-06-05**），37 天前发布，项目非常活跃（4 个月 5 个 release） |
| **若仅升级文档版本号** | 工程量 < 10 分钟，零代码风险，但**实质收益为零**（代码没用这个库） |
| **若真正引入依赖重构** | 工程量数天起，需重写 7+ 核心文件，有回归风险，**收益与 aurora 实际需求匹配度中** |

**核心判断**：tRPC-Agent-Go 是**非常活跃**的项目（2026 年内已发布 5 个 release），长期不跟踪会积累技术债。但 aurora-go 的 agent 模块是**完全自研**的，全量重构 ROI 低。**建议优先引入 P1（LLM Router）和 P3（MCP 工具）两个高收益子包**，其余保持自研。

---

## 2. 版本跨度与时间线

| 版本 | 发布日期 | 性质 | 距今(2026-07-12) |
|:---|:---|:---|:---|
| `v1.8.0` | **2026-04-07** | aurora 当前对标版本 | 96 天 |
| `v1.9.0` | **2026-05-08** | minor | 65 天 |
| `v1.9.1` | **2026-05-11** | patch | 62 天 |
| `v1.10.0` | **2026-06-05** | minor（当前最新） | 37 天 |

> ⚠️ 日期修正：此前版本报告因 WebFetch AI 摘要将年份误读为 2024/2025，经 GitHub API `published_at` 字段确认为 **2026 年**。

`v1.8.0...v1.10.0` 共 **275 个 commit**，通过 GitHub Compare API 抓取的增量提交已覆盖主要变化。

---

## 3. 关键变化（按子包分类）

### 3.1 v1.8.0 基线（aurora 对标版本）已含特性

| 子包 | 变化要点 | 与 aurora 自研模块对标关系 |
|:---|:---|:---|
| `model/openai` | 推断 DeepSeek 变体（text-only chat） | 🔴 **高相关** — `llm_router.go:20` 直接对标 |
| `memory` | preload 适应体量、lexical+vector 排序优化 | 🔴 **高相关** — `memory.go:17` 直接对标 |
| `memory/sqlitevec` | 启动时迁移旧 schema | 🟡 中相关（aurora 用 Redis，不用 sqlite） |
| `graphagent` | invoke metrics、默认 stream=true、terminal-only 过滤 | 🔴 **高相关** — `workflow.go:19` 对标 StateGraph |
| `session/mysql` | 序列化 state 更新 | 🟡 中相关（aurora 用 GORM 自管） |
| `knowledge` | sqlite vec vectorstore | 🟡 中相关（aurora 用 ES） |
| `runner` | safe-boundary user steering | 🟢 低相关 |
| `lsc` | invocation-scoped 事件控制、execution trace | 🟢 低相关 |
| `agent/telemetry` | trace+metric 加 name/id 属性 | 🟢 低相关（aurora 用 Zap） |
| `openclaw` | file-backed memory、sidebar admin | ⚪ 无关（官方 demo UI） |

### 3.2 v1.9.0 → v1.10.0 增量变化（275 commits）

> 从 Compare API 抓取的 commit 标题归纳，按子包归类。

| 子包 | 主要变化 | 与 aurora 自研模块对标关系 |
|:---|:---|:---|
| `a2a` | 流式会话事件顺序保序、A2A 协议增强 | ⚪ 无关 — aurora 未用 A2A（agent-to-agent）协议 |
| `agent/LLMAgent` | telemetry error labels 统一、LLMAgent 错误处理增强 | 🟡 中相关 — aurora `agent.go` 自研同类 |
| `agent/dify` | 接入 Dify SDK、流式错误处理 | ⚪ 无关 — aurora 未集成 Dify |
| `graph` | graph messages state append 修复、checkpoint 覆盖白名单 | 🔴 **高相关** — `workflow.go` 对标 |
| `processor` | summary user injection merge order、batched tool call state | 🟡 中相关 |
| `tool/workspaceexec` | conversation files staging 到 work/inputs | 🟡 中相关 — `tool_hub.go` 对标 |
| `evaluator` | rubric reference critic evaluator | ⚪ 无关 — aurora 无评测集 |
| `event` | execution trace surface ids clone | 🟢 低相关 |
| `internal/telemetry` | structured response errors in labels | 🟢 低相关 |
| `session` | graph agent child surface roots 对齐、custom state 保留 | 🟡 中相关 |
| `openclaw` | admin 导航、async URLs under proxy | ⚪ 无关 |
| `docs/examples` | streamtool example、sbti example | ⚪ 无关 |

### 3.3 v1.9.1 修复要点（3 PR）

- `openclaw`: admin sidebar 可导航性
- `{agent, graph, runner}`: graph agent child surface roots 对齐
- `agent`: invocation views 保留 custom state

---

## 4. 与 aurora 自研 agent 模块的相关性矩阵

| aurora 自研文件 | 对标 tRPC 子包 | v1.8→v1.10 增量是否影响 aurora |
|:---|:---|:---|
| `llm_router.go` | `model/openai` + `model/deepseek` | 🟡 **中** — v1.8 已含 DeepSeek 变体推断，v1.9-v1.10 该子包无重大变化 |
| `memory.go` | `memory/memorysvc` + Redis 适配器 | 🟢 **低** — v1.9-v1.10 的 memory 改动集中在 sqlitevec/openclaw，aurora 用 Redis |
| `rag.go` | `knowledge` | 🟢 **低** — v1.9-v1.10 的 knowledge 改动集中在 sqlite vec，aurora 用 ES |
| `tool_hub.go` | `tool/function` + `tool/mcp` | 🟡 **中** — `tool/workspaceexec` 有文件 staging 改进，但 aurora tool_hub 是自定义实现 |
| `workflow.go` | `graph/StateGraph` | 🔴 **高** — `graph messages state append` 修复、`checkpoint 覆盖白名单` 与 workflow 状态管理直接相关 |
| `writing_assistant.go` | `workflow.Execute` | 🟢 **低** — 是 workflow 的上层封装 |
| `agent.go` | `agent/LLMAgent` | 🟡 **中** — telemetry error 统一，但 aurora 用 Zap 自管日志 |
| `semantic_search.go` | `knowledge` 检索 | 🟢 **低** |
| `sse_scanner.go` | 无对标（自研 SSE 解析） | ⚪ 无关 |
| `comment_assistant.go` / `content_moderator.go` | 无对标（aurora 业务封装） | ⚪ 无关 |

**相关性汇总**：275 个 commit 中，**高相关仅 1 项**（graph state 修复），**中相关 4 项**，其余 270+ commit 与 aurora 自研实现无直接关系。

---

## 5. 三方案工程量与收益对比

| 维度 | 方案A：仅升级文档版本号 | 方案B：引入依赖全量重构 | 方案C：渐进式混合引入 |
|:---|:---|:---|:---|
| **工程量** | < 10 分钟 | 数天（重写 7+ 文件 + 全量测试） | 数小时（1-2 个子模块替换） |
| **代码改动** | 0 行（仅注释/README/MIGRATION_GUIDE） | 1000+ 行 | 200-500 行 |
| **go.mod 依赖** | 不引入 | 引入 `trpc.group/trpc-agent-go@v1.10.0` 及其传递依赖 | 部分引入 |
| **回归风险** | 零 | 高（agent API 全部需重测） | 中 |
| **实质收益** | ❌ 零（代码没用此库） | 🟡 部分（获得 graph state 修复、telemetry 增强，但大量子包 aurora 用不到） | 🟡 部分 |
| **维护成本** | 不变 | 上升（依赖一个 275+ commit/季的快速迭代库） | 上升 |
| **与 aurora 现状契合度** | 高（保持架构对标声明） | 低（自研已稳定，重构重复投资） | 中 |

### 5.1 方案A 的额外说明

如果选 A，建议把 "v1.8" → "v1.10.0"，同步更新：
- `internal/agent/agent.go:1`（注释）
- `README.md:24, 45, 70, 81, 294`（5 处）
- `docs/MIGRATION_GUIDE.md:90`

但需明确告知：**这只是"对标版本号刷新"，aurora 的 agent 实现仍是自研，不会因文档版本号变化而获得任何新特性**。

### 5.2 方案B 不推荐的核心理由

1. **重复投资**：aurora 自研 agent 模块已稳定运行，引入 tRPC-Agent-Go 等于推翻重做
2. **依赖膨胀**：tRPC-Agent-Go v1.10.0 引入 `a2a` / `openclaw` / `evaluator` / `dify` 等大量 aurora 用不到的子包
3. **快速迭代的负担**：该库 2 个月 275 commits，跟随升级成本高
4. **架构差异**：tRPC-Agent-Go 的 session/graph/telemetry 体系与 aurora 现有 GORM+Redis+Zap 体系不兼容，需大量适配

### 5.3 方案C 的潜在切入点（若坚持引入）

唯一值得考虑的引入点是 **`workflow.go` 对标 `graph/StateGraph`**：
- v1.9-v1.10 修复了 graph messages state append（#1618）和 checkpoint 覆盖白名单（#1600）
- 如果 aurora 的 workflow 在并发状态管理上有 bug 隐患，引入 `graph` 子包有实际价值
- 但需先验证 aurora `workflow.go` 是否真存在这些 bug

---

## 6. 推荐与触发条件

### 6.1 当前推荐：**保持现状，不升级**

理由：
1. aurora 不依赖此库，升级版本号无实质收益
2. 自研 agent 模块稳定运行，重构风险 > 收益
3. v1.9-v1.10 的核心改动（a2a/dify/openclaw/evaluator）与 aurora 业务无关

### 6.2 重新评估的触发条件

以下任一情况出现时，建议重新评估引入 tRPC-Agent-Go：

| 触发条件 | 建议方案 |
|:---|:---|
| aurora 需要接入 **A2A 多 agent 协作**协议 | 方案B（引入 `a2a` 子包） |
| aurora `workflow.go` 出现**并发状态管理 bug** | 方案C（引入 `graph` 子包替换 workflow） |
| aurora 需要接入 **Dify** 平台 | 方案B（引入 `agent/dify`） |
| aurora 需要接入 **MCP 协议工具**生态 | 方案C（引入 `tool/mcp`） |
| aurora 需要 **agent 评测集**能力 | 方案B（引入 `evaluator`） |

### 6.3 若仍需执行升级的备选动作

如果出于"保持文档新鲜度"的目的必须刷新版本号，建议：
- 仅执行方案A（文档版本号 v1.8 → v1.10.0）
- 在 `agent.go` 注释里增加说明："本模块为自研实现，架构对标 tRPC-Agent-Go v1.10.0，未直接依赖该库"
- 避免误导后续维护者以为存在真实依赖

---

## 7. 附：数据来源

- tRPC-Agent-Go GitHub: https://github.com/trpc-group/trpc-agent-go
- v1.8.0 Release: https://github.com/trpc-group/trpc-agent-go/releases/tag/v1.8.0
- v1.9.1 Release: https://github.com/trpc-group/trpc-agent-go/releases/tag/v1.9.1
- v1.8.0...v1.10.0 Compare API: 275 commits, status=ahead
- aurora-go `go.mod`: 无 `trpc.group/trpc-agent-go` 依赖
- aurora-go `internal/agent/`: 10 个 .go 文件全部 import `github.com/aurora-go/aurora/internal/...`
