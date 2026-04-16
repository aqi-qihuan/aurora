package strategy

import (
	"fmt"
	"log/slog"

	"github.com/aurora-go/aurora/internal/config"

	"gorm.io/gorm"
)

// ==================== 全局策略上下文单例 ====================

var (
	// GlobalSearchContext 全局搜索策略上下文（单例）
	GlobalSearchContext *SearchContext

	// GlobalUploadContext 全局上传策略上下文（单例）
	GlobalUploadContext *UploadContext

	// globalESClient 全局 ES 客户端（用于避免循环导入）
	globalESClient ESClientInterface
)

// InitStrategies 初始化所有策略上下文 (对标Java Spring自动注入)
//
// 应在 Bootstrap 阶段调用，确保 DB/Redis/ES/MinIO 都已初始化
// 对标Java: @Autowired Map<String, SearchStrategy> searchStrategyMap 的自动装配
func InitStrategies(cfg *config.Config, db *gorm.DB) error {
	slog.Info("=== Initializing Strategy Pattern ===")

	// 1. 初始化搜索策略（含自动降级）
	searchCtx, err := initSearchStrategy(cfg, db)
	if err != nil {
		slog.Error("Failed to initialize search strategy", "error", err)
		// 搜索策略失败不阻断启动，使用MySQL兜底
		searchCtx, _ = NewSearchContext(SearchModeMySQL, nil, db)
	}
	GlobalSearchContext = searchCtx
	slog.Info("[Strategy] Search context ready",
		"mode", GlobalSearchContext.GetMode(),
	)

	// 2. 初始化上传策略
	uploadCtx, err := initUploadStrategy(cfg)
	if err != nil {
		slog.Error("Failed to initialize upload strategy", "error", err)
		// 上传策略失败不阻断启动，后续可延迟初始化
		GlobalUploadContext = nil
	} else {
		GlobalUploadContext = uploadCtx
		slog.Info("[Strategy] Upload context ready",
			"mode", GlobalUploadContext.GetMode(),
		)
	}

	slog.Info("=== Strategy Pattern Initialized ===")
	return nil
}

// initSearchStrategy 初始化搜索策略（含自动降级）
func initSearchStrategy(cfg *config.Config, db *gorm.DB) (*SearchContext, error) {
	mode := cfg.Search.Mode

	// ES 服务指针（如果 ES 未启用则为 nil）
	var esClient ESClientInterface
	if mode == SearchModeElasticsearch {
		// 从全局获取 ES 服务（通过接口避免循环导入）
		esClient = getGlobalESClient()
	}

	return NewSearchContext(mode, esClient, db)
}

// getGlobalESClient 获取全局 ES 客户端（延迟加载，避免循环导入）
func getGlobalESClient() ESClientInterface {
	return globalESClient
}

// SetGlobalESClient 设置全局 ES 客户端（在 main.go 中调用）
func SetGlobalESClient(client ESClientInterface) {
	globalESClient = client
}

// GetGlobalESClient 获取全局 ES 客户端（对外暴露，供 main.go 等使用）
func GetGlobalESClient() ESClientInterface {
	return globalESClient
}

// initUploadStrategy 初始化上传策略
func initUploadStrategy(cfg *config.Config) (*UploadContext, error) {
	mode := cfg.Upload.Mode
	if mode == "" {
		mode = UploadModeMinIO // 默认使用MinIO（对齐Java application-prod.yml: upload.mode=minio）
	}

	switch mode {
	case UploadModeMinIO:
		return NewUploadContext(mode, cfg.MinIO)
	case UploadModeOSS:
		if cfg.OSS == nil {
			return nil, fmt.Errorf("OSS mode selected but oss config is missing")
		}
		// OSS需要特殊处理: NewUploadContext接受MinIOConfig类型
		// 这里创建一个临时的MinIOConfig来传递OSS配置信息
		tmpCfg := config.MinIOConfig{
			Bucket: cfg.OSS.BucketName,
		}
		return NewUploadContext(mode, tmpCfg)
	default:
		return nil, fmt.Errorf("unknown upload mode: %s", mode)
	}
}

// ==================== 便捷访问函数 ====================

// GetSearchContext 获取全局搜索策略上下文（安全访问，nil-safe）
func GetSearchContext() *SearchContext {
	if GlobalSearchContext == nil {
		slog.Warn("GlobalSearchContext is nil! Call InitStrategies first.")
	}
	return GlobalSearchContext
}

// GetUploadContext 获取全局上传策略上下文（安全访问，nil-safe）
func GetUploadContext() *UploadContext {
	if GlobalUploadContext == nil {
		slog.Warn("GlobalUploadContext is nil! Call InitStrategies first.")
	}
	return GlobalUploadContext
}
