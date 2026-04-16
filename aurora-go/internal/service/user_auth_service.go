package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/aurora-go/aurora/internal/constant"
	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/model"
	"github.com/aurora-go/aurora/internal/util"
	"github.com/aurora-go/aurora/internal/vo"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserAuthService 用户认证业务逻辑 (对标 Java UserAuthServiceImpl + UserServiceImpl)
type UserAuthService struct {
	db           *gorm.DB
	rdb          *redis.Client
	emailService interface{} // EmailService（延迟注入，避免循环依赖）
}

func NewUserAuthService(db *gorm.DB, rdb *redis.Client) *UserAuthService {
	return &UserAuthService{db: db, rdb: rdb}
}

// SetEmailService 设置邮件服务（延迟注入，避免循环依赖）
func (s *UserAuthService) SetEmailService(emailService interface{}) {
	s.emailService = emailService
}

// cleanIPAddress 清理 IP 地址中的特殊字符
// 处理：换行符、空格、尾部数字（如 "95.40.12.12 0" → "95.40.12.12"）
func cleanIPAddress(ip string) string {
	if ip == "" {
		return ""
	}
	
	// 1. 先按空格分割，取第一部分（处理 "95.40.12.12 0" 或 "110.184.180. \n10" 等情况）
	// Fields会自动处理多个空格、制表符等空白字符
	parts := strings.Fields(ip)
	if len(parts) > 0 {
		ip = parts[0]
	}
	
	// 2. 移除换行符、回车符和制表符（处理 "110.184.180.\n10" 这种情况）
	ip = strings.ReplaceAll(ip, "\n", "")
	ip = strings.ReplaceAll(ip, "\r", "")
	ip = strings.ReplaceAll(ip, "\t", "")
	
	// 3. 如果包含逗号，取第一部分（处理 X-Forwarded-For 格式）
	if idx := strings.Index(ip, ","); idx > 0 {
		ip = ip[:idx]
	}
	
	// 4. 最后移除首尾空格
	ip = strings.TrimSpace(ip)
	
	return ip
}

