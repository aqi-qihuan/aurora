package agent

import "log/slog"

// ========== Panic恢复工具函数 ==========
// 对标 L4 故障隔离: goroutine+recover包装, Agent panic不杀主进程
//
// 注意: 原 SSEScanner 已在阶段1（LLM Router 替换为 tRPC model/openai）后废弃删除
// tRPC model.GenerateContent 返回 channel，不再需要手搓 SSE 解析

func recoverPanic(operation string) {
	if r := recover(); r != nil {
		slog.Error("Agent panic recovered",
			"operation", operation,
			"error", r,
		)
	}
}
