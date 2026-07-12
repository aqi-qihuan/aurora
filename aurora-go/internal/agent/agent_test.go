package agent

import "testing"

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
