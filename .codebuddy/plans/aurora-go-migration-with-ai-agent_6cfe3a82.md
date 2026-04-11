---
name: aurora-go-migration-with-ai-agent
overview: 将 Aurora 博客系统从 SpringBoot 4.1.0-M4 + Java 25 迁移到 Go 1.26，并在架构中融入 AI Agent 能力，包括技术选型、架构设计、中间件映射、分阶段实施计划
design:
  architecture:
    framework: react
    component: shadcn
  styleKeywords:
    - Cyberpunk Neon
    - Glassmorphism
    - Dark Theme
    - AI-Tech Future
    - Gradient Glow
    - Micro-interaction
  fontSystem:
    fontFamily: "'Inter', 'SF Pro Display', system-ui, sans-serif"
    heading:
      size: 28px
      weight: 700
    subheading:
      size: 20px
      weight: 600
    body:
      size: 15px
      weight: 400
  colorSystem:
    primary:
      - "#00D4FF"
      - "#7B2FBE"
      - "#FF2D95"
    background:
      - "#0A0E17"
      - "#111827"
      - "#1A1F36"
    text:
      - "#E8ECF4"
      - "#9CA3AF"
      - "#6B7280"
    functional:
      - "#00FF88"
      - "#FF4757"
      - "#FFB800"
      - "#3B82F6"
todos:
  - id: init-project
    content: 初始化 aurora-go 项目骨架：go.mod/Dockerfile/Makefile/configs/config.yaml/目录结构和Gin路由框架搭建
    status: completed
  - id: model-layer
    content: 用[golang-pro][golang-patterns]实现24张表的GORM Model定义 + DTO/VO结构体 + 枚举常量 + 错误码体系
    status: completed
    dependencies:
      - init-project
  - id: infra-layer
    content: 实现基础设施层：MySQL/Redis/RabbitMQ/ES/MinIO/SMTP连接池封装 + Viper配置加载 + Zap日志
    status: completed
    dependencies:
      - model-layer
  - id: core-handlers
    content: 用[golang-pro]实现20个Handler(Controller)：Article/UserAuth/Comment/Category/Tag/FriendLink/Talk/Photo等全部REST端点
    status: completed
    dependencies:
      - infra-layer
  - id: core-services
    content: 用[golang-pro]实现26个Service业务层：含事务处理/并发查询(goroutine errgroup)/Redis缓存策略/复杂业务逻辑
    status: completed
    dependencies:
      - core-handlers
  - id: auth-rbac
    content: 实现安全体系：JWT Token签发验证中间件 + RBAC动态权限中间件 + 限流中间件 + QQ OAuth登录
    status: completed
    dependencies:
      - core-services
  - id: mq-consumer
    content: 实现3个RabbitMQ消费者：Maxwell ES同步 + 评论邮件通知 + 文章订阅通知
    status: completed
    dependencies:
      - auth-rbac
  - id: scheduler
    content: 实现robiqu/cron定时任务：UniqueView统计 + 缓存清理 + 地域分布 + 百度SEO + ES同步
    status: completed
    dependencies:
      - mq-consumer
  - id: search-upload-strategy
    content: 实现策略模式：ES/MySQL双搜索策略 + MinIO/OSS双上传策略 + 策略上下文自动选择
    status: completed
    dependencies:
      - scheduler
  - id: agent-engine
    content: 实现Agent引擎核心：LLM多模型Router + Tool工具集Hub + Memory会话记忆 + RAG管道
    status: completed
    dependencies:
      - search-upload-strategy
  - id: agent-features
    content: 实现AI功能：智能写作助手 + 智能评论助手 + 语义搜索 + 内容审核 + 工作流编排
    status: completed
    dependencies:
      - agent-engine
  - id: agent-ui
    content: 实现Agent前端UI：Studio控制台 + 写作助手面板 + AI数据分析图表 + 搜索增强界面
    status: completed
    dependencies:
      - agent-features
  - id: test-optimize
    content: 用[golang-testing]编写核心模块测试 + 性能基准测试 + API Swagger文档 + Docker部署配置
    status: completed
    dependencies:
      - agent-ui
---

## Product Overview

将 Aurora 博客系统从 Java SpringBoot 4.1.0-M4 完整迁移至 Go 1.26 版本，在保留全部现有功能的基础上，融入 AI Agent 智能体能力，打造下一代 AI 驱动的博客系统。这是一个包含 7 个中间件依赖、24 张数据表、20 个 Controller、26 个 Service 的复杂全栈博客系统。

## Core Features

