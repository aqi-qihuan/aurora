# Aurora Go - 博客系统 Go 后端

> 从 Java SpringBoot 4.1.0 完整迁移至 **Go 1.26**，兼容 Aurora 全套前端，支持可选 AI Agent 模块

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.26-00ADD8?style=for-the-badge&logo=go" />
  <img src="https://img.shields.io/badge/Gin-1.10-008ECF?style=for-the-badge" />
  <img src="https://img.shields.io/badge/GORM-1.30-7C4DFF?style=for-the-badge" />
  <img src="https://img.shields.io/badge/ES-8.19-FEC514?style=for-the-badge&logo=elasticsearch" />
  <img src="https://img.shields.io/badge/内存-~29MiB-success?style=for-the-badge" />
</p>

---

## 架构亮点

| 特性 | Java SpringBoot | Aurora Go | 提升 |
|------|-----------------|-----------|------|
| 运行时内存 | ~280 MiB (JVM) | **~29 MiB** | **↓89.6%** |
| 启动时间 | ~8s | **~0.3s** | **↓96%** |
| Docker 镜像 | ~180 MB (JRE) | **~5 MB** (Alpine) | **↓97.2%** |
| API 延迟 (P99) | ~50ms | **~10ms** | **↓80%** |
| 总内存占用 | ~1,587 MiB | **~1,336 MiB** | **↓15.8%** |
| AI Agent | — | tRPC-Agent-Go v1.8 (可选) | **全新** |
| 前端兼容 | ✅ | ✅ **100% API 兼容** | — |

---

## 技术栈

| 分类 | 技术 | 版本 | 说明 |
|:-----|:-----|:-----|:-----|
| Web 框架 | Gin | 1.10 | 高性能 HTTP 框架 |
| ORM | GORM | 1.30 | MySQL 数据访问 |
| 缓存 | go-redis | 9.7 | Redis Stack 客户端 |
| 消息队列 | amqp091-go | 1.10 | RabbitMQ 客户端 |
| 搜索引擎 | go-elasticsearch/v8 | 8.19 | ES 8.x 原生客户端 |
| 对象存储 | minio-go/v7 | 7.0 | MinIO 客户端 |
| 认证 | golang-jwt/v5 | 5.2 | JWT + RBAC 权限 |
| 定时任务 | robfig/cron | 3.0 | Cron 调度器 |
| 配置管理 | Viper | 1.19 | YAML + 环境变量 |
| 日志 | Zap | 1.27 | 结构化日志 |
| HTML 净化 | bluemonday | 1.0 | XSS 防护 |
| IP 地域 | ip2region | — | IP 归属地查询 |
| Agent 引擎 | tRPC-Agent-Go | 1.8 | 腾讯开源，可选插件 |

---

## 功能模块

| 模块 | Handler | Service | Model | 说明 |
|:-----|:-------:|:-------:|:-----:|:-----|
| 文章管理 | ✅ | ✅ | ✅ | CRUD、搜索、归档 |
| 分类管理 | ✅ | ✅ | ✅ | CRUD、层级管理 |
| 标签管理 | ✅ | ✅ | ✅ | CRUD (varchar 50) |
| 评论系统 | ✅ | ✅ | ✅ | 评论、回复、审核 |
| 友链管理 | ✅ | ✅ | ✅ | CRUD、上下架 |
| 说说管理 | ✅ | ✅ | ✅ | CRUD |
| 相册管理 | ✅ | ✅ | ✅ | 相册 + 照片 CRUD |
| 资源管理 | ✅ | ✅ | ✅ | 文件上传管理 |
| 菜单管理 | ✅ | ✅ | ✅ | 动态菜单、路由 |
| 角色权限 | ✅ | ✅ | ✅ | RBAC 角色权限 |
| 用户认证 | ✅ | ✅ | ✅ | 登录/注册/QQ OAuth |
| 网站配置 | ✅ | ✅ | ✅ | 全局配置管理 |
| 关于页面 | ✅ | ✅ | ✅ | 关于我内容管理 |
| 定时任务 | ✅ | ✅ | ✅ | Cron 任务 + 日志 |
| 文件上传 | ✅ | ✅ | — | MinIO/OSS 策略 |
| 全文搜索 | — | ✅ | — | MySQL/ES 策略 |
| 日志记录 | ✅ | ✅ | ✅ | 操作日志 + 异常日志 |
| AI Agent | ✅ | — | — | tRPC-Agent-Go 对话 |
| 作品集 Portfolio | ✅ | ✅ | ✅ | GitHub 仓库同步 + 后台覆盖封面/排序/置顶/可见性 |

---

## 项目结构

```
aurora-go/
├── cmd/
│   └── server/main.go              # 入口文件
├── internal/
│   ├── agent/                      # 🤖 AI Agent (12 files, tRPC-Agent-Go)
│   ├── config/                     # ⚙️ 配置管理 (Viper)
│   ├── constant/                   # 📋 常量和枚举
│   ├── consumer/                   # 📨 MQ 消费者 (Maxwell 文章同步等)
│   ├── dto/                        # 📦 数据传输对象
│   ├── errors/                     # ❌ 错误码定义
│   ├── handler/                    # 🌐 Gin HTTP 处理器 (24 files)
│   │   └── router.go              #    路由注册中心
│   ├── infrastructure/             # 🏗️ 基础设施 (DB/Redis/ES/MinIO 初始化)
│   ├── middleware/                  # 🔐 中间件 (9 files)
│   │   ├── jwt_auth.go            #    JWT 认证
│   │   ├── rbac.go                #    RBAC 权限
│   │   ├── ratelimit.go           #    限流
│   │   ├── cors.go                #    跨域
│   │   ├── recovery.go            #    异常恢复
│   │   ├── access_log.go          #    访问日志
│   │   └── nocache.go             #    禁缓存
│   ├── model/                      # 📊 GORM 数据模型 (17 files, 17张表)
│   ├── scheduler/                  # ⏰ 定时任务 (cron, 9 files)
│   ├── service/                    # 💼 业务逻辑层 (28 files)
│   │   └── registry.go            #    服务注册表
│   ├── strategy/                   # 🎯 策略模式 (9 files)
│   │   ├── search/                #    搜索策略 (MySQL/ES)
│   │   └── upload/                #    上传策略 (MinIO/OSS)
│   ├── util/                       # 🔧 工具函数 (10 files)
│   └── vo/                         # 📤 视图对象 (请求参数校验, 4 files)
├── cmd/
│   ├── server/main.go              # 入口文件
│   └── diag/                       # 🔍 诊断工具 (4 files)
├── configs/
│   └── config.yaml                 # 主配置文件
├── scripts/
│   ├── ip/ip2region.xdb           # IP 归属地数据库
│   ├── portfolio.sql              # 作品集模块初始化脚本
│   └── portfolio_menu_fix.sql     # 作品集菜单修复脚本
├── docs/                           # 📖 文档
│   ├── API.md                      # API 接口文档
│   ├── AGENT_GUIDE.md             # Agent 使用指南
│   ├── PORTFOLIO_GUIDE.md         # 作品集模块指南
│   ├── MIGRATION_GUIDE.md         # Java → Go 迁移指南
│   └── TEST_REPORT.md             # 测试报告
├── .env.example                    # 环境变量模板
├── docker-compose-go.yml           # Docker 编排 (含中间件)
├── Dockerfile                      # 多阶段构建
├── Makefile                        # 构建命令
├── .golangci.yml                   # Lint 配置
└── go.mod                          # Go 模块
```

