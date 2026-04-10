package agent

import (
	"bufio"
	"io"
	"log/slog"
	"strings"
)

// ========== SSE 流式解析器 ==========
// 对标 event.StreamingEvents 的数据解析层
// 解析 Server-Sent Events 格式: data: {...}\n\n

// SSEScanner SSE流扫描器
type SSEScanner struct {
	scanner *bufio.Scanner
	data    string
}

func newSSEScanner(r io.Reader) *SSEScanner {
	return &SSEScanner{
		scanner: bufio.NewScanner(r),
	}
}

// Scan 读取下一个SSE事件
func (s *SSEScanner) Scan() bool {
	for s.scanner.Scan() {
		line := s.scanner.Text()

		// 跳过空行和注释行
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		// 解析 data: 前缀
		if strings.HasPrefix(line, "data: ") {
			s.data = strings.TrimPrefix(line, "data: ")
			return true
		}
		if strings.HasPrefix(line, "data:") {
			s.data = strings.TrimPrefix(line, "data:")
			return true
		}

		// 其他字段(event:, id:, retry:)暂忽略
	}

	return false
}

// Data 返回当前事件的data字段值
func (s *SSEScanner) Data() string {
	return s.data
}

// Err 返回扫描错误
func (s *SSEScanner) Err() error {
	return s.scanner.Err()
}

// ========== Panic恢复工具函数 ==========
// 对标 L4 故障隔离: goroutine+recover包装, Agent panic不杀主进程

func recoverPanic(operation string) {
	if r := recover(); r != nil {
		slog.Error("Agent panic recovered",
			"operation", operation,
			"error", r,
		)
	}
}
