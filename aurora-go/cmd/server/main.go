package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/aurora-go/aurora/internal/agent"
	"github.com/aurora-go/aurora/internal/banner"
	"github.com/aurora-go/aurora/internal/config"
	"github.com/aurora-go/aurora/internal/consumer"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/handler"
	"github.com/aurora-go/aurora/internal/infrastructure"
	"github.com/aurora-go/aurora/internal/infrastructure/mq"
	"github.com/aurora-go/aurora/internal/middleware"
	"github.com/aurora-go/aurora/internal/model"
	"github.com/aurora-go/aurora/internal/service"
	"github.com/aurora-go/aurora/internal/strategy"
	"github.com/aurora-go/aurora/internal/util"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// @title Aurora Blog API
// @version 1.0
// @description Aurora 博客系统 Go 1.26 后端 - 基于 tRPC-Agent-Go 的 AI 驱动博客平台
// @termsOfService https://github.com/nicepkg/aurora

// @contact.name Aurora Team
// @contact.email aurora@example.com

// @license.name Apache 2.0
// @license.url https://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @basePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Token 认证，格式: Bearer <token>

func main() {
	startTime := time.Now()

	// 0. 初始化时区为 Asia/Shanghai (CST, UTC+8)
	// Docker alpine 镜像默认不含 tzdata，TZ 环境变量无效，time.Local 默认 UTC
	// 必须在所有 time.Now() 调用前设置，否则日期/定时任务/MySQL loc=Local 全部偏移 8 小时
	if loc, err := time.LoadLocation("Asia/Shanghai"); err == nil {
		time.Local = loc
	} else {
		// 降级: tzdata 不可用时使用固定偏移量 (UTC+8)
		time.Local = time.FixedZone("CST", 8*3600)
		slog.Warn("tzdata 不可用，使用固定 UTC+8 偏移量", "error", err.Error())
	}

	// 解析命令行参数
	configPath := flag.String("config", "configs/config.yaml", "配置文件路径")
	flag.Parse()

	// 加载 .env 文件（失败不报错，允许无 .env 运行）
	_ = godotenv.Load()

	// 1. 加载配置（Viper AutomaticEnv 会覆盖 .env 中的 AURORA_ 前缀变量）
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 2. 初始化所有基础设施（按顺序: Logger → DB → Redis → MQ → ES → MinIO → Email）
	infrastructure.Bootstrap(cfg)

	// 2.1 初始化 Elasticsearch（避免循环依赖，在此处初始化）
	if len(cfg.ES.URLs) > 0 {
		slog.Info("🚀 开始初始化 Elasticsearch...")
		esService, err := service.NewESService(cfg.ES.URLs, cfg.ES.Username, cfg.ES.Password, cfg.ES.IndexName)
		if err != nil {
			slog.Warn("Elasticsearch 连接失败，将使用 MySQL 搜索", "error", err)
		} else {
			// 设置全局 ES 服务实例
			service.SetGlobalESService(esService)
			slog.Info("✅ Elasticsearch 连接成功")

			// 初始化索引并同步数据
			db := infrastructure.GetDB()
			if db != nil {
				initializer := service.NewESIndexInitializer(esService, db)
				ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
				defer cancel()
				if err := initializer.Initialize(ctx); err != nil {
					slog.Error("ES 索引初始化失败", "error", err)
				}
			}

			// 重新初始化搜索策略（使用新的 ES 客户端）
			strategy.SetGlobalESClient(esService)
			slog.Info("✅ ES 搜索策略已更新（等待 Registry 创建后重新注入）")
		}
	} else {
		slog.Warn("Elasticsearch 未配置，将使用 MySQL 搜索")
	}

	// 2.1 初始化 IP2Region (IP归属地查询)
	if cfg.IP2Region.Enabled {
		dbFile := cfg.IP2Region.DBFile
		// 支持相对路径和绝对路径
		if !filepath.IsAbs(dbFile) {
			dbFile = filepath.Join("scripts", "ip", dbFile)
		}
		if err := util.InitIP2Region(dbFile); err != nil {
			slog.Warn("ip2region初始化失败, 将使用默认IP归属地", "db_file", dbFile, "error", err)
		} else {
			slog.Info("ip2region IP归属地查询已启用", "db_file", dbFile)
		}
	} else {
		slog.Info("ip2region IP归属地查询已禁用")
	}

	// 3. 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode) // 生产模式（Gin默认已处理debug/release）

	// 4. 创建Service注册中心（所有Handler通过Registry获取Service实例）
	db := infrastructure.GetDB()
	rdb := infrastructure.GetRedis()
	registry := service.NewRegistry(db, rdb, *cfg, slog.Default())
	service.SetGlobalRegistry(registry)
	slog.Info("Service registry initialized", "services", 24)

	// 2.1 初始化 RabbitMQ 消费者（对标Java @RabbitListener 自动启动）
	// 注意: 必须在 Bootstrap 后、HTTP Server 前启动
	if mqChannel := mq.GetChannel(); mqChannel != nil {
		siteURL := cfg.Server.GetSiteURL()
		if siteURL == "" {
			siteURL = "https://www.aurora.blog"
		}
		consumerMgr := consumer.NewConsumerManager(mqChannel, db, nil, *cfg, slog.Default())
		if err := consumerMgr.StartAll(); err != nil {
			slog.Warn("⚠️ RabbitMQ消费者启动失败，异步消息功能将不可用", "error", err)
		}
	} else {
		slog.Info("RabbitMQ 未连接，跳过消费者启动")
	}
	if cfg.Server.EnableScheduler {
		slog.Info("🚀 开始初始化定时任务调度器...")
		scheduler := infrastructure.GetScheduler()
		if scheduler != nil {
			scheduler.Start()
			slog.Info("✅ 定时任务调度器已启动")
		} else {
			slog.Warn("⚠️ 调度器未初始化，跳过定时任务")
		}
	} else {
		slog.Info("定时任务调度器已禁用")
	}

	// 4.1 初始化默认网站配置（如果数据库中不存在）
	if err := initDefaultWebsiteConfig(db); err != nil {
		slog.Warn("初始化默认网站配置失败", "error", err)
	}

	// 5. 创建 Gin 引擎并注册全局中间件
	r := gin.New()
	
	// ⭐ 配置信任代理：允许从代理头（X-Forwarded-For/X-Real-IP）获取真实客户端IP
	// nil 表示信任所有来源（适用于Nginx反向代理场景）
	// 如果有固定Nginx IP，可以指定: []string{"172.18.0.1"}
	r.SetTrustedProxies(nil)
	
	slogLogger := slog.Default()
	r.Use(middleware.Recovery(registry, slogLogger))
	r.Use(middleware.Logger(slogLogger))
	r.Use(middleware.CORS())
	r.Use(middleware.NoCache()) // 禁用缓存，确保前端路由切换时获取最新数据
	// r.Use(middleware.RateLimiter(rdb, slog.Default())) // P0-6 限流需Redis客户端

	// 5.1 静态文件服务 (上传的图片等资源)
	r.Static("/uploads", "./uploads")
	slog.Info("Static file server enabled: /uploads -> ./uploads")

	// 6. 健康检查端点（无需认证）- 对标 Spring Actuator /health
	r.GET("/health", func(c *gin.Context) {
		status := infrastructure.HealthCheck()
		allUp := true
		for _, s := range status {
			if s == "DOWN" {
				allUp = false
				break
			}
		}

		httpStatus := http.StatusOK
		if !allUp {
			httpStatus = http.StatusServiceUnavailable
		}

		c.JSON(httpStatus, gin.H{
			"status":     map[string]interface{}{"components": status},
			"timestamp":  time.Now().Unix(),
			"version":    "1.0.0-go",
			"agentReady": cfg.Agent.Enabled,
		})
	})

	// 6.1 Swagger 文档 UI（开发/调试用，可通过 SWAGGER_ENABLED 环境变量控制）
	if os.Getenv("SWAGGER_ENABLED") == "true" || cfg.Server.Mode == "debug" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		slog.Info("Swagger UI 已启用: http://localhost:8080/swagger/index.html")
	}

	// 7. 注册所有路由（公开/受保护/后台管理 - 20个Handler, 80+端点）
	tokenSvc := registry.TokenSvc
	router := handler.NewRouter(registry, tokenSvc, slog.Default())
	router.RegisterRoutes(r)
	slog.Info("All routes registered (80+ endpoints)")

	// 8. Agent 路由 (可选插件 - 5级隔离保证 L2/L3)
	if cfg.Agent.Enabled {
		// 初始化Agent引擎
		if err := agent.InitAgent(&cfg.Agent); err != nil {
			slog.Error("Agent init failed (non-fatal)", "error", err)
		} else {
			registerAgentRoutes(r) // 传入根路由, 内部创建 /api/agent/* 路由组
			slog.Info("Agent 模块已启用")
		}
	} else {
		slog.Info("Agent 模块已禁用（零初始化、零路由、零内存）")
	}

	// 9. 创建 HTTP Server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 10. 打印启动面板
	swaggerEnabled := os.Getenv("SWAGGER_ENABLED") == "true" || cfg.Server.Mode == "debug"
	probes := []banner.ServiceProbe{
		banner.MySQLProbe(db),
		banner.RedisProbe(rdb),
	}
	banner.PrintStartupBanner(cfg.Server.Mode, cfg.Server.Port, startTime, probes, swaggerEnabled)

	// 11. 启动 HTTP 服务
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP服务启动失败", "error", err.Error())
			os.Exit(1)
		}
	}()

	// 11.1 如果启用了 TLS，同时启动 HTTPS 服务
	if cfg.Server.TLS.Enabled {
		slog.Info("TLS/HTTPS 配置已启用",
			"port", cfg.Server.TLS.Port,
			"cert_file", cfg.Server.TLS.CertFile,
			"key_file", cfg.Server.TLS.KeyFile,
		)

		srvTLS := &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Server.TLS.Port),
			Handler:      r,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}

		go func() {
			if err := srvTLS.ListenAndServeTLS(cfg.Server.TLS.CertFile, cfg.Server.TLS.KeyFile); err != nil && err != http.ErrServerClosed {
				slog.Error("HTTPS服务启动失败", "error", err.Error())
				os.Exit(1)
			}
		}()
	}

	// 12. 等待关闭信号 → 优雅关闭所有基础设施
	infrastructure.WaitForSignal()
}

