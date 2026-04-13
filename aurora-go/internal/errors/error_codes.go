package errors

import (
	stderrors "errors"
	"fmt"
)

// AppError 自定义应用错误
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("code=%d message=%s", e.Code, e.Message)
}

// New 创建新的AppError
func New(code int, msg string) *AppError {
	return &AppError{Code: code, Message: msg}
}

// Wrap 包装已有错误
func Wrap(code int, msg string, err error) *AppError {
	return &AppError{Code: code, Message: fmt.Sprintf("%s: %v", msg, err)}
}

// 预定义错误码（对齐Java版异常码）
var (
	// 成功 (200)
	OK = &AppError{200, "操作成功"}

	// 通用错误 (500)
	ErrInternalServer = New(500, "服务器内部错误")
	ErrSystemBusy     = New(501, "系统繁忙，请稍后重试")

	// 参数错误 (400)
	ErrInvalidParams  = New(400, "请求参数错误")
	ErrValidation    = New(401, "数据验证失败")

	// 认证/授权 (401-403)
	ErrUnauthorized   = New(401, "未登录或Token已过期")
	ErrForbidden      = New(403, "无访问权限")
	ErrTokenExpired   = New(404, "Token已过期，请重新登录")

	// 用户相关 (500-599)
	ErrUserNotFound        = New(500, "用户不存在")
	ErrUserDisabled        = New(501, "用户已被禁用")
	ErrUserExists          = New(502, "用户已存在")
	ErrPasswordIncorrect   = New(503, "密码错误")
	ErrEmailAlreadyUsed    = New(504, "邮箱已被使用")
	ErrUsernameExists      = New(505, "用户名已存在")
	ErrEmailExists         = New(506, "邮箱已存在")
	ErrInvalidCredentials  = New(507, "用户名或密码错误")
	ErrAccountDisabled     = New(508, "账号已被禁用")
	ErrInvalidOldPassword  = New(509, "旧密码错误")

	// 文章相关 (600-699)
	ErrArticleNotFound     = New(600, "文章不存在")
	ErrArticlePasswordWrong = New(601, "文章密码错误")

	// 文件上传 (700-799)
	ErrFileUploadFailed = New(700, "文件上传失败")
	ErrFileTypeNotSupported = New(701, "不支持的文件类型")
	ErrFileSizeExceeded = New(702, "文件大小超出限制")

	// 评论相关 (800-899)
	ErrCommentNotFound = New(800, "评论不存在")

	// 分类相关 (810-819)
	ErrCategoryNotFound      = New(810, "分类不存在")
	ErrCategoryNameExists    = New(811, "分类名已存在")
	ErrCategoryHasArticles   = New(812, "分类下存在文章，无法删除")

	// 标签相关 (820-829)
	ErrTagNotFound      = New(820, "标签不存在")
	ErrTagNameExists    = New(821, "标签名已存在")

	// 友链相关 (830-839)
	ErrFriendLinkNotFound = New(830, "友链不存在")

	// 定时任务相关 (840-849)
	ErrJobNotFound = New(840, "定时任务不存在")

	// 菜单相关 (850-859)
	ErrMenuNotFound = New(850, "菜单不存在")

	// 相册/照片相关 (860-869)
	ErrAlbumNotFound = New(860, "相册不存在")
	ErrPhotoNotFound = New(861, "照片不存在")
	ErrAlbumNameExists = New(862, "相册名已存在")

	// 资源相关 (870-879)
	ErrResourceNotFound = New(870, "资源不存在")

	// 角色相关 (880-889)
	ErrRoleNotFound        = New(880, "角色不存在")
	ErrRoleNameExists      = New(881, "角色名已存在")
	ErrCannotDeleteDefaultRole = New(882, "不能删除默认角色")
	ErrRoleHasUsers        = New(883, "角色下存在用户，无法删除")

	// 说说相关 (890-899)
	ErrTalkNotFound = New(890, "说说不存在")

	// Agent 相关 (900-999) - Agent模块专属
	ErrAgentDisabled       = New(900, "AI Agent功能未启用")
	ErrAgentLLMFailed      = New(901, "LLM调用失败")
	ErrAgentToolExecution  = New(902, "工具执行失败")
	ErrAgentContextTooLong = New(903, "上下文超出长度限制")

	// 配置/策略 (1000-1099) - 策略模式专属
	ErrInvalidConfig = New(1000, "配置无效")
	ErrNotImplemented = New(1001, "功能尚未实现（需集成SDK）")
)

// WithMsg 返回新的AppError，使用自定义消息
func (e *AppError) WithMsg(msg string) *AppError {
	return &AppError{Code: e.Code, Message: msg}
}

// Is 判断错误是否匹配
func Is(err error, target *AppError) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == target.Code
	}
	return false
}

// IsStd 判断错误是否匹配标准库error (用于 gorm.ErrRecordNotFound 等)
func IsStd(err, target error) bool {
	return stderrors.Is(err, target)
}
