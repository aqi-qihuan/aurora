package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/aurora-go/aurora/internal/config"
	"github.com/aurora-go/aurora/internal/constant"
	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/util"
)

// QQOAuthService QQ OAuth登录服务 (对标Java QQLoginStrategyImpl)
// 完整OAuth2.0流程: 获取授权URL → 回调验证Token → 获取用户信息 → 自动注册/登录
type QQOAuthService struct {
	qqCfg  config.QQConfig
	userSvc *UserAuthService
	tokenSvc *TokenService
	logger  *slog.Logger
	client  *http.Client
}

// NewQQOAuthService 创建QQ OAuth服务实例
func NewQQOAuthService(
	qqCfg config.QQConfig,
	userSvc *UserAuthService,
	tokenSvc *TokenService,
	logger *slog.Logger,
) *QQOAuthService {
	return &QQOAuthService{
		qqCfg:   qqCfg,
		userSvc: userSvc,
		tokenSvc: tokenSvc,
		logger:  logger,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// ===== 对外接口 (对标Java SocialLoginStrategy) =====

// GetAuthorizationURL 生成QQ OAuth授权URL (前端跳转用)
func (s *QQOAuthService) GetAuthorizationURL(redirectURI string) string {
	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", s.qqCfg.AppID)
	params.Set("redirect_uri", redirectURI)
	params.Set("state", generateState())
	return fmt.Sprintf("%s?%s",
		"https://graph.qq.com/oauth2.0/authorize",
		params.Encode(),
	)
}

// Login 执行完整的QQ OAuth登录流程 (对标Java AbstractSocialLoginStrategyImpl.login)
//
// 流程:
//  1. 验证accessToken和openId的合法性 (调用QQ API校验)
//  2. 根据openId查找本地用户记录
//  3. 存在 → 更新最后登录时间，生成Token
//  4. 不存在 → 自动注册新用户(默认角色=USER)，生成Token
//
// 返回: UserInfoDTO(含Token), 错误
func (s *QQOAuthService) Login(ctx context.Context, req *dto.QQLoginVO) (*dto.UserInfoDTO, error) {
	// Step 1: 校验Token有效性 (对标Java checkQQToken)
	if err := s.checkAccessToken(req.AccessToken, req.OpenID); err != nil {
		return nil, fmt.Errorf("QQ Token校验失败: %w", err)
	}

	// Step 2: 获取QQ用户信息 (对标Java getSocialUserInfo)
	qqUser, err := s.getQQUserInfo(req.AccessToken, req.OpenID)
	if err != nil {
		return nil, fmt.Errorf("获取QQ用户信息失败: %w", err)
	}

	// 优先使用QQ空间100x100头像, 其次普通头像
	qqAvatar := qqUser.FigureurlQQ1
	if qqAvatar == "" {
		qqAvatar = qqUser.Figureurl_1
	}
	if qqAvatar == "" {
		qqAvatar = qqUser.Figureurl
	}

	s.logger.Info("QQ OAuth登录",
		"open_id", req.OpenID,
		"nickname", qqUser.Nickname,
	)

	// Step 3: 查找或创建本地用户 (对标Java getUserAuth / saveUserDetail)
	userInfoDTO, err := s.userSvc.FindOrCreateBySocialLogin(
		ctx,
		req.OpenID,
		constant.LoginTypeQQ,
		qqUser.Nickname,
		qqAvatar,
		req.AccessToken,
	)
	if err != nil {
		return nil, fmt.Errorf("用户注册/登录处理失败: %w", err)
	}

	// Step 4: 生成JWT Token
	userDetails := &dto.UserDetailsDTO{
		ID:       userInfoDTO.ID,
		Nickname: userInfoDTO.Nickname,
		Avatar:   userInfoDTO.Avatar,
		Email:    userInfoDTO.Email,
	}
	token, err := s.tokenSvc.CreateToken(userDetails)
	if err != nil {
		return nil, fmt.Errorf("生成Token失败: %w", err)
	}
	userInfoDTO.Token = token

	return userInfoDTO, nil
}

// ===== 内部方法 =====

// checkAccessToken 验证AccessToken+OpenID是否匹配 (对标Java checkQQToken)
func (s *QQOAuthService) checkAccessToken(accessToken, openID string) error {
	checkURL := fmt.Sprintf("%s?%s",
		s.qqCfg.CheckTokenURL,
		url.Values{"access_token": {accessToken}}.Encode(),
	)

	resp, err := s.client.Get(checkURL)
	if err != nil {
		return fmt.Errorf("请求QQ API失败: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("解析QQ响应失败: %w", err)
	}

	// QQ API返回格式: callback( {"client_id":"xxx","openid":"xxx"} );
	// 检查返回的openId是否与传入的一致
	respOpenID, ok := result["openid"].(string)
	if !ok || respOpenID != openID {
		return fmt.Errorf("OpenID不匹配: expected=%s, got=%s", openID, respOpenID)
	}

	return nil
}

// getQQUserInfo 获取QQ用户信息 (对标Java getSocialUserInfo)
func (s *QQOAuthService) getQQUserInfo(accessToken, openID string) (*dto.QQUserInfoDTO, error) {
	userInfoURL := fmt.Sprintf(
		"%s?%s",
		s.qqCfg.UserInfoURL,
		url.Values{
			"oauth_consumer_key": {s.qqCfg.AppID},
			"openid":             {openID},
			"access_token":       {accessToken},
		}.Encode(),
	)

	resp, err := s.client.Get(userInfoURL)
	if err != nil {
		return nil, fmt.Errorf("请求QQ用户信息API失败: %w", err)
	}
	defer resp.Body.Close()

	var qqUser dto.QQUserInfoDTO
	if err := json.NewDecoder(resp.Body).Decode(&qqUser); err != nil {
		return nil, fmt.Errorf("解析QQ用户信息失败: %w", err)
	}

	if qqUser.Nickname == "" {
		return nil, fmt.Errorf("QQ用户昵称为空")
	}

	return &qqUser, nil
}

// generateState 生成OAuth state参数防CSRF攻击
func generateState() string {
	return util.GenerateRandomStringSimple(16)
}