### 一、现有功能完整迁移（100% 兼容）

- **文章系统**: 草稿/发布/置顶/推荐/密码保护/导入导出(MD)/归档/文章搜索(ES+MySQL双模式)
- **分类标签**: 多级分类管理 + 标签体系 + 文章多对多关联
- **评论系统**: 多类型嵌套评论(文章/说说/友链/关于/留言)/邮件通知(RabbitMQ异步)/审核机制
- **用户认证**: JWT + Redis Session / BCrypt密码加密 / QQ OAuth登录 / 邮箱验证码
- **权限控制**: Resource-Based RBAC 动态权限 / 角色-菜单-资源三级授权
- **文件存储**: MinIO 对象存储 / 阿里云OSS 双策略切换
- **相册管理**: 相册CRUD + 照片上传 + 私密相册
- **友链管理**: 友链申请审核
- **说说(Talk)**: 短内容发布
- **定时任务(Quartz)**: UniqueView统计/缓存清理/用户地域分布/百度SEO推送/ES数据同步/日志清理
- **数据统计**: Redis ZSet浏览量排行/HyperLogLog独立访客/IP地域分布(Geo)
- **网站配置**: 网站信息/社交链接/邮箱通知开关/打赏配置
- **操作日志**: AOP切面记录 + 异常日志
- **RabbitMQ消息队列**: 评论邮件通知/文章订阅通知/Maxwell ES数据同步(3个Consumer)
- **Elasticsearch全文搜索**: ik_max_word分词 + 高亮显示 + 索引自动初始化

### 二、AI Agent 新增能力（创新融合）

- **AI 智能写作助手**: 基于LLM的文章生成/续写/润色/摘要提取/关键词推荐
- **AI 智能评论助手**: 自动识别垃圾评论/智能回复建议/情感分析/敏感词过滤增强
- **AI 搜索增强**: 语义搜索(向量检索 RAG) 替代或补充传统关键词搜索
- **AI 内容审核**: 自动审核文章/评论内容合规性
- **Agent 工作流编排**: 文章发布流水线(AI生成->审核->SEO优化->定时发布)
- **AI 数据分析面板**: 访客行为分析/热门内容预测/运营决策建议

## Tech Stack Selection

### 后端核心（Go 1.26）

| 组件 | 技术选型 | 说明 |
| --- | --- | --- |
| **Web框架** | Gin v1.10+ | 高性能HTTP路由框架 |
| **ORM** | GORM v1.30+ | Go最成熟ORM，支持AutoMigrate/关联/JTX |
| **数据库驱动** | go-sql-driver/mysql | MySQL官方Go驱动 |
| **缓存** | go-redis/v9 | Redis客户端，支持全部Redis数据结构 |
| **消息队列** | rabbitmq/amqp091-go | RabbitMQ原生Go客户端 |
| **全文搜索** | olivere/elastic v7 或 elastic/go-elasticsearch v8 | ES Go客户端 |
| **对象存储** | minio-go/v7 | MinIO官方SDK |
| **认证** | golang-jwt/jwt v5 | JWT token签发验证 |
| **密码** | x/crypto/bcrypt | BCrypt密码哈希 |
| **配置管理** | viper + fsnotify | YAML配置热加载 |
| **参数校验** | go-playground/validator v10 | 结构体验证标签 |
| **日志** | uber-go/zap | 高性能结构化日志 |
| **Swagger** | swaggo/gin-swagger | API文档自动生成 |
| **定时任务** | robfig/cron v3 | Cron表达式解析与调度 |
| **IP库** | ip2region | IP归属地查询 |
| **邮件** | net/smtp + html/template | 标准库SMTP邮件发送 |
| **敏感词** | 自实现或 sensitive-word Go移植 | 敏感词过滤 |
| **AI/LLM SDK** | openai-go / anthropic-go | 多模型统一接入层 |
| **向量数据库** | pgvector (PostgreSQL扩展) 或 qdrant-go | RAG语义检索 |
| **Agent框架** | trpc.group/trpc-go/trpc-agent-go v1.8+ | 腾讯开源AI Agent引擎(替代自研) |


### 架构决策理由

1. **Gin而非Echo/Fiber**: Gin生态最成熟，中间件丰富，社区最大
2. **GORM而非sqlx/ent**: GORM的关联操作和AutoMigrate最适合迁移场景，可1:1映射MyBatis-Plus能力
3. **保留所有中间件**: 不改变基础设施，只替换语言层，降低迁移风险
4. **Go 1.26新特性**: 使用iterators/generics/log/slog等新特性简化代码
5. **Agent采用tRPC-Agent-Go v1.8**(腾讯开源): 替代自研引擎，获多Agent协作/MCP协议/Skill系统/可观测性四大超预期收益，代码量减少79%

