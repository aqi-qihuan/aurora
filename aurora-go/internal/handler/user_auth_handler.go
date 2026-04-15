package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/aurora-go/aurora/internal/constant"
	"github.com/gin-gonic/gin"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/model"
	"github.com/aurora-go/aurora/internal/scheduler"
	"github.com/aurora-go/aurora/internal/service"
	"github.com/aurora-go/aurora/internal/util"
	"github.com/aurora-go/aurora/internal/vo"
)

// UserAuthHandler 用户认证处理器（对标 Java UserController + UserAuthController）
type UserAuthHandler struct {
	registry *service.Registry
}

func NewUserAuthHandler(registry *service.Registry) *UserAuthHandler {
	return &UserAuthHandler{registry: registry}
}

// Register 用户注册
// POST /api/auth/register
func (h *UserAuthHandler) Register(c *gin.Context) {
	var registerVO vo.RegisterVO
	if err := c.ShouldBindJSON(&registerVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	result, err := h.registry.UserAuth.Register(c.Request.Context(), registerVO)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// Login 用户登录
// POST /api/auth/login
func (h *UserAuthHandler) Login(c *gin.Context) {
	var loginVO vo.LoginVO

	// 使用ShouldBind自动检测Content-Type,支持JSON和表单两种格式
	if err := c.ShouldBind(&loginVO); err != nil {
		slog.Warn("登录请求参数解析失败",
			"error", err.Error(),
			"content_type", c.GetHeader("Content-Type"))
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("用户名或密码不能为空"))
		return
	}

	result, err := h.registry.UserAuth.Login(c.Request.Context(), loginVO)
	if err != nil {
		slog.Warn("登录失败", "username", loginVO.Username, "error", err.Error())
		util.ResponseError(c, err)
		return
	}

	// 使用TokenService生成JWT Token（替换临时随机字符串）
	if h.registry.TokenSvc != nil {
		// 从请求头解析 User-Agent 获取浏览器和操作系统信息
		userAgent := c.GetHeader("User-Agent")
		browser := util.ParseBrowser(userAgent)
		os := util.ParseOS(userAgent)
		
		// 获取客户端IP
		clientIP := util.GetClientIP(c)
		
		// 计算IP归属地（对标Java IpUtil.getIpSource(ipAddress)）
		// 注意：不是用数据库中存的IPSource，而是根据当前登录IP实时计算
		ipSource := util.GetIPRegion(clientIP)
		
		// 构造完整的UserDetailsDTO（对标Java版）
		// 关键：ID字段必须使用UserAuth.id（登录认证ID），不是UserInfo.id
		// Java版TokenServiceImpl.createToken()中userId = userDetailsDTO.getId() = UserAuth.id
		// 下线时delLoginUser(userId)也是用UserAuth.id删除，必须一致
		userDetail := &dto.UserDetailsDTO{
			ID:            result.ID,          // UserAuth.id (登录认证ID, 不是UserInfo.id)
			UserInfoID:    result.UserInfoID,  // UserInfo.id
			Email:         result.Email,
			LoginType:     result.LoginType, // int类型
			Username:      result.Username,
			Nickname:      result.Nickname,
			Avatar:        result.Avatar,
			Intro:         result.Intro,
			Website:       result.Website,
			IsSubscribe:   result.IsSubscribe,
			IPAddress:     clientIP,
			IPSource:      ipSource,  // 使用实时计算的IP归属地，不是数据库中的旧值
			IsDisable:     0, // 默认不禁用
			Browser:       browser,
			OS:            os,
			LastLoginTime: time.Now(), // 记录登录时间
			Roles:         []string{"admin"}, // TODO: 从数据库查询实际角色
		}
		
		tokenString, err := h.registry.TokenSvc.CreateToken(userDetail)
		if err != nil {
			slog.Error("生成JWT Token失败", "error", err)
			util.ResponseError(c, errors.ErrInternalServer.WithMsg("Token生成失败"))
			return
		}
		
		// 更新返回结果中的Token
		result.Token = tokenString
		slog.Debug("JWT Token生成成功", "user_id", result.UserInfoID, "browser", browser, "os", os, "ip", clientIP)
	}

	util.ResponseSuccess(c, result)
}

// Logout 用户登出
// POST /api/auth/logout
func (h *UserAuthHandler) Logout(c *gin.Context) {
	// 获取当前用户ID (Gin Context中可能是 uint64 或 uint)
	userID, _ := c.Get("user_id")
	var uid uint
	switch v := userID.(type) {
	case uint:
		uid = v
	case uint64:
		uid = uint(v)
	case float64:
		uid = uint(v)
	default:
		util.ResponseError(c, errors.ErrUnauthorized.WithMsg("无法获取用户ID"))
		return
	}
	if err := h.registry.UserAuth.Logout(c.Request.Context(), uid); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "登出成功")
}

