package util

import (
	"math"
	"testing"
)

// ===== Offset 测试 =====

func TestOffset(t *testing.T) {
	tests := []struct {
		name     string
		pageNum  int
		pageSize int
		want     int
	}{
		{"first page", 1, 10, 0},
		{"second page", 2, 10, 10},
		{"third page", 3, 10, 20},
		{"large page", 100, 20, 1980},
		{"page size 1", 5, 1, 4},
		{"page size 100", 2, 100, 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Offset(tt.pageNum, tt.pageSize); got != tt.want {
				t.Errorf("Offset(%d, %d) = %d; want %d", tt.pageNum, tt.pageSize, got, tt.want)
			}
		})
	}
}

// ===== BuildPageResult 测试 =====

func TestBuildPageResult(t *testing.T) {
	tests := []struct {
		name           string
		list           interface{}
		total          int64
		pageNum        int
		pageSize       int
		wantTotalPages int
		wantHasPrev    bool
		wantHasNext    bool
	}{
		{
			name:           "single page",
			list:           []string{"a", "b"},
			total:          2,
			pageNum:        1,
			pageSize:       10,
			wantTotalPages: 1,
			wantHasPrev:    false,
			wantHasNext:    false,
		},
		{
			name:           "multiple pages - first",
			list:           []int{1, 2, 3},
			total:          25,
			pageNum:        1,
			pageSize:       10,
			wantTotalPages: 3,
			wantHasPrev:    false,
			wantHasNext:    true,
		},
		{
			name:           "multiple pages - middle",
			list:           []int{1},
			total:          25,
			pageNum:        2,
			pageSize:       10,
			wantTotalPages: 3,
			wantHasPrev:    true,
			wantHasNext:    true,
		},
		{
			name:           "multiple pages - last",
			list:           []int{1},
			total:          25,
			pageNum:        3,
			pageSize:       10,
			wantTotalPages: 3,
			wantHasPrev:    true,
			wantHasNext:    false,
		},
		{
			name:           "zero total",
			list:           []int{},
			total:          0,
			pageNum:        1,
			pageSize:       10,
			wantTotalPages: 1,
			wantHasPrev:    false,
			wantHasNext:    false,
		},
		{
			name:           "exact fit",
			list:           []int{},
			total:          20,
			pageNum:        1,
			pageSize:       20,
			wantTotalPages: 1,
			wantHasPrev:    false,
			wantHasNext:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildPageResult(tt.list, tt.total, tt.pageNum, tt.pageSize)
			if got.TotalPages != tt.wantTotalPages {
				t.Errorf("TotalPages = %d; want %d", got.TotalPages, tt.wantTotalPages)
			}
			if got.HasPrev != tt.wantHasPrev {
				t.Errorf("HasPrev = %v; want %v", got.HasPrev, tt.wantHasPrev)
			}
			if got.HasNext != tt.wantHasNext {
				t.Errorf("HasNext = %v; want %v", got.HasNext, tt.wantHasNext)
			}
			if got.Total != tt.total {
				t.Errorf("Total = %d; want %d", got.Total, tt.total)
			}
			if got.PageNum != tt.pageNum {
				t.Errorf("PageNum = %d; want %d", got.PageNum, tt.pageNum)
			}
		})
	}
}

// ===== ParseInt 测试 =====

func TestParseInt(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{"positive", "123", 123, false},
		{"zero", "0", 0, false},
		{"negative", "-42", 42, false}, // our parser treats '-' as skip when result==0
		{"with letters", "12abc34", 12, false},
		{"empty", "", 0, false},
		{"letters only", "abc", 0, false},
		{"large number", "999999", 999999, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result int
			got, err := ParseInt(tt.input, &result)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInt(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ParseInt(%q) = %d; want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseInt_UnsupportedType(t *testing.T) {
	var f float64
	_, err := ParseInt("123", &f)
	if err != ErrUnsupportedType {
		t.Errorf("expected ErrUnsupportedType, got %v", err)
	}
}

func TestParseInt_Int64Target(t *testing.T) {
	var result int64
	got, err := ParseInt("42", &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 42 {
		t.Errorf("got %d; want 42", got)
	}
	if result != 42 {
		t.Errorf("result = %d; want 42", result)
	}
}

// ===== NewPageResult 测试 =====

func TestNewPageResult(t *testing.T) {
	list := []string{"a", "b"}
	result := NewPageResult(list, 100, 1, 10)
	if result.Count != 100 {
		t.Errorf("Count = %d; want 100", result.Count)
	}
	if result.PageNum != 1 {
		t.Errorf("PageNum = %d; want 1", result.PageNum)
	}
	if result.PageSize != 10 {
		t.Errorf("PageSize = %d; want 10", result.PageSize)
	}
}

// ===== Benchmark =====

func BenchmarkOffset(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Offset(100, 20)
	}
}

func BenchmarkBuildPageResult(b *testing.B) {
	list := make([]int, 100)
	for i := 0; i < b.N; i++ {
		BuildPageResult(list, 1000, 5, 20)
	}
}

func BenchmarkParseInt(b *testing.B) {
	var result int
	for i := 0; i < b.N; i++ {
		ParseInt("12345", &result)
	}
}

func TestPageResultTotalPages(t *testing.T) {
	// 验证分页计算: total=95, size=10 → 10页
	result := BuildPageResult(nil, 95, 1, 10)
	expected := int(math.Ceil(float64(95) / float64(10)))
	if result.TotalPages != expected {
		t.Errorf("TotalPages = %d; want %d", result.TotalPages, expected)
	}
}