## Tech Architecture

### 系统架构总览

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          客户端层 (保持不变)                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐                  │
│  │ aurora-blog   │  │ aurora-admin │  │  Agent Dashboard  │ (可选)          │
│  │  (Vue3前台)   │  │ -v3(管理台)  │  │   Studio/Writer   │                  │
│  └──────┬───────┘  └──────┬───────┘  └────────┬─────────┘                  │
└─────────┼─────────────────┼──────────────────┼─────────────────────────────┘
          │  REST API       │  REST API         │  SSE / REST(可选)
          ▼                 ▼                  ▼
╔═══════════════════════════════════════════════════════════════════════════╗
║                      Aurora Go Backend                                     ║
║  ┌─────────────────────────────────────────────────────────────────────┐  ║
║  │              Gateway Layer (Gin路由网关)                             │  ║
║  │    CORS / Recovery / Logger / RateLimit / JWT Auth / RBAC         │  ║
║  ├─────────────────────────────────────────────────────────────────────┤  ║
║  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌───────────┐ ┌─────────┐ │  ║
║  │  │ Article  │ │ User/Auth│ │ Comment  │ │ ...20个   │ │ [Agent] │ │  ║ ← 可插拔模块
║  │  │ Handler  │ │ Handler  │ │ Handler  │ │ Handler   │ │Handler  │ │  ║   (虚线=可选)
║  │  └────┬─────┘ └────┬─────┘ └────┬─────┘ └─────┬─────┘ └────┬────┘ │  ║
║  │       │            │           │             │            │       │  ║
║  │  ┌────▼────────────▼───────────▼─────────────▼────────────▼────┐  │  ║
║  │  │              Service Layer (业务逻辑层)                        │  │  ║
║  │  │  ArticleSvc/UserSvc/CommentSvc/SearchSvc/UploadSvc/...      │  │  ║
║  │  └────┬──────────┬──────────┬──────────┬──────────┬───────────┘  │  ║
║  │       │          │          │          │          │               │  ║
║  │  ┌────▼──┐ ┌────▼──┐ ┌────▼──┐ ┌────▼──┐ ┌────▼──┐ ┌────────┐ │  ║
║  │  │ GORM  │ │ Redis │ │RabbitMQ│ │ ES CLI│ │ MinIO  │ │[tRPC]  │ │  ║ ← tRPC-Go仅Agent用
║  │  │ MySQL │ │ Cache │ │  MQ   │ │Search │ │  OSS   │ │Agent   │ │  ║   其他层零依赖
║  │  └───────┘ └──────┘ └───────┘ └───────┘ └───────┘ └────────┘ │  ║
╚═══════════════════════════════════════════════════════════════════════════╝
                              ↑ 核心系统(必须)     ↑ Agent插件(可停)