// QQLogin QQ OAuth 登录回调
// POST /api/auth/qq/callback
func (h *UserAuthHandler) QQLogin(c *gin.Context) {
	var qqVO dto.QQLoginVO
	if err := c.ShouldBindJSON(&qqVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	if h.registry.QQOAuthSvc == nil {
		util.ResponseError(c, errors.ErrInternalServer.WithMsg("QQ登录功能未启用"))
		return
	}
	result, err := h.registry.QQOAuthSvc.Login(c.Request.Context(), &qqVO)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// SendVerificationCode 发送邮箱验证码
// POST /api/auth/code
func (h *UserAuthHandler) SendVerificationCode(c *gin.Context) {
	var emailVO dto.EmailVO
	if err := c.ShouldBindJSON(&emailVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	if err := h.registry.UserAuth.SendVerificationCode(c.Request.Context(), emailVO.Email); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "验证码已发送，请查收邮件")
}

// UpdatePassword 修改密码
// PUT /api/user/password
func (h *UserAuthHandler) UpdatePassword(c *gin.Context) {
	var passwordVO vo.PasswordVO
	if err := c.ShouldBindJSON(&passwordVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	// 获取当前用户ID (Gin Context中可能是 uint64 或 uint)
	userID, _ := c.Get("user_id")
	var uid uint
	switch v := userID.(type) {
	case uint:
		uid = v
	case uint64:
		uid = uint(v)
	case float64:
		uid = uint(v)
	default:
		util.ResponseError(c, errors.ErrUnauthorized.WithMsg("无法获取用户ID"))
		return
	}
	if err := h.registry.UserAuth.ChangePassword(c.Request.Context(), uid, passwordVO); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "密码修改成功")
}

// ResetPassword 重置密码（通过邮箱验证码）
// PUT /api/auth/password/reset
func (h *UserAuthHandler) ResetPassword(c *gin.Context) {
	var resetVO dto.ResetPasswordVO
	if err := c.ShouldBindJSON(&resetVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	
	// 调用 Service 层重置密码
	if err := h.registry.UserAuth.ResetPassword(
		c.Request.Context(),
		resetVO.Email,
		resetVO.Code,
		resetVO.NewPassword,
	); err != nil {
		util.ResponseError(c, err)
		return
	}
	
	util.ResponseSuccess(c, "密码重置成功")
}

// GetUserInfo 获取当前登录用户信息
// GET /api/user/info
func (h *UserAuthHandler) GetUserInfo(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		util.ResponseError(c, errors.ErrUnauthorized)
		return
	}
	// 兼容 uint64/uint/float64 类型
	var uid uint
	switch v := userID.(type) {
	case uint:
		uid = v
	case uint64:
		uid = uint(v)
	case float64:
		uid = uint(v)
	default:
		util.ResponseError(c, errors.ErrUnauthorized.WithMsg("无法获取用户ID"))
		return
	}
	result, err := h.registry.UserAuth.GetUserInfoByID(c.Request.Context(), uid)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// ==================== 后台管理端点（UserInfoController + UserAuthController） ====================

// ListUsers 查询后台用户列表
// GET /api/admin/users
func (h *UserAuthHandler) ListUsers(c *gin.Context) {
	var condition dto.ConditionVO
	c.ShouldBindQuery(&condition)
	pageNum, pageSize := util.PageQuery(c)
	page := dto.PageVO{PageNum: pageNum, PageSize: pageSize}

	result, err := h.registry.UserAuth.ListUsers(c.Request.Context(), condition, page)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// ListUserAreas 获取用户区域分布
// GET /api/admin/users/area
// 对标 Java UserAuthServiceImpl.listUserAreas()
func (h *UserAuthHandler) ListUserAreas(c *gin.Context) {
	var condition dto.ConditionVO
	c.ShouldBindQuery(&condition)
	
	rdb := h.registry.RDB
	if rdb == nil {
		util.ResponseSuccess(c, []interface{}{})
		return
	}
	
	ctx := c.Request.Context()
	
	// 根据 type 参数选择不同的数据源 (对标Java switch getUserAreaType(conditionVO.getType()))
	// type=1: 用户 - 从 user_area (String JSON) 读取
	// type=2: 游客 - 从 visitor_area (Hash) 读取
	var result []map[string]interface{}
	
	// 处理指针类型，默认为 1（用户）
	areaType := int8(1)
	if condition.Type != nil {
		areaType = *condition.Type
	}
	
	switch areaType {
	case 1: // 用户
		data, err := rdb.Get(ctx, constant.UserArea).Bytes()
		if err != nil {
			if err.Error() != "redis: nil" {
				slog.Warn("获取用户地域分布失败", "error", err.Error())
			} else {
				slog.Info("Redis中user_area不存在，可能定时任务未运行")
			}
			util.ResponseSuccess(c, []interface{}{})
			return
		}
		
		slog.Info("Redis中user_area原始数据", "data", string(data))
		
		// 解析 JSON 数组
		var areaList []struct {
			Name  string `json:"name"`
			Value int64  `json:"value"`
		}
		if err := json.Unmarshal(data, &areaList); err != nil {
			slog.Warn("解析用户地域分布失败", "error", err.Error(), "raw_data", string(data))
			util.ResponseSuccess(c, []interface{}{})
			return
		}
		
		slog.Info("解析用户地域分布成功", "count", len(areaList))
		
		// 转换为前端需要的格式: {province: "北京", count: 5}
		result = make([]map[string]interface{}, len(areaList))
		for i, item := range areaList {
			result[i] = map[string]interface{}{
				"province": item.Name,
				"count":    item.Value,
			}
		}
		
	case 2: // 游客
		// 从 Hash 读取所有字段 (对标Java redisService.hGetAll(VISITOR_AREA))
		visitorArea, err := rdb.HGetAll(ctx, constant.VisitorArea).Result()
		if err != nil {
			slog.Warn("获取访客地域分布失败", "error", err.Error())
			util.ResponseSuccess(c, []interface{}{})
			return
		}
		
		// 转换为前端需要的格式
		result = make([]map[string]interface{}, 0, len(visitorArea))
		for province, countStr := range visitorArea {
			var count int64
			// 使用 strconv 替代 Sscanf，更安全
			if parsedCount, parseErr := strconv.ParseInt(countStr, 10, 64); parseErr == nil {
				count = parsedCount
			} else {
				slog.Warn("解析访客数量失败", "province", province, "value", countStr)
				continue
			}
			result = append(result, map[string]interface{}{
				"province": province,
				"count":    count,
			})
		}
		
	default:
		util.ResponseSuccess(c, []interface{}{})
		return
	}
	
	util.ResponseSuccess(c, result)
}

// TriggerUserAreaStats 手动触发用户地域统计（用于测试）
// POST /api/admin/users/area/trigger
func (h *UserAuthHandler) TriggerUserAreaStats(c *gin.Context) {
	if h.registry.RDB == nil || h.registry.DB == nil {
		util.ResponseError(c, errors.ErrInternalServer.WithMsg("Redis或数据库未初始化"))
		return
	}
	
	ctx := c.Request.Context()
	job := scheduler.NewUserAreaJob(h.registry.DB, h.registry.RDB)
	
	if err := job.Run(ctx); err != nil {
		slog.Error("手动触发用户地域统计失败", "error", err.Error())
		util.ResponseError(c, errors.ErrInternalServer.WithMsg(err.Error()))
		return
	}
	
	util.ResponseSuccess(c, "用户地域统计已更新")
}

// UpdateAdminPassword 修改管理员密码（对标Java UserAuthController.updateAdminPassword）
// PUT /api/admin/users/password
// Java逻辑: 管理员直接重置密码，不需要验证旧密码
func (h *UserAuthHandler) UpdateAdminPassword(c *gin.Context) {
	var passwordVO vo.PasswordVO
	if err := c.ShouldBindJSON(&passwordVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	// 获取当前用户ID (Gin Context中可能是 uint64 或 uint)
	userID, _ := c.Get("user_id")
	var uid uint
	switch v := userID.(type) {
	case uint:
		uid = v
	case uint64:
		uid = uint(v)
	case float64:
		uid = uint(v)
	default:
		util.ResponseError(c, errors.ErrUnauthorized.WithMsg("无法获取用户ID"))
		return
	}
	if err := h.registry.UserAuth.UpdateAdminPassword(c.Request.Context(), uid, passwordVO.NewPassword); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, nil)
}

// UpdateUserRole 修改用户角色和昵称（完全对标Java UserInfoController.updateUserRole）
// PUT /api/admin/users/role
// Java逻辑：1)更新UserInfo.nickname  2)删除旧角色  3)批量插入新角色
func (h *UserAuthHandler) UpdateUserRole(c *gin.Context) {
	var body struct {
		UserInfoID uint   `json:"userInfoId" binding:"required"` // 用户信息ID（不是UserAuth.id）
		Nickname   string `json:"nickname" binding:"required"`   // 昵称
		RoleIDs    []uint `json:"roleIds" binding:"required"`    // 角色ID列表
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}

	if len(body.RoleIDs) == 0 {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("用户角色不能为空"))
		return
	}

	ctx := c.Request.Context()

	// Step 1: 更新用户昵称
	if body.Nickname != "" {
		if err := h.registry.UserAuth.UpdateUserInfo(ctx, body.UserInfoID, vo.UpdateUserVO{
			Nickname: &body.Nickname,
		}); err != nil {
			slog.Error("更新用户昵称失败", "error", err.Error(), "userInfoId", body.UserInfoID)
			util.ResponseError(c, err)
			return
		}
	}

	// Step 2: 更新用户角色（删除旧角色 + 插入新角色）
	if err := h.registry.Role.AssignRoleToUser(ctx, body.UserInfoID, body.RoleIDs); err != nil {
		slog.Error("更新用户角色失败", "error", err.Error(), "userInfoId", body.UserInfoID, "roleIds", body.RoleIDs)
		util.ResponseError(c, err)
		return
	}

	slog.Info("用户角色更新成功", "userInfoId", body.UserInfoID, "nickname", body.Nickname, "roleIds", body.RoleIDs)
	util.ResponseSuccess(c, "角色修改成功")
}

// UpdateUserDisable 修改用户禁用状态（对标Java UserInfoController.updateUserDisable）
// PUT /api/admin/users/disable
// Java逻辑: 1)下线用户(删除Redis Session) 2)更新 isDisable 字段
func (h *UserAuthHandler) UpdateUserDisable(c *gin.Context) {
	var body struct {
		UserID    uint `json:"userId" binding:"required"`
		IsDisable int8 `json:"isDisable" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}

	ctx := c.Request.Context()

	// Step 1: 下线用户（对标Java removeOnlineUser(userDisableVO.getId())）
	if body.IsDisable == 1 {
		if err := h.registry.UserAuth.RemoveOnlineUser(ctx, body.UserID); err != nil {
			slog.Warn("下线用户失败", "error", err.Error(), "userId", body.UserID)
			// 不阻断后续操作
		}
	}

	// Step 2: 更新禁用状态
	if err := h.registry.DB.WithContext(ctx).Model(&model.UserInfo{}).Where("id = ?", body.UserID).Update("is_disable", body.IsDisable).Error; err != nil {
		util.ResponseError(c, err)
		return
	}

	util.ResponseSuccess(c, nil)
}

// ListOnlineUsers 查看在线用户列表
// GET /api/admin/users/online
func (h *UserAuthHandler) ListOnlineUsers(c *gin.Context) {
	var condition dto.ConditionVO
	c.ShouldBindQuery(&condition)
	pageNum, pageSize := util.PageQuery(c)
	page := dto.PageVO{PageNum: pageNum, PageSize: pageSize}

	result, err := h.registry.UserAuth.ListOnlineUsers(c.Request.Context(), condition, page)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// RemoveOnlineUser 下线指定用户
// DELETE /api/admin/users/:id/online
func (h *UserAuthHandler) RemoveOnlineUser(c *gin.Context) {
	userInfoIdStr := c.Param("id")
	if userInfoIdStr == "" {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("用户ID不能为空"))
		return
	}

	var userInfoId uint
	fmt.Sscanf(userInfoIdStr, "%d", &userInfoId)

	if err := h.registry.UserAuth.RemoveOnlineUser(c.Request.Context(), userInfoId); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "用户已下线")
}

// ==================== 用户信息端点（UserInfoController） ====================

// UpdateUserInfo 更新用户信息（对标Java UserInfoController.updateUserInfo）
// PUT /api/users/info
// Java逻辑: 更新当前用户的 nickname, intro, website
func (h *UserAuthHandler) UpdateUserInfo(c *gin.Context) {
	var body struct {
		Nickname string `json:"nickname" binding:"required"`
		Intro    string `json:"intro"`
		Website  string `json:"website"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}

	// 获取当前用户ID (Gin Context中可能是 uint64 或 uint)
	userID, _ := c.Get("user_id")
	var uid uint
	switch v := userID.(type) {
	case uint:
		uid = v
	case uint64:
		uid = uint(v)
	case float64:
		uid = uint(v)
	default:
		util.ResponseError(c, errors.ErrUnauthorized.WithMsg("无法获取用户ID"))
		return
	}

	ctx := c.Request.Context()

	// Step 1: 查询当前用户的 userInfoId (对标Java UserUtil.getUserDetailsDTO().getUserInfoId())
	var auth model.UserAuth
	if err := h.registry.DB.WithContext(ctx).Select("user_info_id").Where("id = ?", uid).First(&auth).Error; err != nil {
		slog.Warn("查询用户信息失败", "error", err.Error(), "userId", uid)
		util.ResponseError(c, errors.ErrUserNotFound)
		return
	}
	userInfoID := auth.UserID

	// Step 2: 更新用户信息（对标Java UserInfoServiceImpl.updateUserInfo）
	if err := h.registry.UserAuth.UpdateUserInfo(ctx, userInfoID, vo.UpdateUserVO{
		Nickname: &body.Nickname,
		Intro:    &body.Intro,
		WebSite:  &body.Website,
	}); err != nil {
		util.ResponseError(c, err)
		return
	}

	util.ResponseSuccess(c, nil)
}

// UpdateUserAvatar 更新用户头像
// POST /api/users/avatar
// 对标Java版: 1)上传到对象存储(MinIO/OSS) 2)更新数据库avatar字段 3)返回完整URL
func (h *UserAuthHandler) UpdateUserAvatar(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("请选择要上传的头像"))
		return
	}

	// 获取当前用户ID (Gin Context中可能是 uint64 或 uint)
	userID, _ := c.Get("user_id")
	var uid uint
	switch v := userID.(type) {
	case uint:
		uid = v
	case uint64:
		uid = uint(v)
	case float64:
		uid = uint(v)
	default:
		util.ResponseError(c, errors.ErrUnauthorized.WithMsg("无法获取用户ID"))
		return
	}

	// 使用 FileService 上传头像 (对标Java版 uploadStrategyContext.executeUploadStrategy)
	avatarURL, err := h.registry.File.UploadAvatar(c.Request.Context(), file, uid)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, avatarURL)
}

