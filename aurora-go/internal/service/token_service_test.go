package service

import (
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/aurora-go/aurora/internal/config"
	"github.com/aurora-go/aurora/internal/constant"
	"github.com/golang-jwt/jwt/v5"
)

// ===== ExtractToken 测试 =====

func TestExtractToken(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
		want       string
	}{
		{"valid bearer", "Bearer abc123", "abc123"},
		{"bearer with long token", "Bearer eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U", "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"},
		{"no bearer prefix", "abc123", ""},
		{"empty string", "", ""},
		{"bearer only", "Bearer ", ""},
		{"lowercase bearer", "bearer abc", ""}, // case sensitive
		{"basic auth", "Basic abc123", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractToken(tt.authHeader)
			if got != tt.want {
				t.Errorf("ExtractToken(%q) = %q; want %q", tt.authHeader, got, tt.want)
			}
		})
	}
}

// ===== generateUUID 测试 =====

func TestGenerateUUID(t *testing.T) {
	uuid1 := generateUUID()
	uuid2 := generateUUID()

	if uuid1 == "" {
		t.Error("generateUUID() should not return empty string")
	}
	if uuid1 == uuid2 {
		t.Error("consecutive generateUUID() should return different values")
	}
	// UUID格式: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	if len(uuid1) < 10 {
		t.Errorf("UUID too short: %q", uuid1)
	}
}

func TestGenerateUUID_Uniqueness(t *testing.T) {
	uuids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := generateUUID()
		if uuids[id] {
			t.Errorf("duplicate UUID: %q", id)
		}
		uuids[id] = true
	}
}

// ===== createTokenBySubject + ParseToken 往返测试 =====
// 这两个方法不依赖 Redis，可独立测试

func newTestTokenService(secret string) *TokenService {
	return NewTokenService(
		config.JWTConfig{
			Secret:     secret,
			ExpireTime: 168, // 7 天
			Issuer:     "aurora-go-test",
		},
		nil, // redis 不参与 token 生成与解析
		slog.Default(),
	)
}

func TestCreateAndParseToken_RoundTrip(t *testing.T) {
	svc := newTestTokenService("test-secret-key")

	token, err := svc.createTokenBySubject("12345")
	if err != nil {
		t.Fatalf("createTokenBySubject failed: %v", err)
	}
	if token == "" {
		t.Fatal("token should not be empty")
	}
	if !strings.HasPrefix(token, "eyJ") {
		t.Errorf("token should be JWT format (starts with eyJ): %s", token[:10])
	}

	// 解析回来
	claims, err := svc.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}

	// 验证 claims
	userID, ok := claims[constant.JwtClaimUserID].(string)
	if !ok {
		t.Fatalf("user_id claim is not string: %v", claims[constant.JwtClaimUserID])
	}
	if userID != "12345" {
		t.Errorf("user_id = %q; want 12345", userID)
	}

	// 验证 issuer
	iss, _ := claims["iss"].(string)
	if iss != "aurora-go-test" {
		t.Errorf("iss = %q; want aurora-go-test", iss)
	}

	// 验证 jti 存在
	jti, ok := claims["jti"]
	if !ok || jti == "" {
		t.Error("jti claim should exist and be non-empty")
	}

	// 验证签发时间与过期时间
	iat, _ := claims[constant.JwtClaimIssueAt].(float64)
	exp, _ := claims[constant.JwtClaimExpireAt].(float64)
	if iat <= 0 {
		t.Errorf("iat should be positive unix timestamp, got %v", iat)
	}
	if exp <= iat {
		t.Errorf("exp (%v) should be greater than iat (%v)", exp, iat)
	}
	// expire_time=168 小时 = 604800 秒
	if int64(exp-iat) != int64(168*60*60) {
		t.Errorf("token lifetime = %d sec; want %d", int64(exp-iat), 168*60*60)
	}
}

func TestParseToken_InvalidToken(t *testing.T) {
	svc := newTestTokenService("test-secret-key")

	tests := []struct {
		name        string
		token       string
		wantErr     bool
		errContains string
	}{
		{"空字符串", "", true, ""},
		{"非 JWT 格式", "not-a-jwt", true, ""},
		{"乱码", "abc.def.ghi", true, ""},
		{"只有两部分", "abc.def", true, ""},
		{"空 payload", "abc..def", true, ""},
		{"篡改的 token", "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjMifQ.invalid-signature", true, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.ParseToken(tt.token)
			if !tt.wantErr {
				if err != nil {
					t.Errorf("ParseToken(%q) unexpected error: %v", tt.token, err)
				}
				return
			}
			if err == nil {
				t.Errorf("ParseToken(%q) expected error, got nil", tt.token)
			}
		})
	}
}

func TestParseToken_SecretMismatch(t *testing.T) {
	// 用 secret A 签发，用 secret B 解析，应失败
	svcA := newTestTokenService("secret-a")
	svcB := newTestTokenService("secret-b")

	token, err := svcA.createTokenBySubject("99")
	if err != nil {
		t.Fatalf("createToken failed: %v", err)
	}

	_, err = svcB.ParseToken(token)
	if err == nil {
		t.Error("ParseToken with mismatched secret should fail")
	}
}

func TestParseToken_InvalidSigningAlgorithm(t *testing.T) {
	svc := newTestTokenService("test-secret-key")

	// 用 none 算法构造 token，应被拒绝（防止 alg=none 攻击）
	claims := jwt.MapClaims{
		constant.JwtClaimUserID: "attacker",
		"exp":                   time.Now().Add(time.Hour).Unix(),
	}
	unsignedToken := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	// jwt/v5 用 UnsafeAllowNoneSignatureType 作为密钥参数来签名 none 算法 token
	tokenString, err := unsignedToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("construct none-alg token failed: %v", err)
	}

	_, err = svc.ParseToken(tokenString)
	if err == nil {
		t.Error("ParseToken should reject none-algorithm token")
	}
}

// ===== Benchmark =====

func BenchmarkExtractToken(b *testing.B) {
	header := "Bearer eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0In0.signature"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ExtractToken(header)
	}
}

func BenchmarkGenerateUUID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		generateUUID()
	}
}
