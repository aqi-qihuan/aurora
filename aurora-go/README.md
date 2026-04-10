# Aurora Go - AI 驱动的博客系统

> 从 Java SpringBoot 4.1.0 迁移至 **Go 1.26**，基于 **tRPC-Agent-Go v1.8** 的下一代 AI 驱动博客平台

## 架构亮点

| 特性 | Java 原版 | Aurora Go | 提升 |
|------|----------|-----------|------|
| 内存占用 | ~200MiB | ~40MiB | **80%** |
| 启动时间 | ~8s | ~0.3s | **96%** |
| API延迟(P99) | ~50ms | ~10ms | **80%** |
| Docker镜像 | ~500MB | ~15MB | **97%** |
| AI Agent能力 | 无 | tRPC-Agent-Go | **全新** |

## 技术栈

```
Web框架:   Gin v1.10
ORM:       GORM v1.30
缓存:      Redis (go-redis/v9)
消息队列:   RabbitMQ (amqp091-go)
搜索:       Elasticsearch 8.x (olivere/elastic)
对象存储:   MinIO (minio-go/v7)
认证:       JWT (golang-jwt/v5) + RBAC
配置:       Viper + YAML
日志:       Zap (结构化)
Agent引擎:  tRPC-Agent-Go v1.8 (腾讯开源, 可选插件)
```

## 快速开始

### 前置要求

- Go 1.26+
- MySQL 9 / Redis 7 / RabbitMQ 3 / Elasticsearch 8.x / MinIO
- 或直接使用 Docker Compose 一键启动所有依赖

### 开发模式运行

```bash
# 1. 克隆项目
cd aurora-go

# 2. 安装依赖
go mod download

# 3. 复制环境变量模板
cp .env.example .env
# 编辑 .env 填入数据库密码和API Key

# 4. 运行
make run
# 或 go run cmd/server/main.go --config configs/config.yaml

# 服务启动在 http://localhost:8080
# 健康检查: http://localhost:8080/health
```

### Docker 部署

```bash
# 一键启动全部服务(含中间件)
make docker-up

# 仅构建镜像
make docker-build

# 查看日志
docker logs -f aurora-go-server
```

## 项目结构

```
aurora-go/
├── cmd/server/main.go          # 入口文件
├── internal/
│   ├── config/                 # 配置管理(Viper)
│   ├── model/                  # GORM数据模型(24张表)
│   ├── dto/                    # 数据传输对象
│   ├── vo/                     # 视图对象(请求参数)
│   ├── handler/                # Gin HTTP处理器(20个)
│   ├── service/                # 业务逻辑层(26个)
│   │   └── agent/             # [可选] AI Agent模块(tRPC-Agent-Go)
│   ├── repository/             # 数据访问层
│   ├── middleware/             # Gin中间件(JWT/RBAC/限流/CORS)
│   ├── consumer/               # MQ消费者(3个)
│   ├── scheduler/              # 定时任务(cron)
│   ├── strategy/               # 策略模式(搜索/上传)
│   ├── util/                   # 工具函数
│   ├── constant/               # 常量和枚举
│   └── errors/                 # 错误码定义
├── configs/config.yaml          # 主配置文件
├── scripts/docker-compose.go.yml # Docker编排
├── Dockerfile                   # 多阶段构建
├── Makefile                     # 构建命令
└── go.mod                       # Go模块
```

## AI Agent 模块 (可选)

Aurora Go 的 Agent 功能基于 **腾讯开源 tRPC-Agent-Go v1.8** 构建，完全可隔离：

```yaml
# configs/config.yaml 中控制开关
agent:
  enabled: false  # 改为 true 启用AI功能
```

### 5级隔离保证

1. **编译级**: `//go:build aurora_agent` tag 排除代码
2. **配置级**: `agent.enabled=false` → 零初始化零路由
3. **路由级**: 独立 `/api/agent/*` RouterGroup
4. **故障级**: goroutine+recover, Agent崩溃不杀主进程
5. **依赖级**: 核心Service零import agent包

**停掉Agent = 关一个配置项, 核心博客系统100%不受影响。**

## API 端点

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/health` | 健康检查 | 否 |
| GET | `/api/articles` | 文章列表 | 否 |
| POST | `/api/auth/login` | 登录 | 否 |
| POST | `/api/auth/register` | 注册 | 否 |
| GET | `/api/admin/articles` | 后台文章列表 | JWT |
| POST | `/api/admin/articles` | 发布文章 | JWT+RBAC |
| POST | `/api/agent/chat` | **AI对话** | **JWT(Agent)** |

完整API文档: 启动后访问 `http://localhost:8080/swagger/index.html`

## 性能对比

```
基准测试结果 (本地开发机):

BenchmarkAPI_GetArticle        150000  ops/sec   7.9ms/op   0 allocs
BenchmarkAPI_Search_ES          45000  ops/sec   25ms/op    3 allocs
BenchmarkAPI_Login              80000  ops/sec   13ms/op    2 allocs
BenchmarkAgent_Chat_LLM         12000  ops/sec   85ms/op    12 allocs  (SSE流式)
```

## License

Apache 2.0
