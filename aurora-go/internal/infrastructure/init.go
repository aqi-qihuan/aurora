package infrastructure

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"gorm.io/gorm"

	"github.com/aurora-go/aurora/internal/config"
	"github.com/aurora-go/aurora/internal/infrastructure/database"
	"github.com/aurora-go/aurora/internal/infrastructure/email"
	"github.com/aurora-go/aurora/internal/infrastructure/logger"
	"github.com/aurora-go/aurora/internal/infrastructure/mq"
	"github.com/aurora-go/aurora/internal/infrastructure/search"
	"github.com/aurora-go/aurora/internal/infrastructure/storage"
	"github.com/aurora-go/aurora/internal/scheduler"
	"github.com/aurora-go/aurora/internal/strategy"
)

// Bootstrap 按顺序初始化所有基础设施组件
// 初始化顺序: Logger → DB → Redis → RabbitMQ → ES → MinIO → Email → 策略模式 → 调度器
// 对标 Java 的 @Configuration + @Bean 依赖注入
// 降级模式: 核心依赖(MySQL/Redis)不可用时仅 warn，路由层仍可运行
func Bootstrap(cfg *config.Config) {
	slog.Info("=== Aurora Go Infrastructure Bootstrap Start ===")

	// 1. 日志系统（最先初始化，其他组件依赖日志输出）
	logger.InitZapLogger(&cfg.Log)
	slog.Info("[1/10] Zap logger initialized", "level", cfg.Log.Level, "format", cfg.Log.Format)

	// 2. MySQL 数据库连接（降级: 失败不退出，仅路由层可用）
	if err := database.InitMySQL(&cfg.MySQL); err != nil {
		slog.Warn("[2/10] MySQL connection failed (degraded mode - API stubs only)", "error", err)
	} else {
		slog.Info("[2/10] MySQL connected")
	}

	// 3. Redis 缓存连接（降级: 失败不退出）
	if err := database.InitRedis(&cfg.Redis); err != nil {
		slog.Warn("[3/10] Redis connection failed (caching disabled)", "error", err)
	} else {
		slog.Info("[3/10] Redis connected")
	}

	// 4. RabbitMQ 消息队列（可选）
	if cfg.RabbitMQ.Host != "" {
		if err := mq.InitRabbitMQ(&cfg.RabbitMQ); err != nil {
			slog.Warn("[4/10] RabbitMQ connection failed (async messaging disabled)", "error", err)
		} else {
			slog.Info("[4/10] RabbitMQ connected")
		}
	} else {
		slog.Warn("[4/10] RabbitMQ not configured, skipping")
	}

	// 5. Elasticsearch 全文搜索（可选）- 延迟到 main.go 中初始化以避免循环依赖
	if len(cfg.ES.URLs) > 0 {
		slog.Info("[5/10] Elasticsearch configured, will initialize in main.go")
	} else {
		slog.Warn("[5/10] Elasticsearch not configured, falling back to MySQL search")
	}

	// 6. MinIO 对象存储（可选，OSS为替代方案）
	if cfg.MinIO.Endpoint != "" {
		if err := storage.InitMinIO(&cfg.MinIO); err != nil {
			slog.Warn("[6/10] MinIO connection failed (file upload disabled)", "error", err)
		} else {
			slog.Info("[6/10] MinIO connected")
		}
	} else if cfg.OSS != nil && cfg.OSS.Endpoint != "" {
		slog.Info("[6/10] MinIO not configured, using OSS instead")
	} else {
		slog.Warn("[6/10] No storage backend configured (MinIO/OSS)")
	}

	// 7. 邮件发送服务（延迟初始化）
	email.InitEmailService(&cfg.Email)
	slog.Info("[7/10] Email service ready")

	// 8. 策略模式 (ES/MySQL双搜索 + MinIO/OSS双上传, 对标Java StrategyContext)
	db := database.GetDB()
	if db != nil {
		if err := strategy.InitStrategies(cfg, db); err != nil {
			slog.Error("[8/10] Strategy init failed (non-fatal), using defaults", "error", err)
		} else {
			if sc := strategy.GetSearchContext(); sc != nil {
				slog.Info("[8/10] Search+Upload strategies initialized",
					"search_mode", sc.GetMode(),
				)
			}
		}
	} else {
		slog.Warn("[8/10] Skipped strategy init (no DB connection)")
	}

	// 9. 定时任务调度器 (robfig/cron, 对标Java Quartz)
	if db != nil {
		rdb := database.GetRedis()
		InitScheduler(cfg, db, rdb)
		slog.Info("[9/10] Scheduler initialized from database")
	} else {
		slog.Warn("[9/10] Skipped scheduler init (no DB connection)")
	}

	slog.Info("=== Aurora Go Infrastructure Bootstrap Complete ===")
}

