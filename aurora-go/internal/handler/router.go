package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/aurora-go/aurora/internal/middleware"
	"github.com/aurora-go/aurora/internal/service"
)

// Router 路由注册器（对标Java的WebMvcConfigurer/路由配置）
// 将所有Handler端点注册到Gin Engine上
type Router struct {
	// Handler 实例
	ArticleHandler       *ArticleHandler
	UserAuthHandler      *UserAuthHandler
	CommentHandler       *CommentHandler
	CategoryHandler      *CategoryHandler
	TagHandler           *TagHandler
	FriendLinkHandler    *FriendLinkHandler
	TalkHandler          *TalkHandler
	PhotoHandler         *PhotoHandler
	PhotoAlbumHandler    *PhotoAlbumHandler
	RoleHandler          *RoleHandler
	MenuHandler          *MenuHandler
	JobHandler           *JobHandler
	JobLogHandler        *JobLogHandler
	OperationLogHandler  *OperationLogHandler
	ExceptionLogHandler  *ExceptionLogHandler
	AuroraInfoHandler    *AuroraInfoHandler
	WebsiteConfigHandler *WebsiteConfigHandler
	FileHandler          *FileHandler
	ResourceHandler      *ResourceHandler
	AboutHandler         *AboutHandler
	// JWT认证服务（用于admin路由）
	tokenSvc *service.TokenService
	logger   *slog.Logger
}

// NewRouter 创建路由器并初始化所有Handler实例
// 所有Handler通过Registry获取Service实例, 确保单例共享
func NewRouter(registry *service.Registry, tokenSvc *service.TokenService, logger *slog.Logger) *Router {
	return &Router{
		ArticleHandler:       NewArticleHandler(registry.Article, registry.File),
		UserAuthHandler:      NewUserAuthHandler(registry),
		CommentHandler:       NewCommentHandler(registry.Comment),
		CategoryHandler:      NewCategoryHandler(registry.Category),
		TagHandler:           NewTagHandler(registry.Tag),
		FriendLinkHandler:    NewFriendLinkHandler(registry.FriendLink),
		TalkHandler:          NewTalkHandler(registry.Talk, registry.File),
		PhotoHandler:         NewPhotoHandler(registry.Photo, registry.UploadSvc),
		PhotoAlbumHandler:    NewPhotoAlbumHandler(registry.PhotoAlbum, registry.UploadSvc),
		RoleHandler:          NewRoleHandler(registry.Role),
		MenuHandler:          NewMenuHandler(registry.Menu),
		JobHandler:           NewJobHandler(registry.Job),
		JobLogHandler:        NewJobLogHandler(registry.JobLog),
		OperationLogHandler:  NewOperationLogHandler(registry.OperationLog),
		ExceptionLogHandler:  NewExceptionLogHandler(registry.ExceptionLog),
		AuroraInfoHandler:    NewAuroraInfoHandler(registry.AuroraInfo, registry.StatsService),
		WebsiteConfigHandler: NewWebsiteConfigHandler(registry),
		FileHandler:          NewFileHandler(registry.File),
		ResourceHandler:      NewResourceHandler(registry.Resource),
		AboutHandler:         NewAboutHandler(registry.About),
		tokenSvc:             tokenSvc,
		logger:               logger,
	}
}

// RegisterRoutes 注册所有路由（在main.go中调用）
// 路径完全对齐 Java SpringBoot 原始 API 路径
func (r *Router) RegisterRoutes(engine *gin.Engine) {
	// 静态文件服务 - 服务上传的资源文件（对标Java版 Spring Boot static-resource）
	// 访问 http://localhost:8080/aurora/articles/xxx.jpg 可以访问上传的图片
	engine.Static("/aurora", "./aurora")

	api := engine.Group("/api")

	// ==================== 公开路由（无需认证）====================
	public := api.Group("")
	r.registerPublicRoutes(public)

	// ==================== 受保护路由（需JWT认证）====================
	protected := api.Group("")
	protected.Use(middleware.JWTAuth())

	r.registerProtectedRoutes(protected)

	// ==================== 后台管理路由（需JWT + 管理员角色）====================
	admin := api.Group("/admin")
	// 使用完整的JWT认证中间件（解析Token，获取用户ID，自动续期）
	admin.Use(middleware.JWTAuthEnhanced(r.tokenSvc, r.logger))
	// 使用RBAC权限控制中间件（对标Java的FilterInvocationSecurityMetadataSource + AccessDecisionManager）
	admin.Use(middleware.RBAC(service.GetGlobalRegistry()))
	// 使用操作日志中间件（对标Java @OptLog AOP）
	admin.Use(middleware.AccessLog(service.GetGlobalRegistry(), r.logger))

	r.registerAdminRoutes(admin)
}