---

## 快速开始

### 前置要求

- Go 1.26+
- MySQL 8+ / Redis 7+ / RabbitMQ 3+ / Elasticsearch 8.x / MinIO
- 或直接使用 Docker Compose 一键启动

### 开发模式

```bash
cd aurora-go

# 1. 安装依赖
go mod download

# 2. 配置
cp .env.example .env
# 编辑 .env 填入数据库密码等敏感信息

# 3. 运行
make run
# 或直接: go run cmd/server/main.go --config configs/config.yaml

# 服务启动在 http://localhost:8080
# 健康检查: curl http://localhost:8080/health
```

### 交叉编译 (Windows → Linux)

```powershell
# PowerShell
$env:CGO_ENABLED="0"
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -ldflags="-w -s" -o aurora-server-linux ./cmd/server
```

```bash
# Linux/macOS
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make build-linux
```

### Docker 部署

```bash
# 方式 1: 使用 Makefile
make docker-build     # 构建镜像
make docker-up        # 启动全部服务 (含中间件)
make docker-down      # 停止

# 方式 2: 手动
docker compose -f docker-compose-go.yml up -d
docker logs -f aurora-go
```

### 仅部署 Go 二进制 (Alpine)

```bash
# 1. 编译
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o aurora-server ./cmd/server

# 2. 上传到服务器
scp aurora-server configs/config.yaml user@server:/opt/aurora/app/

# 3. 启动容器
docker run -d --name aurora-go \
  --network aurora-network \
  -p 8080:8080 \
  -v /opt/aurora/app/aurora-server:/app/aurora-server \
  -v /opt/aurora/app/configs:/app/configs \
  -e TZ=Asia/Shanghai \
  --restart=always \
  alpine:3.20 \
  /app/aurora-server --config configs/config.yaml
```

---

## 配置说明

### 配置方式

支持 **YAML 文件** + **环境变量** 双配置，环境变量优先级更高：

```
# 环境变量映射规则 (Viper AutomaticEnv):
# AURORA_ 前缀 + 下划线分隔层级
# 例: mysql.password → AURORA_MYSQL_PASSWORD
# 例: redis.host     → AURORA_REDIS_HOST
# 例: es.username    → AURORA_ELASTICSEARCH_USERNAME
```

### 核心配置项

| 配置项 | 环境变量 | 默认值 | 说明 |
|:-------|:---------|:-------|:-----|
| `server.port` | `AURORA_SERVER_PORT` | 8080 | 服务端口 |
| `server.mode` | `AURORA_SERVER_MODE` | release | gin 模式: debug/release |
| `search.mode` | `AURORA_SEARCH_MODE` | elasticsearch | 搜索: mysql/elasticsearch |
| `upload.mode` | `AURORA_UPLOAD_MODE` | minio | 上传: minio/oss |
| `agent.enabled` | `AURORA_AGENT_ENABLED` | false | AI Agent 开关 |
| `github.enabled` | `AURORA_GITHUB_ENABLED` | false | GitHub 作品集同步开关 |
| `github.username` | `AURORA_GITHUB_USERNAME` | — | GitHub 用户名 |
| `github.token` | `AURORA_GITHUB_TOKEN` | — | GitHub PAT（建议环境变量注入） |
| `github.exclude` | `AURORA_GITHUB_EXCLUDE` | — | 排除仓库名（逗号分隔） |

### 环境变量模板

详见 [`.env.example`](.env.example)，包含完整的配置项说明。

> ⚠️ 敏感信息（密码、API Key）请通过环境变量注入，不要写在 config.yaml 中！

---

## API 端点

> 路径与 `internal/handler/router.go` 完全对齐，分三类：公开（无需认证）/ 受保护（JWT 登录）/ 后台（JWT + RBAC）。

#### 健康检查（根路径）

| 方法 | 路径 | 说明 |
|:-----|:-----|:-----|
| GET | `/health` | 健康检查（无需认证） |

### 公开 API（`/api`，无需认证）

#### 文章 / 归档

| 方法 | 路径 | 说明 |
|:-----|:-----|:-----|
| GET | `/api/articles/topAndFeatured` | 置顶 + 推荐文章 |
| GET | `/api/articles/all` | 文章列表（分页） |
| GET | `/api/articles/categoryId` | 按分类查询文章 |
| GET | `/api/articles/tagId` | 按标签查询文章 |
| GET | `/api/articles/search` | 搜索文章 |
| GET | `/api/articles/:id` | 文章详情 |
| POST | `/api/articles/access` | 验证加密文章密码 |
| GET | `/api/archives/all` | 归档列表 |

#### 用户认证

| 方法 | 路径 | 说明 |
|:-----|:-----|:-----|
| POST | `/api/users/login` | 登录 |
| POST | `/api/users/register` | 注册 |
| GET | `/api/users/code` | 发送邮箱验证码 |
| PUT | `/api/users/password` | 修改密码（通过 code 区分重置/修改） |
| POST | `/api/users/password/reset` | 重置密码 |
| POST | `/api/users/oauth/qq` | QQ OAuth 登录 |

