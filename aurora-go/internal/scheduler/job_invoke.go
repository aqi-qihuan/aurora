package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

// TaskRegistry 任务函数注册表（对标Java Spring Bean）
// Java: @Component("auroraQuartz") public class AuroraQuartz { ... }
// Go: 用map映射 beanName.methodName → 函数
var TaskRegistry = make(map[string]TaskFunc)

// TaskFunc 任务函数签名
type TaskFunc func(ctx context.Context) error

// InitTaskRegistry 初始化所有内置任务函数
// 对标Java AuroraQuartz类中的所有方法
func InitTaskRegistry(scheduler *Scheduler) {
	// 注册 "auroraQuartz.xxx" 格式的任务
	prefix := "auroraQuartz"
	
	TaskRegistry[prefix+".saveUniqueView"] = NewUniqueViewJob(scheduler.db, scheduler.rdb).Run
	TaskRegistry[prefix+".clear"] = NewClearCacheJob(scheduler.rdb).Run
	TaskRegistry[prefix+".statisticalUserArea"] = NewUserAreaJob(scheduler.db, scheduler.rdb).Run
	TaskRegistry[prefix+".baiduSeo"] = NewBaiduSeoJob(scheduler.db, scheduler.siteURL).Run
	TaskRegistry[prefix+".clearJobLogs"] = NewCleanLogJob(scheduler.db).Run
	TaskRegistry[prefix+".importDataIntoES"] = NewESSyncJob(scheduler.db).Run
	
	slog.Info("任务函数注册表初始化完成", "count", len(TaskRegistry))
}

// InvokeMethod 解析并调用任务方法（对标Java JobInvokeUtil.invokeMethod）
// invokeTarget格式:
//   - "auroraQuartz.clearJobLogs"          → 调用 TaskRegistry["auroraQuartz.clearJobLogs"]
//   - "com.aurora.quartz.CustomJob.run"    → 不支持（需要额外实现）
func InvokeMethod(ctx context.Context, invokeTarget string) error {
	if invokeTarget == "" {
		return fmt.Errorf("invokeTarget不能为空")
	}

	// 1. 检查是否在注册表中（对标Java SpringUtil.getBean(beanName)）
	if fn, ok := TaskRegistry[invokeTarget]; ok {
		return fn(ctx)
	}

	// 2. 尝试解析为 beanName.methodName 格式
	// Java: beanName = "auroraQuartz", methodName = "clearJobLogs"
	if idx := strings.LastIndex(invokeTarget, "."); idx > 0 {
		beanName := invokeTarget[:idx]
		methodName := invokeTarget[idx+1:]
		
		// 查找注册的函数（支持完整路径匹配）
		key := beanName + "." + methodName
		if fn, ok := TaskRegistry[key]; ok {
			return fn(ctx)
		}
	}

	return fmt.Errorf("未找到任务函数: %s (请检查任务名是否在TaskRegistry中注册)", invokeTarget)
}
