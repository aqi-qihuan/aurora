package dto

import (
	"time"

	"github.com/aurora-go/aurora/internal/model"
)

// ===== 用户相关 DTO =====

type UserDTO struct {
	ID        uint      `json:"id"`
	Email     string    `json:"email"`
	Nickname  string    `json:"nickname"`
	Avatar    string    `json:"avatar"`
	Intro     string    `json:"intro"`
	WebSite   string    `json:"webSite"`
	IsDisable int8      `json:"isDisable"`
	CreateTime time.Time `json:"createTime,omitempty"`
}

type UserAdminDTO struct {
	ID            uint     `json:"id"`
	Email         string   `json:"email"`
	Nickname      string   `json:"nickname"`
	Avatar        string   `json:"avatar"`
	IsDisable     int8     `json:"isDisable"`
	Roles         []string `json:"roles"`
	ArticleCount  int      `json:"articleCount"`
	CategoryCount int      `json:"categoryCount"`
	TagCount      int      `json:"tagCount"`
	CreateTime    time.Time `json:"createTime"`
}

type LoginVO struct {
	UserID    uint   `json:"userId"`
	Token     string `json:"token"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Email     string `json:"email"`
	IsDisable int8   `json:"isDisable"`
}

// ===== 分类 DTO =====

type CategoryDTO struct {
	ID            uint      `json:"id"`
	CategoryName  string    `json:"categoryName"`
	Alias         string    `json:"alias"`
	Description    string    `json:"description"`
	ArticleCount  int       `json:"articleCount"`
	Sort          int       `json:"sort"`
	CreateTime     time.Time `json:"createTime,omitempty"`
}

type OptionDTO struct {
	Label string `json:"label"`
	Value uint   `json:"value"`
}

// ===== 标签 DTO =====

type TagDetailDTO struct {
	ID           uint             `json:"id"`
	TagName       string           `json:"tagName"`
	ArticleCount int              `json:"articleCount"`
	Articles     []ArticleCardDTO `json:"articles"`
}

// ===== 友链 DTO =====

type FriendLinkDTO struct {
	ID          uint      `json:"id"`
	Nickname    string    `json:"nickname,omitempty"` // 申请者昵称
	LinkName    string    `json:"linkName"`
	LinkAvatar  string    `json:"linkAvatar"`
	LinkAddress string    `json:"linkAddress"`
	LinkIntro   string    `json:"linkIntro"`
	CreateTime   time.Time `json:"createTime,omitempty"`
}

type FriendLinkAdminDTO struct {
	ID          uint      `json:"id"`
	UserID      *uint     `json:"userId,omitempty"`
	Nickname    string    `json:"nickname,omitempty"`
	LinkName    string    `json:"linkName"`
	LinkAvatar  string    `json:"linkAvatar"`
	LinkAddress string    `json:"linkAddress"`
	LinkIntro   string    `json:"linkIntro"`
	Status      int8      `json:"status"`
	CreateTime   time.Time `json:"createTime"`
}

// ===== 说说 DTO =====

type TalkDTO struct {
	ID         uint      `json:"id"`
	UserID     uint      `json:"userId"`
	Nickname   string    `json:"nickname,omitempty"`
	Avatar     string    `json:"avatar,omitempty"`
	Content    string    `json:"content"`
	IsTop      int8      `json:"isTop"`
	Status     int8      `json:"status"`
	LikeCount  int64     `json:"likeCount"`
	CommentCount int     `json:"commentCount"`
	CreateTime time.Time `json:"createTime"`
}

type TalkDetailDTO struct {
	TalkDTO
}

// ===== 相册/照片 DTO =====

type PhotoDTO struct {
	ID  uint   `json:"id"`
	URL string `json:"url"`
}

type AlbumDTO struct {
	ID          uint      `json:"id"`
	AlbumName   string    `json:"albumName"`
	AlbumCover  string    `json:"albumCover"`
	Info        string    `json:"info"`
	PhotoCount  int       `json:"photoCount"`
	IsPrivate   bool      `json:"isPrivate"`
	CreateTime   time.Time `json:"createTime"`
}

// ===== 角色/菜单 DTO =====

type RoleDTO struct {
	ID          uint   `json:"id"`
	RoleName    string `json:"roleName"`
	RoleLabel   string `json:"roleLabel"`
	Description  string `json:"description"`
	IsDisable   int8   `json:"isDisable"`
	IsDefault   int8   `json:"isDefault"`
	MenuIDs     []uint `json:"menuIds"`
	CreateTime   time.Time `json:"createTime"`
}

type RoleDetailDTO struct {
	ID          uint     `json:"id"`
	RoleName    string   `json:"roleName"`
	RoleLabel   string   `json:"roleLabel"`
	Description  string   `json:"description"`
	IsDisable   int8     `json:"isDisable"`
	IsDefault   int8     `json:"isDefault"`
	MenuIDs     []uint   `json:"menuIds"`
	Menus       []model.Menu `json:"menus,omitempty"` // 直接返回Menu用于前端渲染
}

type MenuDTO struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Component   string    `json:"component"`
	Icon        string    `json:"icon"`
	Sort        int       `json:"sort"`
	Type        int8      `json:"type"`
	Permission  string    `json:"permission"`
	Hidden      int8      `json:"hidden"`
	OrderNum    *int      `json:"orderNum"`
	ParentID    *uint     `json:"parentId"`
	CreateTime   time.Time `json:"createTime"`
}

type MenuTreeDTO struct {
	ID          uint          `json:"id"`
	Name        string        `json:"name"`
	Path        string        `json:"path"`
	Component   string        `json:"component"`
	Icon        string        `json:"icon"`
	Type        int8          `json:"type"`
	Permission  string        `json:"permission"`
	Hidden      int8          `json:"hidden"`
	OrderNum    *int          `json:"orderNum"`
	Sort        int           `json:"sort"`
	ParentID    *uint         `json:"parentId"`
	Children    []MenuTreeDTO `json:"children"`
}

// ===== 定时任务 DTO =====

type JobDTO struct {
	ID              uint       `json:"id"`
	JobName         string     `json:"jobName"`
	JobGroup        string     `json:"jobGroup,omitempty"`
	InvokeTarget    string     `json:"invokeTarget"`
	CronExpression  string     `json:"cronExpression,omitempty"`
	Status          int8       `json:"status"`
	Duration        *int64     `json:"duration,omitempty"`
	ErrorMsg        string     `json:"errorMsg,omitempty"`
	LastRunTime     *time.Time `json:"lastRunTime,omitempty"`
	CreateTime      time.Time  `json:"createTime"`
}

type JobLogDTO struct {
	ID        uint      `json:"id"`
	JobID     uint      `json:"jobId"`
	JobName   string    `json:"jobName"`
	JobGroup  string    `json:"jobGroup,omitempty"`
	Status    int8      `json:"status"`
	Duration  *int64    `json:"duration,omitempty"`
	ErrorMsg  string    `json:"errorMsg,omitempty"`
	CreateTime time.Time `json:"createTime"`
}

type JobDetailDTO struct {
	ID              uint       `json:"id"`
	JobName         string     `json:"jobName"`
	JobGroup        string     `json:"jobGroup,omitempty"`
	InvokeTarget    string     `json:"invokeTarget"`
	CronExpression  string     `json:"cronExpression"`
	Status          int8       `json:"status"`
	LastRunTime     *time.Time `json:"lastRunTime"`
	CreateTime      time.Time  `json:"createTime"`
}

// ===== 日志 DTO =====

type OperationLogVO struct {
	UserID    uint   `json:"userId"`
	Module    string `json:"module"`
	Operation string `json:"operation"`
	Method    string `json:"method"`
	URL       string `json:"url"`
	IP        string `json:"ip"`
	Duration  *int64 `json:"duration"`
	Status    int8   `json:"status"`
	ErrorMsg  string `json:"errorMsg"`
}

type OperationLogDTO struct {
	ID         uint      `json:"id"`
	UserID     uint      `json:"userId"`
	Nickname   string    `json:"nickname,omitempty"`
	Module     string    `json:"module"`
	Operation  string    `json:"operation"`
	Method     string    `json:"method"`
	URL        string    `json:"url"`
	IP         string    `json:"ip"`
	Duration   *int64    `json:"duration"`
	Status     int8      `json:"status"`
	ErrorMsg   string    `json:"errorMsg"`
	CreateTIme time.Time `json:"createTime"`
}

type ExceptionLogVO struct {
	UserID     uint   `json:"userId"`
	URL        string `json:"url"`
	Method     string `json:"method"`
	IP         string `json:"ip"`
	ErrorMsg   string `json:"errorMsg"`
	Stacktrace string `json:"stacktrace"`
}

type ExceptionLogDTO struct {
	ID         uint      `json:"id"`
	UserID     uint      `json:"userID"`
	URL        string    `json:"url"`
	Method     string    `json:"method"`
	IP         string    `json:"ip"`
	ErrorMsg   string    `json:"errorMsg"`
	Stacktrace string    `json:"stacktrace"`
	Status     int8      `json:"status"`
	CreateTime  time.Time `json:"createTime"`
}

// ===== 首页聚合 DTO =====

type HomeInfoDTO struct {
	TopArticles    []ArticleCardDTO `json:"topArticles"`
	LatestArticles []ArticleCardDTO `json:"latestArticles"`
	Categories     []CategoryDTO    `json:"categories"`
	Tags           []TagDTO         `json:"tags"`
	FriendLinks    []FriendLinkDTO  `json:"friendLinks"`
	Talks          []TalkDTO        `json:"talks"`
}

type AdminDashboardDTO struct {
	TotalArticles   int64 `json:"totalArticles"`
	TotalUsers      int64 `json:"totalUsers"`
	TotalComments   int64 `json:"totalComments"`
	TotalViews      int64 `json:"totalViews"`
	PendingReviews  int64 `json:"pendingReviews"`
	TodayArticles   int64 `json:"todayArticles"`
	UniqueVisitors  int64 `json:"uniqueVisitors"`
}

// ===== 网站配置 DTO =====

type WebsiteConfigDTO struct {
	ID                     uint  `json:"id,omitempty"`
	SiteName               string `json:"siteName"`
	SiteURL                string `json:"siteUrl"`
	AuthorName             string `json:"authorName"`
	AuthorAvatar           string `json:"authorAvatar"`
	Logo                   string `json:"logo"`
	Favicon                string `json:"favicon"`
	SiteIntro              string `json:"siteIntro"`
	Notice                 string `json:"notice"`
	FooterInfo             string `json:"footerInfo"`
	IcpNumber              string `json:"icpNumber"`
	BaiduPushURL           string `json:"baiduPushUrl"`
	GAID                   string `json:"gaId"`
	WechatQRCode           string `json:"wechatQrcode"`
	AlipayQRCode           string `json:"alipayQrcode"`
	CommentNotifyEnabled   bool   `json:"commentNotifyEnabled"`
	RegisterEnabled        bool   `json:"registerEnabled"`
	RewardEnabled           bool   `json:"rewardEnabled"`
}

// ===== 文件上传 DTO =====

type FileUploadDTO struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
}

// ===== 资源权限 DTO =====

type ResourceDTO struct {
	ID            uint      `json:"id"`
	ResourceName  string    `json:"resourceName"`
	URL           string    `json:"url"`
	RequestMethod string    `json:"requestMethod"`
	CreateTime     time.Time `json:"createTime"`
}

// ===== 评论 DTO (补充完整) =====

type CommentDTO struct {
	ID            uint      `json:"id"`
	UserID        uint      `json:"userId"`
	Nickname      string    `json:"nickname,omitempty"`
	Avatar        string    `json:"avatar,omitempty"`
	Content       string    `json:"content"`
	Type          int8      `json:"type"`
	ParentID      uint      `json:"parentId"`
	ReplyNickname string    `json:"replyNickname,omitempty"`
	LikeCount     int64     `json:"likeCount"`
	IsReview      int8      `json:"isReview"`
	Location      string    `json:"location,omitempty"`
	CreateTime    time.Time `json:"createTime"`
}

type CommentTreeDTO struct {
	CommentDTO
	Replies []CommentTreeDTO `json:"replies"`
}

type CommentAdminDTO struct {
	ID         uint      `json:"id"`
	UserID     uint      `json:"userId"`
	Nickname   string    `json:"nickname,omitempty"`
	Content    string    `json:"content"`
	Type       int8      `json:"type"`
	ParentID   uint      `json:"parentId"`
	IsReview   int8      `json:"isReview"`
	IP         string    `json:"ip"`
	Location   string    `json:"location"`
	LikeCount  int64     `json:"likeCount"`
	CreateTime time.Time `json:"createTime"`
}

// ===== RBAC权限 DTO =====

// ResourceRoleDTO 资源-角色映射DTO (对标Java ResourceRoleDTO)
// 用于RBAC权限控制: 每个API路径+方法 → 允许的角色列表
type ResourceRoleDTO struct {
	ID            uint     `json:"id"`
	URL           string   `json:"url"`             // API路径 (如 /api/admin/articles)
	RequestMethod string   `json:"requestMethod"`    // HTTP方法 (GET/POST/PUT/DELETE)
	RoleList      []string `json:"roleList"`         // 允许的角色列表
}
