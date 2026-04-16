package strategy

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/aurora-go/aurora/internal/config"
	"github.com/aurora-go/aurora/internal/errors"

	"gorm.io/gorm"
)

// ==================== 常量定义 (对标Java CommonConstant) ====================

const (
	PreTag = "<mark>"  // 高亮开始标记 (对标Java PRE_TAG)
	PostTag = "</mark>" // 高亮结束标记 (对标Java POST_TAG)
)

// ==================== 搜索模式枚举 (对标Java SearchModeEnum) ====================

// SearchMode 搜索模式常量 (对标Java SearchModeEnum)
const (
	SearchModeMySQL        = "mysql"         // MySQL LIKE模糊搜索
	SearchModeElasticsearch = "elasticsearch" // ES全文搜索(ik_max_word分词)
)

// ==================== 上传模式枚举 (对标Java UploadModeEnum) ====================

// UploadMode 上传模式常量 (对标Java UploadModeEnum)
const (
	UploadModeMinIO = "minio" // MinIO对象存储
	UploadModeOSS   = "oss"   // 阿里云OSS
)

// ==================== 策略上下文 (工厂模式 + 自动选择) ====================

// SearchContext 搜索策略上下文 (对标Java SearchStrategyContext)
// 根据 search.mode 配置自动选择 ES 或 MySQL 搜索策略
//
// 对标Java:
//   @Service
//   public class SearchStrategyContext {
//       @Value("${search.mode}") private String searchMode;
//       @Autowired private Map<String, SearchStrategy> searchStrategyMap;
//       public List<ArticleSearchDTO> executeSearchStrategy(String keywords) { ... }
//   }
type SearchContext struct {
	strategy SearchStrategy
	mode     string
}

// NewSearchContext 创建搜索策略上下文（根据模式自动选择实现）
//
// 选择逻辑:
//  - elasticsearch: 使用 EsSearchStrategy (需要ES服务不为nil)
//  - mysql: 使用 MySQLSearchStrategy
//  - 默认: 如果ES可用优先ES，否则降级到MySQL
//
// Go增强点:
//  - ES不可用时自动降级到MySQL（Java版会NPE）
//  - 支持运行时动态切换模式
func NewSearchContext(mode string, esClient ESClientInterface, db *gorm.DB) (*SearchContext, error) {
	var strat SearchStrategy
	actualMode := mode

	switch mode {
	case SearchModeElasticsearch:
		// ES模式: 需要ES服务初始化成功
		if esClient == nil {
			slog.Warn("Elasticsearch requested but not available, falling back to MySQL",
				"mode", mode,
			)
			strat = NewMySQLSearchStrategy(db)
			actualMode = SearchModeMySQL
		} else {
			strat = NewEsSearchStrategy(esClient)
		}
	case SearchModeMySQL, "":
		strat = NewMySQLSearchStrategy(db)
		actualMode = SearchModeMySQL
	default:
		return nil, fmt.Errorf("%w: unknown search mode=%q (supported: %s, %s)",
			errors.ErrInvalidConfig, mode, SearchModeElasticsearch, SearchModeMySQL)
	}

	slog.Info("Search strategy initialized", "mode", actualMode, "strategy", fmt.Sprintf("%T", strat))
	return &SearchContext{strategy: strat, mode: actualMode}, nil
}

// ExecuteSearch 执行搜索（委托给当前策略） (对标Java executeSearchStrategy)
func (ctx *SearchContext) ExecuteSearch(c context.Context, keywords string, current, size int) ([]map[string]interface{}, int, error) {
	if ctx.strategy == nil {
		return []map[string]interface{}{}, 0, fmt.Errorf("search strategy not initialized")
	}
	return ctx.strategy.SearchArticles(c, keywords, current, size)
}

// GetMode 获取当前搜索模式
func (ctx *SearchContext) GetMode() string {
	return ctx.mode
}

// ==================== 上传策略上下文 ====================

// ReadSeekerCloser 接口组合: 同时支持读取和关闭操作（用于文件上传流）
// 对标Java MultipartFile.getInputStream()
type ReadSeekerCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

// UploadContext 上传策略上下文 (对标Java UploadStrategyContext)
// 根据 upload.mode 配置自动选择 MinIO 或 OSS 上传策略
type UploadContext struct {
	strategy UploadStrategy
	mode     string
}

// NewUploadContext 创建上传策略上下文（根据模式自动选择实现）
//
// 选择逻辑:
//  - minio: 使用 MinIOUploadStrategy (需要MinIO配置)
//  - oss: 使用 OSSUploadStrategy (需要OSS配置)
//  - 默认: 使用MinIO
func NewUploadContext(mode string, minioCfg config.MinIOConfig) (*UploadContext, error) {
	var strat UploadStrategy
	actualMode := mode

	switch mode {
	case UploadModeOSS:
		// OSS模式: 当前为预留实现（需引入aliyun-oss-go-sdk）
		slog.Warn("OSS upload mode selected but SDK not integrated, using placeholder",
			"hint", "请引入 github.com/aliyun/aliyun-oss-go-sdk/oss 以启用完整功能",
		)
		return nil, fmt.Errorf("%w: OSS upload requires aliyun-oss-go-sdk integration", errors.ErrNotImplemented)
	case UploadModeMinIO, "":
		strategy, err := NewMinIOUploadStrategy(minioCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create MinIO upload strategy: %w", err)
		}
		strat = strategy
		actualMode = UploadModeMinIO
	default:
		return nil, fmt.Errorf("%w: unknown upload mode=%q (supported: %s, %s)",
			errors.ErrInvalidConfig, mode, UploadModeMinIO, UploadModeOSS)
	}

	slog.Info("Upload strategy initialized", "mode", actualMode, "strategy", fmt.Sprintf("%T", strat))
	return &UploadContext{strategy: strat, mode: actualMode}, nil
}

// ExecuteUploadFromFile 从文件流上传（带MD5去重） (对标Java executeUploadStrategy(MultipartFile, path))
func (ctx *UploadContext) ExecuteUploadFromFile(c context.Context, fileName string, reader ReadSeekerCloser, path string) (string, error) {
	if ctx.strategy == nil {
		return "", fmt.Errorf("upload strategy not initialized")
	}

	base := &BaseUploadStrategy{}
	return base.UploadFileFromStream(c, ctx.strategy, fileName, reader, path)
}

// GetMode 获取当前上传模式
func (ctx *UploadContext) GetMode() string {
	return ctx.mode
}
