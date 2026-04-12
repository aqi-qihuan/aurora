package dto

// ===== Handler层请求绑定用VO (从vo包复用或定义新类型) =====
// handler中的 ShouldBindJSON/ShouldBindQuery 使用这些类型

// ===== 文章相关 =====

// ArticleVO 文章发布/编辑请求（对齐Java版ArticleVO）
type ArticleVO struct {
	ID              uint     `json:"id,omitempty"`
	ArticleTitle    string   `json:"articleTitle" binding:"required,max=200"`
	ArticleContent  string   `json:"articleContent" binding:"required"`
	ArticleAbstract string   `json:"articleAbstract,omitempty"`
	ArticleCover    string   `json:"articleCover,omitempty"`
	CategoryName    string   `json:"categoryName,omitempty"` // 分类名称（对齐Java）
	TagNames        []string `json:"tagNames,omitempty"`     // 标签名称列表（对齐Java）
	Type            int8     `json:"type" binding:"omitempty,oneof=1 2 3"`
	OriginalURL     string   `json:"originalUrl,omitempty"`
	IsTop           int8     `json:"isTop" binding:"omitempty,oneof=0 1"`
	IsFeatured      int8     `json:"isFeatured" binding:"omitempty,oneof=0 1"`
	Status          *int8    `json:"status" binding:"omitempty,oneof=0 1 2 3"`
	Password        string   `json:"password,omitempty"`
}

// ArticleStatusUpdateVO 文章状态更新
type ArticleStatusUpdateVO struct {
	ID         uint `json:"id" binding:"required"`
	IsTop      *int8 `json:"isTop" binding:"omitempty,oneof=0 1"`
	IsFeatured *int8 `json:"isFeatured" binding:"omitempty,oneof=0 1"`
	Status     *int8 `json:"status" binding:"omitempty,oneof=0 1 2 3"`
}

// ArticlePasswordVO 文章密码设置
type ArticlePasswordVO struct {
	Password string `json:"password"`
}

// ArticlePasswordVerifyVO 文章密码验证
type ArticlePasswordVerifyVO struct {
	Password string `json:"password" binding:"required"`
}

// ===== 用户相关 =====

// RegisterVO 用户注册
type RegisterVO struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email,max=50"`
	Password string `json:"password" binding:"required,min=6,max=100"`
	Nickname string `json:"nickname" binding:"required,min=1,max=30"`
}

// 注意: LoginVO 已在 extra_dto.go 中定义
// 注意: QQLoginVO 已在 auth_dto.go 中定义

// EmailVO 邮箱绑定/修改
type EmailVO struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required"`
}

// PasswordVO 密码修改
type PasswordVO struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6,max=100"`
}

// ResetPasswordVO 密码重置
type ResetPasswordVO struct {
	Email       string `json:"email" binding:"required,email"`
	Code        string `json:"code" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6,max=100"`
}

// ===== 分类/标签 =====

// CategoryVO 分类请求
type CategoryVO struct {
	CategoryName string `json:"categoryName" binding:"required,max=20"`
	Alias        string `json:"alias,omitempty"`
	Description  string `json:"description,omitempty"`
	Sort         int    `json:"sort"`
}

// TagVO 标签请求
type TagVO struct {
	TagName string `json:"tagName" binding:"required,max=20"`
}

// ===== 友链 =====

// FriendLinkVO 友链请求
type FriendLinkVO struct {
	LinkName    string `json:"linkName" binding:"required,max=50"`
	LinkAvatar  string `json:"linkAvatar,omitempty"`
	LinkAddress string `json:"linkAddress" binding:"required,url,max=500"`
	LinkIntro   string `json:"linkIntro,omitempty"`
}

// ===== 评论 =====

// CommentVO 评论请求
type CommentVO struct {
	ArticleID    uint   `json:"articleId,omitempty"`
	TalkID       uint   `json:"talkId,omitempty"`
	FriendLinkID uint   `json:"friendLinkId,omitempty"`
	AboutID      uint   `json:"aboutId,omitempty"`
	Type         int8   `json:"type" binding:"required,oneof=1 2 3 4 5"`
	ParentID     uint   `json:"parentId"`
	ReplyUserID  *uint  `json:"replyUserId,omitempty"`
	Content      string `json:"content" binding:"required,max=2000"`
}

