package util

import (
	"encoding/hex"
	"strings"
	"testing"
)

// ===== BCrypt 测试 =====

func TestBCryptHash(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"normal password", "mypassword123", false},
		{"empty password", "", false},
		{"long password", strings.Repeat("a", 72), false}, // bcrypt max 72 bytes
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := BCryptHash(tt.password)
			if (err != nil) != tt.wantErr {
				t.Fatalf("BCryptHash() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !strings.HasPrefix(hash, "$2a$") && !strings.HasPrefix(hash, "$2b$") {
				t.Errorf("BCryptHash() = %q; want bcrypt prefix", hash)
			}
		})
	}
}

func TestBCryptCheck(t *testing.T) {
	password := "test_password_123"
	hash, err := BCryptHash(password)
	if err != nil {
		t.Fatalf("BCryptHash() failed: %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{"correct password", password, hash, true},
		{"wrong password", "wrong_password", hash, false},
		{"empty password", "", hash, false},
		{"invalid hash", password, "invalid_hash", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BCryptCheck(tt.password, tt.hash); got != tt.want {
				t.Errorf("BCryptCheck() = %v; want %v", got, tt.want)
			}
		})
	}
}

// ===== MD5Hex 测试 =====

func TestMD5Hex(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string // MD5 hex
	}{
		{"empty", "", "d41d8cd98f00b204e9800998ecf8427e"},
		{"hello", "hello", "5d41402abc4b2a76b9719d911017c592"},
		{"abc", "abc", "900150983cd24fb0d6963f7d28e17f72"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MD5Hex(tt.input)
			if got != tt.want {
				t.Errorf("MD5Hex(%q) = %q; want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestMD5Hex_Length(t *testing.T) {
	got := MD5Hex("any string")
	if len(got) != 32 { // MD5 hex = 16 bytes = 32 hex chars
		t.Errorf("MD5Hex() length = %d; want 32", len(got))
	}
}

// ===== SHA256Hex 测试 =====

func TestSHA256Hex(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{"hello", "hello", "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SHA256Hex(tt.input)
			if got != tt.want {
				t.Errorf("SHA256Hex(%q) = %q; want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSHA256Hex_Length(t *testing.T) {
	got := SHA256Hex("any string")
	if len(got) != 64 { // SHA256 hex = 32 bytes = 64 hex chars
		t.Errorf("SHA256Hex() length = %d; want 64", len(got))
	}
}

// ===== Base64 测试 =====

func TestBase64RoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{"simple", []byte("hello world")},
		{"empty", []byte{}},
		{"binary", []byte{0x00, 0x01, 0x02, 0xff}},
		{"unicode", []byte("你好世界")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := Base64Encode(tt.input)
			decoded, err := Base64Decode(encoded)
			if err != nil {
				t.Fatalf("Base64Decode() error = %v", err)
			}
			if string(decoded) != string(tt.input) {
				t.Errorf("round trip failed: got %v; want %v", decoded, tt.input)
			}
		})
	}
}

func TestBase64Decode_Invalid(t *testing.T) {
	_, err := Base64Decode("!!!invalid!!!")
	if err == nil {
		t.Error("expected error for invalid base64, got nil")
	}
}

// ===== GenerateRandomString 测试 =====

func TestGenerateRandomString(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"8 chars", 8},
		{"16 chars", 16},
		{"32 chars", 32},
		{"1 char", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateRandomString(tt.length)
			if err != nil {
				t.Fatalf("GenerateRandomString() error = %v", err)
			}
			if len(got) != tt.length {
				t.Errorf("length = %d; want %d", len(got), tt.length)
			}
		})
	}
}

func TestGenerateRandomString_Uniqueness(t *testing.T) {
	results := make(map[string]bool)
	for i := 0; i < 100; i++ {
		s, _ := GenerateRandomString(16)
		if results[s] {
			t.Errorf("duplicate random string: %q", s)
		}
		results[s] = true
	}
}

func TestGenerateRandomStringSimple(t *testing.T) {
	got := GenerateRandomStringSimple(10)
	if len(got) != 10 {
		t.Errorf("length = %d; want 10", len(got))
	}
}

// ===== GenerateCode 测试 =====

func TestGenerateCode(t *testing.T) {
	code := GenerateCode(6)
	if len(code) != 6 {
		t.Errorf("code length = %d; want 6", len(code))
	}
	// 验证全是数字
	for _, ch := range code {
		if ch < '0' || ch > '9' {
			t.Errorf("non-digit char in code: %c", ch)
		}
	}
}

func TestGenerateCode_Uniqueness(t *testing.T) {
	codes := make(map[string]bool)
	for i := 0; i < 50; i++ {
		code := GenerateCode(6)
		if codes[code] {
			t.Errorf("duplicate code: %q", code)
		}
		codes[code] = true
	}
}

// ===== GenerateTokenID 测试 =====

func TestGenerateTokenID(t *testing.T) {
	tokenID, err := GenerateTokenID()
	if err != nil {
		t.Fatalf("GenerateTokenID() error = %v", err)
	}
	// 格式: {timestamp}_{random8chars}
	if !strings.Contains(tokenID, "_") {
		t.Errorf("tokenID = %q; want underscore separator", tokenID)
	}
	parts := strings.SplitN(tokenID, "_", 2)
	if len(parts[0]) == 0 {
		t.Error("timestamp part is empty")
	}
	if len(parts[1]) != 8 {
		t.Errorf("random part length = %d; want 8", len(parts[1]))
	}
}

// ===== Benchmark =====

func BenchmarkMD5Hex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		MD5Hex("benchmark test string")
	}
}

func BenchmarkSHA256Hex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SHA256Hex("benchmark test string")
	}
}

func BenchmarkGenerateRandomString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateRandomString(16)
	}
}

func BenchmarkBCryptHash(b *testing.B) {
	password := "benchmark_password"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BCryptHash(password)
	}
}

func BenchmarkBCryptCheck(b *testing.B) {
	password := "benchmark_password"
	hash, _ := BCryptHash(password)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BCryptCheck(password, hash)
	}
}

// 确保导入hex包
var _ = hex.EncodeToString
