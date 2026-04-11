package vo

// ===== 用户相关 VO =====

type RegisterVO struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email,max=50"`
	Password string `json:"password" binding:"required,min=6,max=100"`
	Nickname string `json:"nickname" binding:"required,min=1,max=30"`
}

type LoginVO struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
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
	Alias        string `json:"alias,omitempty" binding:"omitempty,max=50"`
	Description  string `json:"description,omitempty" binding:"omitempty,max=500"`
	Sort         int    `json:"sort"`
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

// ===== 说说 VO =====

type TalkVO struct {
	Content string `json:"content" binding:"required"`
	Status  int8   `json:"status" binding:"omitempty,oneof=1 2"`
}

// ===== 相册 VO =====

type PhotoAlbumVO struct {
	AlbumName  string `json:"albumName" binding:"required,max=50"`
	AlbumCover string `json:"albumCover,omitempty" binding:"omitempty,max=1024"`
	Info       string `json:"info,omitempty" binding:"omitempty,max=500"`
	Status     int8   `json:"status" binding:"omitempty,oneof=1 2"` // 1公开 2私密
}

// ===== 角色 VO =====

type RoleVO struct {
	RoleName    string `json:"roleName" binding:"required,max=50"`
	RoleLabel   string `json:"roleLabel" binding:"required,max=50"`
	Description string `json:"description,omitempty" binding:"omitempty,max=200"`
	Sort        int    `json:"sort"`
	MenuIDs     []uint `json:"menuIds"`
}

// ===== 菜单 VO =====

type MenuVO struct {
	Name       string `json:"name" binding:"required,max=50"`
	Path       string `json:"path" binding:"required,max=255"`
	Component  string `json:"component,omitempty" binding:"omitempty,max=255"`
	Icon       string `json:"icon,omitempty" binding:"omitempty,max=100"`
	Sort       int    `json:"sort"`
	Type       int8   `json:"type" binding:"required,oneof=0 1 2"` // 0目录 1菜单 2按钮
	Permission string `json:"permission,omitempty" binding:"omitempty,max=100"`
	ParentID   uint   `json:"parentId,omitempty"`
	Hidden     *int8  `json:"hidden,omitempty"`
	OrderNum   *int   `json:"orderNum,omitempty"`
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
	ResourceName  string `json:"resourceName" binding:"required,max=50"`
	URL           string `json:"url" binding:"required,url,max=255"`
	RequestMethod string `json:"requestMethod" binding:"required,oneof=GET POST PUT DELETE PATCH"`
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
