package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
)

// TaskRegistry 任务函数注册表（对标Java Spring Bean）
// Java: @Component("auroraQuartz") public class AuroraQuartz { ... }
// Go: 用map映射 beanName.methodName → 函数
var TaskRegistry = make(map[string]TaskFunc)

// TaskFunc 任务函数签名
type TaskFunc func(ctx context.Context, params ...interface{}) error

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
//   - "auroraQuartz.saveUniqueView('param')" → 带参数调用
//   - "com.aurora.quartz.CustomJob.run"    → 不支持（需要额外实现）
func InvokeMethod(ctx context.Context, invokeTarget string) error {
	if invokeTarget == "" {
		return fmt.Errorf("invokeTarget不能为空")
	}

	// 1. 解析方法名和参数（对标Java getBeanName/getMethodName/getMethodParams）
	beanName, methodName, methodParams, err := parseInvokeTarget(invokeTarget)
	if err != nil {
		return fmt.Errorf("解析invokeTarget失败: %w", err)
	}

	// 2. 构建查找key
	key := beanName + "." + methodName

	// 3. 检查是否在注册表中（对标Java SpringUtil.getBean(beanName)）
	fn, ok := TaskRegistry[key]
	if !ok {
		return fmt.Errorf("未找到任务函数: %s (请检查任务名是否在TaskRegistry中注册)", key)
	}

	// 4. 调用任务函数（对标Java method.invoke(bean, params)）
	return fn(ctx, methodParams...)
}

// parseInvokeTarget 解析invokeTarget字符串
// 对标Java: getBeanName() + getMethodName() + getMethodParams()
// 示例: "auroraQuartz.saveUniqueView('test',123,true)" → bean="auroraQuartz", method="saveUniqueView", params=["test", 123, true]
func parseInvokeTarget(invokeTarget string) (string, string, []interface{}, error) {
	// 提取括号内的参数部分
	paramStart := strings.Index(invokeTarget, "(")
	paramEnd := strings.LastIndex(invokeTarget, ")")

	var paramStr string
	var hasParams bool
	if paramStart > 0 && paramEnd > paramStart {
		paramStr = invokeTarget[paramStart+1 : paramEnd]
		hasParams = true
		invokeTarget = invokeTarget[:paramStart] // 移除参数部分
	}

	// 解析 beanName.methodName
	lastDot := strings.LastIndex(invokeTarget, ".")
	if lastDot <= 0 {
		return "", "", nil, fmt.Errorf("无效的invokeTarget格式: %s (期望: beanName.methodName)", invokeTarget)
	}

	beanName := invokeTarget[:lastDot]
	methodName := invokeTarget[lastDot+1:]

	// 解析参数（对标Java getMethodParams）
	var params []interface{}
	if hasParams && strings.TrimSpace(paramStr) != "" {
		var err error
		params, err = parseMethodParams(paramStr)
		if err != nil {
			return "", "", nil, fmt.Errorf("解析参数失败: %w", err)
		}
	}

	return beanName, methodName, params, nil
}

// parseMethodParams 解析方法参数字符串
// 对标Java JobInvokeUtil.getMethodParams()
// 支持的类型:
//   - 字符串: 'value' 或 "value"
//   - 布尔值: true/false
//   - Long: 123L
//   - Double: 123.45D
//   - Integer: 123
func parseMethodParams(paramStr string) ([]interface{}, error) {
	// 按逗号分割参数
	paramParts := strings.Split(paramStr, ",")
	params := make([]interface{}, 0, len(paramParts))

	for _, part := range paramParts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// 字符串类型: 'value' 或 "value"
		if strings.HasPrefix(part, "'") && strings.HasSuffix(part, "'") {
			// 单引号字符串
			params = append(params, part[1:len(part)-1])
		} else if strings.HasPrefix(part, "\"") && strings.HasSuffix(part, "\"") {
			// 双引号字符串
			params = append(params, part[1:len(part)-1])
		} else if strings.EqualFold(part, "true") || strings.EqualFold(part, "false") {
			// 布尔类型
			params = append(params, strings.EqualFold(part, "true"))
		} else if strings.HasSuffix(strings.ToUpper(part), "L") {
			// Long类型
			numStr := strings.TrimSuffix(part, "L")
			numStr = strings.TrimSuffix(numStr, "l")
			num, err := strconv.ParseInt(strings.TrimSpace(numStr), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("解析Long参数失败 '%s': %w", part, err)
			}
			params = append(params, num)
		} else if strings.HasSuffix(strings.ToUpper(part), "D") {
			// Double类型
			numStr := strings.TrimSuffix(part, "D")
			numStr = strings.TrimSuffix(numStr, "d")
			num, err := strconv.ParseFloat(strings.TrimSpace(numStr), 64)
			if err != nil {
				return nil, fmt.Errorf("解析Double参数失败 '%s': %w", part, err)
			}
			params = append(params, num)
		} else {
			// 默认Integer类型
			num, err := strconv.Atoi(strings.TrimSpace(part))
			if err != nil {
				return nil, fmt.Errorf("解析Integer参数失败 '%s': %w", part, err)
			}
			params = append(params, num)
		}
	}

	return params, nil
}
