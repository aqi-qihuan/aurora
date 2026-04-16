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

// LabelOptionDTO 标签选项DTO（对标Java LabelOptionDTO）
// 用于树形下拉框（菜单选择、资源选择等）
type LabelOptionDTO struct {
	ID       uint              `json:"id"`
	Label    string            `json:"label"`
	Children []LabelOptionDTO  `json:"children,omitempty"`
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
	LinkName    string    `json:"linkName"`
	LinkAvatar  string    `json:"linkAvatar"`
	LinkAddress string    `json:"linkAddress"`
	LinkIntro   string    `json:"linkIntro"`
	CreateTime  time.Time `json:"createTime,omitempty"`
}

type FriendLinkAdminDTO struct {
	ID          uint      `json:"id"`
	LinkName    string    `json:"linkName"`
	LinkAvatar  string    `json:"linkAvatar"`
	LinkAddress string    `json:"linkAddress"`
	LinkIntro   string    `json:"linkIntro"`
	CreateTime  time.Time `json:"createTime"`
}

// ===== 相册/照片 DTO =====

type PhotoDTO struct {
	ID  uint   `json:"id"`
	URL string `json:"url"`
}

// PhotoAdminDTO 后台管理照片DTO（对标Java PhotoAdminDTO）
type PhotoAdminDTO struct {
	ID        uint   `json:"id"`
	AlbumID   uint   `json:"albumId"`
	PhotoName string `json:"photoName"`
	PhotoDesc string `json:"photoDesc,omitempty"`
	PhotoSrc  string `json:"photoSrc"`
	IsDelete  int8   `json:"isDelete"`
}

// PhotoAlbumInfoDTO 前台相册信息DTO（对标Java PhotoDTO）
type PhotoAlbumInfoDTO struct {
	PhotoAlbumCover string   `json:"photoAlbumCover"` // 相册封面
	PhotoAlbumName  string   `json:"photoAlbumName"`  // 相册名称
	Photos          []string `json:"photos"`           // 照片URL列表
}

type AlbumDTO struct {
	ID         uint      `json:"id"`
	AlbumName  string    `json:"albumName"`
	AlbumDesc  string    `json:"albumDesc,omitempty"`  // 对齐Java PhotoAlbumDTO.albumDesc
	AlbumCover string    `json:"albumCover"`
	Status     int8      `json:"status,omitempty"`     // 1公开 2私密（前端用此字段统计）
	PhotoCount int       `json:"photoCount"`
	Info       string    `json:"info,omitempty"`       // 兼容旧版（前台用）
	IsPrivate  bool      `json:"isPrivate,omitempty"`
	CreateTime time.Time `json:"createTime,omitempty"`
}

// ===== 角色/菜单 DTO =====

type RoleDTO struct {
	ID          uint      `json:"id"`
	RoleName    string    `json:"roleName"`
	IsDisable   int8      `json:"isDisable"`
	MenuIDs     []uint    `json:"menuIds"`
	CreateTime  time.Time `json:"createTime"`
}

type RoleDetailDTO struct {
	ID          uint     `json:"id"`
	RoleName    string   `json:"roleName"`
	IsDisable   int8     `json:"isDisable"`
	MenuIDs     []uint   `json:"menuIds"`
	Menus       []model.Menu `json:"menus,omitempty"` // 直接返回Menu用于前端渲染
}

type MenuDTO struct {
	ID         uint        `json:"id"`
	Name       string      `json:"name"`
	Path       string      `json:"path"`
	Component  string      `json:"component"`
	Icon       string      `json:"icon"`
	OrderNum   int         `json:"orderNum"`  // 排序号（数据库对应 order_num 字段）
	IsHidden   int8        `json:"isHidden"`  // 是否隐藏 0否1是（数据库对应 is_hidden 字段）
	ParentID   *uint       `json:"parentId"`
	Children   []MenuDTO   `json:"children,omitempty"` // 子菜单（树形结构）
	CreateTime time.Time   `json:"createTime"`
	UpdateTime *time.Time  `json:"updateTime,omitempty"`
}

type MenuTreeDTO struct {
	ID          uint          `json:"id"`
	Name        string        `json:"name"`          // 侧边栏菜单用
	Label       string        `json:"label"`         // 角色权限树用（el-tree默认用label）
	Path        string        `json:"path"`
	Component   string        `json:"component"`
	Icon        string        `json:"icon"`
	IsHidden    int8          `json:"isHidden"`      // 是否隐藏 0否1是（数据库对应 is_hidden 字段）
	OrderNum    int           `json:"orderNum"`      // 排序号（数据库对应 order_num 字段）
	ParentID    *uint         `json:"parentId"`
	Children    []MenuTreeDTO `json:"children"`
	CreateTime  time.Time     `json:"createTime,omitempty"`
	UpdateTime  *time.Time    `json:"updateTime,omitempty"`
}

// ===== 定时任务 DTO =====

