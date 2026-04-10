package enum

// CommentType 评论类型枚举
type CommentType int8

const (
	CommentTypeArticle  CommentType = 1 // 文章评论
	CommentTypeTalk     CommentType = 2 // 说说评论
	CommentTypeLink     CommentType = 3 // 友链评论
	CommentTypeAbout    CommentType = 4 // 关于评论
	CommentTypeMessage  CommentType = 5 // 留言板
)

// LoginType 登录方式枚举
type LoginType int8

const (
	LoginTypeEmail   LoginType = 1 // 邮箱登录
	LoginTypeQQ      LoginType = 2 // QQ登录
	LoginTypeMobile  LoginType = 3 // 手机号
)

// SearchMode 搜索模式枚举
type SearchMode string

const (
	SearchModeES    SearchMode = "elasticsearch" // ES全文搜索
	SearchModeMySQL SearchMode = "mysql"          // MySQL模糊搜索
)

// UploadMode 上传模式枚举
type UploadMode string

const (
	UploadModeMinio UploadMode = "minio"
	UploadModeOSS   UploadMode = "oss"
)