// Shutdown 优雅关闭所有基础设施（逆序关闭）
// 关闭顺序: Email → MQ → ES → MinIO → Redis → DB → Logger
// 对标 Spring 的 DisposableBean.destroy() 或 ShutdownHook
func Shutdown() {
	slog.Info("=== Aurora Go Graceful Shutdown Start ===")

	// 创建超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 使用 channel 等待所有清理完成
	done := make(chan struct{})

	go func() {
		defer close(done)

		// 1. 关闭邮件服务（无状态，无需关闭）

		// 2. 关闭 RabbitMQ 连接
		if err := mq.CloseRabbitMQ(); err != nil {
			slog.Error("Failed to close RabbitMQ", "error", err)
		} else {
			slog.Info("RabbitMQ connection closed")
		}

		// 3. Elasticsearch 无状态客户端，无需关闭

		// 4. MinIO 无状态客户端，无需关闭

		// 5. 关闭 Redis
		if err := database.CloseRedis(); err != nil {
			slog.Error("Failed to close Redis", "error", err)
		} else {
			slog.Info("Redis connection closed")
		}

		// 6. 关闭 MySQL
		if err := database.Close(); err != nil {
			slog.Error("Failed to close MySQL", "error", err)
		} else {
			slog.Info("MySQL connection closed")
		}

		// 7. 刷新日志缓冲区
		if err := logger.Sync(); err != nil {
			slog.Error("Failed to sync logger", "error", err)
		} else {
			slog.Info("Logger flushed")
		}
	}()

	select {
	case <-done:
		slog.Info("=== Aurora Go Graceful Shutdown Complete ===")
	case <-ctx.Done():
		slog.Error("Shutdown timeout exceeded, forcing exit...")
	}
}

// WaitForSignal 等待系统信号触发优雅关闭
// 监听: SIGINT(CTRL+C) / SIGTERM(kill) / SIGHUP
func WaitForSignal() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	sig := <-sigChan
	slog.Info("Received shutdown signal", "signal", sig.String())
	Shutdown()
	os.Exit(0)
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return database.GetDB()
}

// GetRedis 获取Redis客户端
func GetRedis() *redis.Client {
	return database.RDB
}

// GetES 获取ES客户端
func GetES() *search.ESClient {
	return search.Client
}

// GetEmailService 获取邮件服务实例
func GetEmailService() interface{} {
	return email.GetEmailService()
}

// HealthCheck 健康检查（返回各中间件连接状态）
// 用于: /api/actuator/health 接口
type HealthStatus map[string]string

func HealthCheck() HealthStatus {
	status := make(map[string]string)

	// 检查MySQL
	if db := database.GetDB(); db != nil {
		sqlDB, err := db.DB()
		if err == nil && sqlDB.Ping() == nil {
			status["mysql"] = "UP"
		} else {
			status["mysql"] = "DOWN"
		}
	} else {
		status["mysql"] = "DOWN"
	}

	// 检查Redis
	if rdb := database.GetRedis(); rdb != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		if rdb.Ping(ctx).Err() == nil {
			status["redis"] = "UP"
			cancel()
		} else {
			status["redis"] = "DOWN"
			cancel()
		}
	} else {
		status["redis"] = "DOWN"
	}

	// 检查RabbitMQ
	if ch := mq.GetChannel(); ch != nil && !ch.IsClosed() {
		status["rabbitmq"] = "UP"
	} else {
		status["rabbitmq"] = "DISABLED"
	}

	// 检查ES
	if es := search.Client; es != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		if _, err := es.Health(ctx); err == nil {
			status["elasticsearch"] = "UP"
			cancel()
		} else {
			status["elasticsearch"] = "DOWN"
			cancel()
		}
	} else {
		status["elasticsearch"] = "DISABLED"
	}

	// 检查MinIO
	if mc := storage.MinIOClient; mc != nil {
		status["minio"] = "UP"
	} else {
		status["minio"] = "DISABLED"
	}

	return status
}

// ==================== 调度器管理 ====================

var (
	appScheduler *scheduler.Scheduler
)

// InitScheduler 初始化定时任务调度器 (robfig/cron, 对标Java Quartz Scheduler)
// 从数据库动态加载所有任务并注册到调度器
func InitScheduler(cfg *config.Config, db *gorm.DB, rdb *redis.Client) {
	// 创建调度器（传入Redis客户端）
	appScheduler = scheduler.NewScheduler(db, rdb, cfg.Server.GetSiteURL())

	// 初始化任务函数注册表（对标Java Spring Bean扫描）
	scheduler.InitTaskRegistry(appScheduler)

	// 从数据库加载所有任务（对标Java @PostConstruct init()）
	if err := appScheduler.InitFromDatabase(); err != nil {
		slog.Error("从数据库加载定时任务失败", "error", err)
	}

	// 启动调度器
	appScheduler.Start()
}

// GetScheduler 获取调度器实例
func GetScheduler() *scheduler.Scheduler {
	return appScheduler
}

// StopScheduler 停止调度器
func StopScheduler() {
	if appScheduler != nil && appScheduler.IsRunning() {
		ctx := appScheduler.Stop()
		<-ctx.Done()
	}
}