#### 评论 / 分类 / 标签

| 方法 | 路径 | 说明 |
|:-----|:-----|:-----|
| GET | `/api/comments` | 评论列表 |
| GET | `/api/comments/topSix` | 最新 6 条评论 |
| GET | `/api/comments/:id/replies` | 评论回复列表 |
| GET | `/api/categories/all` | 分类列表 |
| GET | `/api/tags/all` | 标签列表 |
| GET | `/api/tags/topTen` | Top 10 标签 |

#### 友链 / 说说 / 相册

| 方法 | 路径 | 说明 |
|:-----|:-----|:-----|
| GET | `/api/links` | 友链列表 |
| GET | `/api/talks` | 说说列表 |
| GET | `/api/talks/:id` | 说说详情 |
| GET | `/api/photos/albums` | 相册列表 |
| GET | `/api/albums/:albumId/photos` | 相册照片列表 |

#### 首页 / 关于 / 配置 / 作品集

| 方法 | 路径 | 说明 |
|:-----|:-----|:-----|
| GET | `/api/` | 首页信息聚合 |
| GET | `/api/about` | 关于页 |
| GET | `/api/website/config` | 网站配置（前台只读） |
| GET | `/api/portfolios/featured` | 首页置顶作品集（6 条） |
| GET | `/api/portfolios` | 作品集分页列表 |
| POST | `/api/report` | 访客上报 |

### 受保护 API（`/api`，需 JWT 登录）

#### 评论互动

| 方法 | 路径 | 说明 |
|:-----|:-----|:-----|
| POST | `/api/comments/save` | 新增评论 |
| POST | `/api/comments/:id/reply` | 回复评论 |
| POST | `/api/comments/:id/like` | 点赞评论 |

#### 用户信息

| 方法 | 路径 | 说明 |
|:-----|:-----|:-----|
| PUT | `/api/users/info` | 更新用户信息 |
| POST | `/api/users/avatar` | 更新头像 |
| PUT | `/api/users/email` | 绑定邮箱 |
| PUT | `/api/users/subscribe` | 更新订阅状态 |
| POST | `/api/users/logout` | 登出 |
| GET | `/api/users/info/:id` | 获取用户信息 |

### 后台 API（`/api/admin`，需 JWT + RBAC）

#### 系统信息 / 配置

| 方法 | 路径 | 说明 |
|:-----|:-----|:-----|
| GET | `/api/admin/` | 后台首页信息 |
| GET | `/api/admin/website/config` | 获取网站配置（平铺） |
| PUT | `/api/admin/website/config` | 更新网站配置 |
| PUT | `/api/admin/about` | 更新关于页 |
| POST | `/api/admin/config/images` | 上传配置图片 |

#### 用户管理

| 方法 | 路径 | 说明 |
|:-----|:-----|:-----|
| GET | `/api/admin/users` | 用户列表 |
| GET | `/api/admin/users/area` | 用户地域统计 |
| POST | `/api/admin/users/area/trigger` | 手动触发地域统计 |
| PUT | `/api/admin/users/password` | 修改密码 |
| PUT | `/api/admin/users/role` | 修改用户角色 |
| PUT | `/api/admin/users/disable` | 禁用/启用用户 |
| GET | `/api/admin/users/online` | 在线用户列表 |
| DELETE | `/api/admin/users/:id/online` | 踢用户下线 |
| GET | `/api/admin/users/role` | 角色列表（编辑用户用） |

#### 文章管理

| 方法 | 路径 | 说明 |
|:-----|:-----|:-----|
| GET | `/api/admin/articles` | 后台文章列表 |
| POST | `/api/admin/articles` | 保存文章（统一入口，按 id 区分新增/更新） |
| POST | `/api/admin/articles/save` | 新增文章（兼容 Vue3 前端） |
| POST | `/api/admin/articles/update` | 更新文章（兼容 Vue3 前端） |
| PUT | `/api/admin/articles/topAndFeatured` | 更新置顶/推荐 |
| PUT | `/api/admin/articles` | 更新删除状态（软删除） |
| DELETE | `/api/admin/articles/delete` | 物理删除文章 |
| POST | `/api/admin/articles/images` | 上传文章图片 |
| GET | `/api/admin/articles/:id` | 文章详情 |
| POST | `/api/admin/articles/import` | 导入文章 |
| POST | `/api/admin/articles/export` | 导出文章 |

#### 分类 / 标签 / 评论管理

| 方法 | 路径 | 说明 |
|:-----|:-----|:-----|
| GET | `/api/admin/categories` | 分类列表 |
| GET | `/api/admin/categories/search` | 搜索分类 |
| POST | `/api/admin/categories` | 新增/更新分类 |
| DELETE | `/api/admin/categories` | 删除分类 |
| GET | `/api/admin/tags` | 标签列表 |
| GET | `/api/admin/tags/search` | 搜索标签 |
| POST | `/api/admin/tags` | 新增/更新标签 |
| DELETE | `/api/admin/tags` | 删除标签 |
| GET | `/api/admin/comments` | 评论列表 |
| PUT | `/api/admin/comments/review` | 审核评论 |
| DELETE | `/api/admin/comments` | 删除评论 |

#### 友链 / 说说 / 相册 / 照片管理