// registerAgentRoutes 注册 Agent 相关路由 (P0-10 阶段实现)
// 隔离保证: 仅在 agent.enabled=true 时注册 /api/agent/* 路由组
// L3 路由隔离: 独立RouterGroup, 不影响其他路由
func registerAgentRoutes(r *gin.Engine) {
	agentHandler := handler.NewAgentHandler()

	// 创建独立路由组, 带JWT认证
	agentGroup := r.Group("/api/agent")
	agentGroup.Use(middleware.JWTAuth())

	// 核心端点 (4个)
	agentGroup.GET("/chat", agentHandler.Chat)       // SSE流式对话 (支持GET/POST)
	agentGroup.POST("/chat", agentHandler.Chat)      // 同步对话
	agentGroup.POST("/write", agentHandler.Write)    // AI写作助手
	agentGroup.POST("/search", agentHandler.Search)  // AI语义搜索
	agentGroup.POST("/analyze", agentHandler.Analyze) // 数据分析+洞察
	agentGroup.GET("/sessions", agentHandler.Sessions)// 会话列表

	slog.Info("Agent routes registered: /api/agent/{chat,write,search,analyze,sessions}")
}

// initDefaultWebsiteConfig 初始化默认网站配置（如果数据库中不存在）
func initDefaultWebsiteConfig(db *gorm.DB) error {
	var config model.WebsiteConfig
	err := db.First(&config, 1).Error
	
	// 如果记录存在，直接返回
	if err == nil {
		slog.Info("网站配置已存在", "id", config.ID)
		return nil
	}
	
	// 如果不是 RecordNotFound 错误，返回错误
	if !errors.IsStd(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("查询网站配置失败: %w", err)
	}
	
	// 创建默认配置
	defaultConfig := `{
		"name": "Aurora Blog",
		"englishName": "Aurora",
		"author": "Aurora",
		"authorAvatar": "https://static.aqi125.cn/aurora/config/avatar.jpg",
		"authorIntro": "欢迎来到我的博客",
		"logo": "https://static.aqi125.cn/aurora/config/logo.png",
		"notice": "欢迎来到 Aurora 博客系统",
		"websiteCreateTime": "2024-01-01 00:00:00",
		"touristAvatar": "https://static.aqi125.cn/aurora/config/tourist.png",
		"userAvatar": "https://static.aqi125.cn/aurora/config/user.png"
	}`
	
	newConfig := model.WebsiteConfig{
		ID:     1,
		Config: defaultConfig,
	}
	
	if err := db.Create(&newConfig).Error; err != nil {
		return fmt.Errorf("创建默认网站配置失败: %w", err)
	}
	
	slog.Info("默认网站配置已创建")
	return nil
}
