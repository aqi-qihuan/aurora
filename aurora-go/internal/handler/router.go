package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/aurora-go/aurora/internal/config"
	"github.com/aurora-go/aurora/internal/middleware"
)

// Router 路由注册器（对标Java的WebMvcConfigurer/路由配置）
// 将所有Handler端点注册到Gin Engine上
type Router struct {
	cfg *config.Config

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
}

// NewRouter 创建路由器并初始化所有Handler实例
func NewRouter(cfg *config.Config) *Router {
	return &Router{
		cfg: cfg,
		ArticleHandler:       NewArticleHandler(),
		UserAuthHandler:      NewUserAuthHandler(),
		CommentHandler:       NewCommentHandler(),
		CategoryHandler:      NewCategoryHandler(),
		TagHandler:           NewTagHandler(),
		FriendLinkHandler:    NewFriendLinkHandler(),
		TalkHandler:          NewTalkHandler(),
		PhotoHandler:         NewPhotoHandler(),
		PhotoAlbumHandler:    NewPhotoAlbumHandler(),
		RoleHandler:          NewRoleHandler(),
		MenuHandler:          NewMenuHandler(),
		JobHandler:           NewJobHandler(),
		JobLogHandler:        NewJobLogHandler(),
		OperationLogHandler:  NewOperationLogHandler(),
		ExceptionLogHandler:  NewExceptionLogHandler(),
		AuroraInfoHandler:    NewAuroraInfoHandler(),
		WebsiteConfigHandler: NewWebsiteConfigHandler(),
		FileHandler:          NewFileHandler(),
		ResourceHandler:      NewResourceHandler(),
		AboutHandler:         NewAboutHandler(),
	}
}

// RegisterRoutes 注册所有路由（在main.go中调用）
// 对标 Java 的 @RequestMapping + @GetMapping/@PostMapping 等注解
func (r *Router) RegisterRoutes(engine *gin.Engine) {
	api := engine.Group("/api")

	// ==================== 公开路由（无需认证）====================
	public := api.Group("")
	r.registerPublicRoutes(public)

	// ==================== 受保护路由（需JWT认证）====================
	protected := api.Group("")
	protected.Use(middleware.JWTAuth())
	// protected.Use(middleware.RBAC()) // P0-6 添加RBAC权限控制

	r.registerProtectedRoutes(protected)

	// ==================== 后台管理路由（需JWT + 管理员角色）====================
	admin := api.Group("/admin")
	admin.Use(middleware.JWTAuth())
	// admin.Use(middleware.RequireRole("admin")) // P0-6 角色过滤

	r.registerAdminRoutes(admin)
}

// registerPublicRoutes 公开API路由（前台访问）
func (r *Router) registerPublicRoutes(rg *gin.RouterGroup) {
	// --- 文章 ---
	rg.GET("/articles", r.ArticleHandler.ListArticles)
	rg.GET("/articles/search", r.ArticleHandler.SearchArticles)
	rg.GET("/articles/topAndFeatured", r.ArticleHandler.TopAndFeaturedArticles)
	rg.GET("/articles/archives", r.ArticleHandler.GetArchives)
	rg.GET("/articles/:id", r.ArticleHandler.GetArticleById)
	rg.GET("/articles/:id/export", r.ArticleHandler.ExportArticle)

	// --- 认证 ---
	auth := rg.Group("/auth")
	{
		auth.POST("/register", r.UserAuthHandler.Register)
		auth.POST("/login", r.UserAuthHandler.Login)
		auth.POST("/code", r.UserAuthHandler.SendVerificationCode)
		auth.POST("/password/reset", r.UserAuthHandler.ResetPassword)
		auth.POST("/qq/callback", r.UserAuthHandler.QQLogin)
	}
	rg.POST("/auth/logout", r.UserAuthHandler.Logout)

	// --- 评论 ---
	rg.GET("/articles/:articleId/comments", r.CommentHandler.ListComments)
	rg.POST("/articles/:articleId/comments", r.CommentHandler.AddComment)
	rg.POST("/comments/:id/reply", r.CommentHandler.ReplyComment)
	rg.POST("/comments/:id/like", r.CommentHandler.LikeComment)

	// --- 分类标签 ---
	rg.GET("/categories", r.CategoryHandler.ListCategories)
	rg.GET("/categories/:id", r.CategoryHandler.GetCategoryById)
	rg.GET("/tags", r.TagHandler.ListTags)
	rg.GET("/tags/search", r.TagHandler.SearchTags)
	rg.GET("/tags/:id", r.TagHandler.GetTagById)
	rg.GET("/articles/:articleId/tags", r.TagHandler.ListTagsByArticleId)

	// --- 友链 ---
	rg.GET("/links", r.FriendLinkHandler.ListFriendLinks)
	rg.POST("/links", r.FriendLinkHandler.SaveFriendLink)

	// --- 说说 ---
	rg.GET("/talks", r.TalkHandler.ListTalks)
	rg.GET("/talks/:id", r.TalkHandler.GetTalkById)
	rg.POST("/talks/:id/like", r.TalkHandler.LikeTalk)
	rg.POST("/talks/:id/comments", r.TalkHandler.AddTalkComment)

	// --- 相册 ---
	rg.GET("/albums", r.PhotoAlbumHandler.ListAlbums)
	rg.GET("/albums/:id", r.PhotoAlbumHandler.GetAlbumById)
	rg.GET("/albums/:albumId/photos", r.PhotoHandler.ListPhotos)

	// --- 关于页 ---
	rg.GET("/about", r.AboutHandler.GetAbout)

	// --- 首页信息聚合 ---
	rg.GET("/home/info", r.AuroraInfoHandler.GetHomeInfo)

	// --- 网站配置(前台只读) ---
	rg.GET("/website/config", r.WebsiteConfigHandler.GetWebsiteConfig)

	// --- 文章密码验证 ---
	rg.POST("/articles/:id/password/verify", r.ArticleHandler.VerifyArticlePassword)
}