| 方法 | 路径 | 说明 |
|:-----|:-----|:-----|
| GET | `/api/admin/links` | 友链列表 |
| POST | `/api/admin/links` | 新增/更新友链 |
| PUT | `/api/admin/links` | 更新友链（兼容前端） |
| DELETE | `/api/admin/links` | 删除友链 |
| GET | `/api/admin/talks` | 说说列表 |
| GET | `/api/admin/talks/:id` | 说说详情 |
| POST | `/api/admin/talks` | 新增/更新说说 |
| POST | `/api/admin/talks/images` | 上传说说图片 |
| DELETE | `/api/admin/talks` | 删除说说 |
| GET | `/api/admin/photos/albums` | 相册列表 |
| GET | `/api/admin/photos/albums/info` | 相册信息列表 |
| GET | `/api/admin/photos/albums/:id/info` | 相册详情 |
| POST | `/api/admin/photos/albums` | 新增/更新相册 |
| POST | `/api/admin/photos/albums/upload` | 上传相册封面 |
| DELETE | `/api/admin/photos/albums/:id` | 删除相册 |
| GET | `/api/admin/photos` | 照片列表 |
| POST | `/api/admin/photos/upload` | 上传照片 |
| POST | `/api/admin/photos` | 批量保存照片 |
| PUT | `/api/admin/photos` | 更新照片 |
| PUT | `/api/admin/photos/album` | 移动照片到相册 |
| PUT | `/api/admin/photos/delete` | 更新照片删除状态 |
| DELETE | `/api/admin/photos` | 删除照片 |

#### 角色 / 资源 / 菜单管理

| 方法 | 路径 | 说明 |
|:-----|:-----|:-----|
| GET | `/api/admin/roles` | 角色列表 |
| POST | `/api/admin/role` | 新增/更新角色 |
| DELETE | `/api/admin/roles` | 删除角色 |
| GET | `/api/admin/role/menus` | 菜单选项（编辑角色用） |
| GET | `/api/admin/role/resources` | 资源选项（编辑角色用） |
| GET | `/api/admin/resources` | 资源列表 |
| POST | `/api/admin/resources` | 新增/更新资源 |
| DELETE | `/api/admin/resources/:id` | 删除资源 |
| GET | `/api/admin/menus` | 菜单列表 |
| POST | `/api/admin/menus` | 新增/更新菜单 |
| PUT | `/api/admin/menus/isHidden` | 更新菜单隐藏状态 |
| DELETE | `/api/admin/menus/:id` | 删除菜单 |
| GET | `/api/admin/user/menus` | 当前用户菜单 |

#### 定时任务 / 日志

| 方法 | 路径 | 说明 |
|:-----|:-----|:-----|
| GET | `/api/admin/jobs` | 任务列表 |
| POST | `/api/admin/jobs` | 新增任务 |
| PUT | `/api/admin/jobs` | 更新任务 |
| DELETE | `/api/admin/jobs` | 删除任务 |
| GET | `/api/admin/jobs/:id` | 任务详情 |
| PUT | `/api/admin/jobs/status` | 更新任务状态 |
| PUT | `/api/admin/jobs/run` | 立即运行一次 |
| GET | `/api/admin/jobs/jobGroups` | 任务分组列表 |
| GET | `/api/admin/jobLogs` | 任务日志列表 |
| DELETE | `/api/admin/jobLogs` | 删除任务日志 |
| DELETE | `/api/admin/jobLogs/clean` | 清空任务日志 |
| GET | `/api/admin/jobLogs/jobGroups` | 日志分组列表 |
| GET | `/api/admin/operation/logs` | 操作日志列表 |
| DELETE | `/api/admin/operation/logs` | 删除操作日志 |
| GET | `/api/admin/exception/logs` | 异常日志列表 |
| DELETE | `/api/admin/exception/logs` | 删除异常日志 |

#### 文件上传 / 作品集管理

| 方法 | 路径 | 说明 |
|:-----|:-----|:-----|
| POST | `/api/admin/upload` | 上传文件 |
| POST | `/api/admin/upload/batch` | 批量上传 |
| POST | `/api/admin/upload/image` | 上传图片 |
| GET | `/api/admin/portfolios` | 作品集列表（含隐藏项） |
| PUT | `/api/admin/portfolios` | 编辑作品集（封面/排序/置顶/可见性） |
| DELETE | `/api/admin/portfolios` | 批量删除作品集 |
| POST | `/api/admin/portfolios/sync` | 手动触发 GitHub 同步 |

### Agent API（可选，`/api/agent`，需 JWT + Agent 启用）

| 方法 | 路径 | 说明 |
|:-----|:-----|:-----|
| POST | `/api/agent/chat` | AI 对话（SSE 流式） |

> 📖 完整 API 文档见 [`docs/API.md`](docs/API.md)，作品集详情见 [`docs/PORTFOLIO_GUIDE.md`](docs/PORTFOLIO_GUIDE.md)

---

## AI Agent 模块 (可选)

Aurora Go 的 Agent 功能基于 **腾讯开源 tRPC-Agent-Go v1.8** 构建，完全可隔离：

```yaml
# configs/config.yaml 中控制开关
agent:
  enabled: false  # 改为 true 启用AI功能
```

### 5 级隔离保证

| 级别 | 机制 | 说明 |
|:-----|:-----|:-----|
| 编译级 | `//go:build aurora_agent` tag | 代码编译排除 |
| 配置级 | `agent.enabled=false` | 零初始化零路由 |
| 路由级 | 独立 `/api/agent/*` RouterGroup | 路由隔离 |
| 故障级 | goroutine + recover | Agent 崩溃不杀主进程 |
| 依赖级 | 核心Service零import agent包 | 依赖隔离 |

**停掉 Agent = 关一个配置项，核心博客系统 100% 不受影响。**

### 支持的 LLM

| Provider | 模型 | 配置项 |
|:---------|:-----|:-------|
| OpenAI | gpt-4o-mini | `OPENAI_API_KEY` / `OPENAI_BASE_URL` |
| DeepSeek | deepseek-chat | `DEEPSEEK_API_KEY` / `DEEPSEEK_BASE_URL` |
| 通义千问 | qwen-plus | `QWEN_API_KEY` |
| Claude | claude-sonnet-4 | `CLAUDE_API_KEY` |

> 📖 详见 [`docs/AGENT_GUIDE.md`](docs/AGENT_GUIDE.md)

---

## 策略模式

搜索和上传均采用策略模式，支持灵活切换：

### 搜索策略

| 模式 | 实现 | 说明 |
|:-----|:-----|:-----|
| `mysql` | MySQL LIKE 查询 | 无需 ES，适合小数据量 |
| `elasticsearch` | ES 8.x 全文检索 | 推荐，支持中文分词 |

### 上传策略

| 模式 | 实现 | 说明 |
|:-----|:-----|:-----|
| `minio` | MinIO 对象存储 | 自托管，推荐 |
| `oss` | 阿里云 OSS | 云存储 |

