# Portfolio 作品集模块指南

> 模块定位：首页/作品集页展示个人 GitHub 仓库，通过定时任务自动同步仓库快照，后台可覆盖封面/排序/置顶/可见性。

---

## 一、模块概览

| 项 | 说明 |
|:---|:---|
| **数据表** | `t_portfolio`（GitHub 仓库快照缓存） |
| **业务定位** | 个人作品集展示页，对标 Java 版 `PortfolioController` |
| **同步来源** | GitHub REST API `/users/{username}/repos` |
| **同步方式** | 定时任务（每天 03:00）+ 后台手动触发 |
| **缓存策略** | Redis key `portfolio:featured`，写操作后 invalidate |
| **新增模块** | aurora-go 自研，Java 版无此模块（仅在 aurora-go 中实现） |

---

## 二、数据表结构（t_portfolio）

| 字段 | 类型 | 说明 |
|:---|:---|:---|
| `id` | int PK | 主键 |
| `repo_id` | bigint UNIQUE | GitHub 仓库 ID（唯一键，upsert 依据） |
| `name` | varchar(128) | 仓库名 |
| `full_name` | varchar(255) | `owner/repo` |
| `description` | varchar(500) | 仓库描述 |
| `html_url` | varchar(500) | 仓库地址 |
| `homepage` | varchar(500) | 演示地址 |
| `language` | varchar(64) | 主语言 |
| `stargazers_count` | int | star 数 |
| `forks_count` | int | fork 数 |
| `topics` | text | 话题标签 JSON 数组 |
| `repo_created_at` / `repo_updated_at` | datetime | 仓库时间 |
| `cover` | varchar(500) | **自定义封面**（后台覆盖，同步时不触碰） |
| `sort` | int | **排序权重**（越大越靠前，同步时不触碰） |
| `is_featured` | tinyint(1) | **是否首页置顶**（同步时不触碰） |
| `is_visible` | tinyint(1) | **是否展示**（0 隐藏 1 展示，同步时不触碰） |
| `create_time` / `update_time` | datetime | 记录时间 |

**索引**：`uk_repo_id`（唯一）、`idx_visible_featured`（is_visible, is_featured, sort）

**人工配置字段**（同步时保留）：`cover` / `sort` / `is_featured` / `is_visible`
**动态同步字段**：其余字段每次同步覆盖

---

## 三、API 接口

### 3.1 前台接口（`/api` 前缀，无需认证）

| 方法 | 路径 | 说明 | Handler |
|:---|:---|:---|:---|
| GET | `/api/portfolios/featured` | 首页置顶作品集（最多 6 条） | `PortfolioHandler.ListFeatured` |
| GET | `/api/portfolios` | 作品集分页列表（仅可见） | `PortfolioHandler.ListAll` |

**排序规则**：`is_featured DESC, sort DESC, stargazers_count DESC, repo_updated_at DESC`

**ListFeatured 响应**（`PortfolioDTO`）：

```json
{
  "id": 1,
  "name": "aurora-blog",
  "fullName": "aqi/aurora-blog",
  "description": "博客前端",
  "htmlUrl": "https://github.com/aqi/aurora-blog",
  "homepage": "https://www.aqi125.cn",
  "language": "Vue",
  "stargazersCount": 10,
  "forksCount": 2,
  "topics": ["vue", "typescript"],
  "repoUpdatedAt": "2026-06-01T12:00:00Z",
  "cover": "https://ws.aqi125.cn/cover.png",
  "sort": 100,
  "isFeatured": 1
}
```

### 3.2 后台接口（`/api/admin` 前缀，需认证 + RBAC）

| 方法 | 路径 | 说明 | Handler |
|:---|:---|:---|:---|
| GET | `/api/admin/portfolios` | 分页查询（含隐藏项，支持关键词） | `PortfolioHandler.ListAdmin` |
| PUT | `/api/admin/portfolios` | 编辑作品（封面/排序/置顶/可见性） | `PortfolioHandler.Update` |
| DELETE | `/api/admin/portfolios` | 批量删除（body: `[id1, id2, ...]`） | `PortfolioHandler.Delete` |
| POST | `/api/admin/portfolios/sync` | 手动触发 GitHub 同步 | `PortfolioHandler.Sync` |

**ListAdmin 查询参数**（`ConditionVO`）：
- `keywords`：模糊匹配 name / full_name / description
- `pageNum` / `pageSize`：分页

**Update 请求体**（`PortfolioVO`）：

```json
{
  "id": 1,
  "cover": "https://ws.aqi125.cn/new-cover.png",
  "sort": 200,
  "isFeatured": 1,
  "isVisible": 1
}
```

> `isFeatured` / `isVisible` 为指针类型，`null` 表示不更新；`oneof=0 1` 校验仅允许 0 或 1。