// registerPublicRoutes 公开API路由（前台访问）
// 路径对齐 Java: context-path=/api + Controller原始路径
func (r *Router) registerPublicRoutes(rg *gin.RouterGroup) {
	// --- 文章（ArticleController） ---
	rg.GET("/articles/topAndFeatured", r.ArticleHandler.TopAndFeaturedArticles)
	rg.GET("/articles/all", r.ArticleHandler.ListArticles)
	rg.GET("/articles/categoryId", r.ArticleHandler.ListArticlesByCategoryId)
	rg.GET("/articles/:id", r.ArticleHandler.GetArticleById)
	rg.POST("/articles/access", r.ArticleHandler.VerifyArticlePassword)
	rg.GET("/articles/tagId", r.ArticleHandler.ListArticlesByTagId)
	rg.GET("/articles/search", r.ArticleHandler.SearchArticles)
	rg.GET("/archives/all", r.ArticleHandler.GetArchives)

	// --- 用户认证（UserAuthController） ---
	rg.POST("/users/login", r.UserAuthHandler.Login)
	rg.POST("/users/register", r.UserAuthHandler.Register)
	rg.GET("/users/code", r.UserAuthHandler.SendVerificationCode)
	rg.PUT("/users/password", r.UserAuthHandler.UpdatePassword) // 公开接口：内部通过code区分重置/修改
	rg.POST("/users/oauth/qq", r.UserAuthHandler.QQLogin)
	rg.POST("/users/password/reset", r.UserAuthHandler.ResetPassword)
	// --- 评论（CommentController） ---
	rg.GET("/comments", r.CommentHandler.ListComments)
	rg.POST("/comments/save", r.CommentHandler.AddComment)
	rg.GET("/comments/:id/replies", r.CommentHandler.ReplyComment)
	rg.GET("/comments/topSix", r.CommentHandler.ListTopSixComments)

	// --- 分类（CategoryController） ---
	rg.GET("/categories/all", r.CategoryHandler.ListCategories)

	// --- 标签（TagController） ---
	rg.GET("/tags/all", r.TagHandler.ListTags)
	rg.GET("/tags/topTen", r.TagHandler.ListTopTenTags)

	// --- 友链（FriendLinkController） ---
	rg.GET("/links", r.FriendLinkHandler.ListFriendLinks)

	// --- 说说（TalkController） ---
	rg.GET("/talks", r.TalkHandler.ListTalks)
	rg.GET("/talks/:id", r.TalkHandler.GetTalkById)

	// --- 相册（PhotoAlbumController + PhotoController 前台） ---
	rg.GET("/photos/albums", r.PhotoAlbumHandler.ListAlbums)
	rg.GET("/albums/:albumId/photos", r.PhotoHandler.ListPhotosByAlbumId)

	// --- 关于页（AuroraInfoController） ---
	rg.GET("/about", r.AboutHandler.GetAbout)

	// --- 首页信息聚合（AuroraInfoController） ---
	rg.GET("/", r.AuroraInfoHandler.GetHomeInfo)

	// --- 网站配置（AuroraInfoController 前台只读） ---
	rg.GET("/admin/website/config", r.WebsiteConfigHandler.GetWebsiteConfig)

	// --- 访客上报（AuroraInfoController） ---
	rg.POST("/report", r.AuroraInfoHandler.Report)
}

// registerProtectedRoutes 受保护路由（需JWT登录）
func (r *Router) registerProtectedRoutes(rg *gin.RouterGroup) {
	// --- 用户信息（UserInfoController） ---
	rg.PUT("/users/info", r.UserAuthHandler.UpdateUserInfo)
	rg.POST("/users/avatar", r.UserAuthHandler.UpdateUserAvatar)
	rg.PUT("/users/email", r.UserAuthHandler.BindUserEmail)
	rg.PUT("/users/subscribe", r.UserAuthHandler.UpdateUserSubscribe)
	rg.POST("/users/logout", r.UserAuthHandler.Logout)
	rg.GET("/users/info/:id", r.UserAuthHandler.GetUserInfoById)
}

