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
	ID          uint   `json:"id"`                                     // 更新时必填，创建时为空
	LinkName    string `json:"linkName" binding:"required,max=20"`     // 对标Java：@NotBlank
	LinkAvatar  string `json:"linkAvatar" binding:"required,max=255"`  // 对标Java：@NotBlank
	LinkAddress string `json:"linkAddress" binding:"required,max=50"`  // 对标Java：@NotBlank
	LinkIntro   string `json:"linkIntro" binding:"required,max=100"`   // 对标Java：@NotBlank
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
	JobName        string `json:"jobName,omitempty"`        // 对标Java：无验证标签
	JobGroup       string `json:"jobGroup,omitempty"`       // 对标Java：无验证标签（前端可能不传）
	InvokeTarget   string `json:"invokeTarget,omitempty"`   // 对标Java：无验证标签
	CronExpression string `json:"cronExpression,omitempty"` // 对标Java：无验证标签
	MisfirePolicy  *int   `json:"misfirePolicy,omitempty"`  // 对标Java：Integer类型
	Concurrent     *int   `json:"concurrent,omitempty"`     // 对标Java：Integer类型
	Status         *int   `json:"status,omitempty"`         // 对标Java：Integer类型
	Remark         string `json:"remark,omitempty"`         // 对标Java：无验证标签
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
// 对标Java WebsiteConfigVO，字段名必须与前端期望完全一致

type WebsiteConfigVO struct {
	Name                 string `json:"name"`                 // 对标Java：网站名称
	EnglishName          string `json:"englishName"`          // 对标Java：网站英文名称
	Author               string `json:"author"`               // 对标Java：网站作者
	AuthorAvatar         string `json:"authorAvatar"`         // 对标Java：网站头像
	AuthorIntro          string `json:"authorIntro"`          // 对标Java：网站作者介绍
	Logo                 string `json:"logo"`                 // 对标Java：网站logo
	MultiLanguage        *int   `json:"multiLanguage"`        // 对标Java：Integer（0关闭1开启）
	Notice               string `json:"notice"`               // 对标Java：网站公告
	WebsiteCreateTime    string `json:"websiteCreateTime"`    // 对标Java：网站创建时间
	BeianNumber          string `json:"beianNumber"`          // 对标Java：工信部备案号
	QqLogin              *int   `json:"qqLogin"`              // 对标Java：Integer（0关闭1开启）
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
	IsCommentReview      *int   `json:"isCommentReview"`      // 对标Java：Integer（0关闭1开启）
	IsEmailNotice        *int   `json:"isEmailNotice"`        // 对标Java：Integer（0关闭1开启）
	IsReward             *int   `json:"isReward"`             // 对标Java：Integer（0关闭1开启）
	WeiXinQRCode         string `json:"weiXinQRCode"`         // 对标Java：微信二维码
	AlipayQRCode         string `json:"alipayQRCode"`         // 对标Java：支付宝二维码
	Favicon              string `json:"favicon"`              // 对标Java：favicon
	WebsiteTitle         string `json:"websiteTitle"`         // 对标Java：网页标题
	GonganBeianNumber    string `json:"gonganBeianNumber"`    // 对标Java：公安部备案编号
}
