package vo

// ===== 用户相关 VO =====

type RegisterVO struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email,max=50"`
	Password string `json:"password" binding:"required,min=6,max=100"`
	Nickname string `json:"nickname" binding:"required,min=1,max=30"`
}

// LoginVO 登录请求VO (对标Java LoginVO)
type LoginVO struct {
	Username string `json:"username" form:"username" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}

type UpdateUserVO struct {
	Nickname *string `json:"nickname,omitempty" binding:"omitempty,max=30"`
	Intro    *string `json:"intro,omitempty" binding:"omitempty,max=500"`
	WebSite  *string `json:"webSite,omitempty" binding:"omitempty,max=255"`
	Avatar   *string `json:"avatar,omitempty" binding:"omitempty,max=1024"`
}

type PasswordVO struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6,max=100"`
}

type QQLoginVO struct {
	OpenID   string `json:"openId" binding:"required"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// ===== 分类 VO =====

type CategoryVO struct {
	CategoryName string `json:"categoryName" binding:"required,max=20"`
}

// ===== 标签 VO =====

type TagVO struct {
	TagName string `json:"tagName" binding:"required,max=20"`
}

// ===== 友链 VO =====

type FriendLinkVO struct {
	LinkName    string `json:"linkName" binding:"required,max=50"`
	LinkAvatar  string `json:"linkAvatar" binding:"omitempty,max=1024"`
	LinkAddress string `json:"linkAddress" binding:"required,url,max=500"`
	LinkIntro   string `json:"linkIntro" binding:"omitempty,max=500"`
}

// ===== 相册 VO =====

type PhotoAlbumVO struct {
	ID         uint   `json:"id"`                                  // 更新时必填，创建时为空
	AlbumName  string `json:"albumName" binding:"required,max=20"`
	AlbumCover string `json:"albumCover,omitempty" binding:"omitempty,max=1024"`
	Info       string `json:"albumDesc,omitempty" binding:"omitempty,max=50"`  // Java用albumDesc，Go前端传albumDesc
	Status     int8   `json:"status" binding:"omitempty,oneof=1 2"` // 1公开 2私密
}

// ===== 角色 VO =====

type RoleVO struct {
	RoleName    string `json:"roleName" binding:"required,max=20"`
	MenuIDs     []uint `json:"menuIds"`
}

// ===== 菜单 VO =====

type MenuVO struct {
	Name       string `json:"name" binding:"required,max=20"`
	Path       string `json:"path" binding:"required,max=50"`
	Component  string `json:"component" binding:"required,max=50"`
	Icon       string `json:"icon" binding:"required,max=50"`
	OrderNum   int    `json:"orderNum" binding:"required"`  // 排序号（数据库对应 order_num 字段）
	ParentID   uint   `json:"parentId,omitempty"`
	IsHidden   *int8  `json:"isHidden,omitempty"`  // 是否隐藏 0否1是（数据库对应 is_hidden 字段）
}

// ===== 评论 VO =====

type CommentVO struct {
	ArticleID    uint   `json:"articleId,omitempty"`    // 文章评论时传
	TalkID       uint   `json:"talkId,omitempty"`       // 说说评论时传
	FriendLinkID uint   `json:"friendLinkId,omitempty"` // 友链评论时传
	AboutID      uint   `json:"aboutId,omitempty"`      // 关于页评论时传
	Type         int8   `json:"type" binding:"required,oneof=1 2 3 4 5"`
	ParentID     uint   `json:"parentId"`               // 回复的父评论ID(0=顶级)
	ReplyUserID  *uint  `json:"replyUserId,omitempty"`   // 被回复用户ID
	Content      string `json:"content" binding:"required,max=2000"`
}

// ===== 定时任务 VO =====

type JobVO struct {
	JobName        string `json:"jobName" binding:"required,max=50"`
	JobGroup       string `json:"jobGroup" binding:"required,max=50"`
	InvokeTarget   string `json:"invokeTarget" binding:"required,max=200"`
	CronExpression string `json:"cronExpression" binding:"required,max=100"`
}

// ===== 资源权限 VO =====

type ResourceVO struct {
	ID            uint   `json:"id"`                                  // 更新时必填，创建时为空
	ResourceName  string `json:"resourceName" binding:"required,max=50"`
	URL           string `json:"url" binding:"omitempty,max=255"`      // 模块不需要URL
	RequestMethod string `json:"requestMethod" binding:"omitempty,oneof=GET POST PUT DELETE PATCH"`
	ParentID      *uint  `json:"parentId"`                            // 父模块ID，模块为null
	IsAnonymous   *int8  `json:"isAnonymous"`                         // 是否匿名 0否1是
}

// ===== 网站配置 VO =====

type WebsiteConfigVO struct {
	SiteName             string `json:"siteName" binding:"omitempty,max=50"`
	SiteURL              string `json:"siteUrl" binding:"omitempty,max=255"`
	AuthorName           string `json:"authorName" binding:"omitempty,max=30"`
	AuthorAvatar         string `json:"authorAvatar" binding:"omitempty,max=1024"`
	Logo                 string `json:"logo" binding:"omitempty,max=1024"`
	Favicon              string `json:"favicon" binding:"omitempty,max=1024"`
	SiteIntro            string `json:"siteIntro" binding:"omitempty,max=500"`
	Notice               string `json:"notice" binding:"omitempty,max=500"`
	FooterInfo           string `json:"footerInfo" binding:"omitempty,max=500"`
	IcpNumber            string `json:"icpNumber" binding:"omitempty,max=50"`
	BaiduPushURL         string `json:"baiduPushUrl" binding:"omitempty,max=255"`
	GAID                 string `json:"gaId" binding:"omitempty,max=100"`
	WechatQRCode         string `json:"wechatQrcode" binding:"omitempty,max=1024"`
	AlipayQRCode         string `json:"alipayQrcode" binding:"omitempty,max=1024"`
	CommentNotifyEnabled bool   `json:"commentNotifyEnabled"`
	RegisterEnabled      bool   `json:"registerEnabled"`
	RewardEnabled        bool   `json:"rewardEnabled"`
}