// registerAdminRoutes 后台管理路由
// 路径对齐 Java: /admin + Controller原始路径
func (r *Router) registerAdminRoutes(rg *gin.RouterGroup) {
	// --- 系统信息（AuroraInfoController） ---
	rg.GET("/", r.AuroraInfoHandler.GetAdminInfo)
	rg.PUT("/website/config", r.WebsiteConfigHandler.UpdateWebsiteConfig)
	rg.PUT("/about", r.AboutHandler.SaveOrUpdate)
	rg.POST("/config/images", r.WebsiteConfigHandler.UploadConfigImage)

	// --- 用户管理（UserAuthController + UserInfoController） ---
	rg.GET("/users", r.UserAuthHandler.ListUsers)
	rg.GET("/users/area", r.UserAuthHandler.ListUserAreas)
	rg.POST("/users/area/trigger", r.UserAuthHandler.TriggerUserAreaStats) // 手动触发统计（测试用）
	rg.PUT("/users/password", r.UserAuthHandler.UpdateAdminPassword)
	rg.PUT("/users/role", r.UserAuthHandler.UpdateUserRole)
	rg.PUT("/users/disable", r.UserAuthHandler.UpdateUserDisable)
	rg.GET("/users/online", r.UserAuthHandler.ListOnlineUsers)
	rg.DELETE("/users/:id/online", r.UserAuthHandler.RemoveOnlineUser)
	rg.GET("/users/role", r.RoleHandler.ListRoles) // 获取角色列表（用于编辑用户时选择角色）

	// --- 文章管理（ArticleController） ---
	rg.GET("/articles", r.ArticleHandler.ListAdminArticles)
	rg.POST("/articles", r.ArticleHandler.SaveArticle)             // 统一入口（对标Java：通过id区分新增/更新）
	rg.POST("/articles/save", r.ArticleHandler.SaveArticle)       // 新增文章（兼容Vue3前端）
	rg.POST("/articles/update", r.ArticleHandler.SaveArticle)     // 更新文章（兼容Vue3前端）
	rg.PUT("/articles/topAndFeatured", r.ArticleHandler.UpdateArticleTopAndFeatured)
	rg.PUT("/articles", r.ArticleHandler.UpdateArticleDelete)
	rg.DELETE("/articles/delete", r.ArticleHandler.DeleteArticle) // 彻底删除（物理删除）
	rg.POST("/articles/images", r.ArticleHandler.UploadArticleImage)
	rg.GET("/articles/:id", r.ArticleHandler.GetAdminArticleById)
	rg.POST("/articles/import", r.ArticleHandler.ImportArticle)
	rg.POST("/articles/export", r.ArticleHandler.ExportArticle)

	// --- 分类管理（CategoryController） ---
	rg.GET("/categories", r.CategoryHandler.ListAdminCategories)
	rg.GET("/categories/search", r.CategoryHandler.SearchCategories)
	rg.DELETE("/categories", r.CategoryHandler.DeleteCategory)
	rg.POST("/categories", r.CategoryHandler.SaveOrUpdate)

	// --- 标签管理（TagController） ---
	rg.GET("/tags", r.TagHandler.ListAdminTags)
	rg.GET("/tags/search", r.TagHandler.SearchTags)
	rg.POST("/tags", r.TagHandler.SaveOrUpdate)
	rg.DELETE("/tags", r.TagHandler.DeleteTags)

	// --- 评论管理（CommentController） ---
	rg.GET("/comments", r.CommentHandler.ListAdminComments)
	rg.PUT("/comments/review", r.CommentHandler.UpdateCommentReview)
	rg.DELETE("/comments", r.CommentHandler.DeleteComment)

	// --- 友链管理（FriendLinkController） ---
	rg.GET("/links", r.FriendLinkHandler.ListAdminFriendLinks)
	rg.POST("/links", r.FriendLinkHandler.SaveOrUpdateFriendLink)  // 新增或更新（对标Java: saveOrUpdateFriendLink）
	rg.PUT("/links", r.FriendLinkHandler.UpdateFriendLink)        // 更新（兼容前端）
	rg.DELETE("/links", r.FriendLinkHandler.DeleteFriendLink)

	// --- 说说管理（TalkController） ---
	rg.GET("/talks", r.TalkHandler.ListAdminTalks)
	rg.GET("/talks/:id", r.TalkHandler.GetAdminTalkById)
	rg.POST("/talks", r.TalkHandler.SaveOrUpdate)
	rg.POST("/talks/images", r.TalkHandler.UploadTalkImage)
	rg.DELETE("/talks", r.TalkHandler.DeleteTalks)

	// --- 相册管理（PhotoAlbumController） ---
	rg.GET("/photos/albums", r.PhotoAlbumHandler.ListAdminAlbums)
	rg.GET("/photos/albums/info", r.PhotoAlbumHandler.ListAlbumInfos)
	rg.GET("/photos/albums/:id/info", r.PhotoAlbumHandler.GetAlbumById)
	rg.POST("/photos/albums", r.PhotoAlbumHandler.SaveOrUpdate)
	rg.POST("/photos/albums/upload", r.PhotoAlbumHandler.UploadAlbumCover)
	rg.DELETE("/photos/albums/:id", r.PhotoAlbumHandler.DeleteAlbum)

	// --- 照片管理（PhotoController） ---
	rg.GET("/photos", r.PhotoHandler.ListAdminPhotos)
	rg.POST("/photos/upload", r.PhotoHandler.UploadPhoto)
	rg.POST("/photos", r.PhotoHandler.SavePhotos)
	rg.PUT("/photos", r.PhotoHandler.UpdatePhoto)
	rg.PUT("/photos/album", r.PhotoHandler.MovePhotosAlbum)
	rg.PUT("/photos/delete", r.PhotoHandler.UpdatePhotoDelete)
	rg.DELETE("/photos", r.PhotoHandler.DeletePhotos)

	// --- 角色管理（RoleController） ---
	rg.GET("/roles", r.RoleHandler.ListRoles)
	rg.POST("/role", r.RoleHandler.SaveOrUpdate)
	rg.DELETE("/roles", r.RoleHandler.DeleteRoles)
	rg.GET("/role/menus", r.MenuHandler.ListMenuOptions)
	rg.GET("/role/resources", r.ResourceHandler.ListResourceOptions)

	// --- 资源管理（ResourceController） ---
	rg.GET("/resources", r.ResourceHandler.ListResources)
	rg.POST("/resources", r.ResourceHandler.SaveOrUpdate)
	rg.DELETE("/resources/:id", r.ResourceHandler.DeleteResource)

	// --- 菜单管理（MenuController） ---
	rg.GET("/menus", r.MenuHandler.ListMenus)
	rg.POST("/menus", r.MenuHandler.SaveOrUpdate)
	rg.PUT("/menus/isHidden", r.MenuHandler.UpdateMenuIsHidden)
	rg.DELETE("/menus/:id", r.MenuHandler.DeleteMenu)
	rg.GET("/user/menus", r.MenuHandler.GetUserMenus)

	// --- 定时任务管理（JobController） ---
	rg.GET("/jobs", r.JobHandler.ListJobs)
	rg.POST("/jobs", r.JobHandler.SaveJob)
	rg.PUT("/jobs", r.JobHandler.UpdateJob)
	rg.DELETE("/jobs", r.JobHandler.DeleteJob)
	rg.GET("/jobs/:id", r.JobHandler.GetJobById)
	rg.PUT("/jobs/status", r.JobHandler.UpdateJobStatus)
	rg.PUT("/jobs/run", r.JobHandler.RunJobOnce)  // Java使用PUT
	rg.GET("/jobs/jobGroups", r.JobHandler.ListJobGroups)

	// --- 定时任务日志（JobLogController） ---
	rg.GET("/jobLogs", r.JobLogHandler.ListJobLogs)
	rg.DELETE("/jobLogs", r.JobLogHandler.DeleteJobLogs)
	rg.DELETE("/jobLogs/clean", r.JobLogHandler.CleanJobLogs)
	rg.GET("/jobLogs/jobGroups", r.JobLogHandler.ListJobLogGroups)

	// --- 操作日志（OperationLogController） ---
	rg.GET("/operation/logs", r.OperationLogHandler.ListOperationLogs)
	rg.DELETE("/operation/logs", r.OperationLogHandler.DeleteOperationLogs)

	// --- 异常日志（ExceptionLogController） ---
	rg.GET("/exception/logs", r.ExceptionLogHandler.ListExceptionLogs)
	rg.DELETE("/exception/logs", r.ExceptionLogHandler.DeleteExceptionLogs)

	// --- 文件上传 ---
	rg.POST("/upload", r.FileHandler.UploadFile)
	rg.POST("/upload/batch", r.FileHandler.BatchUpload)
	rg.POST("/upload/image", r.FileHandler.UploadImage)
}
