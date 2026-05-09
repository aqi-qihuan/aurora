package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/aurora-go/aurora/internal/constant"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/middleware"
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

// @Summary 用户注册
// @Tags 用户
// @Description 注册新用户账号
// @Accept json
// @Produce json
// @Param registerVO body vo.RegisterVO true "注册信息"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 500 {object} object "服务器内部错误"
// @Router /api/users/register [post]
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

// @Summary 用户登录
// @Tags 用户
// @Description 用户登录获取Token
// @Accept json
// @Produce json
// @Param loginVO body vo.LoginVO true "登录信息"
// @Success 200 {object} object "成功返回登录结果和Token"
// @Failure 400 {object} object "请求参数错误"
// @Failure 500 {object} object "服务器内部错误"
// @Router /api/users/login [post]
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

	// 获取客户端IP和归属地（对标Java IpUtil.getIpAddress/getIpSource）
	clientIP := util.GetClientIP(c)
	ipSource := util.GetIPRegion(clientIP)

	result, err := h.registry.UserAuth.Login(c.Request.Context(), loginVO, clientIP, ipSource)
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
			ID:            result.ID,         // UserAuth.id (登录认证ID, 不是UserInfo.id)
			UserInfoID:    result.UserInfoID, // UserInfo.id
			Email:         result.Email,
			LoginType:     result.LoginType, // int类型
			Username:      result.Username,
			Nickname:      result.Nickname,
			Avatar:        result.Avatar,
			Intro:         result.Intro,
			Website:       result.Website,
			IsSubscribe:   result.IsSubscribe,
			IPAddress:     clientIP,
			IPSource:      ipSource, // 使用实时计算的IP归属地，不是数据库中的旧值
			IsDisable:     0,        // 默认不禁用
			Browser:       browser,
			OS:            os,
			LastLoginTime: time.Now(),        // 记录登录时间
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

// @Summary 用户登出
// @Tags 用户
// @Description 用户登出（需要登录）
// @Accept json
// @Produce json
// @Success 200 {object} object "成功响应"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/users/logout [post]
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

// @Summary QQ OAuth登录
// @Tags 用户
// @Description 通过QQ OAuth登录
// @Accept json
// @Produce json
// @Param qqLoginVO body dto.QQLoginVO true "QQ登录信息"
// @Success 200 {object} object "成功返回登录结果"
// @Failure 400 {object} object "请求参数错误"
// @Failure 500 {object} object "服务器内部错误"
// @Router /api/users/oauth/qq [post]
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

// @Summary 发送邮箱验证码
// @Tags 用户
// @Description 发送邮箱验证码
// @Accept json
// @Produce json
// @Param username query string true "邮箱地址"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 500 {object} object "服务器内部错误"
// @Router /api/users/code [get]
func (h *UserAuthHandler) SendVerificationCode(c *gin.Context) {
	email := c.Query("username") // 前端传参名是 username（实际是邮箱）
	if email == "" {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("邮箱不能为空"))
		return
	}
	if err := h.registry.UserAuth.SendVerificationCode(c.Request.Context(), email); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "验证码已发送，请查收邮件")
}