---

## 四、GitHub 同步机制

### 4.1 同步流程

```
定时任务/手动触发
  → SyncFromGitHub(ctx)
  → fetchGitHubRepos(username, token)  // GET /users/{username}/repos?sort=updated&per_page=100&type=owner
  → 过滤 archived 仓库 + exclude 配置
  → upsertPortfolio(repo)  // 按 repo_id upsert
  → 清理数据库中已不在 GitHub 列表的陈旧记录
  → invalidateCache()  // 删 Redis key
```

### 4.2 upsert 策略

- **新增仓库**：插入完整记录，`is_visible=1`（默认可见）、`is_featured=0`、`sort=0`、`cover=""`
- **已存在仓库**：仅更新动态字段（name/description/star/fork/topics/...），**保留** `cover`/`sort`/`is_featured`/`is_visible`
- **GitHub 已删除的仓库**：自动从 `t_portfolio` 删除

### 4.3 GitHub API 限制

- 未配置 token：匿名请求，速率限制 60 次/小时/IP
- 配置 token：5000 次/小时，建议生产环境配置
- 一次拉取最多 100 个仓库（`per_page=100`）

### 4.4 配置项

```yaml
# configs/config.yaml
github:
  enabled: true              # 是否启用作品集同步
  username: "your-github"    # GitHub 用户名
  token: ""                  # PAT，建议通过 AURORA_GITHUB_TOKEN 注入
  exclude: "repo1,repo2"     # 排除的仓库名（逗号分隔，不区分大小写）
```

**环境变量覆盖**：

| 配置键 | 环境变量 |
|:---|:---|
| `github.enabled` | `AURORA_GITHUB_ENABLED` |
| `github.username` | `AURORA_GITHUB_USERNAME` |
| `github.token` | `AURORA_GITHUB_TOKEN` |
| `github.exclude` | `AURORA_GITHUB_EXCLUDE` |

---

## 五、定时任务

| 项 | 值 |
|:---|:---|
| **任务名** | `GitHub作品集同步` |
| **invoke_target** | `auroraQuartz.syncGitHubRepos` |
| **cron** | `0 0 3 * * ?`（每天 03:00） |
| **misfire_policy** | 3（立即执行） |
| **concurrent** | 1（禁止并发） |
| **注册位置** | `scheduler/job_invoke.go:31` |
| **任务实现** | `scheduler/github_sync_job.go` |

**装配方式**（避免 scheduler→service 循环依赖）：

```go
// main.go 装配阶段
scheduler.GitHubSyncFunc = registry.Portfolio.SyncFromGitHub
```

---

## 六、文件清单

| 文件 | 职责 |
|:---|:---|
| `internal/model/portfolio.go` | GORM 实体定义 |
| `internal/dto/portfolio_dto.go` | 前台/后台展示 DTO |
| `internal/vo/portfolio_vo.go` | 后台编辑请求 VO（含校验） |
| `internal/service/portfolio_service.go` | 业务逻辑（查询/编辑/删除/同步） |
| `internal/handler/portfolio_handler.go` | HTTP 处理器 |
| `internal/scheduler/github_sync_job.go` | 定时任务入口 |
| `internal/scheduler/job_invoke.go` | 任务注册 |
| `internal/config/config.go` | `GitHubConfig` 配置结构 |
| `scripts/portfolio.sql` | 建表 + 定时任务 + 菜单初始化 |
| `scripts/portfolio_menu_fix.sql` | 菜单修复脚本 |

---

## 七、数据库初始化

```bash
# 1. 建表 + 注册定时任务 + 注册后台菜单
mysql -u root -p aurora < scripts/portfolio.sql

# 2. （可选）菜单修复
mysql -u root -p aurora < scripts/portfolio_menu_fix.sql
```

**菜单注册**（`scripts/portfolio.sql`）：

```sql
INSERT INTO `t_menu` (`name`, `path`, `component`, `icon`, `order_num`, `is_hidden`, `parent_id`)
VALUES ('作品集管理', '/portfolios', 'portfolio/Portfolio.vue', 'el-icon-myimage-fill', 6, 0, 4);
```

> `component` 字段对应 aurora-admin-v3 的 `src/views/portfolio/Portfolio.vue`，前端需同步实现该组件。

---

## 八、错误码

| 错误码 | 常量 | 说明 |
|:---|:---|:---|
| - | `errors.ErrPortfolioNotFound` | 作品不存在（更新时 RowsAffected=0） |
| - | `errors.ErrPortfolioSyncFailed` | GitHub 同步失败（API 错误/网络错误） |

---

## 九、相关文档

- [README.md](../README.md) — 项目总览
- [API.md](API.md) — API 文档
- [MIGRATION_GUIDE.md](MIGRATION_GUIDE.md) — Java→Go 迁移对照