// BindUserEmail 绑定用户邮箱（对标Java UserInfoController.saveUserEmail）
// PUT /api/users/email
// Java逻辑: 1)校验验证码 2)更新当前用户邮箱
func (h *UserAuthHandler) BindUserEmail(c *gin.Context) {
	var emailVO dto.EmailVO
	if err := c.ShouldBindJSON(&emailVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}

	if emailVO.Code == "" || emailVO.Email == "" {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("邮箱和验证码不能为空"))
		return
	}

	// 获取当前用户ID (Gin Context中可能是 uint64 或 uint)
	userID, _ := c.Get("user_id")
	var uid uint
	switch v := userID.(type) {
	case uint:
		uid = v
	case uint64:
		uid = uint(v)
	case float64:
		uid = uint(v)
	default:
		util.ResponseError(c, errors.ErrUnauthorized.WithMsg("无法获取用户ID"))
		return
	}

	ctx := c.Request.Context()

	// Step 1: 查询当前用户的 userInfoId
	var auth model.UserAuth
	if err := h.registry.DB.WithContext(ctx).Select("user_info_id").Where("id = ?", uid).First(&auth).Error; err != nil {
		util.ResponseError(c, errors.ErrUserNotFound)
		return
	}
	userInfoID := auth.UserID

	// Step 2: 校验验证码（对标Java redisService.get(USER_CODE_KEY + emailVO.getEmail())）
	if h.registry.RDB != nil {
		codeKey := "user_code:" + emailVO.Email
		storedCode, err := h.registry.RDB.Get(ctx, codeKey).Result()
		if err != nil || storedCode != emailVO.Code {
			util.ResponseError(c, errors.ErrInvalidParams.WithMsg("验证码错误"))
			return
		}
		// 删除验证码（一次性使用）
		h.registry.RDB.Del(ctx, codeKey)
	}

	// Step 3: 更新邮箱（对标Java UserInfoServiceImpl.saveUserEmail - 直接updateById）
	if err := h.registry.DB.WithContext(ctx).Model(&model.UserInfo{}).Where("id = ?", userInfoID).Update("email", emailVO.Email).Error; err != nil {
		util.ResponseError(c, err)
		return
	}

	util.ResponseSuccess(c, nil)
}

