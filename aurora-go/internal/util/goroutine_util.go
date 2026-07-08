package util

import (
	"log/slog"
	"runtime/debug"
	"sync"
)

// SafeGo 启动带 recover 的 goroutine（WaitGroup 模式）
// 防止 goroutine 内 panic 导致整个进程崩溃（gin.Recovery 无法捕获 goroutine 内的 panic）
//
// 用法:
//
//	var wg sync.WaitGroup
//	util.SafeGo(&wg, func() {
//	    // 你的逻辑（无需手动 defer wg.Done()）
//	})
//	wg.Wait()
func SafeGo(wg *sync.WaitGroup, fn func()) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer recoverGoroutine()
		fn()
	}()
}

// SafeGoAsync 启动带 recover 的 fire-and-forget goroutine（无 WaitGroup）
// 用于异步通知、日志上报等不阻塞主流程的场景
//
// 用法:
//
//	util.SafeGoAsync(func() {
//	    // 异步逻辑
//	})
func SafeGoAsync(fn func()) {
	go func() {
		defer recoverGoroutine()
		fn()
	}()
}

// recoverGoroutine goroutine panic 恢复，记录堆栈信息便于排查
func recoverGoroutine() {
	if r := recover(); r != nil {
		slog.Error("goroutine panic recovered",
			"panic", r,
			"stack", string(debug.Stack()),
		)
	}
}