```

### Agent 模块隔离架构（基于 tRPC-Agent-Go v1.8）

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    internal/agent/  (完全独立，可卸载)                   │
│                                                                         │
│  ┌───────────────────────────────────────────────────────────────────┐ │
│  │              AuroraAgentFactory.go  (唯一入口, ~80行)               │ │
│  │  功能: 组装tRPC组件 → 创建Runner → 对外暴露Run()方法               │ │
│  └────────────────────────┬──────────────────────────────────────────┘ │
│                           │                                            │
│  ┌────────────────────────┼──────────────────────────────────────┐    │
│  │            基于 tRPC-Agent-Go 组件 (import trpc-agent-go)      │    │
│  │                                                         ↓     │    │
│  │  ┌─────────────┐  ┌──────────────┐  ┌───────────────────┐  │    │
│  │  │LLM Router   │  │ Tool Hub      │  │ Memory Service    │  │    │
│  │  │model/openai │  │tool/function │  │memory/memorysvc   │  │    │
│  │  │+deepseek    │  │+mcp          │  │+Redis持久化适配器  │  │    │
│  │  └─────────────┘  └──────────────┘  └───────────────────┘  │    │
│  │                                                          │    │
│  │  ┌─────────────┐  ┌──────────────┐  ┌───────────────────┐  │    │
│  │  │Planner      │  │ Graph Workflow│  │ Knowledge(RAG)    │  │    │
│  │  │planner pkg  │  │graph.StateGph │  │knowledge包        │  │    │
│  │  │ReAct循环    │  │DAG多条件路由  │  │ES混合检索适配器   │  │    │
│  │  └─────────────┘  └──────────────┘  └───────────────────┘  │    │
│  └─────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  ┌───────────────────────────────────────────────────────────────────┐ │
│  │  agent_handler.go  (Gin→tRPC桥接, ~100行)                         │ │
│  │  POST /api/agent/chat   → runner.Run() + SSE流式输出               │ │
│  │  POST /api/agent/write  → workflow.Execute(ArticlePublishFlow)    │ │
│  │  POST /api/agent/search → ragPipeline.Retrieve() + LLM生成        │ │
│  │  GET  /api/agent/stats  → analysisTool.Run(StatsQuery)             │ │
│  └───────────────────────────────────────────────────────────────────┘ │
│                                                                         │
│  ┌───────────────────────────────────────────────────────────────────┐ │
│  │  aurora_tools.go  (业务工具注册, ~200行)                            │ │
│  │  SearchArticlesTool → 调用 SearchService (ES/MySQL)                │ │
│  │  WriteArticleTool  → 调用 ArticleService (CRUD事务)                 │ │
│  │  AnalyzeStatsTool  → 调用 RedisService + StatisticsRepo            │ │
│  │  ReviewContentTool → 调用敏感词过滤 + LLM审核                       │ │
│  │  ManageTagsTool    → 调用 TagService                                │ │
│  └───────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────┘

═══════════════════════════════════════════════════════════════════════
  隔离保证（5级防护）:
  ① 编译隔离: //go:build aurora_agent tag 控制是否编译
  ② 配置隔离: agent.enabled=false → 零初始化,零路由注册
  ③ 路由隔离: 独立RouterGroup /api/agent/*, 不影响其他路由
  ④ 故障隔离: goroutine+recover包装, Agent panic不杀主进程
  ⑤ 依赖隔离: 核心代码零import agent包, 仅接口回调
═══════════════════════════════════════════════════════════════════════
```

### 目录结构