type JobDTO struct {
	ID             uint       `json:"id"`
	JobName        string     `json:"jobName"`
	JobGroup       string     `json:"jobGroup,omitempty"`
	InvokeTarget   string     `json:"invokeTarget"`
	CronExpression string     `json:"cronExpression,omitempty"`
	MisfirePolicy  int        `json:"misfirePolicy,omitempty"`
	Concurrent     int        `json:"concurrent,omitempty"`
	Status         int        `json:"status"`
	Remark         string     `json:"remark,omitempty"`
	CreateTime     time.Time  `json:"createTime"`
	UpdateTime     time.Time  `json:"updateTime,omitempty"`
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
	ID             uint       `json:"id"`
	JobName        string     `json:"jobName"`
	JobGroup       string     `json:"jobGroup,omitempty"`
	InvokeTarget   string     `json:"invokeTarget"`
	CronExpression string     `json:"cronExpression"`
	MisfirePolicy  int        `json:"misfirePolicy,omitempty"`
	Concurrent     int        `json:"concurrent,omitempty"`
	Status         int        `json:"status"`
	Remark         string     `json:"remark,omitempty"`
	CreateTime     time.Time  `json:"createTime"`
	UpdateTime     time.Time  `json:"updateTime,omitempty"`
}

// ===== 日志 DTO =====

// OperationLogVO 操作日志VO（用于中间件创建日志记录，对标Java OperationLog实体）
type OperationLogVO struct {
	UserID        uint   `json:"userId"`
	Module        string `json:"module"`         // 操作模块（如"文章模块"）
	Operation     string `json:"operation"`      // 操作类型（如"ADD","UPDATE","DELETE"）
	Method        string `json:"method"`         // 操作方法（如"ArticleController.addArticle"）
	URL           string `json:"url"`            // 请求路径
	IP            string `json:"ip"`             // 操作IP
	RequestMethod string `json:"requestMethod"`  // 请求方式（GET/POST等）
	RequestParam  string `json:"requestParam"`   // 请求参数
	ResponseData  string `json:"responseData"`   // 返回数据
	OptDesc       string `json:"optDesc"`        // 操作描述
	Nickname      string `json:"nickname"`       // 用户昵称
	IpSource      string `json:"ipSource"`       // IP来源
}

// OperationLogDTO 操作日志DTO（对标Java OperationLogDTO）
// 注意：JSON标签必须与Java完全一致，否则前端无法渲染
type OperationLogDTO struct {
	ID             uint      `json:"id"`
	OptModule      string    `json:"optModule"`                // 系统模块（对标Java）
	OptUri         string    `json:"optUri"`                   // 请求接口（对标Java）
	OptType        string    `json:"optType"`                  // 操作类型（对标Java）
	OptMethod      string    `json:"optMethod"`                // 请求方法（对标Java）
	OptDesc        string    `json:"optDesc"`                  // 操作描述（对标Java）
	RequestMethod  string    `json:"requestMethod"`            // 请求方式（对标Java）
	RequestParam   string    `json:"requestParam"`             // 请求参数（对标Java）
	ResponseData   string    `json:"responseData"`             // 返回数据（对标Java）
	Nickname       string    `json:"nickname"`                 // 操作人员（对标Java）
	IpAddress      string    `json:"ipAddress"`                // 登录IP（对标Java）
	IpSource       string    `json:"ipSource"`                 // 登录地址（对标Java）
	CreateTime     time.Time `json:"createTime"`               // 操作日期（对标Java）
}

// ExceptionLogVO 异常日志VO（用于中间件创建异常记录，对标Java ExceptionLog实体）
type ExceptionLogVO struct {
	URL           string `json:"url"`            // 请求接口
	Method        string `json:"method"`         // 请求方法
	RequestMethod string `json:"requestMethod"`  // 请求方式
	RequestParam  string `json:"requestParam"`   // 请求参数
	OptDesc       string `json:"optDesc"`        // 操作描述
	ExceptionInfo string `json:"exceptionInfo"`  // 异常堆栈信息
	IP            string `json:"ip"`             // 操作IP
	IpSource      string `json:"ipSource"`       // IP来源
}

