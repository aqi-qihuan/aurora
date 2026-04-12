package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/model"
	"github.com/aurora-go/aurora/internal/util"
	"github.com/aurora-go/aurora/internal/vo"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserAuthService 用户认证业务逻辑 (对标 Java UserAuthServiceImpl + UserServiceImpl)
type UserAuthService struct {
	db *gorm.DB
}

func NewUserAuthService(db *gorm.DB) *UserAuthService {
	return &UserAuthService{db: db}
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
	// TODO: P0-6 清除Redis Session Key
	// redis.Del(constant.LoginUser + strconv.FormatUint(uint64(userID), 10))
	
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
	updates := make(map[string]interface{})
	
	if vo.Nickname != nil && *vo.Nickname != "" {
		updates["nickname"] = *vo.Nickname
	}
	if vo.Intro != nil && *vo.Intro != "" {
		updates["intro"] = *vo.Intro
	}
	if vo.WebSite != nil && *vo.WebSite != "" {
		updates["web_site"] = *vo.WebSite
	}
	if vo.Avatar != nil && *vo.Avatar != "" {
		updates["avatar"] = *vo.Avatar
	}

	result := s.db.WithContext(ctx).
		Model(&model.UserInfo{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("更新用户信息失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.ErrUserNotFound
	}

	slog.Info("用户信息更新", "user_id", id)
	return nil
}

// ChangePassword 修改密码
func (s *UserAuthService) ChangePassword(ctx context.Context, id uint, vo vo.PasswordVO) error {
	var auth model.UserAuth

	if err := s.db.WithContext(ctx).Where("user_id = ? AND login_type = ?", id, 1).First(&auth).Error; err != nil {
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

	slog.Info("密码修改成功", "user_id", id)
	return nil
}

// ListUsers 后台分页查询用户列表
func (s *UserAuthService) ListUsers(ctx context.Context, cond dto.ConditionVO, page dto.PageVO) (*dto.PageResultDTO, error) {
	var users []model.UserInfo
	var count int64

	baseQuery := s.db.WithContext(ctx).Model(&model.UserInfo{})

	if cond.Keywords != "" {
		baseQuery = baseQuery.Where("nickname LIKE ? OR email LIKE ?", "%"+cond.Keywords+"%", "%"+cond.Keywords+"%")
	}

	if err := baseQuery.Count(&count).Error; err != nil {
		return nil, fmt.Errorf("统计用户数失败: %w", err)
	}

	offset := page.GetOffset()
	if err := baseQuery.
		Preload("Roles").
		Order("create_time DESC").
		Limit(page.PageSize).
		Offset(offset).
		Find(&users).Error; err != nil {
		return nil, fmt.Errorf("查询用户列表失败: %w", err)
	}

	list := make([]dto.UserAdminDTO, len(users))
	for i, u := range users {
		list[i] = s.toUserAdminDTO(&u)
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
func (s *UserAuthService) SendVerificationCode(ctx context.Context, email string) error {
	code := util.GenerateCode(6)

	// TODO: 存入Redis并设置5分钟TTL
	// redis.SetEX(constant.UserAuthCode+email, 5*time.Minute, code)

	// TODO: P0-7 调用 EmailService.SendVerificationEmail(email, code)
	_ = code // 占位, 待集成邮件服务

	slog.Info("发送验证码", "email", email)
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
		ID:        u.ID,
		Email:     u.Email,
		Nickname:  u.Nickname,
		Avatar:    u.Avatar,
		Intro:     u.Intro,
		WebSite:   u.Website,
		IsDisable: u.IsDisable,
		CreateTime: u.CreateTime,
	}
}

func (s *UserAuthService) toUserAdminDTO(u *model.UserInfo) dto.UserAdminDTO {
	roles := make([]string, len(u.Roles))
	for i, r := range u.Roles {
		roles[i] = r.RoleName
	}

	return dto.UserAdminDTO{
		ID:        u.ID,
		Email:     u.Email,
		Nickname:  u.Nickname,
		Avatar:    u.Avatar,
		IsDisable: u.IsDisable,
		Roles:     roles,
		CreateTime: u.CreateTime,
	}
}
