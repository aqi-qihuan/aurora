package errors

import (
	"errors"
	"testing"
)

// ===== AppError 创建和格式化 =====

func TestNew(t *testing.T) {
	err := New(400, "参数错误")
	if err.Code != 400 {
		t.Errorf("Code = %d; want 400", err.Code)
	}
	if err.Message != "参数错误" {
		t.Errorf("Message = %q; want %q", err.Message, "参数错误")
	}
}

func TestAppError_Error(t *testing.T) {
	err := New(500, "服务器错误")
	errStr := err.Error()
	if errStr == "" {
		t.Error("Error() should not be empty")
	}
	// 应包含错误码和消息
	if !contains(errStr, "500") || !contains(errStr, "服务器错误") {
		t.Errorf("Error() = %q; should contain code and message", errStr)
	}
}

func TestWrap(t *testing.T) {
	innerErr := errors.New("connection refused")
	err := Wrap(500, "数据库连接失败", innerErr)
	if err.Code != 500 {
		t.Errorf("Code = %d; want 500", err.Code)
	}
	if !contains(err.Message, "数据库连接失败") {
		t.Errorf("Message should contain original msg, got %q", err.Message)
	}
	if !contains(err.Message, "connection refused") {
		t.Errorf("Message should contain inner error, got %q", err.Message)
	}
}

// ===== Is 函数测试 =====

func TestIs(t *testing.T) {
	err := New(401, "未授权")
	tests := []struct {
		name   string
		err    error
		target *AppError
		want   bool
	}{
		{"matching code", err, New(401, "其他消息"), true},
		{"non-matching code", err, New(403, "未授权"), false},
		{"nil error", nil, err, false},
		{"non-AppError error", errors.New("standard error"), err, false},
		{"same instance", err, err, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Is(tt.err, tt.target); got != tt.want {
				t.Errorf("Is() = %v; want %v", got, tt.want)
			}
		})
	}
}

// ===== 预定义错误常量测试 =====

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name    string
		err     *AppError
		wantCode int
	}{
		{"OK", OK, 200},
		{"ErrInternalServer", ErrInternalServer, 500},
		{"ErrInvalidParams", ErrInvalidParams, 400},
		{"ErrUnauthorized", ErrUnauthorized, 401},
		{"ErrForbidden", ErrForbidden, 403},
		{"ErrUserNotFound", ErrUserNotFound, 500},
		{"ErrArticleNotFound", ErrArticleNotFound, 600},
		{"ErrFileUploadFailed", ErrFileUploadFailed, 700},
		{"ErrCommentNotFound", ErrCommentNotFound, 800},
		{"ErrAgentDisabled", ErrAgentDisabled, 900},
		{"ErrInvalidConfig", ErrInvalidConfig, 1000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.wantCode {
				t.Errorf("%s.Code = %d; want %d", tt.name, tt.err.Code, tt.wantCode)
			}
			if tt.err.Message == "" {
				t.Errorf("%s.Message should not be empty", tt.name)
			}
		})
	}
}

// ===== AppError 实现 error 接口 =====

func TestAppError_ImplementsError(t *testing.T) {
	var err error = New(400, "test")
	if err.Error() == "" {
		t.Error("AppError should implement error interface")
	}
}

// ===== Benchmark =====

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		New(400, "benchmark error")
	}
}

func BenchmarkWrap(b *testing.B) {
	inner := errors.New("inner")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Wrap(500, "wrapper", inner)
	}
}

func BenchmarkIs(b *testing.B) {
	err := New(401, "unauthorized")
	target := New(401, "other")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Is(err, target)
	}
}

// helper
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