```
aurora-go/
├── cmd/
│   └── server/
│       └── main.go                    # [NEW] 入口文件，服务启动/优雅关闭/诊断
├── internal/                           # [NEW] 私有应用代码
│   ├── config/                        # [NEW] 配置管理
│   │   ├── config.go                  # Viper配置加载(YAML/ENV)
│   │   ├── mysql.go                   # MySQL连接池配置
│   │   ├── redis.go                   # Redis连接配置
│   │   ├── rabbitmq.go                # RabbitMQ连接配置
│   │   ├── elasticsearch.go           # ES客户端配置
│   │   ├── minio.go                   # MinIO客户端配置
│   │   ├── jwt.go                     # JWT密钥配置
│   │   └── log.go                     # Zap日志配置
│   │
│   ├── model/                         # [NEW] 数据模型(GORM实体)
│   │   ├── article.go                 # Article/ArticleTag实体
│   │   ├── user.go                    # UserInfo/UserAuth/UserRole
│   │   ├── category.go                # Category实体
│   │   ├── tag.go                     # Tag实体
│   │   ├── comment.go                 # Comment实体
│   │   ├── friend_link.go             # FriendLink实体
│   │   ├── talk.go                    # Talk实体
│   │   ├── photo.go                   # Photo/PhotoAlbum实体
│   │   ├── resource.go                # Resource/RoleResource实体
│   │   ├── role.go                    # Role/RoleMenu实体
│   │   ├── menu.go                    # Menu实体
│   │   ├── job.go                     # Job/JobLog实体
│   │   ├── log.go                    # OperationLog/ExceptionLog实体
│   │   ├── about.go                  # About实体
│   │   ├── website_config.go          # WebsiteConfig实体
│   │   └── unique_view.go             # UniqueView实体
│   │
│   ├── dto/                           # [NEW] 数据传输对象
│   │   ├── article_dto.go             # ArticleDTO/ArticleCardDTO/ArticleSearchDTO等
│   │   ├── user_dto.go                # UserDTO/UserDetailsDTO/UserAdminDTO等
│   │   ├── comment_dto.go             # CommentDTO/ReplyDTO/CommentAdminDTO等
│   │   ├── common_dto.go             # PageResultDTO/ConditionVO/ResultVO等
│   │   ├── site_dto.go               # AuroraHomeInfoDTO/AuroraAdminInfoDTO等
│   │   └── ...                       # 其他DTO按模块分组
│   │
│   ├── vo/                            # [NEW] 视图对象(请求参数)
│   │   ├── article_vo.go              # ArticleVO/ArticleTopFeaturedVO等
│   │   ├── user_vo.go                 # UserVO/PasswordVO/QQLoginVO等
│   │   ├── comment_vo.go              # CommentVO/ReviewVO等
│   │   └── common_vo.go              # ConditionVO/DeleteVO/EmailVO等
│   │
│   ├── handler/                       # [NEW] HTTP处理器(Controller)
│   │   ├── article_handler.go         # 15个端点: CRUD/搜索/归档/导出
│   │   ├── user_auth_handler.go       # 8个端点: 注册/登录/登出/OAuth
│   │   ├── user_info_handler.go       # 用户信息管理
│   │   ├── comment_handler.go         # 8个端点: 评论CRUD/回复/审核
│   │   ├── category_handler.go        # 分类CRUD
│   │   ├── tag_handler.go             # 标签CRUD
│   │   ├── friend_link_handler.go     # 友链管理
│   │   ├── talk_handler.go           # 说说管理
│   │   ├── photo_handler.go          # 相册/照片管理
│   │   ├── photo_album_handler.go     # 相册管理
│   │   ├── resource_handler.go        # 资源权限管理
│   │   ├── role_handler.go           # 角色管理
│   │   ├── menu_handler.go           # 菜单管理
│   │   ├── job_handler.go            # 定时任务管理
│   │   ├── job_log_handler.go        # 调度日志
│   │   ├── operation_log_handler.go   # 操作日志
│   │   ├── exception_log_handler.go   # 异常日志
│   │   ├── aurora_info_handler.go     # 首页/后台信息聚合
│   │   ├── website_config_handler.go  # 网站配置
│   │   ├── file_handler.go           # 文件上传
│   │   └── agent_handler.go          # [NEW] AI Agent端点
│   │
│   ├── service/                       # [NEW] 业务逻辑层
│   │   ├── article_service.go         # 文章核心逻辑(含事务)
│   │   ├── user_auth_service.go       # 认证注册逻辑
│   │   ├── user_info_service.go       # 用户信息管理
│   │   ├── comment_service.go         # 评论逻辑(含嵌套/通知)
│   │   ├── category_service.go        # 分类管理
│   │   ├── tag_service.go            # 标签管理
│   │   ├── friend_link_service.go     # 友链管理
│   │   ├── talk_service.go           # 说说管理
│   │   ├── photo_service.go          # 相册照片管理
│   │   ├── resource_service.go       # 资源权限管理
│   │   ├── role_service.go           # 角色权限管理
│   │   ├── menu_service.go           # 菜单管理
│   │   ├── job_service.go            # 定时任务调度
│   │   ├── aurora_info_service.go     # 信息聚合(首页/后台统计数据)
│   │   ├── website_config_service.go  # 网站配置
│   │   ├── search_service.go         # 搜索策略(ES/MySQL)
│   │   ├── upload_service.go         # 上传策略(MinIO/OSS)
│   │   ├── email_service.go          # 邮件发送服务
│   │   ├── redis_service.go          # Redis封装(全部数据结构操作)
│   │   ├── es_service.go             # Elasticsearch封装
│   │   ├── token_service.go          # JWT Token管理
│   │   │
│   │   └── [agent/]                   # [可选插件] 基于tRPC-Agent-Go v1.8
│   │       ├── aurora_agent.go        # Agent工厂: 组装LLM/Tool/Memory/Runner (~80行)
│   │       ├── aurora_tools.go        # 6个Aurora业务Tool注册 (~200行)
│   │       ├── aurora_handler.go      # Gin→tRPC桥接 + SSE流式输出 (~100行)
│   │       ├── aurora_workflow.go     # 预置工作流(发布/审核/分析DAG) (~150行)
│   │       └── memory_adapter.go      # Redis持久化Memory适配器 (~60行)
│   │
│   ├── repository/                   # [NEW] 数据访问层
│   │   ├── mysql_repo.go             # GORM通用CRUD/分页/条件构建
│   │   ├── article_repo.go          # Article自定义查询
│   │   ├── comment_repo.go           # Comment嵌套查询
│   │   ├── user_repo.go             # 用户相关查询
│   │   ├── statistics_repo.go        # 统计聚合查询
│   │   └── es_repo.go               # ES查询构建
│   │
│   ├── middleware/                    # [NEW] Gin中间件
│   │   ├── jwt_auth.go              # JWT认证中间件(对应Java Filter)
│   │   ├── rbac.go                  # RBAC权限中间件(动态权限)
│   │   ├── ratelimit.go             # 接口限流中间件(Redis滑动窗口)
│   │   ├── logger.go                # 请求日志中间件
│   │   ├── recovery.go              # 全局异常恢复中间件
│   │   ├── cors.go                  # CORS跨域中间件
│   │   └── access_log.go            # 操作日志AOP中间件
│   │
│   ├── consumer/                     # [NEW] MQ消费者
│   │   ├── maxwell_consumer.go       # ES数据同步消费者
│   │   ├── email_consumer.go         # 评论邮件通知消费者
│   │   └── subscribe_consumer.go     # 文章订阅通知消费者
│   │
│   ├── scheduler/                    # [NEW] 定时任务
│   │   ├── scheduler.go             # 任务调度器入口
│   │   ├── unique_view_job.go       # 每日独立访客统计
│   │   ├── clear_cache_job.go       # 缓存清理
│   │   ├── user_area_job.go         # 用户地域分布统计
│   │   ├── baidu_seo_job.go         # 百度SEO推送
│   │   ├── es_sync_job.go           # ES全量同步
│   │   └── clean_log_job.go         # 日志清理
│   │
│   ├── strategy/                     # [NEW] 策略模式
│   │   ├── search_strategy.go        # 搜索策略接口
│   │   ├── es_search_strategy.go     # ES搜索实现
│   │   ├── mysql_search_strategy.go  # MySQL搜索实现
│   │   ├── upload_strategy.go       # 上传策略接口
│   │   ├── minio_upload_strategy.go # MinIO上传实现
│   │   └── oss_upload_strategy.go   # OSS上传实现
│   │
│   ├── util/                         # [NEW] 工具函数
│   │   ├── page_util.go             # 分页参数处理
│   │   ├── response_util.go         # 统一响应格式(ResultVO)
│   │   ├── ip_util.go               # IP解析(ip2region)
│   │   ├── html_util.go             # HTML过滤/XSS防护
│   │   ├── file_util.go             # 文件类型判断
│   │   ├── bean_copy_util.go        # Map/Struct拷贝
│   │   └── crypto_util.go           # 加密/Hash工具
│   │
│   ├── constant/                     # [NEW] 常量定义
│   │   ├── redis_key.go             # Redis Key常量
│   │   ├── rabbitmq_const.go        # MQ队列/交换机常量
│   │   ├── auth_const.go            # 认证相关常量
│   │   ├── opt_type_const.go        # 操作类型常量
│   │   └── enum/                    # 枚举定义
│   │       ├── article_status_enum.go
│   │       ├── comment_type_enum.go
│   │       ├── search_mode_enum.go
│   │       ├── upload_mode_enum.go
│   │       ├── login_type_enum.go
│   │       └── status_code_enum.go
│   │
│   └── errors/                       # [NEW] 错误处理
│       ├── app_error.go              # 自定义错误类型
│       └── error_codes.go            # 错误码枚举
│
├── pkg/                              # [NEW] 可导出公共库(无Agent自研代码)
│   └── (tRPC-Agent-Go通过go.mod直接引入, 无需本地pkg封装)
│
├── [internal/agent/]                 # [可选插件目录, agent.enabled=true时加载]
│   └── (仅5个文件, ~590行, 见上方service/agent/说明)
│
├── configs/                          # [NEW] 配置文件
│   ├── config.yaml                   # 主配置文件(合并原application-prod.yml)
│   └── config.example.yaml          # 配置模板
│
├── scripts/                          # [NEW] 脚本
│   ├── build.sh                      # 交叉编译脚本
│   ├── docker-compose.go.yml         # Docker编排(替代java版)
│   └── migrate.sh                    # 数据迁移辅助脚本
│
├── api/                              # [NEW] API定义
│   └── openapi.yaml                  # OpenAPI 3.0规范文档
│
├── docs/                             # [NEW] 文档
│   ├── MIGRATION_GUIDE.md            # Java→Go 迁移指南
│   ├── AGENT_GUIDE.md                # Agent开发指南
│   └── API.md                        # API文档(由swag生成)
│
├── test/                             # [NEW] 测试
│   ├── integration/                  # 集成测试
│   │   ├── article_test.go
│   │   ├── auth_test.go
│   │   └── agent_test.go
│   └── benchmark/                    # 性能基准测试
│       └── api_benchmark_test.go
│
├── go.mod                            # [NEW] Go模块定义
├── go.sum
├── Makefile                          # [NEW] 构建任务
├── Dockerfile                        # [NEW] 多阶段构建(极简镜像)
└── README.md                         # [NEW] 项目说明
```