// @Summary 修改/重置密码
// @Tags 用户
// @Description 通过验证码修改或重置密码
// @Accept json
// @Produce json
// @Param body body object true "邮箱、验证码和新密码(username, code, password)"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 500 {object} object "服务器内部错误"
// @Router /api/users/password [put]
func (h *UserAuthHandler) UpdatePassword(c *gin.Context) {
	var userVO struct {
		Username string `json:"username" binding:"required"`
		Code     string `json:"code" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&userVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}

	// 调用 Service 层重置密码 (对标Java userAuthService.updatePassword)
	if err := h.registry.UserAuth.ResetPassword(
		c.Request.Context(),
		userVO.Username,
		userVO.Code,
		userVO.Password,
	); err != nil {
		util.ResponseError(c, err)
		return
	}

	util.ResponseSuccess(c, "密码修改成功")
}

// @Summary 重置密码
// @Tags 用户
// @Description 通过邮箱验证码重置密码
// @Accept json
// @Produce json
// @Param resetVO body dto.ResetPasswordVO true "重置密码信息"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 500 {object} object "服务器内部错误"
// @Router /api/users/password/reset [put]
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

// @Summary 获取当前登录用户信息
// @Tags 用户
// @Description 获取当前登录用户的信息（需要登录）
// @Accept json
// @Produce json
// @Success 200 {object} object "成功返回用户信息"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/users/info/{id} [get]
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

// @Summary 查询后台用户列表
// @Tags 用户
// @Description 后台查询用户列表（分页）
// @Accept json
// @Produce json
// @Param keyword query string false "搜索关键词"
// @Param current query int false "当前页码"
// @Param size query int false "每页数量"
// @Success 200 {object} object "成功返回用户列表"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/users [get]
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

// @Summary 获取用户区域分布
// @Tags 用户
// @Description 获取用户地域分布统计
// @Accept json
// @Produce json
// @Param type query int false "区域类型(1用户/2游客)"
// @Success 200 {object} object "成功返回地域分布"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/users/area [get]
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

// @Summary 手动触发用户地域统计
// @Tags 用户
// @Description 手动触发用户地域分布统计
// @Accept json
// @Produce json
// @Success 200 {object} object "成功响应"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/users/area/trigger [post]
func (h *UserAuthHandler) TriggerUserAreaStats(c *gin.Context) {
	if h.registry.RDB == nil || h.registry.DB == nil {
		util.ResponseError(c, errors.ErrInternalServer.WithMsg("Redis或数据库未初始化"))
		return
	}

	// 使用独立的 Context，避免 HTTP 请求结束后被取消
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	job := scheduler.NewUserAreaJob(h.registry.DB, h.registry.RDB)

	if err := job.Run(ctx); err != nil {
		slog.Error("手动触发用户地域统计失败", "error", err.Error())
		util.ResponseError(c, errors.ErrInternalServer.WithMsg(err.Error()))
		return
	}

	util.ResponseSuccess(c, "用户地域统计已更新")
}

// @Summary 修改管理员密码
// @Tags 用户
// @Description 修改管理员密码（无需旧密码）
// @Accept json
// @Produce json
// @Param passwordVO body vo.PasswordVO true "新密码"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/users/password [put]
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

// @Summary 修改用户角色和昵称
// @Tags 用户
// @Description 修改用户角色和昵称
// @Accept json
// @Produce json
// @Param body body object true "用户信息(userInfoId, nickname, roleIds)"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/users/role [put]
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

// @Summary 修改用户禁用状态
// @Tags 用户
// @Description 禁用或启用用户（同时下线）
// @Accept json
// @Produce json
// @Param body body object true "用户ID和禁用状态(id, isDisable)"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/users/disable [put]
func (h *UserAuthHandler) UpdateUserDisable(c *gin.Context) {
	var body struct {
		ID        uint `json:"id" binding:"required"`             // 对标Java UserDisableVO.id
		IsDisable int8 `json:"isDisable" binding:"min=0,max=1"`   // 对标Java UserDisableVO.isDisable (0正常,1禁用)
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}

	ctx := c.Request.Context()

	// Step 1: 下线用户（对标Java removeOnlineUser(userDisableVO.getId())）
	if body.IsDisable == 1 {
		if err := h.registry.UserAuth.RemoveOnlineUser(ctx, body.ID); err != nil {
			slog.Warn("下线用户失败", "error", err.Error(), "userId", body.ID)
			// 不阻断后续操作
		}
	}

	// Step 2: 更新禁用状态
	if err := h.registry.DB.WithContext(ctx).Model(&model.UserInfo{}).Where("id = ?", body.ID).Update("is_disable", body.IsDisable).Error; err != nil {
		util.ResponseError(c, err)
		return
	}

	util.ResponseSuccess(c, nil)
}

// @Summary 查看在线用户列表
// @Tags 用户
// @Description 查看当前在线用户列表
// @Accept json
// @Produce json
// @Param keyword query string false "搜索关键词"
// @Param current query int false "当前页码"
// @Param size query int false "每页数量"
// @Success 200 {object} object "成功返回在线用户列表"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/users/online [get]
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

// @Summary 下线指定用户
// @Tags 用户
// @Description 强制指定用户下线
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/users/{id}/online [delete]
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

// @Summary 更新用户信息
// @Tags 用户
// @Description 更新当前用户的昵称、简介和网站
// @Accept json
// @Produce json
// @Param body body object true "用户信息(nickname, intro, website)"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/users/info [put]
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

// @Summary 更新用户头像
// @Tags 用户
// @Description 上传并更新用户头像
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "头像文件"
// @Success 200 {object} object "成功返回头像URL"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/users/avatar [post]
func (h *UserAuthHandler) UpdateUserAvatar(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("请选择要上传的头像"))
		return
	}

	// 获取 t_user_info.id (对标Java UserUtil.getUserDetailsDTO().getUserInfoId())
	userInfoID, _ := c.Get("user_info_id")
	if userInfoID == nil {
		// 回退：从 Gin Context 获取 user_id (t_user_auth.id)，再查询 user_info_id
		userID := middleware.GetUserID(c)
		if userID == 0 {
			util.ResponseError(c, errors.ErrUnauthorized.WithMsg("无法获取用户ID"))
			return
		}
		var auth model.UserAuth
		if err := h.registry.DB.Select("user_info_id").Where("id = ?", userID).First(&auth).Error; err != nil {
			util.ResponseError(c, errors.ErrUserNotFound)
			return
		}
		userInfoID = auth.UserID
	}

	var uid uint
	switch v := userInfoID.(type) {
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

	// 使用 FileService 上传头像 (对标Java uploadStrategyContext.executeUploadStrategy)
	avatarURL, err := h.registry.File.UploadAvatar(c.Request.Context(), file, uid)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, avatarURL)
}

// @Summary 绑定用户邮箱
// @Tags 用户
// @Description 绑定当前用户的邮箱（需要验证码）
// @Accept json
// @Produce json
// @Param emailVO body dto.EmailVO true "邮箱和验证码"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/users/email [put]
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

// @Summary 修改用户订阅状态
// @Tags 用户
// @Description 修改当前用户的邮件订阅状态
// @Accept json
// @Produce json
// @Param body body object true "用户ID和订阅状态(userId, isSubscribe)"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/users/subscribe [put]
func (h *UserAuthHandler) UpdateUserSubscribe(c *gin.Context) {
	var body struct {
		UserID      uint `json:"userId" binding:"required"`
		IsSubscribe int8 `json:"isSubscribe"` // 移除 required,允许 0 值(对标Java Integer类型)
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}

	// 校验 IsSubscribe 值范围 (对标Java: 0或1)
	if body.IsSubscribe != 0 && body.IsSubscribe != 1 {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("订阅状态值无效，必须为0或1"))
		return
	}

	ctx := c.Request.Context()

	// Step 1: 检查用户是否存在并获取邮箱（对标Java userInfoMapper.selectOne）
	var email string
	err := h.registry.DB.WithContext(ctx).Model(&model.UserInfo{}).Select("email").Where("id = ?", body.UserID).First(&email).Error
	if err != nil {
		// 区分记录不存在和其他错误
		if err == gorm.ErrRecordNotFound {
			util.ResponseError(c, errors.ErrUserNotFound.WithMsg(fmt.Sprintf("用户ID %d 不存在", body.UserID)))
		} else {
			util.ResponseError(c, errors.ErrInternalServer.WithMsg("查询用户信息失败"))
		}
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

// @Summary 根据ID获取用户信息
// @Tags 用户
// @Description 根据ID获取用户信息
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} object "成功返回用户信息"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/users/info/{id} [get]
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
