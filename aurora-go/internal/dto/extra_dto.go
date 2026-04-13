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
	ID            uint           `json:"id"`
	UserInfoId    uint           `json:"userInfoId"`               // 对标Java: user_info_id
	Avatar        string         `json:"avatar"`
	Nickname      string         `json:"nickname"`
	Roles         []UserRoleDTO  `json:"roles"`                     // 角色列表 (对标Java: List<UserRoleDTO>)
	LoginType     int8           `json:"loginType"`                 // 登录类型 1邮箱 2QQ 3Gitee 4Github 5微博
	IpAddress     string         `json:"ipAddress"`                 // 登录IP地址
	IpSource      string         `json:"ipSource"`                  // IP归属地
	CreateTime    *time.Time     `json:"createTime,omitempty"`      // 注册时间 (可为空)
	LastLoginTime *time.Time     `json:"lastLoginTime,omitempty"`   // 最后登录时间（可为空）
	IsDisable     int8           `json:"isDisable"`                 // 禁用状态 0正常 1禁用
	Status        *int8          `json:"status,omitempty"`          // 用户状态
}

// UserRoleDTO 用户角色DTO（对标Java UserRoleDTO）
type UserRoleDTO struct {
	ID       uint   `json:"id"`
	RoleName string `json:"roleName"`
}

// UserOnlineDTO 在线用户DTO（对标Java UserOnlineDTO）
// 数据源：Redis login_user Hash（存储UserDetailsDTO）
type UserOnlineDTO struct {
	UserInfoId    uint       `json:"userInfoId"`    // 用户信息ID
	Nickname      string     `json:"nickname"`       // 昵称
	Avatar        string     `json:"avatar"`         // 头像
	IpAddress     string     `json:"ipAddress"`      // 登录IP
	IpSource      string     `json:"ipSource"`       // IP归属地
	Browser       string     `json:"browser"`        // 浏览器
	Os            string     `json:"os"`             // 操作系统
	LastLoginTime *time.Time `json:"lastLoginTime"`  // 最后登录时间
}

// LoginVO 登录响应DTO (完全对标Java UserInfoDTO)
// 用于: POST /api/auth/login 成功后的返回
type LoginVO struct {
	ID            uint   `json:"id"`             // 用户认证ID (UserAuth.id, 对标Java UserDetailsDTO.id)
	UserInfoID    uint   `json:"userInfoId"`     // 用户信息ID (UserInfo.id, 对标Java UserDetailsDTO.userInfoId)
	Email         string `json:"email"`
	LoginType     int    `json:"loginType"`      // 登录类型 1邮箱 3QQ (对标Java UserDetailsDTO.loginType)
	Username      string `json:"username"`        // 用户名/登录名
	Nickname      string `json:"nickname"`        // 用户昵称
	Avatar        string `json:"avatar"`          // 头像URL
	Intro         string `json:"intro"`           // 个人简介
	Website       string `json:"website"`         // 个人网站
	IsSubscribe   int8   `json:"isSubscribe"`     // 是否订阅
	IPAddress     string `json:"ipAddress"`       // 登录IP
	IPSource      string `json:"ipSource"`        // IP归属地
	LastLoginTime string `json:"lastLoginTime,omitempty"` // 最后登录时间(ISO8601格式)
	Token         string `json:"token"`           // JWT令牌
}

// ===== 分类 DTO =====

type CategoryDTO struct {
	ID            uint      `json:"id"`
	CategoryName  string    `json:"categoryName"`
	ArticleCount  int       `json:"articleCount"` // t_category 表没有此字段，需动态统计
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
	OrderNum    int       `json:"orderNum"`  // 排序号（数据库对应 order_num 字段）
	IsHidden    int8      `json:"isHidden"`  // 是否隐藏 0否1是（数据库对应 is_hidden 字段）
	ParentID    *uint     `json:"parentId"`
	CreateTime   time.Time `json:"createTime"`
	UpdateTime  *time.Time `json:"updateTime,omitempty"`
}

type MenuTreeDTO struct {
	ID          uint          `json:"id"`
	Name        string        `json:"name"`
	Path        string        `json:"path"`
	Component   string        `json:"component"`
	Icon        string        `json:"icon"`
	IsHidden    int8          `json:"isHidden"`  // 是否隐藏 0否1是（数据库对应 is_hidden 字段）
	OrderNum    int           `json:"orderNum"`  // 排序号（数据库对应 order_num 字段）
	ParentID    *uint         `json:"parentId"`
	Children    []MenuTreeDTO `json:"children"`
	CreateTime  time.Time     `json:"createTime,omitempty"`
	UpdateTime  *time.Time    `json:"updateTime,omitempty"`
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
	ID            uint       `json:"id"`
	JobID         uint       `json:"jobId"`
	JobName       string     `json:"jobName"`
	JobGroup      string     `json:"jobGroup,omitempty"`
	InvokeTarget  string     `json:"invokeTarget"`
	JobMessage    string     `json:"jobMessage,omitempty"`
	Status        int8       `json:"status"`
	ExceptionInfo string     `json:"exceptionInfo,omitempty"`
	StartTime     *time.Time `json:"startTime,omitempty"`
	EndTime       *time.Time `json:"endTime,omitempty"`
	CreateTime    time.Time  `json:"createTime"`
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

// ===== 后台首页聚合 DTO (对标 Java AuroraAdminInfoDTO) =====

type AuroraAdminInfoDTO struct {
	ViewsCount          int                    `json:"viewsCount"`
	MessageCount        int                    `json:"messageCount"`
	UserCount           int                    `json:"userCount"`
	ArticleCount        int                    `json:"articleCount"`
	CategoryDTOs        []CategoryDTO          `json:"categoryDTOs,omitempty"`
	TagDTOs             []TagDTO               `json:"tagDTOs,omitempty"`
	ArticleStatistics   []ArticleStatisticsDTO `json:"articleStatisticsDTOs,omitempty"`
	UniqueViewDTOs      []UniqueViewDTO        `json:"uniqueViewDTOs,omitempty"`
	ArticleRankDTOs     []ArticleRankDTO       `json:"articleRankDTOs,omitempty"`
}

// ArticleStatisticsDTO 文章统计 (对标 Java ArticleStatisticsDTO)
type ArticleStatisticsDTO struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// UniqueViewDTO 独立访客统计 (对标 Java UniqueViewDTO)
type UniqueViewDTO struct {
	Day       string `json:"day"`
	ViewsCount int   `json:"viewsCount"`
}

// ArticleRankDTO 文章排行 (对标 Java ArticleRankDTO)
type ArticleRankDTO struct {
	ArticleTitle string `json:"articleTitle"`
	ViewsCount   int    `json:"viewsCount"`
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
	ID            uint      `json:"id"`
	UserID        uint      `json:"userId"`
	Nickname      string    `json:"nickname,omitempty"`
	Avatar        string    `json:"avatar,omitempty"`          // 评论人头像
	CommentContent string   `json:"commentContent"`            // 评论内容（对齐前端期望）
	Type          int8      `json:"type"`
	TopicID       *uint     `json:"topicId,omitempty"`
	ArticleTitle  string    `json:"articleTitle,omitempty"`    // 文章/说说标题
	ReplyUserID   *uint     `json:"replyUserId,omitempty"`
	ReplyNickname string    `json:"replyNickname,omitempty"`   // 回复人昵称
	ParentID      uint      `json:"parentId"`
	IsReview      int8      `json:"isReview"`
	LikeCount     int64     `json:"likeCount"`
	CreateTime    time.Time `json:"createTime"`
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
