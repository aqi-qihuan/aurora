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
│   ├── handler/                    # 🌐 Gin HTTP 处理器 (23 files)
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
│   ├── model/                      # 📊 GORM 数据模型 (16 files, 16张表)
│   ├── scheduler/                  # ⏰ 定时任务 (cron)
│   ├── service/                    # 💼 业务逻辑层 (27 files)
│   │   └── registry.go            #    服务注册表
│   ├── strategy/                   # 🎯 策略模式 (9 files)
│   │   ├── search/                #    搜索策略 (MySQL/ES)
│   │   └── upload/                #    上传策略 (MinIO/OSS)
│   ├── util/                       # 🔧 工具函数 (10 files)
│   └── vo/                         # 📤 视图对象 (请求参数校验)
├── configs/
│   └── config.yaml                 # 主配置文件
├── scripts/
│   └── ip/ip2region.xdb           # IP 归属地数据库
├── docs/                           # 📖 文档
│   ├── API.md                      # API 接口文档
│   ├── AGENT_GUIDE.md             # Agent 使用指南
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

### 环境变量模板

详见 [`.env.example`](.env.example)，包含完整的配置项说明。

> ⚠️ 敏感信息（密码、API Key）请通过环境变量注入，不要写在 config.yaml 中！

---

## API 端点

### 前台 API

| 方法 | 路径 | 说明 | 认证 |
|:-----|:-----|:-----|:-----|
| GET | `/health` | 健康检查 | 否 |
| GET | `/api/articles` | 文章列表 | 否 |
| GET | `/api/articles/:id` | 文章详情 | 否 |
| GET | `/api/categories` | 分类列表 | 否 |
| GET | `/api/tags` | 标签列表 | 否 |
| GET | `/api/links` | 友链列表 | 否 |
| GET | `/api/talks` | 说说列表 | 否 |
| GET | `/api/photos/albums` | 相册列表 | 否 |
| GET | `/api/comments` | 评论列表 | 否 |
| GET | `/api/website/config` | 网站配置 | 否 |
| GET | `/api/about` | 关于页面 | 否 |
| POST | `/api/auth/login` | 登录 | 否 |
| POST | `/api/auth/register` | 注册 | 否 |
| POST | `/api/auth/qqLogin` | QQ 登录 | 否 |
| POST | `/api/search` | 全文搜索 | 否 |

### 后台 API

| 方法 | 路径 | 说明 | 认证 |
|:-----|:-----|:-----|:-----|
| GET | `/api/admin/articles` | 后台文章列表 | JWT |
| POST | `/api/admin/articles` | 发布文章 | JWT+RBAC |
| PUT | `/api/admin/articles` | 更新文章 | JWT+RBAC |
| DELETE | `/api/admin/articles` | 删除文章 | JWT+RBAC |
| GET/POST/PUT/DELETE | `/api/admin/categories` | 分类管理 | JWT+RBAC |
| GET/POST/PUT/DELETE | `/api/admin/tags` | 标签管理 | JWT+RBAC |
| GET/POST/PUT/DELETE | `/api/admin/comments` | 评论管理 | JWT+RBAC |
| GET/POST/PUT/DELETE | `/api/admin/links` | 友链管理 | JWT+RBAC |
| GET/POST/PUT/DELETE | `/api/admin/talks` | 说说管理 | JWT+RBAC |
| GET/POST/PUT/DELETE | `/api/admin/photos/albums` | 相册管理 | JWT+RBAC |
| GET/POST/PUT/DELETE | `/api/admin/resources` | 资源管理 | JWT+RBAC |
| GET/POST/PUT/DELETE | `/api/admin/roles` | 角色管理 | JWT+RBAC |
| GET/POST/PUT/DELETE | `/api/admin/menus` | 菜单管理 | JWT+RBAC |
| GET/PUT | `/api/admin/website/config` | 网站配置 | JWT+RBAC |
| GET/PUT | `/api/admin/about` | 关于配置 | JWT+RBAC |
| GET/POST/PUT/DELETE | `/api/admin/jobs` | 定时任务管理 | JWT+RBAC |
| GET/DELETE | `/api/admin/jobLogs` | 任务日志 | JWT+RBAC |
| POST | `/api/admin/files/upload` | 文件上传 | JWT |

### Agent API (可选)

| 方法 | 路径 | 说明 | 认证 |
|:-----|:-----|:-----|:-----|
| POST | `/api/agent/chat` | AI 对话 (SSE 流式) | JWT (Agent) |

> 📖 完整 API 文档见 [`docs/API.md`](docs/API.md)

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
| [`docs/MIGRATION_GUIDE.md`](docs/MIGRATION_GUIDE.md) | Java → Go 迁移指南 |
| [`docs/TEST_REPORT.md`](docs/TEST_REPORT.md) | 测试报告 |
| [`.env.example`](.env.example) | 环境变量模板 |
| [`MinIO上传功能实现说明.md`](MinIO上传功能实现说明.md) | MinIO 上传实现说明 |

---

## License

Apache 2.0
