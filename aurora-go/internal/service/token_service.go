package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/aurora-go/aurora/internal/config"
	"github.com/aurora-go/aurora/internal/constant"
	"github.com/aurora-go/aurora/internal/dto"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

// TokenService JWT Token 管理服务 (对标Java TokenServiceImpl)
// 职责: Token签发/解析/验证/续期/Redis Session管理
type TokenService struct {
	cfg    config.JWTConfig
	rdb    *redis.Client
	logger *slog.Logger
}

// NewTokenService 创建TokenService实例
func NewTokenService(cfg config.JWTConfig, rdb *redis.Client, logger *slog.Logger) *TokenService {
	return &TokenService{
		cfg:    cfg,
		rdb:    rdb,
		logger: logger,
	}
}

// ===== 核心方法 (对标Java TokenServiceImpl) =====

// CreateToken 生成JWT Token并存入Redis Session
func (s *TokenService) CreateToken(userDetails *dto.UserDetailsDTO) (string, error) {
	s.RefreshToken(userDetails) // 先存Redis

	tokenString, err := s.createTokenBySubject(fmt.Sprintf("%d", userDetails.ID))
	if err != nil {
		return "", fmt.Errorf("create jwt token: %w", err)
	}
	return tokenString, nil
}

// createTokenBySubject 根据subject生成纯JWT字符串
func (s *TokenService) createTokenBySubject(subject string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		constant.JwtClaimUserID:   subject,
		"jti":                    generateUUID(),
		"iss":                     s.cfg.Issuer,
		constant.JwtClaimIssueAt: now.Unix(),
		constant.JwtClaimExpireAt: now.Add(time.Duration(s.cfg.ExpireTime) * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.Secret))
}

// RefreshToken 将用户详情写入Redis Hash (对标Java refreshToken)
// Redis结构: login_user → Hash, field=userId, value=UserDetailsDTO JSON
func (s *TokenService) RefreshToken(userDetails *dto.UserDetailsDTO) {
	ctx := context.Background()
	userDetails.ExpireTime = time.Now().Add(constant.TokenExpireTime * time.Second)
	userId := fmt.Sprintf("%d", userDetails.ID)

	data, err := json.Marshal(userDetails)
	if err != nil {
		s.logger.Error("序列化用户Session失败", "userId", userId, "error", err)
		return
	}

	// 使用HSet存储到Hash (对标Java redisService.hSet(LOGIN_USER, userId, userDetails, EXPIRE_TIME))
	if err := s.rdb.HSet(ctx, constant.LoginUser, userId, data).Err(); err != nil {
		s.logger.Error("写入Redis Session失败", "key", constant.LoginUser, "field", userId, "error", err)
		return
	}

	// 设置整个Hash的过期时间
	if err := s.rdb.Expire(ctx, constant.LoginUser, constant.TokenExpireTime*time.Second).Err(); err != nil {
		s.logger.Warn("设置Redis过期时间失败", "error", err)
	}
}

// RenewToken 自动续期 (对标Java renewToken)
// 当距离过期时间 ≤ TokenRenewThreshold(20分钟) 时，自动续期
func (s *TokenService) RenewToken(userDetails *dto.UserDetailsDTO) {
	if userDetails == nil || userDetails.ExpireTime.IsZero() {
		return
	}
	remaining := time.Until(userDetails.ExpireTime)
	if remaining <= time.Minute*constant.TokenRenewThreshold {
		s.RefreshToken(userDetails)
	}
}

// ParseToken 解析并验证JWT Token (对标Java parseToken)
func (s *TokenService) ParseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(s.cfg.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}

// GetUserDetailDTO 从请求Header中提取Token，验证后返回用户详情 (对标Java getUserDetailDTO)
// 返回值: 用户详情(含角色), 错误
func (s *TokenService) GetUserDetailDTO(tokenString string) (*dto.UserDetailsDTO, error) {
	if tokenString == "" || tokenString == "null" {
		return nil, nil // 无Token不算错误
	}

	claims, err := s.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	userIdStr, ok := claims[constant.JwtClaimUserID].(string)
	if !ok || userIdStr == "" {
		return nil, fmt.Errorf("无效的user_id claim")
	}

	// 从Redis Hash获取完整用户信息 (对标Java redisService.hGet(LOGIN_USER, userId))
	ctx := context.Background()
	data, err := s.rdb.HGet(ctx, constant.LoginUser, userIdStr).Bytes()
	if err == redis.Nil {
		return nil, nil // Session不存在(已过期或被删除)
	}
	if err != nil {
		return nil, fmt.Errorf("读取Redis Session: %w", err)
	}

	var user dto.UserDetailsDTO
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, fmt.Errorf("反序列化用户Session: %w", err)
	}

	return &user, nil
}

// DeleteLoginUser 删除用户的登录Session (登出时调用) (对标Java delLoginUser)
func (s *TokenService) DeleteLoginUser(userID uint) error {
	ctx := context.Background()
	userIdStr := fmt.Sprintf("%d", userID)
	// 从Hash中删除指定field (对标Java redisService.hDel(LOGIN_USER, String.valueOf(userId)))
	return s.rdb.HDel(ctx, constant.LoginUser, userIdStr).Err()
}

// ValidateToken 验证Token是否有效（用于需要手动验证的场景）
// 返回: 用户详情, 是否有效, 错误
func (s *TokenService) ValidateToken(tokenString string) (*dto.UserDetailsDTO, bool, error) {
	user, err := s.GetUserDetailDTO(tokenString)
	if err != nil {
		return nil, false, err
	}
	if user == nil {
		return nil, false, nil
	}
	// 检查账号是否被禁用
	if user.IsDisable == 1 {
		return nil, false, nil
	}
	return user, true, nil
}

// ExtractToken 从Authorization Header提取Bearer Token
func ExtractToken(authHeader string) string {
	if len(authHeader) > len(constant.TokenPrefix) &&
		authHeader[:len(constant.TokenPrefix)] == constant.TokenPrefix {
		return authHeader[len(constant.TokenPrefix):]
	}
	return ""
}

func generateUUID() string {
	b := make([]byte, 16)
	// 使用时间戳+随机数模拟UUID (生产环境可用google/uuid)
	for i := range b {
		b[i] = byte(time.Now().UnixNano()>>uint(i*3)) ^ byte(i*37+5)
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