---

## Docker Compose 编排

`docker-compose-go.yml` 包含完整的中间件 + Go 后端：

| 服务 | 镜像 | 端口 | 说明 |
|:-----|:-----|:-----|:-----|
| aurora-go | alpine:3.20 | 8080 | Go 后端 |
| aurora-mysql | mysql:8.0.32 | 3306 | 数据库 |
| aurora-redis | redis/redis-stack-server | 6379 | 缓存 |
| aurora-rabbitmq | rabbitmq:3.11.9-management | 5672/15672 | 消息队列 |
| aurora-elasticsearch | elasticsearch:8.19.14 | 9200 | 全文检索 |
| aurora-minio | bitnami/minio:2023.12.7 | 9000/9001 | 对象存储 |
| aurora-nginx | nginx:1.23.3 | 80/443 | 反向代理 |
| aurora-maxwell | zendesk/maxwell:latest | — | MySQL→MQ 同步 |

> ES 8.19.14 已深度优化：禁用 GeoIP/ML/Watcher，内存限制 640M

### 内存占用优化（实测）

| 服务 | 内存使用 | 内存占比 | 说明 |
|:-----|:---------|:---------|:-----|
| **aurora-go** | **~29 MiB** | **0.85%** | Go 静态编译，极致轻量 |
| aurora-mysql | ~264 MiB | 7.76% | 优化配置 (innodb_buffer_pool_size=128M) |
| aurora-redis | ~5 MiB | 0.15% | 极致优化 (maxmemory 32mb) |
| aurora-rabbitmq | ~110 MiB | 3.22% | 限制 Erlang VM 内存 |
| aurora-elasticsearch | ~579 MiB / 640 MiB | 90.44% | 限制 640M，正常缓存使用 |
| aurora-minio | ~158 MiB | 4.64% | 对象存储 |
| aurora-maxwell | ~180 MiB | 5.30% | MySQL binlog 同步 |
| aurora-nginx | ~11 MiB | 0.34% | 反向代理 |

**Go 模式总内存：~1,336 MiB** vs **Java 模式：~1,587 MiB**，节省 **~251 MiB (15.8%)**

> 💡 Go 后端内存仅 ~29 MiB，相比 Java 后端 (~280 MiB) 节省 **~251 MiB (89.6%)**

---

## Makefile 命令

```bash
make build         # 构建二进制
make run           # 运行开发服务器
make test          # 运行测试
make bench         # 基准测试
make lint          # golangci-lint 检查
make fmt           # 格式化代码
make clean         # 清理构建产物
make docker-build  # Docker 构建镜像
make docker-up     # Docker Compose 启动
make docker-down   # Docker Compose 停止
make build-linux   # 交叉编译 Linux amd64
make build-arm64   # 交叉编译 Linux arm64
make docs          # 生成 Swagger 文档
make help          # 查看所有命令
```

---

## 使用示例

### cURL 示例

#### 1. 用户登录

```bash
# 登录获取 JWT Token
curl -X POST http://localhost:8080/api/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "your_password"
  }'

# 响应示例:
# {
#   "code": 200,
#   "data": {
#     "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
#     "user": {
#       "id": 1,
#       "username": "admin",
#       "nickname": "管理员",
#       "roleList": ["admin"]
#     }
#   }
# }
```

#### 2. 获取文章列表 (公开)

```bash
# 无需认证即可访问
curl -X GET "http://localhost:8080/api/articles/all?current=1&size=10" \
  -H "Content-Type: application/json"

# 带关键词搜索
curl -X GET "http://localhost:8080/api/articles/all?keyword=Go语言&current=1&size=10" \
  -H "Content-Type: application/json"
```

#### 3. 发布文章 (需要认证)

```bash
# 使用获取的 Token
TOKEN="your_jwt_token_here"

curl -X POST http://localhost:8080/api/admin/articles \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "articleTitle": "Go语言入门指南",
    "articleContent": "# Go语言简介\n\nGo是一门开源编程语言...",
    "articleCover": "https://example.com/cover.jpg",
    "categoryId": 1,
    "tagName": ["Go", "后端"],
    "status": 1,
    "isTop": false,
    "isFeatured": false,
    "type": 1
  }'
```

#### 4. 上传图片 (需要认证)

```bash
# 上传文章封面或图片
curl -X POST http://localhost:8080/api/admin/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@/path/to/image.jpg"

# 响应示例:
# {
#   "code": 200,
#   "data": {
#     "url": "https://minio.example.com/aurora/2026/04/image.jpg",
#     "filename": "image.jpg",
#     "size": 102400
#   }
# }
```

#### 5. AI 对话 (需要认证 + Agent 启用)

```bash
# SSE 流式对话
curl -X POST http://localhost:8080/api/agent/chat \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept: text/event-stream" \
  -d '{
    "message": "帮我写一篇关于Gin框架的博客",
    "model": "deepseek-chat",
    "stream": true
  }'

# 非流式对话
curl -X POST http://localhost:8080/api/agent/chat \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "message": "Go语言的优缺点是什么？",
    "stream": false
  }'
```

#### 6. 搜索文章

```bash
# MySQL 搜索模式
curl -X GET "http://localhost:8080/api/articles/search?keyword=Gin&mode=mysql" \
  -H "Content-Type: application/json"

# Elasticsearch 搜索模式 (推荐)
curl -X GET "http://localhost:8080/api/articles/search?keyword=Go语言&mode=es" \
  -H "Content-Type: application/json"
```

---

### Go 客户端示例

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type LoginResponse struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    struct {
        AccessToken string `json:"accessToken"`
        User        struct {
            ID       int      `json:"id"`
            Username string   `json:"username"`
            RoleList []string `json:"roleList"`
        } `json:"user"`
    } `json:"data"`
}

