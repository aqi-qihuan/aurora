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

	"github.com/aurora-go/aurora/internal/agent"
	"github.com/aurora-go/aurora/internal/config"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/handler"
	"github.com/aurora-go/aurora/internal/infrastructure"
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
	// 解析命令行参数
	configPath := flag.String("config", "configs/config.yaml", "配置文件路径")
	flag.Parse()

	// 1. 加载配置
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

	// 4.1 初始化默认网站配置（如果数据库中不存在）
	if err := initDefaultWebsiteConfig(db); err != nil {
		slog.Warn("初始化默认网站配置失败", "error", err)
	}

	// 5. 创建 Gin 引擎并注册全局中间件
	r := gin.New()
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

	// 10. 启动 HTTP 服务
	go func() {
		slog.Info("Aurora Go 服务启动 (HTTP)",
			"addr", srv.Addr,
			"mode", cfg.Server.Mode,
			"agent_enabled", cfg.Agent.Enabled,
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP服务启动失败", "error", err.Error())
			os.Exit(1)
		}
	}()

	// 10.1 如果启用了 TLS，同时启动 HTTPS 服务
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
			slog.Info("Aurora Go 服务启动 (HTTPS)",
				"addr", srvTLS.Addr,
				"mode", cfg.Server.Mode,
			)
			if err := srvTLS.ListenAndServeTLS(cfg.Server.TLS.CertFile, cfg.Server.TLS.KeyFile); err != nil && err != http.ErrServerClosed {
				slog.Error("HTTPS服务启动失败", "error", err.Error())
				os.Exit(1)
			}
		}()
	}

	// 11. 等待关闭信号 → 优雅关闭所有基础设施
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