// ReplyVO 回复请求
type ReplyVO struct {
	ParentID    uint   `json:"parentId" binding:"required"`
	Content     string `json:"content" binding:"required,max=2000"`
	ReplyUserID *uint  `json:"replyUserId,omitempty"`
}

// ReviewVO 审核请求 (通用)
type ReviewVO struct {
	ID         uint `json:"id" binding:"required"`
	IsApproved bool `json:"isApproved" binding:"required"`
}

// ===== 说说 =====

// TalkVO 说说请求
type TalkVO struct {
	Content string `json:"content" binding:"required"`
	Status  int8   `json:"status" binding:"omitempty,oneof=1 2"`
}

// ===== 相册 =====

// PhotoAlbumVO 相册请求
type PhotoAlbumVO struct {
	AlbumName  string `json:"albumName" binding:"required,max=50"`
	AlbumCover string `json:"albumCover,omitempty"`
	Info       string `json:"info,omitempty"`
	Status     int8   `json:"status" binding:"omitempty,oneof=1 2"`
}

// ===== 角色/菜单 =====

// RoleVO 角色请求
type RoleVO struct {
	RoleName    string `json:"roleName" binding:"required,max=50"`
	RoleLabel   string `json:"roleLabel" binding:"required,max=50"`
	Description string `json:"description,omitempty"`
	Sort        int    `json:"sort"`
	MenuIDs     []uint `json:"menuIds"`
}

// MenuVO 菜单请求
type MenuVO struct {
	Name       string `json:"name" binding:"required,max=50"`
	Path       string `json:"path" binding:"required,max=255"`
	Component  string `json:"component,omitempty"`
	Icon       string `json:"icon,omitempty"`
	Sort       int    `json:"sort"`
	Type       int8   `json:"type" binding:"required,oneof=0 1 2"`
	Permission string `json:"permission,omitempty"`
	ParentID   uint   `json:"parentId,omitempty"`
	Hidden     *int8  `json:"hidden,omitempty"`
	OrderNum   *int   `json:"orderNum,omitempty"`
}

// MenuIdsVO 菜单ID列表
type MenuIdsVO struct {
	MenuIDs []uint `json:"menuIds" binding:"required,min=1"`
}

// ===== 定时任务 =====

// JobVO 定时任务请求
type JobVO struct {
	JobName        string `json:"jobName" binding:"required,max=50"`
	JobGroup       string `json:"jobGroup" binding:"required,max=50"`
	InvokeTarget   string `json:"invokeTarget" binding:"required,max=200"`
	CronExpression string `json:"cronExpression" binding:"required,max=100"`
}

// StatusUpdateVO 状态更新请求 (通用)
type StatusUpdateVO struct {
	ID     uint `json:"id" binding:"required"`
	Status int8 `json:"status" binding:"required"`
}

// ===== 资源 =====

// ResourceVO 资源请求
type ResourceVO struct {
	ResourceName  string `json:"resourceName" binding:"required,max=50"`
	URL           string `json:"url" binding:"required,url,max=255"`
	RequestMethod string `json:"requestMethod" binding:"required,oneof=GET POST PUT DELETE PATCH"`
}

// ResourceIdsVO 资源ID列表
type ResourceIdsVO struct {
	ResourceIDs []uint `json:"resourceIds" binding:"required,min=1"`
}

// ===== 网站配置 =====

// WebsiteConfigVO 网站配置请求
type WebsiteConfigVO struct {
	SiteName             string `json:"siteName,omitempty"`
	SiteURL              string `json:"siteUrl,omitempty"`
	AuthorName           string `json:"authorName,omitempty"`
	AuthorAvatar         string `json:"authorAvatar,omitempty"`
	Logo                 string `json:"logo,omitempty"`
	Favicon              string `json:"favicon,omitempty"`
	SiteIntro            string `json:"siteIntro,omitempty"`
	Notice               string `json:"notice,omitempty"`
	FooterInfo           string `json:"footerInfo,omitempty"`
	IcpNumber            string `json:"icpNumber,omitempty"`
	BaiduPushURL         string `json:"baiduPushUrl,omitempty"`
	GAID                 string `json:"gaId,omitempty"`
	WechatQRCode         string `json:"wechatQrcode,omitempty"`
	AlipayQRCode         string `json:"alipayQrcode,omitempty"`
	CommentNotifyEnabled bool   `json:"commentNotifyEnabled"`
	RegisterEnabled      bool   `json:"registerEnabled"`
	RewardEnabled        bool   `json:"rewardEnabled"`
}
