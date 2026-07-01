package service

import (
	"testing"
)

// ===== getCategoryID 测试 =====

func TestGetCategoryID(t *testing.T) {
	validID := uint(42)
	zeroID := uint(0)

	tests := []struct {
		name       string
		categoryID *uint
		want       uint
	}{
		{"正常值", &validID, 42},
		{"零值指针", &zeroID, 0},
		{"nil 指针", nil, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getCategoryID(tt.categoryID)
			if got != tt.want {
				t.Errorf("getCategoryID(%v) = %d; want %d", tt.categoryID, got, tt.want)
			}
		})
	}
}

// ===== int8Ptr 测试 =====

func TestInt8Ptr(t *testing.T) {
	v := int8(1)
	p := int8Ptr(v)
	if p == nil {
		t.Fatal("int8Ptr returned nil")
	}
	if *p != v {
		t.Errorf("int8Ptr(%d) = %d; want %d", v, *p, v)
	}

	// 验证多次调用返回不同地址
	p2 := int8Ptr(v)
	if p == p2 {
		t.Error("int8Ptr should return different pointer addresses on each call")
	}
}
