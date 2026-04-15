package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"strings"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/model"
	"github.com/aurora-go/aurora/internal/service"
	"github.com/gin-gonic/gin"
)

// OptLogMeta 操作日志元数据 (对标Java @OptLog注解)
type OptLogMeta struct {
	Module string // 系统模块
	Type   string // 操作类型
	Desc   string // 操作描述
}

// OptLogRegistry 路由到操作日志元数据的映射表
// Key: "METHOD /api/admin/..." (注意: 必须包含 /api 前缀，因为路由组定义为 api.Group)
var OptLogRegistry = map[string]OptLogMeta{
	// 网站配置
	"PUT /api/admin/website/config":     {Module: "aurora信息", Type: "修改", Desc: "更新网站配置"},
	"PUT /api/admin/about":              {Module: "aurora信息", Type: "修改", Desc: "更新关于页面"},
	"POST /api/admin/config/images":     {Module: "aurora信息", Type: "上传", Desc: "上传配置图片"},

	// 文章管理
	"POST /api/admin/articles":                {Module: "文章模块", Type: "新增", Desc: "发布文章"},
	"PUT /api/admin/articles":                 {Module: "文章模块", Type: "修改", Desc: "修改文章"},
	"PUT /api/admin/articles/topAndFeatured":  {Module: "文章模块", Type: "修改", Desc: "修改文章属性"},
	"DELETE /api/admin/articles/delete":       {Module: "文章模块", Type: "删除", Desc: "彻底删除文章"},
	"POST /api/admin/articles/images":         {Module: "文章模块", Type: "上传", Desc: "上传文章图片"},
	"POST /api/admin/articles/import":         {Module: "文章模块", Type: "导入", Desc: "导入文章"},
	"POST /api/admin/articles/export":         {Module: "文章模块", Type: "导出", Desc: "导出文章"},

	// 分类管理
	"POST /api/admin/categories":              {Module: "分类管理", Type: "新增/修改", Desc: "保存分类"},
	"DELETE /api/admin/categories":            {Module: "分类管理", Type: "删除", Desc: "删除分类"},

	// 标签管理
	"POST /api/admin/tags":                    {Module: "标签管理", Type: "新增/修改", Desc: "保存标签"},
	"DELETE /api/admin/tags":                  {Module: "标签管理", Type: "删除", Desc: "删除标签"},

	// 评论管理
	"PUT /api/admin/comments/review":          {Module: "评论管理", Type: "修改", Desc: "审核评论"},
	"DELETE /api/admin/comments":              {Module: "评论管理", Type: "删除", Desc: "删除评论"},

	// 友链管理
	"POST /api/admin/links":                   {Module: "友链管理", Type: "新增/修改", Desc: "保存友链"},
	"PUT /api/admin/links":                    {Module: "友链管理", Type: "修改", Desc: "修改友链"},
	"DELETE /api/admin/links":                 {Module: "友链管理", Type: "删除", Desc: "删除友链"},

	// 说说管理
	"POST /api/admin/talks":                   {Module: "说说模块", Type: "新增/修改", Desc: "发布说说"},
	"POST /api/admin/talks/images":            {Module: "说说模块", Type: "上传", Desc: "上传说说图片"},
	"DELETE /api/admin/talks":                 {Module: "说说模块", Type: "删除", Desc: "删除说说"},

	// 相册管理
	"POST /api/admin/photos/albums":           {Module: "相册管理", Type: "新增/修改", Desc: "保存相册"},
	"POST /api/admin/photos/albums/upload":    {Module: "相册管理", Type: "上传", Desc: "上传相册封面"},
	"DELETE /api/admin/photos/albums/:id":     {Module: "相册管理", Type: "删除", Desc: "删除相册"},

	// 照片管理
	"POST /api/admin/photos/upload":           {Module: "照片管理", Type: "上传", Desc: "上传照片"},
	"POST /api/admin/photos":                  {Module: "照片管理", Type: "保存", Desc: "保存照片信息"},
	"PUT /api/admin/photos":                   {Module: "照片管理", Type: "修改", Desc: "修改照片信息"},
	"PUT /api/admin/photos/album":             {Module: "照片管理", Type: "移动", Desc: "移动照片相册"},
	"PUT /api/admin/photos/delete":            {Module: "照片管理", Type: "修改", Desc: "标记删除照片"},
	"DELETE /api/admin/photos":                {Module: "照片管理", Type: "删除", Desc: "彻底删除照片"},

	// 用户管理
	"PUT /api/admin/users/role":               {Module: "用户管理", Type: "修改", Desc: "修改用户角色"},
	"PUT /api/admin/users/disable":            {Module: "用户管理", Type: "修改", Desc: "修改用户状态"},
	"PUT /api/admin/users/password":           {Module: "用户管理", Type: "修改", Desc: "修改用户密码"},
	"DELETE /api/admin/users/:id/online":      {Module: "用户管理", Type: "下线", Desc: "踢出用户"},

	// 角色管理
	"POST /api/admin/role":                    {Module: "角色管理", Type: "新增/修改", Desc: "保存角色"},
	"DELETE /api/admin/roles":                 {Module: "角色管理", Type: "删除", Desc: "删除角色"},

	// 资源管理
	"POST /api/admin/resources":               {Module: "资源管理", Type: "新增/修改", Desc: "保存资源"},
	"DELETE /api/admin/resources/:id":         {Module: "资源管理", Type: "删除", Desc: "删除资源"},

	// 菜单管理
	"POST /api/admin/menus":                   {Module: "菜单管理", Type: "新增/修改", Desc: "保存菜单"},
	"PUT /api/admin/menus/isHidden":           {Module: "菜单管理", Type: "修改", Desc: "修改菜单显隐"},
	"DELETE /api/admin/menus/:id":             {Module: "菜单管理", Type: "删除", Desc: "删除菜单"},

	// 定时任务管理
	"POST /api/admin/jobs":                    {Module: "任务管理", Type: "新增/修改", Desc: "保存定时任务"},
	"PUT /api/admin/jobs":                     {Module: "任务管理", Type: "修改", Desc: "修改定时任务"},
	"DELETE /api/admin/jobs":                  {Module: "任务管理", Type: "删除", Desc: "删除定时任务"},
	"PUT /api/admin/jobs/status":              {Module: "任务管理", Type: "修改", Desc: "修改任务状态"},
	"PUT /api/admin/jobs/run":                 {Module: "任务管理", Type: "执行", Desc: "执行一次任务"},
	"DELETE /api/admin/jobLogs":               {Module: "任务管理", Type: "删除", Desc: "删除调度日志"},
	"DELETE /api/admin/jobLogs/clean":         {Module: "任务管理", Type: "清空", Desc: "清空调度日志"},

	// 系统日志
	"DELETE /api/admin/operation/logs":        {Module: "系统日志", Type: "删除", Desc: "删除操作日志"},
	"DELETE /api/admin/exception/logs":        {Module: "系统日志", Type: "删除", Desc: "删除异常日志"},

	// 文件上传
	"POST /api/admin/upload":                  {Module: "文件管理", Type: "上传", Desc: "上传文件"},
	"POST /api/admin/upload/batch":            {Module: "文件管理", Type: "上传", Desc: "批量上传文件"},
	"POST /api/admin/upload/image":            {Module: "文件管理", Type: "上传", Desc: "上传图片"},
}