func main() {
    // 1. 登录获取 Token
    loginReq := LoginRequest{
        Username: "admin",
        Password: "your_password",
    }

    reqBody, _ := json.Marshal(loginReq)
    resp, err := http.Post(
        "http://localhost:8080/api/users/login",
        "application/json",
        bytes.NewBuffer(reqBody),
    )
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)
    var loginResp LoginResponse
    json.Unmarshal(body, &loginResp)

    if loginResp.Code != 200 {
        panic("登录失败: " + loginResp.Message)
    }

    token := loginResp.Data.AccessToken
    fmt.Println("登录成功! Token:", token[:20]+"...")

    // 2. 使用 Token 访问受保护接口
    req, _ := http.NewRequest("GET", "http://localhost:8080/api/admin/articles", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err = client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    body, _ = io.ReadAll(resp.Body)
    fmt.Println("文章列表:", string(body))
}
```

---

### JavaScript/TypeScript 示例

```typescript
// api-client.ts
class AuroraAPI {
  private baseURL: string;
  private token: string | null = null;

  constructor(baseURL: string = 'http://localhost:8080') {
    this.baseURL = baseURL;
  }

  // 登录
  async login(username: string, password: string): Promise<string> {
    const response = await fetch(`${this.baseURL}/api/users/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password }),
    });

    const data = await response.json();
    if (data.code !== 200) {
      throw new Error(data.message);
    }

    this.token = data.data.accessToken;
    return this.token;
  }

  // 获取文章列表
  async getArticles(params: {
    current?: number;
    size?: number;
    keyword?: string;
    categoryId?: number;
  } = {}) {
    const query = new URLSearchParams();
    if (params.current) query.set('current', params.current.toString());
    if (params.size) query.set('size', params.size.toString());
    if (params.keyword) query.set('keyword', params.keyword);
    if (params.categoryId) query.set('categoryId', params.categoryId.toString());

    const response = await fetch(
      `${this.baseURL}/api/articles?${query.toString()}`
    );
    return response.json();
  }

  // 发布文章 (需要认证)
  async createArticle(article: {
    articleTitle: string;
    articleContent: string;
    categoryId: number;
    tagName?: string[];
    status?: number;
  }) {
    if (!this.token) throw new Error('请先登录');

    const response = await fetch(`${this.baseURL}/api/admin/articles`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${this.token}`,
      },
      body: JSON.stringify(article),
    });

    return response.json();
  }

  // AI 对话 (SSE 流式)
  async chat(message: string, onToken: (token: string) => void): Promise<void> {
    if (!this.token) throw new Error('请先登录');

    const response = await fetch(`${this.baseURL}/api/agent/chat`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${this.token}`,
        'Accept': 'text/event-stream',
      },
      body: JSON.stringify({
        message,
        stream: true,
      }),
    });

    const reader = response.body?.getReader();
    const decoder = new TextDecoder();

    while (reader) {
      const { done, value } = await reader.read();
      if (done) break;

      const chunk = decoder.decode(value);
      const lines = chunk.split('\n');

      for (const line of lines) {
        if (line.startsWith('data:')) {
          const data = JSON.parse(line.slice(5));
          if (data.type === 'token') {
            onToken(data.content);
          }
        }
      }
    }
  }
}

// 使用示例
async function main() {
  const api = new AuroraAPI();

  // 登录
  await api.login('admin', 'password123');

  // 获取文章列表
  const articles = await api.getArticles({ current: 1, size: 10 });
  console.log('文章列表:', articles);

  // AI 对话
  await api.chat('Go语言有什么优势?', (token) => {
    process.stdout.write(token); // 实时输出 AI 回复
  });
}

main();
```

---

### Postman/Insomnia 集合

项目提供了 Postman 集合文件,可直接导入:

```bash
# 下载 Postman 集合
curl -O https://raw.githubusercontent.com/your-repo/aurora-go/main/docs/postman_collection.json

# 或手动导入 docs/postman_collection.json
```

**集合包含:**
- 所有 API 端点
- 自动 Token 管理 (登录后自动提取 Token)
- 环境变量配置
- 测试用例

---

## 贡献指南

感谢你对 Aurora Go 项目的关注!我们欢迎任何形式的贡献,包括但不限于:提交代码、改进文档、报告 Bug、提出新功能建议。

### 📋 目录

1. [行为准则](#行为准则)
2. [如何贡献](#如何贡献)
3. [报告 Bug](#报告-bug)
4. [提出功能建议](#提出功能建议)
5. [提交代码](#提交代码)
6. [代码规范](#代码规范)
7. [测试要求](#测试要求)
8. [提交信息格式](#提交信息格式)
9. [代码审查](#代码审查)

---

### 行为准则

参与本项目的所有贡献者均须遵守以下准则:

- **尊重他人**: 使用友好、包容的语言
- **接受建设性批评**: 以专业态度对待反馈
- **关注问题本身**: 对事不对人
- **维护社区和谐**: 禁止骚扰、歧视、攻击性言论

---

### 如何贡献

#### 1. Fork 项目仓库

```bash
# 在 GitHub 上 Fork 项目后,克隆你的 Fork
git clone https://github.com/your-username/aurora-go.git
cd aurora-go

# 添加上游仓库
git remote add upstream https://github.com/original-owner/aurora-go.git
```

#### 2. 创建功能分支

```bash
# 从 develop 分支创建新分支
git checkout develop
git pull upstream develop
git checkout -b feat/add-comment-reply-notification
```

#### 3. 进行开发

- 编写代码并添加测试
- 确保测试通过: `make test`
- 确保 Lint 通过: `make lint`
- 更新相关文档

#### 4. 提交代码

```bash
# 提交前运行格式化
make fmt

# 提交 (使用 Conventional Commits 格式)
git add .
git commit -m "feat: add comment reply notification feature"

# 推送到你的 Fork
git push origin feat/add-comment-reply-notification
```

#### 5. 创建 Pull Request

- 在 GitHub 上创建 PR
- 填写 PR 模板
- 等待代码审查

---

### 报告 Bug

#### 在提交 Issue 前,请确认:

- [ ] 已搜索现有 Issues,确认没有重复
- [ ] 使用的是最新版本
- [ ] 提供了完整的复现步骤

#### Issue 模板

```markdown
**Bug 描述**
简明描述遇到的问题

**复现步骤**
1. 执行 '...'
2. 配置 '...'
3. 看到错误

**预期行为**
描述你期望发生什么

**实际行为**
描述实际发生了什么

**环境信息**
- OS: [e.g., Windows 11, Ubuntu 22.04]
- Go 版本: [e.g., 1.26]
- Aurora Go 版本: [e.g., 1.0.0]
- 数据库: [e.g., MySQL 8.0]

**日志输出**
粘贴相关错误日志 (如有)

**截图**
如果适用,添加截图
```

---

### 提出功能建议

我们欢迎新功能建议!请在使用 Issues 提交时包含:

- **功能描述**: 清晰描述你希望添加的功能
- **使用场景**: 解释为什么需要这个功能
- **替代方案**: 描述你考虑过的其他方案
- **示例**: 如果可能,提供伪代码或示例

---

### 提交代码

#### 分支命名规范

| 类型 | 前缀 | 示例 |
|------|------|------|
| 新功能 | `feat/` | `feat/add-oauth-github` |
| Bug 修复 | `fix/` | `fix/login-token-expiry` |
| 文档 | `docs/` | `docs/update-api-reference` |
| 性能优化 | `perf/` | `perf/optimize-article-query` |
| 重构 | `refactor/` | `refactor/simplify-handler` |
| 测试 | `test/` | `test/add-user-service-tests` |
| CI/CD | `ci/` | `ci/add-github-actions` |
| 样式修复 | `style/` | `style/fix-golint-warnings` |

#### Pull Request 规范

**PR 标题格式** (遵循 Conventional Commits):

```
feat: add GitHub OAuth login support
fix: resolve article pagination issue
docs: update API documentation for /upload endpoint
perf: optimize Elasticsearch query performance
```

**PR 描述模板**:

```markdown
## 变更类型
- [ ] Bug 修复
- [ ] 新功能
- [ ] 性能优化
- [ ] 重构
- [ ] 文档更新
- [ ] 测试更新

## 变更描述
详细描述本次变更的内容

## 关联 Issue
Closes #123
Relates to #456

## 测试计划
描述如何测试这些变更

## 检查清单
- [ ] 代码符合项目规范
- [ ] 已添加必要的测试
- [ ] 所有测试通过
- [ ] 已更新相关文档
- [ ] 没有产生新的警告
```

---

### 代码规范

#### Go 代码规范

1. **遵循 Go 官方规范**
   - 使用 `gofmt` 或 `goimports` 格式化代码
   - 运行 `make fmt`  before提交

2. **命名规范**
   ```go
   // ✅ 正确示例
   type UserService interface {}
   func GetUserByID(id int64) (*User, error)
   var userCount int
   
   // ❌ 错误示例
   type userService interface {}
   func getuserbyid(id int64) (*User, error)
   var user_count int
   ```

3. **注释规范**
   ```go
   // UserService defines business logic for user management.
   // It provides methods for creating, updating, and retrieving users.
   type UserService interface {
       // GetUserByID retrieves a user by their unique identifier.
       // Returns ErrUserNotFound if the user does not exist.
       GetUserByID(ctx context.Context, id int64) (*User, error)
   }
   ```

4. **错误处理**
   ```go
   // ✅ 正确: 显式处理错误
   user, err := userService.GetUserByID(ctx, userID)
   if err != nil {
       return nil, fmt.Errorf("failed to get user: %w", err)
   }
   
   // ❌ 错误: 忽略错误
   user, _ := userService.GetUserByID(ctx, userID)
   ```

5. **依赖注入**
   ```go
   // ✅ 正确: 使用接口 + 依赖注入
   type ArticleHandler struct {
       articleService service.ArticleService
       logger        *zap.Logger
   }
   
   func NewArticleHandler(as service.ArticleService, logger *zap.Logger) *ArticleHandler {
       return &ArticleHandler{
           articleService: as,
           logger:        logger,
       }
   }
   ```

#### Gin 处理规范

```go
// Handler 示例
func (h *ArticleHandler) GetArticle(c *gin.Context) {
    // 1. 提取路径参数
    articleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        util.ResponseError(c, ErrInvalidArticleID)
        return
    }

    // 2. 调用 Service 层
    article, err := h.articleService.GetArticleByID(c.Request.Context(), articleID)
    if err != nil {
        util.ResponseError(c, err)
        return
    }

    // 3. 返回成功响应
    util.ResponseSuccess(c, article)
}
```

#### 数据库模型规范

```go
// Model 示例
type Article struct {
    ID            int64     `gorm:"primaryKey;autoIncrement" json:"id"`
    ArticleTitle   string    `gorm:"type:varchar(200);not null" json:"articleTitle" binding:"required,max=200"`
    ArticleContent string    `gorm:"type:text" json:"articleContent"`
    ArticleCover   string    `gorm:"type:varchar(500)" json:"articleCover"`
    CategoryID    int64     `gorm:"index" json:"categoryId"`
    Status        int       `gorm:"default:1" json:"status"` // 1-公开 2-草稿 3-密码
    IsTop         bool      `gorm:"default:false" json:"isTop"`
    IsFeatured    bool      `gorm:"default:false" json:"isFeatured"`
    ViewCount     int       `gorm:"default:0" json:"viewCount"`
    LikeCount     int       `gorm:"default:0" json:"likeCount"`
    CreatedAt     time.Time `json:"createTime"`
    UpdatedAt     time.Time `json:"updateTime"`
}
```

---

### 测试要求

#### 1. 单元测试

- **覆盖率要求**: 核心业务逻辑覆盖率 ≥ 80%
- **测试文件命名**: `*.test.go` (e.g., `article_service_test.go`)
- **测试函数命名**: `Test<FunctionName>_<Scenario>` (e.g., `TestGetUserByID_NotFound`)

**测试示例**:

```go
func TestGetArticleByID_Success(t *testing.T) {
    // Arrange
    mockRepo := new(MockArticleRepository)
    mockRepo.On("FindByID", int64(1)).Return(&Article{
        ID:          1,
        ArticleTitle: "Test Article",
    }, nil)

    service := NewArticleService(mockRepo)

    // Act
    article, err := service.GetArticleByID(context.Background(), 1)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, int64(1), article.ID)
    assert.Equal(t, "Test Article", article.ArticleTitle)
    mockRepo.AssertExpectations(t)
}

func TestGetArticleByID_NotFound(t *testing.T) {
    // Arrange
    mockRepo := new(MockArticleRepository)
    mockRepo.On("FindByID", int64(999)).Return(nil, ErrArticleNotFound)

    service := NewArticleService(mockRepo)

    // Act
    article, err := service.GetArticleByID(context.Background(), 999)

    // Assert
    assert.Error(t, err)
    assert.Nil(t, article)
    assert.Equal(t, ErrArticleNotFound, err)
    mockRepo.AssertExpectations(t)
}
```

#### 2. 集成测试

```go
// integration_test.go
func TestArticleAPI_E2E(t *testing.T) {
    // 启动测试服务器
    router := setupTestRouter()
    defer cleanupTestDB()

    // 登录获取 Token
    token := loginTestUser(t, router)

    // 测试创建文章
    req := httptest.NewRequest("POST", "/api/admin/articles", strings.NewReader(`{
        "articleTitle": "Test Article",
        "articleContent": "Content here",
        "categoryId": 1
    }`))
    req.Header.Set("Authorization", "Bearer "+token)
    req.Header.Set("Content-Type", "application/json")

    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, 200, w.Code)
}
```

#### 3. 运行测试

```bash
# 运行所有测试
make test

# 运行特定测试
go test -v ./internal/service -run TestGetArticleByID

# 查看覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# 运行基准测试
make bench
```

---

### 提交信息格式

遵循 [Conventional Commits](https://www.conventionalcommits.org/) 规范:

#### 提交类型

| 类型 | 说明 | 示例 |
|------|------|------|
| `feat` | 新功能 | `feat: add user registration API` |
| `fix` | Bug 修复 | `fix: resolve token expiry issue` |
| `docs` | 文档更新 | `docs: update API.md with new endpoints` |
| `style` | 代码格式 (不影响功能) | `style: format code with gofmt` |
| `refactor` | 重构 | `refactor: simplify error handling in service` |
| `perf` | 性能优化 | `perf: optimize article list query` |
| `test` | 测试相关 | `test: add unit tests for UserService` |
| `chore` | 构建/工具相关 | `chore: upgrade gin to v1.10` |
| `ci` | CI/CD 相关 | `ci: add GitHub Actions workflow` |

#### 提交格式

```
<type>(<scope>): <subject>

<body>

<footer>
```

**示例**:

```commit
feat(auth): add GitHub OAuth login support

- Add GitHub OAuth strategy
- Update User model to store GitHub ID
- Add migration for github_id column

Closes #123
```

```commit
fix(article): resolve pagination issue when page size is 0

When size=0 was passed, the API returned all records instead of
returning a validation error. Now returns 400 with error message.

Fixes #456
```

---

### 代码审查

#### 审查者在审查代码时应关注:

1. **功能正确性**
   - 代码是否实现了预期功能?
   - 边界条件是否处理?
   - 错误处理是否完善?

2. **代码质量**
   - 命名是否清晰?
   - 函数是否过长? (建议 ≤ 50 行)
   - 是否有重复代码可以抽取?

3. **性能**
   - 是否有 N+1 查询问题?
   - 是否有不必要的内存分配?
   - 数据库查询是否优化?

4. **安全性**
   - 是否有 SQL 注入风险?
   - 输入验证是否充分?
   - 敏感信息是否泄露?

5. **测试**
   - 是否有足够的测试覆盖?
   - 测试用例是否合理?
   - 是否包含边界情况?

#### 审查流程

1. **提交 PR** → 2. **自动检查** (CI/CD) → 3. **代码审查** → 4. **修改反馈** → 5. **批准合并**

#### 审查建议语气

- ✅ "建议考虑使用 `errors.Is()` 来检查错误,更符合 Go 习惯"
- ❌ "你这样写错误检查是错的"

- ✅ "这个函数有点长,是否可以拆分成几个小函数以提高可读性?"
- ❌ "函数太长了,重写"

---

### 发布流程

#### 版本号规则 (Semantic Versioning)

```
MAJOR.MINOR.PATCH (e.g., 1.2.3)

MAJOR: 不兼容的 API 修改
MINOR: 向后兼容的功能新增
PATCH: 向后兼容的 bug 修复
```

#### 发布步骤

1. 更新 `CHANGELOG.md`
2. 创建版本标签: `git tag -a v1.2.0 -m "Release v1.2.0"`
3. 推送标签: `git push origin v1.2.0`
4. GitHub Actions 自动构建并发布 Release

---

### 联系

- **作者**: 七七
- **QQ**: 2316364297
- **网站**: https://www.aqi125.cn
- **GitHub**: [aqi-qihuan/aurora: 基于SpringBoot4.1.X+Vue3开发的个人博客系统](https://github.com/aqi-qihuan/aurora))

---

### 致谢

感谢所有贡献者的付出! ❤️

<a href="https://github.com/your-repo/aurora-go/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=your-repo/aurora-go" />
</a>

---

## 数据库注意事项

| 表 | 列 | 类型 | 说明 |
|:---|:---|:-----|:-----|
| `t_tag` | `tag_name` | `varchar(50)` | 已从 20 扩展至 50，Go 验证规则 `max=50` |
| `t_category` | `category_name` | `varchar(50)` | 已从 20 扩展至 50，Go 验证规则 `max=50` |

---

## 相关文档

| 文档 | 说明 |
|:-----|:-----|
| [`docs/API.md`](docs/API.md) | 完整 API 接口文档 |
| [`docs/AGENT_GUIDE.md`](docs/AGENT_GUIDE.md) | AI Agent 使用指南 |
| [`docs/PORTFOLIO_GUIDE.md`](docs/PORTFOLIO_GUIDE.md) | 作品集模块指南 |
| [`docs/DIAG_TOOLS.md`](docs/DIAG_TOOLS.md) | cmd/diag 诊断工具集说明 |
| [`docs/MIGRATION_GUIDE.md`](docs/MIGRATION_GUIDE.md) | Java → Go 迁移指南 |
| [`docs/TEST_REPORT.md`](docs/TEST_REPORT.md) | 测试报告 |
| [`.env.example`](.env.example) | 环境变量模板 |
| [`MinIO上传功能实现说明.md`](MinIO上传功能实现说明.md) | MinIO 上传实现说明 |

---

## License

Apache 2.0
