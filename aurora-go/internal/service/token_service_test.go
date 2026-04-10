package service

import (
	"testing"
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
