package constant

// RedisKey Redis Key 常量定义 (对齐Java版 RedisConstant)
const (
	// ===== 文章相关 =====
	ArticleViewsRanking  = "article_views_count"    // 文章浏览排行 (ZSet, 对齐Java版 ARTICLE_VIEWS_COUNT)
	ArticleAccess        = "article_access:"         // 密码文章访问权限 (Set)

	// ===== 博客统计 =====
	BlogViewsCount       = "blog_views_count"         // 总浏览量 (String, 对标Java BLOG_VIEWS_COUNT)
	UniqueVisitor        = "unique_visitor"           // 每日独立访客 (Set, 对标Java UNIQUE_VISITOR)
	VisitorArea          = "visitor_area"             // 访客地域分布 (Hash, 对标Java VISITOR_AREA)
	UserArea             = "user_area"                // 用户地域分布 (String JSON, 对标Java USER_AREA)

	// ===== 用户会话 =====
	LoginUser            = "login:user:"                // 登录用户Session (Hash)
	UserAuthCode         = "user:auth_code:"            // 验证码 (String, TTL=5min)
	RegisterCodePrefix   = "register:code:"             // 注册验证码

	// ===== 搜索缓存 =====
	SearchResultCache    = "search:result:"              // 搜索结果缓存 (Hash, TTL=10min)
	ArticleArchiveCache  = "article:archive:"             // 文章归档缓存 (Hash, TTL=1h)

	// ===== 站点数据 =====
	SiteViewCount        = "site_view_count"          // 站点总访问量
	TodayViewCount       = "today_view_count"         // 今日访问量
	UniqueViewToday      = "unique_view_today"        // 今日独立访客数
)

// RabbitMQ 常量
const (
	MQExchangeMaxwell = "maxwell.exchange"
	MQQueueMaxwell     = "maxwell.queue"
	MQRouteKeyMaxwell  = "maxwell.route"

	MQExchangeNotice  = "notice.exchange"
	MQQueueEmail      = "notice.email.queue"
	MQRouteKeyEmail   = "notice.email"

	MQQueueSubscribe  = "notice.subscribe.queue"
	MQRouteKeySubscribe = "notice.subscribe"
)

// 认证常量 (补充 auth_const.go 未定义的)
const (
	JwtClaimRole = "role"
	JwtIssuer    = "aurora-go"
)