## Implementation Notes

### 关键技术映射表

| Java/SpringBoot | Go/Aurora-Go | 备注 |
| --- | --- | --- |
| Spring MVC/Gin | gin.Context + HandlerFunc | 路由注册方式不同，功能完全对等 |
| MyBatis-Plus | GORM | AutoMigrate替代建表，Session替代事务注解 |
| Spring Security | 自定义JWT中间件+RBAC中间件 | FilterChain→Middleware Chain |
| @Autowired注入 | 依赖注入(DI容器或手动) | Go常用构造函数注入 |
| @OptLog AOP | AccessLog中间件 | Gin中间件可拦截请求/响应 |
| CompletableFuture | goroutine + channel/errgroup | 并发查询改为goroutine并发 |
| RabbitTemplate | amqp091-go | 发布/消费模式一致 |
| RedisTemplate | go-redis/v9 | 所有数据结构操作1:1映射 |
| ElasticsearchClient | olivere/elastic | Builder模式类似 |
| Quartz | robfig/cron | 表达式兼容 |
| @Valid | go-playground/validator | 结构体验证标签 |
| Knife4j/Swagger | swaggo/gin-swagger | 注释声明式API文档 |
| Lombok @Data | Go struct tag | JSON/db/gorm/validation tag |


### Agent层技术映射（自研 → tRPC-Agent-Go v1.8）

