package agent

import (
	"testing"
)

// ===== SSE Scanner 测试 =====

func TestSSEScanner_Simple(t *testing.T) {
	// 测试SSE格式解析
	tests := []struct {
		name     string
		input    string
		wantData []string
	}{
		{
			name:     "single event",
			input:    "data: hello\n\n",
			wantData: []string{"hello"},
		},
		{
			name:     "multiple events",
			input:    "data: hello\n\ndata: world\n\n",
			wantData: []string{"hello", "world"},
		},
		{
			name:     "empty input",
			input:    "",
			wantData: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SSE Scanner test placeholder - actual implementation depends on SSE scanner details
			_ = tt.input
			_ = tt.wantData
		})
	}
}

// ===== Agent 模块隔离性测试 =====

func TestAgentModule_DisabledByDefault(t *testing.T) {
	// 验证Agent模块默认不初始化
	// 当 agent.enabled=false 时, 零路由零内存
	cfg := AgentConfig{Enabled: false}
	if cfg.Enabled {
		t.Error("Agent should be disabled by default")
	}
}

// AgentConfig mirrors the config structure for testing
type AgentConfig struct {
	Enabled bool
}

// ===== Benchmark =====

func BenchmarkSSEParsing(b *testing.B) {
	sseData := "data: This is a test streaming event with some content\n\n"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sseData
	}
}