// ExceptionLogDTO 异常日志DTO（对标Java ExceptionLogDTO）
// 注意：JSON标签必须与Java完全一致，否则前端无法渲染
type ExceptionLogDTO struct {
	ID             uint      `json:"id"`
	OptUri         string    `json:"optUri"`                   // 请求接口（对标Java）
	OptMethod      string    `json:"optMethod"`                // 请求方法（对标Java）
	RequestMethod  string    `json:"requestMethod"`            // 请求方式（对标Java）
	RequestParam   string    `json:"requestParam"`             // 请求参数（对标Java）
	OptDesc        string    `json:"optDesc"`                  // 操作描述（对标Java）
	ExceptionInfo  string    `json:"exceptionInfo"`            // 异常堆栈（对标Java）
	IpAddress      string    `json:"ipAddress"`                // 登录IP（对标Java）
	IpSource       string    `json:"ipSource"`                 // 登录地址（对标Java）
	CreateTime     time.Time `json:"createTime"`               // 异常时间（对标Java）
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
	TopArticles    []ArticleCardDTO  `json:"topArticles"`
	LatestArticles []ArticleCardDTO  `json:"latestArticles"`
	Categories     []CategoryDTO     `json:"categories"`
	Tags           []TagDTO          `json:"tags"`
	FriendLinks    []FriendLinkDTO   `json:"friendLinks"`
	Talks          []TalkDTO         `json:"talks"`
	WebsiteConfig  *WebsiteConfigDTO `json:"websiteConfigDTO,omitempty"` // 网站配置（对标Java websiteConfigDTO）
	ViewCount      int               `json:"viewCount"`               // 总浏览量（对标Java viewCount）
	ArticleCount   int               `json:"articleCount"`            // 文章总数（对标Java articleCount）
	CategoryCount  int               `json:"categoryCount"`           // 分类总数（对标Java categoryCount）
	TagCount       int               `json:"tagCount"`                // 标签总数（对标Java tagCount）
	TalkCount      int               `json:"talkCount"`               // 说说总数（对标Java talkCount）
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
// 对标Java WebsiteConfigDTO，字段名必须与前端期望完全一致

type WebsiteConfigDTO struct {
	Name                 string `json:"name"`                 // 对标Java：网站名称
	EnglishName          string `json:"englishName"`          // 对标Java：网站英文名称
	Author               string `json:"author"`               // 对标Java：网站作者
	AuthorAvatar         string `json:"authorAvatar"`         // 对标Java：作者头像
	AuthorIntro          string `json:"authorIntro"`          // 对标Java：作者介绍
	Logo                 string `json:"logo"`                 // 对标Java：网站logo
	MultiLanguage        *int   `json:"multiLanguage"`        // 对标Java：多语言（0关闭1开启）
	Notice               string `json:"notice"`               // 对标Java：网站公告
	WebsiteCreateTime    string `json:"websiteCreateTime"`    // 对标Java：网站创建时间
	BeianNumber          string `json:"beianNumber"`          // 对标Java：工信部备案号
	QqLogin              *int   `json:"qqLogin"`              // 对标Java：QQ登录（0关闭1开启）
	Github               string `json:"github"`               // 对标Java：github
	Gitee                string `json:"gitee"`                // 对标Java：gitee
	Qq                   string `json:"qq"`                   // 对标Java：qq
	WeChat               string `json:"weChat"`               // 对标Java：微信
	Weibo                string `json:"weibo"`                // 对标Java：微博
	Csdn                 string `json:"csdn"`                 // 对标Java：csdn
	Zhihu                string `json:"zhihu"`                // 对标Java：zhihu
	Juejin               string `json:"juejin"`               // 对标Java：juejin
	Twitter              string `json:"twitter"`              // 对标Java：twitter
	Stackoverflow        string `json:"stackoverflow"`        // 对标Java：stackoverflow
	TouristAvatar        string `json:"touristAvatar"`        // 对标Java：游客头像
	UserAvatar           string `json:"userAvatar"`           // 对标Java：用户头像
	IsCommentReview      *int   `json:"isCommentReview"`      // 对标Java：评论审核（0关闭1开启）
	IsEmailNotice        *int   `json:"isEmailNotice"`        // 对标Java：邮箱通知（0关闭1开启）
	IsReward             *int   `json:"isReward"`             // 对标Java：打赏（0关闭1开启）
	WeiXinQRCode         string `json:"weiXinQRCode"`         // 对标Java：微信二维码
	AlipayQRCode         string `json:"alipayQRCode"`         // 对标Java：支付宝二维码
	Favicon              string `json:"favicon"`              // 对标Java：favicon
	WebsiteTitle         string `json:"websiteTitle"`         // 对标Java：网页标题
	GonganBeianNumber    string `json:"gonganBeianNumber"`    // 对标Java：公安部备案编号
}

// AboutDTO 关于页面DTO（对标Java AboutDTO）
type AboutDTO struct {
	Content string `json:"content"` // 关于页面内容（HTML/Markdown）
}

// ===== 文件上传 DTO =====

type FileUploadDTO struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
}

// ===== 资源权限 DTO =====

type ResourceDTO struct {
	ID            uint          `json:"id"`
	ResourceName  string        `json:"resourceName"`
	URL           string        `json:"url"`
	RequestMethod string        `json:"requestMethod"`
	ParentID      *uint         `json:"parentId"`
	IsAnonymous   int8          `json:"isAnonymous"`
	CreateTime    time.Time     `json:"createTime"`
	UpdateTime    *time.Time    `json:"updateTime,omitempty"`
	Children      []ResourceDTO `json:"children,omitempty"` // 子资源（树形结构）
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