| 原自研方案(已废弃) | tRPC-Agent-Go 替代 | 说明 |
| --- | --- | --- |
| `pkg/agent/llm/` (~400行) | `model/openai` + `model/deepseek` | 配置即插, 支持VariantDeepSeek |
| `pkg/agent/tool/` (~500行) | `tool/function.NewFunctionTool()` | 任意Go函数→Tool, 一行封装 |
| `pkg/agent/memory/` (~300行) | `memory/memorysvc` + Redis Adapter | 内置InMemory+自定义Redis持久化 |
| `ai/agent/engine.go` ReAct循环 | `planner` + `cycleagent.CycleAgent` | 官方ReAct实现, 更健壮 |
| `ai/workflow/pipeline.go` DAG | `graph.StateGraph` | 类型安全多条件路由, 等价LangGraph |
| `ai/rag/pipeline.go` RAG管道 | `knowledge` 包 | 文档检索增强生成内置支持 |
| 自建 SSE streaming (~200行) | `event.StreamingEvents` | Server-Sent Events 原生输出 |
| (无) 多Agent协作 | `chainagent` / `parallelagent` | **新增能力**: 链式/并行编排 |
| (无) MCP协议支持 | `tool/mcptool` | **新增能力**: Model Context Protocol |
| (无) Skill系统 | `skill` 包(SKILL.md) | **新增能力**: 可复用技能加载 |
| (无) OpenTelemetry | `telemetry` 包 | **新增能力**: 追踪+指标 |
| (无) A2A互操作 | `server/a2a` | **新增能力**: 跨Agent系统通信 |


**Agent模块代码量对比：自研 ~3200行 → tRPC适配 ~590行，减少 82%**

### Agent 隔离设计（5级防护）

```
级别   机制                     效果
───────────────────────────────────────────────────
L1    //go:build aurora_agent   编译时完全排除Agent代码
L2    agent.enabled=false       运行时零初始化, 零内存占用
L3    独立RouterGroup           /api/agent/* 不注册, 其他路由不受影响
L4    goroutine+recover         Agent panic不杀主进程, 自动降级返回503
L5    接口回调解耦              核心Service不import任何agent包
```

| CompletableFuture | goroutine + channel/errgroup | 并发查询改为goroutine并发 |
| --- | --- | --- |
| RabbitTemplate | amqp091-go | 发布/消费模式一致 |
| RedisTemplate | go-redis/v9 | 所有数据结构操作1:1映射 |
| ElasticsearchClient | olivere/elastic | Builder模式类似 |
| Quartz | robfig/cron | 表达式兼容 |
| @Valid | go-playground/validator | 结构体tag验证 |
| Knife4j/Swagger | swaggo/gin-swagger | 注释声明式API文档 |
| Lombok @Data | Go struct tag | JSON/db/gorm/validation tag |


### 性能预期提升

- **内存占用**: 从 200-400MB(JVM) 降至 30-60MB(Go二进制)，降低约80%
- **启动时间**: 从 8-15秒(SpringBoot) 降至 0.3-0.8秒(Go)，提升约95%
- **响应延迟**: P99从50-100ms降至10-30ms(无JVM预热/GC停顿)
- **部署体积**: JAR 80MB+ → Go binary 15-25MB(静态编译)