// UpdateUserSubscribe 修改用户订阅状态（对标Java UserInfoController.updateUserSubscribe）
// PUT /api/users/subscribe
// Java逻辑: 1)检查用户是否绑定邮箱 2)更新 isSubscribe 字段
func (h *UserAuthHandler) UpdateUserSubscribe(c *gin.Context) {
	var body struct {
		UserID      uint `json:"userId" binding:"required"`
		IsSubscribe int8 `json:"isSubscribe" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}

	ctx := c.Request.Context()

	// Step 1: 检查用户是否绑定邮箱（对标Java StringUtils.isEmpty(temp.getEmail())）
	var email string
	if err := h.registry.DB.WithContext(ctx).Select("email").Where("id = ?", body.UserID).Table("t_user_info").First(&email).Error; err != nil {
		util.ResponseError(c, errors.ErrUserNotFound)
		return
	}
	if email == "" {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("邮箱未绑定！"))
		return
	}

	// Step 2: 更新订阅状态
	if err := h.registry.DB.WithContext(ctx).Model(&model.UserInfo{}).Where("id = ?", body.UserID).Update("is_subscribe", body.IsSubscribe).Error; err != nil {
		util.ResponseError(c, err)
		return
	}

	util.ResponseSuccess(c, nil)
}

// GetUserInfoById 根据ID获取用户信息
// GET /api/users/info/:id
func (h *UserAuthHandler) GetUserInfoById(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的用户ID"))
		return
	}
	result, err := h.registry.UserAuth.GetUserInfoByID(c.Request.Context(), uint(id))
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}