// Register 注册用户 (事务: UserInfo + UserAuth)
func (s *UserAuthService) Register(ctx context.Context, reg vo.RegisterVO) (*dto.UserDTO, error) {
	var user model.UserInfo
	var auth model.UserAuth

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 检查用户名是否已存在
		var count int64
		if tx.Model(&model.UserAuth{}).Where("username = ?", reg.Username).Count(&count); count > 0 {
			return errors.ErrUsernameExists
		}
		
		// 检查邮箱
		if tx.Model(&model.UserInfo{}).Where("email = ?", reg.Email).Count(&count); count > 0 {
			return errors.ErrEmailExists
		}

		// BCrypt密码加密 (cost=10, 对标Java BCryptPasswordEncoder)
		hashedPwd, err := bcrypt.GenerateFromPassword([]byte(reg.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("密码加密失败: %w", err)
		}

		// 创建用户信息
		user = model.UserInfo{
			Email:    reg.Email,
			Nickname: reg.Nickname,
		}
		if err := tx.Create(&user).Error; err != nil {
			return fmt.Errorf("创建用户信息失败: %w", err)
		}

		// 创建认证信息
		auth = model.UserAuth{
			UserID:   user.ID,
			Username: reg.Username,
			Password: string(hashedPwd),
			LoginType: 1, // 邮箱登录
		}
		if err := tx.Create(&auth).Error; err != nil {
			return fmt.Errorf("创建认证信息失败: %w", err)
		}

		// 分配默认角色(普通用户)
		var defaultRole model.Role
		if err := tx.Where("is_default = ?", 1).First(&defaultRole).Error; err == nil {
			tx.Exec("INSERT INTO t_user_role(user_id, role_id) VALUES (?, ?)", user.ID, defaultRole.ID)
		}

		slog.Info("用户注册成功", "user_id", user.ID, "username", reg.Username)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.toUserDTO(&user), nil
}

// Login 登录验证 (对标 Java LoginStrategy + TokenService)
func (s *UserAuthService) Login(ctx context.Context, login vo.LoginVO) (*dto.LoginVO, error) {
	// 第一步: 快速查询 UserAuth (不关联 UserInfo)
	var auth model.UserAuth
	if err := s.db.WithContext(ctx).
		Select("id", "user_info_id", "username", "password", "login_type").
		Where("username = ? AND login_type = ?", login.Username, 1).
		First(&auth).Error; err != nil {
		if errors.IsStd(err, gorm.ErrRecordNotFound) {
			slog.Warn("登录失败: 用户不存在", "username", login.Username)
			return nil, errors.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("查询认证失败: %w", err)
	}

	// 第二步: 验证BCrypt密码 (失败则直接返回，无需查询 UserInfo)
	if err := bcrypt.CompareHashAndPassword([]byte(auth.Password), []byte(login.Password)); err != nil {
		slog.Warn("登录失败: 密码错误", "username", login.Username)
		return nil, errors.ErrInvalidCredentials
	}

	// 第三步: 仅当密码正确时才查询 UserInfo + Roles
	var userInfo model.UserInfo
	if err := s.db.WithContext(ctx).
		Preload("Roles").
		First(&userInfo, auth.UserID).Error; err != nil {
		slog.Error("登录失败: 用户信息缺失", "user_id", auth.UserID)
		return nil, errors.ErrInternalServer.WithMsg("用户数据异常，请联系管理员")
	}

	// 检查账户状态
	if userInfo.IsDisable == 1 {
		return nil, errors.ErrAccountDisabled
	}

	// TODO: P0-6 JWT签发 → 返回Token
	token, _ := util.GenerateRandomString(32) // 临时token, P0-6替换为JWT

	// 异步更新最后登录时间 (不阻塞登录响应)
	go func(userID uint) {
		s.db.Model(&model.UserInfo{}).Where("id = ?", userID).Update("update_time", time.Now())
	}(auth.UserID)

	slog.Info("用户登录成功", "user_id", auth.UserID, "username", login.Username)

	// 构造最后登录时间字符串 (ISO8601格式, 对标Java LocalDateTime序列化)
	lastLoginTime := ""
	if auth.LastLoginTime != nil {
		lastLoginTime = auth.LastLoginTime.Format("2006-01-02T15:04:05")
	}

	// 完全对标Java版: 返回UserInfoDTO (从UserDetailsDTO拷贝)
	return &dto.LoginVO{
		ID:            auth.ID,
		UserInfoID:    auth.UserID,
		Email:         userInfo.Email,
		LoginType:     int(auth.LoginType),
		Username:      auth.Username,
		Nickname:      userInfo.Nickname,
		Avatar:        userInfo.Avatar,
		Intro:         userInfo.Intro,
		Website:       userInfo.Website,
		IsSubscribe:   userInfo.IsSubscribe,
		IPAddress:     auth.IPAddress,
		IPSource:      auth.IPSource,
		LastLoginTime: lastLoginTime,
		Token:         token,
	}, nil
}

// Logout 登出 (清除Redis Session)
func (s *UserAuthService) Logout(ctx context.Context, userID uint) error {
	// 删除 Redis Session (对标Java tokenService.delLoginUser)
	if s.rdb != nil {
		if err := s.rdb.HDel(ctx, constant.LoginUser, fmt.Sprintf("%d", userID)).Err(); err != nil {
			slog.Warn("清除Redis Session失败", "userId", userID, "error", err.Error())
		}
	}

	slog.Info("用户登出", "user_id", userID)
	return nil
}

// GetUserInfoByID 根据ID获取用户详情
func (s *UserAuthService) GetUserInfoByID(ctx context.Context, id uint) (*dto.UserDTO, error) {
	var user model.UserInfo
	
	err := s.db.WithContext(ctx).
		Preload("Roles").
		First(&user, id).Error

	if err != nil {
		if errors.IsStd(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrUserNotFound
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	return s.toUserDTO(&user), nil
}

// UpdateUserInfo 更新用户信息
func (s *UserAuthService) UpdateUserInfo(ctx context.Context, id uint, vo vo.UpdateUserVO) error {
	// Step 1: 先检查用户是否存在
	var userInfo model.UserInfo
	if err := s.db.WithContext(ctx).First(&userInfo, id).Error; err != nil {
		if errors.IsStd(err, gorm.ErrRecordNotFound) {
			return errors.ErrUserNotFound
		}
		return fmt.Errorf("查询用户失败: %w", err)
	}

	// Step 2: 构建更新字段
	updates := make(map[string]interface{})
	
	if vo.Nickname != nil && *vo.Nickname != "" {
		updates["nickname"] = *vo.Nickname
	}
	if vo.Intro != nil && *vo.Intro != "" {
		updates["intro"] = *vo.Intro
	}
	if vo.WebSite != nil && *vo.WebSite != "" {
		updates["website"] = *vo.WebSite
	}
	if vo.Avatar != nil && *vo.Avatar != "" {
		updates["avatar"] = *vo.Avatar
	}

	// 如果没有需要更新的字段，直接返回
	if len(updates) == 0 {
		return nil
	}

	// Step 3: 执行更新
	result := s.db.WithContext(ctx).
		Model(&model.UserInfo{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("更新用户信息失败: %w", result.Error)
	}

	slog.Info("用户信息更新成功", "user_id", id, "updated_fields", updates, "rows_affected", result.RowsAffected)
	return nil
}

// ChangePassword 修改密码（用户端，需验证旧密码）
func (s *UserAuthService) ChangePassword(ctx context.Context, id uint, vo vo.PasswordVO) error {
	var auth model.UserAuth

	// id 是 UserAuth.id（从JWT的user_id获取），直接按主键查
	if err := s.db.WithContext(ctx).First(&auth, id).Error; err != nil {
		slog.Warn("查询UserAuth失败", "id", id, "error", err.Error())
		return errors.ErrUserNotFound
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(auth.Password), []byte(vo.OldPassword)); err != nil {
		return errors.ErrInvalidOldPassword
	}

	// 加密新密码
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(vo.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("新密码加密失败: %w", err)
	}

	if err := s.db.WithContext(ctx).Model(&auth).Update("password", string(hashedPwd)).Error; err != nil {
		return fmt.Errorf("修改密码失败: %w", err)
	}

	slog.Info("密码修改成功", "user_info_id", id)
	return nil
}

// ResetPassword 通过邮箱验证码重置密码（对标Java UserAuthServiceImpl.updatePassword）
// Java 流程: 校验验证码 → 检查邮箱是否存在(t_user_auth.username) → 直接更新密码
func (s *UserAuthService) ResetPassword(ctx context.Context, email string, code string, newPassword string) error {
	// 1. 从 Redis 获取验证码
	if s.rdb == nil {
		return errors.ErrInternalServer.WithMsg("系统异常")
	}

	redisKey := constant.UserAuthCode + email
	storedCode, err := s.rdb.Get(ctx, redisKey).Result()
	if err != nil {
		if errors.IsStd(err, redis.Nil) {
			return errors.ErrInvalidParams.WithMsg("验证码已过期或不存在")
		}
		slog.Error("读取验证码失败", "error", err)
		return errors.ErrInternalServer.WithMsg("系统异常")
	}

	// 2. 校验验证码
	if storedCode != code {
		return errors.ErrInvalidParams.WithMsg("验证码错误")
	}

	// 3. 删除已使用的验证码（一次性有效）
	s.rdb.Del(ctx, redisKey)

	// 4. 检查邮箱是否已注册（对标Java checkUser: t_user_auth.username = email）
	var auth model.UserAuth
	if err := s.db.WithContext(ctx).
		Select("id").
		Where("username = ? AND login_type = 1", email).
		First(&auth).Error; err != nil {
		if errors.IsStd(err, gorm.ErrRecordNotFound) {
			return errors.ErrInvalidParams.WithMsg("邮箱尚未注册")
		}
		return fmt.Errorf("查询用户失败: %w", err)
	}

	// 5. BCrypt加密新密码
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 6. 直接更新密码（对标Java: userAuthMapper.update WHERE username = email）
	if err := s.db.WithContext(ctx).
		Model(&model.UserAuth{}).
		Where("username = ?", email).
		Update("password", string(hashedPwd)).Error; err != nil {
		return fmt.Errorf("修改密码失败: %w", err)
	}

	slog.Info("密码重置成功", "email", email, "user_auth_id", auth.ID)
	return nil
}

// UpdateAdminPassword 管理员修改密码（后台，不需要旧密码）
// 对标Java: UserAuthServiceImpl.updateAdminPassword - 管理员直接重置密码，不需要验证旧密码
func (s *UserAuthService) UpdateAdminPassword(ctx context.Context, id uint, newPassword string) error {
	var auth model.UserAuth

	// id 是 UserAuth.id（从JWT的user_id获取），直接按主键查
	if err := s.db.WithContext(ctx).First(&auth, id).Error; err != nil {
		return errors.ErrUserNotFound
	}

	// 直接加密新密码（不需要验证旧密码）
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("新密码加密失败: %w", err)
	}

	if err := s.db.WithContext(ctx).Model(&auth).Update("password", string(hashedPwd)).Error; err != nil {
		return fmt.Errorf("修改密码失败: %w", err)
	}

	slog.Info("管理员密码修改成功", "user_info_id", id)
	return nil
}

// ListUsers 后台分页查询用户列表（完全对标Java UserAuthMapper.listUsers）
// SQL逻辑：从 t_user_info 筛选 → LEFT JOIN t_user_auth + t_user_role + t_role
// 支持 loginType 和 keywords 条件过滤
func (s *UserAuthService) ListUsers(ctx context.Context, cond dto.ConditionVO, page dto.PageVO) (*dto.PageResultDTO, error) {
	type UserRow struct {
		ID            uint
		UserInfoId    uint
		Avatar        string
		Nickname      string
		LoginType     int8
		IpAddress     string
		IpSource      string
		CreateTime    time.Time
		LastLoginTime *time.Time
		IsDisable     int8
		RoleId        uint
		RoleName      string
	}

	// 构建子查询：从 t_user_info 中根据条件筛选
	baseQuery := s.db.WithContext(ctx).Table("t_user_info ui").
		Select("ui.id, ui.avatar, ui.nickname, ui.is_disable")

	// 条件过滤：loginType
	if cond.LoginType != nil && *cond.LoginType > 0 {
		baseQuery = baseQuery.Where("ui.id IN (SELECT user_info_id FROM t_user_auth WHERE login_type = ?)", *cond.LoginType)
	}

	// 条件过滤：keywords（昵称模糊搜索）
	if cond.Keywords != "" {
		baseQuery = baseQuery.Where("ui.nickname LIKE ?", "%"+cond.Keywords+"%")
	}

	// 分页
	offset := page.GetOffset()
	baseQuery = baseQuery.Limit(page.PageSize).Offset(offset)

	// 外层查询：LEFT JOIN 获取 user_auth 和 role 信息
	type ResultRow struct {
		ID            uint
		UserInfoId    uint
		Avatar        string
		Nickname      string
		LoginType     int8
		IpAddress     string
		IpSource      string
		CreateTime    time.Time
		LastLoginTime *time.Time
		IsDisable     int8
		RoleId        *uint
		RoleName      *string
	}

	var rows []ResultRow
	err := s.db.WithContext(ctx).
		Table("(?) ui", baseQuery).
		Select(`
			ua.id,
			ui.id as user_info_id,
			ui.avatar,
			ui.nickname,
			ua.login_type,
			ua.ip_address,
			ua.ip_source,
			ua.create_time,
			ua.last_login_time,
			ui.is_disable,
			r.id as role_id,
			r.role_name
		`).
		Joins("LEFT JOIN t_user_auth ua ON ua.user_info_id = ui.id").
		Joins("LEFT JOIN t_user_role ur ON ui.id = ur.user_id").
		Joins("LEFT JOIN t_role r ON ur.role_id = r.id").
		Find(&rows).Error

	if err != nil {
		return nil, fmt.Errorf("查询用户列表失败: %w", err)
	}

	// 统计总数（对标Java countUser）
	var count int64
	countQuery := s.db.WithContext(ctx).Table("t_user_auth ua").
		Joins("LEFT JOIN t_user_info ui ON ua.user_info_id = ui.id")

	if cond.Keywords != "" {
		countQuery = countQuery.Where("ui.nickname LIKE ?", "%"+cond.Keywords+"%")
	}
	if cond.LoginType != nil && *cond.LoginType > 0 {
		countQuery = countQuery.Where("ua.login_type = ?", *cond.LoginType)
	}
	countQuery.Count(&count)

	if count == 0 {
		return &dto.PageResultDTO{
			List:     []dto.UserAdminDTO{},
			Count:    0,
			PageNum:  page.PageNum,
			PageSize: page.PageSize,
		}, nil
	}

	// 将结果按 userInfoId 聚合，收集角色信息
	userMap := make(map[uint]*dto.UserAdminDTO)
	var order []uint // 保持顺序

	for _, row := range rows {
		if _, exists := userMap[row.UserInfoId]; !exists {
			// 处理 CreateTime 零值问题
			var createTime *time.Time
			if !row.CreateTime.IsZero() {
				createTime = &row.CreateTime
			}
			
			// 清理 IP 地址中的换行符、空格和尾部数字
			ipAddress := cleanIPAddress(row.IpAddress)
			ipSource := cleanIPAddress(row.IpSource)
			
			userMap[row.UserInfoId] = &dto.UserAdminDTO{
				ID:            row.ID,
				UserInfoId:    row.UserInfoId,
				Avatar:        row.Avatar,
				Nickname:      row.Nickname,
				LoginType:     row.LoginType,
				IpAddress:     ipAddress,
				IpSource:      ipSource,
				CreateTime:    createTime,
				LastLoginTime: row.LastLoginTime,
				IsDisable:     row.IsDisable,
				Roles:         []dto.UserRoleDTO{},
			}
			order = append(order, row.UserInfoId)
		}
		if row.RoleName != nil && *row.RoleName != "" {
			userMap[row.UserInfoId].Roles = append(userMap[row.UserInfoId].Roles, dto.UserRoleDTO{
				ID:       *row.RoleId,
				RoleName: *row.RoleName,
			})
		}
	}

	// 构建有序列表
	list := make([]dto.UserAdminDTO, 0, len(order))
	for _, uid := range order {
		list = append(list, *userMap[uid])
	}

	return &dto.PageResultDTO{
		List:     list,
		Count:    count,
		PageNum:  page.PageNum,
		PageSize: page.PageSize,
	}, nil
}

// QQLoginOrRegister QQ OAuth登录 (自动注册或关联账号)
func (s *UserAuthService) QQLoginOrRegister(ctx context.Context, qqInfo vo.QQLoginVO) (*dto.LoginVO, error) {
	var userAuth model.UserAuth

	err := s.db.WithContext(ctx).
		Where("username = ? AND login_type = ?", qqInfo.OpenID, 2). // 适配表结构，OpenID 存入 username
		First(&userAuth).Error

	if err == nil {
		// 已有QQ绑定 → 直接登录
		s.db.Preload("UserInfo").First(&userAuth, userAuth.ID)
		token, _ := util.GenerateRandomString(32) // TODO: P0-6 替换JWT
		
		// 构造最后登录时间字符串 (ISO8601格式)
		lastLoginTime := ""
		if userAuth.LastLoginTime != nil {
			lastLoginTime = userAuth.LastLoginTime.Format("2006-01-02T15:04:05")
		}
		
		// 完全对标Java版: 返回完整用户信息
		return &dto.LoginVO{
			ID:            userAuth.ID,
			UserInfoID:    userAuth.UserID,
			Email:         userAuth.UserInfo.Email,
			LoginType:     int(userAuth.LoginType),
			Username:      userAuth.Username,
			Nickname:      userAuth.UserInfo.Nickname,
			Avatar:        userAuth.UserInfo.Avatar,
			Intro:         userAuth.UserInfo.Intro,
			Website:       userAuth.UserInfo.Website,
			IsSubscribe:   userAuth.UserInfo.IsSubscribe,
			IPAddress:     userAuth.IPAddress,
			IPSource:      userAuth.IPSource,
			LastLoginTime: lastLoginTime,
			Token:         token,
		}, nil
	}

	// 新QQ用户 → 自动注册
	var user model.UserInfo
	txErr := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		user = model.UserInfo{
			Nickname: qqInfo.Nickname,
			Avatar:   qqInfo.Avatar,
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		newAuth := model.UserAuth{
			UserID:    user.ID,
			LoginType: 2,
		}
		if err := tx.Create(&newAuth).Error; err != nil {
			return err
		}

		// 默认角色
		var defaultRole model.Role
		if err := tx.Where("is_default = ?", 1).First(&defaultRole).Error; err == nil {
			tx.Exec("INSERT INTO t_user_role(user_id, role_id) VALUES (?, ?)", user.ID, defaultRole.ID)
		}

		slog.Info("QQ用户注册成功", "user_id", user.ID, "openid", qqInfo.OpenID[:20]+"...")
		return nil
	})

	if txErr != nil {
		return nil, fmt.Errorf("QQ注册失败: %w", txErr)
	}

	token, _ := util.GenerateRandomString(32)
	
	// 构造最后登录时间字符串 (ISO8601格式)
	lastLoginTime := ""
	if userAuth.LastLoginTime != nil {
		lastLoginTime = userAuth.LastLoginTime.Format("2006-01-02T15:04:05")
	}
	
	// 完全对标Java版: 返回完整用户信息
	return &dto.LoginVO{
		ID:            userAuth.ID,
		UserInfoID:    user.ID,
		Email:         user.Email,
		LoginType:     int(userAuth.LoginType),
		Username:      userAuth.Username,
		Nickname:      user.Nickname,
		Avatar:        user.Avatar,
		Intro:         user.Intro,
		Website:       user.Website,
		IsSubscribe:   user.IsSubscribe,
		IPAddress:     userAuth.IPAddress,
		IPSource:      userAuth.IPSource,
		LastLoginTime: lastLoginTime,
		Token:         token,
	}, nil
}

// SendVerificationCode 发送验证码 (邮箱验证码)
// 对标 Java UserAuthServiceImpl.sendCode
func (s *UserAuthService) SendVerificationCode(ctx context.Context, email string) error {
	// 1. 生成6位随机验证码
	code := util.GenerateCode(6)

	// 2. 存入 Redis，设置5分钟TTL（对标Java redisService.setEX(USER_AUTH_CODE + email, 5, code)）
	if s.rdb != nil {
		redisKey := constant.UserAuthCode + email
		if err := s.rdb.SetEx(ctx, redisKey, code, 5*time.Minute).Err(); err != nil {
			slog.Error("验证码存入Redis失败", "email", email, "error", err)
			return errors.ErrInternalServer.WithMsg("验证码存储失败")
		}
	} else {
		slog.Warn("Redis未初始化，无法存储验证码")
		return errors.ErrInternalServer.WithMsg("系统异常")
	}

	// 3. 调用 EmailService 发送邮件
	if s.emailService != nil {
		// 类型断言获取 SendVerificationCode 函数
		type EmailServiceInterface interface {
			SendVerificationCode(toEmail string, code string) error
		}
		if emailSvc, ok := s.emailService.(EmailServiceInterface); ok {
			if err := emailSvc.SendVerificationCode(email, code); err != nil {
				slog.Error("发送验证码邮件失败", "email", email, "error", err)
				// 删除 Redis 中的验证码（避免脏数据）
				s.rdb.Del(ctx, constant.UserAuthCode+email)
				return errors.ErrInternalServer.WithMsg("邮件发送失败，请稍后重试")
			}
		} else {
			slog.Warn("EmailService 接口不匹配")
			return errors.ErrInternalServer.WithMsg("邮件服务异常")
		}
	} else {
		slog.Warn("EmailService 未注入，跳过邮件发送")
		// 开发环境可以返回成功，生产环境应报错
		return errors.ErrInternalServer.WithMsg("邮件服务未配置")
	}

	slog.Info("验证码发送成功", "email", email)
	return nil
}

// FindOrCreateBySocialLogin 根据社交登录信息查找或创建用户 (对标Java AbstractSocialLoginStrategyImpl)
// 用于QQ/GitHub/WeChat等第三方OAuth登录
//
// 流程:
//  1. 通过 openID + loginType 查找 UserAuth 记录
//  2. 存在 → 更新最后登录时间，返回用户详情
//  3. 不存在 → 自动注册新用户(UserInfo + UserAuth + 默认角色)，返回用户详情
func (s *UserAuthService) FindOrCreateBySocialLogin(
	ctx context.Context,
	openID string,
	loginType int,
	nickname string,
	avatar string,
	token string,
) (*dto.UserInfoDTO, error) {
	var userAuth model.UserAuth

	// Step 1: 查找已有账号
	err := s.db.WithContext(ctx).
		Where("username = ? AND login_type = ?", openID, loginType).
		First(&userAuth).Error

	if err == nil {
		// 已有账号 → 更新登录时间并返回
		now := time.Now()
		s.db.WithContext(ctx).Model(&userAuth).Updates(map[string]interface{}{
			"last_login_time": now,
			"password":        token,
		})

		var userInfo model.UserInfo
		s.db.WithContext(ctx).First(&userInfo, userAuth.UserID)

		return &dto.UserInfoDTO{
			ID:       userInfo.ID,
			Nickname: userInfo.Nickname,
			Avatar:   userInfo.Avatar,
		}, nil
	}

	// Step 2: 新用户 → 自动注册 (事务)
	var userInfo model.UserInfo
	txErr := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 创建用户基本信息
		userInfo = model.UserInfo{
			Nickname: nickname,
			Avatar:   avatar,
		}
		if err := tx.Create(&userInfo).Error; err != nil {
			return fmt.Errorf("创建UserInfo失败: %w", err)
		}

		// 创建认证记录
		newAuth := model.UserAuth{
			UserID:      userInfo.ID,
			Username:    openID,
			Password:    token,
			LoginType:   int8(loginType),
			LastLoginTime: func() *time.Time { t := time.Now(); return &t }(),
		}
		if err := tx.Create(&newAuth).Error; err != nil {
			return fmt.Errorf("创建UserAuth失败: %w", err)
		}

		// 分配默认角色
		var defaultRole model.Role
		if err := tx.Where("is_default = ?", 1).First(&defaultRole).Error; err == nil {
			tx.Exec("INSERT INTO t_user_role(user_id, role_id) VALUES (?, ?)", userInfo.ID, defaultRole.ID)
		}

		slog.Info("社交登录自动注册",
			"user_id", userInfo.ID,
			"type", loginType,
			"nickname", nickname,
		)
		return nil
	})

	if txErr != nil {
		return nil, fmt.Errorf("社交注册失败: %w", txErr)
	}

	return &dto.UserInfoDTO{
		ID:       userInfo.ID,
		Nickname: userInfo.Nickname,
		Avatar:   userInfo.Avatar,
	}, nil
}

// ===== DTO转换 =====

func (s *UserAuthService) toUserDTO(u *model.UserInfo) *dto.UserDTO {
	return &dto.UserDTO{
		ID:         u.ID,
		Email:      u.Email,
		Nickname:   u.Nickname,
		Avatar:     u.Avatar,
		Intro:      u.Intro,
		WebSite:    u.Website,
		IsDisable:  u.IsDisable,
		CreateTime: u.CreateTime,
	}
}

// ListOnlineUsers 查看在线用户列表（对标Java UserInfoServiceImpl.listOnlineUsers）
// 数据源：Redis Hash login_user（存储 UserDetailsDTO）
func (s *UserAuthService) ListOnlineUsers(ctx context.Context, cond dto.ConditionVO, page dto.PageVO) (*dto.PageResultDTO, error) {
	if s.rdb == nil {
		return &dto.PageResultDTO{
			List:     []dto.UserOnlineDTO{},
			Count:    0,
			PageNum:  page.PageNum,
			PageSize: page.PageSize,
		}, nil
	}

	// 从 Redis Hash 获取所有在线用户 (对标Java redisService.hGetAll(LOGIN_USER))
	userMaps, err := s.rdb.HGetAll(ctx, constant.LoginUser).Result()
	if err != nil {
		slog.Warn("获取在线用户失败", "error", err.Error())
		return &dto.PageResultDTO{
			List:     []dto.UserOnlineDTO{},
			Count:    0,
			PageNum:  page.PageNum,
			PageSize: page.PageSize,
		}, nil
	}

	// 解析 UserDetailsDTO
	var userDetailsList []*dto.UserDetailsDTO
	for _, dataStr := range userMaps {
		if dataStr == "" {
			continue
		}
		var user dto.UserDetailsDTO
		if err := json.Unmarshal([]byte(dataStr), &user); err != nil {
			slog.Warn("解析用户Session失败", "error", err.Error())
			continue
		}
		userDetailsList = append(userDetailsList, &user)
	}

	// 转换为 UserOnlineDTO (对标Java BeanCopyUtil.copyList)
	var onlineUsers []dto.UserOnlineDTO
	for _, user := range userDetailsList {
		onlineUsers = append(onlineUsers, dto.UserOnlineDTO{
			UserInfoId:    user.UserInfoID,
			Nickname:      user.Nickname,
			Avatar:        user.Avatar,
			IpAddress:     user.IPAddress,
			IpSource:      user.IPSource,
			Browser:       user.Browser,
			Os:            user.OS,
			LastLoginTime: &user.LastLoginTime,
		})
	}

	// 过滤：keywords (对标Java StringUtils.isBlank(conditionVO.getKeywords()) || item.getNickname().contains(...))
	if cond.Keywords != "" {
		filtered := make([]dto.UserOnlineDTO, 0)
		for _, user := range onlineUsers {
			if strings.Contains(user.Nickname, cond.Keywords) {
				filtered = append(filtered, user)
			}
		}
		onlineUsers = filtered
	}

	// 排序：按 lastLoginTime 倒序 (对标Java sorted(Comparator.comparing(UserOnlineDTO::getLastLoginTime).reversed()))
	sort.Slice(onlineUsers, func(i, j int) bool {
		if onlineUsers[i].LastLoginTime == nil && onlineUsers[j].LastLoginTime == nil {
			return false
		}
		if onlineUsers[i].LastLoginTime == nil {
			return false
		}
		if onlineUsers[j].LastLoginTime == nil {
			return true
		}
		return onlineUsers[i].LastLoginTime.After(*onlineUsers[j].LastLoginTime)
	})

	// 手动分页 (对标Java subList)
	totalCount := len(onlineUsers)
	fromIndex := page.GetOffset()
	toIndex := fromIndex + page.PageSize
	if toIndex > totalCount {
		toIndex = totalCount
	}

	var pagedList []dto.UserOnlineDTO
	if fromIndex < totalCount {
		pagedList = onlineUsers[fromIndex:toIndex]
	} else {
		pagedList = []dto.UserOnlineDTO{}
	}

	return &dto.PageResultDTO{
		List:     pagedList,
		Count:    int64(totalCount),
		PageNum:  page.PageNum,
		PageSize: page.PageSize,
	}, nil
}

// RemoveOnlineUser 下线指定用户（对标Java UserInfoServiceImpl.removeOnlineUser）
// 流程：1. 根据 userInfoId 查询 userId  2. 删除 Redis Session
func (s *UserAuthService) RemoveOnlineUser(ctx context.Context, userInfoId uint) error {
	// Step 1: 根据 userInfoId 查询 userId (对标Java userAuthMapper.selectOne)
	var auth model.UserAuth
	if err := s.db.WithContext(ctx).
		Select("id").
		Where("user_info_id = ?", userInfoId).
		First(&auth).Error; err != nil {
		if errors.IsStd(err, gorm.ErrRecordNotFound) {
			return errors.ErrUserNotFound
		}
		return fmt.Errorf("查询用户认证信息失败: %w", err)
	}

	// Step 2: 删除 Redis Session (对标Java tokenService.delLoginUser(userId))
	if s.rdb != nil {
		if err := s.rdb.HDel(ctx, constant.LoginUser, fmt.Sprintf("%d", auth.ID)).Err(); err != nil {
			slog.Warn("删除Redis Session失败", "userId", auth.ID, "error", err.Error())
			// 不返回错误，继续执行
		}
	}

	slog.Info("用户已下线", "userInfoId", userInfoId, "userId", auth.ID)
	return nil
}