### 迁移风险控制

1. **API契约不变**: 所有REST端点路径/方法/请求响应格式100%兼容，前端零改动
2. **数据库Schema不变**: 复用现有24张表，GORM AutoMigrate仅做增量
3. **中间件不变**: MySQL/Redis/RabbitMQ/ES/MinIO 配置直接迁移
4. **渐进式迁移**: 可并行运行Java和Go版本，通过Nginx灰度切流
5. **Agent完全隔离(5级防护)**:

- L1 编译级: `//go:build aurora_agent` tag 可完全排除Agent代码
- L2 配置级: `agent.enabled: false` → 零初始化零路由
- L3 路由级: Agent Handler 独立 RouterGroup, 不影响其他端点
- L4 故障级: goroutine+recover包装, Agent崩溃不杀主进程
- L5 依赖级: 核心Service层零import agent包, 纯接口回调
- **结论**: 停掉Agent = 关一个配置项, 核心博客系统100%不受影响

## 设计概述

Aurora Go 后端的 Agent 能力需要配套一个现代化的 AI 交互界面。本设计聚焦于 **Aurora Agent Studio** -- 一个嵌入到 aurora-admin-v3 管理后台中的 AI 工作台页面，以及博客前台的 AI 辅助交互组件。

设计风格采用 **Cyberpunk Neon UI 结合 Glassmorphism** 的科技感风格，体现 "AI + 博客" 的未来感。深色基调配合霓虹色调的强调色，毛玻璃质感的卡片布局，营造沉浸式的 AI 协作氛围。

### 页面规划（6屏）

#### Page 1: Agent Studio 控制台（管理后台 - 新增页面）

AI 工作台主面板，集成对话、写作、分析三大核心能力。左侧为对话历史列表，中央为主交互区（支持 Markdown 实时渲染），右侧为 Agent 工具面板和参数调节。

#### Page 2: AI 写作助手弹窗（管理后台 - 文章编辑器内嵌）

在文章编辑器(Markdown编辑器)侧边或底部嵌入 AI 面板，提供：续写、润色、摘要提取、关键词推荐、SEO优化建议。采用可拖拽分割面板设计。

#### Page 3: AI 评论管理面板（管理后台 - 评论页扩展）

在现有评论列表上方增加 AI 分析栏：情感分布饼图、垃圾评论识别标记、批量智能回复建议。使用 ECharts 可视化。

#### Page 4: AI 数据洞察仪表盘（管理后台 - Home页扩展）

在后台首页增加 AI 洞察卡片：访问趋势预测、热门内容推荐、用户行为聚类分析、运营建议自然语言生成。

#### Page 5: 博客前台 AI 搜索（aurora-blog - SearchMode.vue 增强）

在现有全局搜索框增加 AI 语义搜索 Tab：输入自然语言问题，返回精准答案(非传统关键词匹配)。结果卡片展示引用来源和高亮片段。

#### Page 6: Agent 配置中心（管理后台 - Setting页扩展）

LLM 提供商配置(API Key/Model/Temperature)、Agent 行为规则设定、工具调用权限管理、使用量监控和Token消耗统计。

### Skill 扩展

- **golang-pro**
- Purpose: 提供 Go 1.26 最佳实践指导，包括并发模式(goroutines/channels)、gRPC/REST 微服务设计、性能优化(pprof)、泛型和接口设计
- Expected outcome: 确保 Go 代码符合 idiomatic Go 惯例，正确使用 Go 1.26 新特性，高性能并发处理 HTTP 请求和数据库操作

- **golang-patterns**
- Purpose: 提供 Go 设计模式和架构模式参考，包括项目布局(standard Go project layout)、错误处理最佳实践、接口抽象设计
- Expected outcome: 保证项目代码组织清晰，分层合理，遵循 Go 社区公认的最佳实践

- **golang-testing**
- Purpose: 提供测试编写指导，包括 table-driven tests、基准测试(benchmarks)、模糊测试(fuzzing)、集成测试策略
- Expected outcome: 为核心 Service 层和 Handler 层编写高覆盖率的单元测试和集成测试

### SubAgent 扩展

- **code-explorer**
- Purpose: 在实施阶段深度探索 Java 源码中复杂的业务逻辑细节，确保迁移不遗漏任何边界情况和特殊处理逻辑
- Expected outcome: 准确理解每个 Service 方法的完整实现细节，确保 Go 版本功能100%对等