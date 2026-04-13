package service

import (
	"log/slog"

	"github.com/aurora-go/aurora/internal/config"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Registry 服务注册中心 (依赖注入容器)
// 对标 Java Spring @Service + @Autowired 自动装配
// 所有Handler通过Registry获取所需Service实例, 确保单例共享
type Registry struct {
	DB  *gorm.DB
	RDB *redis.Client // Redis连接池 (TokenService/限流/缓存用)

	// 核心业务服务
	Article      *ArticleService
	UserAuth     *UserAuthService
	Comment      *CommentService
	Category     *CategoryService
	Tag          *TagService
	FriendLink   *FriendLinkService
	Talk         *TalkService
	Photo        *PhotoService
	PhotoAlbum   *PhotoAlbumService
	Role         *RoleService
	Menu         *MenuService
	Job          *JobService
	JobLog       *JobLogService
	OperationLog *OperationLogService
	ExceptionLog *ExceptionLogService

	// 聚合/辅助服务
	AuroraInfo    *AuroraInfoService
	WebsiteConfig *WebsiteConfigService
	File          *FileService
	Resource      *ResourceService
	About         *AboutService

	// 安全/认证服务 (P0-6新增)
	TokenSvc    *TokenService       // JWT Token管理(含Redis Session)
	QQOAuthSvc  *QQOAuthService     // QQ OAuth登录
}

// NewRegistry 创建服务注册中心 (所有Service共享同一个DB+Redis连接池)
// 调用时机: main.go Bootstrap阶段, 在DB和Redis初始化完成后立即调用
func NewRegistry(db *gorm.DB, rdb *redis.Client, cfg config.Config, logger *slog.Logger) *Registry {
	logger.Info("初始化Service注册中心...")

	// 创建 Redis 统计服务（所有 Service 共享）
	statsService := NewRedisStatsService(rdb)

	r := &Registry{
		DB:  db,
		RDB: rdb,
	}

	// ===== 基础服务 (依赖DB+Redis) =====
	r.Article = NewArticleService(db, statsService)
	r.UserAuth = NewUserAuthService(db)
	r.Comment = NewCommentService(db, statsService)
	r.Category = NewCategoryService(db)
	r.Tag = NewTagService(db)
	r.FriendLink = NewFriendLinkService(db)
	r.Talk = NewTalkService(db)
	r.Photo = NewPhotoService(db)
	r.PhotoAlbum = NewPhotoAlbumService(db)
	r.Role = NewRoleService(db)
	r.Menu = NewMenuService(db)
	r.Job = NewJobService(db)
	r.JobLog = NewJobLogService(db)
	r.OperationLog = NewOperationLogService(db)
	r.ExceptionLog = NewExceptionLogService(db)
	r.AuroraInfo = NewAuroraInfoService(db, statsService)
	r.WebsiteConfig = NewWebsiteConfigService(db)
	r.File = NewFileService(db)
	r.Resource = NewResourceService(db)
	r.About = NewAboutService(db)

	// ===== 安全认证服务 (依赖DB+Redis+配置) =====
	r.TokenSvc = NewTokenService(cfg.JWT, rdb, logger)
	r.QQOAuthSvc = NewQQOAuthService(cfg.QQ, r.UserAuth, r.TokenSvc, logger)

	total := 24 // 24个Service实例 (+2安全服务)
	logger.Info("Service注册中心初始化完成", "services", total, "has_redis", rdb != nil)
	return r
}

// Close 关闭所有服务资源 (预留清理接口)
func (r *Registry) Close() error {
	slog.Info("关闭Service注册中心...")
	return nil
}

// ===== 全局单例访问 (供Agent Tool等外部模块使用) =====

var globalRegistry *Registry

// SetGlobalRegistry 设置全局Registry实例 (main.go Bootstrap阶段调用)
func SetGlobalRegistry(reg *Registry) {
	globalRegistry = reg
}

// GetGlobalRegistry 获取全局Registry实例
func GetGlobalRegistry() *Registry {
	return globalRegistry
}

// GetArticleService 获取文章服务快捷方法
func GetArticleService() *ArticleService {
	if globalRegistry != nil {
		return globalRegistry.Article
	}
	return nil
}

// GetTagService 获取标签服务快捷方法
func GetTagService() *TagService {
	if globalRegistry != nil {
		return globalRegistry.Tag
	}
	return nil
}

// GetCategoryService 获取分类服务快捷方法
func GetCategoryService() *CategoryService {
	if globalRegistry != nil {
		return globalRegistry.Category
	}
	return nil
}