// AccessLog 操作日志中间件 (对标Java @OptLog AOP)
func AccessLog(registry *service.Registry, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method

		// 读取请求体
		var body []byte
		if c.Request.Body != nil {
			body, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		}

		c.Next()

		statusCode := c.Writer.Status()
		go func() {
			// 只记录产生数据变更的“写操作”，忽略“读操作”（GET）
			if method == "GET" {
				return
			}
			if statusCode >= 400 && statusCode != 401 && statusCode != 403 {
				return
			}
			if shouldSkipPath(path) {
				return
			}

			userID := GetUserID(c)
			// 构建路由键（提前定义，供后续使用）
			routeKey := method + " " + path

			// 文件上传接口不记录原始请求体（含二进制数据）
			bodyStr := "[文件上传]"
			if !strings.HasPrefix(routeKey, "POST /api/admin/talks/images") &&
				!strings.HasPrefix(routeKey, "POST /api/admin/articles/images") &&
				!strings.HasPrefix(routeKey, "POST /api/admin/photos/upload") &&
				!strings.HasPrefix(routeKey, "POST /api/admin/upload") {
				bodyStr = string(body)
				if len(bodyStr) > 1000 {
					bodyStr = bodyStr[:1000] + "...(truncated)"
				}
			}

			// 1. 匹配路由元数据 (支持 QueryString 的模糊匹配)
			var meta OptLogMeta
			meta = OptLogRegistry[routeKey]

			// 如果精确匹配失败，尝试忽略 QueryString 再次匹配
			if meta.Module == "" {
				baseKey := method + " " + strings.Split(path, "?")[0]
				meta = OptLogRegistry[baseKey]
				if meta.Module == "" {
					// 使用 Info 级别确保能看到调试日志（日志级别为 Info）
					logger.Info("[操作日志] 未匹配到路由元数据", "routeKey", routeKey, "baseKey", baseKey)
				}
			}

			// 2. 获取真实 IP (支持代理环境)
			realIP := c.ClientIP()
			if realIP == "127.0.0.1" || realIP == "::1" {
				if fwd := c.GetHeader("X-Forwarded-For"); fwd != "" {
					realIP = strings.Split(fwd, ",")[0]
				} else if real := c.GetHeader("X-Real-IP"); real != "" {
					realIP = real
				}
			}

			// 3. 获取用户昵称
			nickname := "未知用户"
			if userID > 0 && registry.DB != nil {
				var userInfo model.UserInfo
				if err := registry.DB.Select("nickname").Where("id = (SELECT user_info_id FROM t_user_auth WHERE id = ?)", userID).First(&userInfo).Error; err == nil && userInfo.Nickname != "" {
					nickname = userInfo.Nickname
				}
			}

			logEntry := &dto.OperationLogVO{
				UserID:        userID,
				Nickname:      nickname,
				Module:        meta.Module,
				Operation:     meta.Type,
				OptDesc:       meta.Desc,
				Method:        routeKey,
				URL:           path,
				RequestMethod: method,
				IP:            realIP,
				RequestParam:  bodyStr,
			}

			if err := registry.OperationLog.Save(logEntry); err != nil {
				logger.Error("保存操作日志失败", "error", err, "path", path)
			}
		}()
	}
}

// shouldSkipPath 判断是否跳过记录的路径
func shouldSkipPath(path string) bool {
	skipPrefixes := []string{
		"/health",
		"/metrics",
		"/favicon.ico",
		"/static/",
		"/api/home/info",
		"/api/articles/search",
	}
	for _, prefix := range skipPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}