// registerProtectedRoutes 受保护路由（需JWT登录）
func (r *Router) registerProtectedRoutes(rg *gin.RouterGroup) {
	user := rg.Group("/user")
	{
		user.GET("/info", r.UserAuthHandler.GetUserInfo)
		user.PUT("/password", r.UserAuthHandler.UpdatePassword)
	}
}

// registerAdminRoutes 后台管理路由
func (r *Router) registerAdminRoutes(rg *gin.RouterGroup) {
	// --- 文章管理 ---
	rg.GET("/articles", r.ArticleHandler.ListAdminArticles)
	rg.POST("/articles", r.ArticleHandler.SaveArticle)
	rg.PUT("/articles/:id", r.ArticleHandler.SaveArticle)        // 更新
	rg.DELETE("/articles/:ids", r.ArticleHandler.DeleteArticle)    // 批量删除
	rg.PUT("/articles/:id/status", r.ArticleHandler.UpdateArticleStatus)
	rg.PUT("/articles/:id/password", r.ArticleHandler.UpdateArticlePassword)
	rg.POST("/articles/import", r.ArticleHandler.ImportArticle)

	// --- 用户管理 ---
	rg.PUT("/user/password", r.UserAuthHandler.UpdatePassword)

	// --- 分类管理 ---
	rg.GET("/categories/options", r.CategoryHandler.ListCategoriesOption)
	rg.POST("/categories", r.CategoryHandler.SaveOrUpdate)
	rg.PUT("/categories/:id", r.CategoryHandler.SaveOrUpdate)
	rg.DELETE("/categories/:id", r.CategoryHandler.DeleteCategory)

	// --- 标签管理 ---
	rg.POST("/tags", r.TagHandler.SaveOrUpdate)
	rg.PUT("/tags/:id", r.TagHandler.SaveOrUpdate)
	rg.DELETE("/tags", r.TagHandler.DeleteTags)
	rg.PUT("/tags/count/sync", r.TagHandler.UpdateTagArticleCount)

	// --- 友链管理 ---
	rg.GET("/links", r.FriendLinkHandler.ListAdminFriendLinks)
	rg.PUT("/links/:id", r.FriendLinkHandler.UpdateFriendLink)
	rg.DELETE("/links/:id", r.FriendLinkHandler.DeleteFriendLink)
	rg.PUT("/links/:id/review", r.FriendLinkHandler.ReviewFriendLink)
	rg.PUT("/links/:id/toggle", r.FriendLinkHandler.ToggleOnline)

	// --- 说说管理 ---
	rg.POST("/talks", r.TalkHandler.SaveOrUpdate)
	rg.PUT("/talks/:id", r.TalkHandler.SaveOrUpdate)
	rg.DELETE("/talks/:id", r.TalkHandler.DeleteTalk)

	// --- 相册管理 ---
	rg.POST("/albums", r.PhotoAlbumHandler.SaveOrUpdate)
	rg.PUT("/albums/:id", r.PhotoAlbumHandler.SaveOrUpdate)
	rg.DELETE("/albums/:id", r.PhotoAlbumHandler.DeleteAlbum)
	rg.POST("/albums/:albumId/photos", r.PhotoHandler.UploadPhoto)
	rg.DELETE("/photos/:id", r.PhotoHandler.DeletePhoto)

	// --- 评论管理 ---
	rg.GET("/comments", r.CommentHandler.ListAdminComments)
	rg.PUT("/comments/:id/review", r.CommentHandler.UpdateCommentReview)
	rg.DELETE("/comments/:ids", r.CommentHandler.DeleteComment)
	rg.GET("/comments/stats", r.CommentHandler.GetCommentStats)

	// --- 角色管理 ---
	rg.GET("/roles", r.RoleHandler.ListRoles)
	rg.POST("/roles", r.RoleHandler.SaveOrUpdate)
	rg.PUT("/roles/:id", r.RoleHandler.SaveOrUpdate)
	rg.DELETE("/roles", r.RoleHandler.DeleteRoles)
	rg.GET("/roles/:id", r.RoleHandler.GetRoleById)
	rg.PUT("/roles/:id/menus", r.RoleHandler.UpdateRoleMenu)

	// --- 菜单管理 ---
	rg.GET("/menus", r.MenuHandler.ListMenus)
	rg.GET("/user/menus", r.MenuHandler.GetUserMenus)
	rg.POST("/menus", r.MenuHandler.SaveOrUpdate)
	rg.PUT("/menus/:id", r.MenuHandler.SaveOrUpdate)
	rg.DELETE("/menus/:id", r.MenuHandler.DeleteMenu)

	// --- 定时任务管理 ---
	rg.GET("/jobs", r.JobHandler.ListJobs)
	rg.POST("/jobs", r.JobHandler.SaveOrUpdate)
	rg.PUT("/jobs/:id", r.JobHandler.SaveOrUpdate)
	rg.DELETE("/jobs/:id", r.JobHandler.DeleteJob)
	rg.PUT("/jobs/:id/status", r.JobHandler.UpdateJobStatus)
	rg.POST("/jobs/:id/run", r.JobHandler.RunJobOnce)

	// --- 日志管理 ---
	rg.GET("/jobLogs", r.JobLogHandler.ListJobLogs)
	rg.DELETE("/jobLogs", r.JobLogHandler.DeleteJobLogs)
	rg.GET("/operationLogs", r.OperationLogHandler.ListOperationLogs)
	rg.DELETE("/operationLogs", r.OperationLogHandler.DeleteOperationLogs)
	rg.GET("/exceptionLogs", r.ExceptionLogHandler.ListExceptionLogs)
	rg.DELETE("/exceptionLogs", r.ExceptionLogHandler.DeleteExceptionLogs)

	// --- 网站配置 ---
	rg.GET("/info", r.AuroraInfoHandler.GetAdminInfo)
	rg.PUT("/website/config", r.WebsiteConfigHandler.UpdateWebsiteConfig)
	rg.POST("/website/config/images", r.WebsiteConfigHandler.UploadConfigImage)
	rg.POST("/about", r.AboutHandler.SaveOrUpdate)
	rg.PUT("/about/:id", r.AboutHandler.SaveOrUpdate)

	// --- 文件上传 ---
	rg.POST("/upload", r.FileHandler.UploadFile)
	rg.POST("/upload/batch", r.FileHandler.BatchUpload)
	rg.POST("/upload/image", r.FileHandler.UploadImage)

	// --- 资源权限管理 ---
	rg.GET("/resources", r.ResourceHandler.ListResources)
	rg.POST("/resources", r.ResourceHandler.SaveOrUpdate)
	rg.PUT("/resources/:id", r.ResourceHandler.SaveOrUpdate)
	rg.DELETE("/resources", r.ResourceHandler.DeleteResources)
	rg.PUT("/roles/:id/resources", r.ResourceHandler.UpdateRoleResource)
}
