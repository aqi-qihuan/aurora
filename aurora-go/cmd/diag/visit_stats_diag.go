package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/aurora-go/aurora/internal/constant"
	"github.com/aurora-go/aurora/internal/model"
)

func main() {
	ctx := context.Background()

	// 连接 Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     "134.175.206.158:6379",
		Password: "aqi1015",
		DB:       0,
	})

	// 连接 MySQL
	db, err := gorm.Open(mysql.Open("root:aqi1015@tcp(134.175.206.158:3306)/aurora?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{})
	if err != nil {
		log.Fatal("MySQL连接失败:", err)
	}

	fmt.Println("=== 访客数据统计诊断工具（修复后） ===\n")

	// 1. 检查 Redis 中的按天 key
	fmt.Println("【1. Redis 按天 Key 检查】")
	for i := 6; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		key := fmt.Sprintf("%s:%s", constant.UniqueVisitor, date)
		count, _ := rdb.SCard(ctx, key).Result()
		status := "✅"
		if count == 0 {
			status = "❌"
		}
		fmt.Printf("  %s %s = %d 个IP\n", status, key, count)
	}

	// 2. 检查数据库记录
	fmt.Println("\n【2. 数据库 t_unique_view 记录（最近7天）】")
	var uniqueViews []model.UniqueView
	startTime := time.Now().AddDate(0, 0, -7)
	db.Where("create_time >= ?", startTime).Order("create_time ASC").Find(&uniqueViews)

	for _, uv := range uniqueViews {
		date := uv.CreateTime.Format("2006-01-02")
		status := "✅"
		if uv.ViewsCount == 0 {
			status = "❌"
		}
		fmt.Printf("  %s %s: views_count=%d\n", status, date, uv.ViewsCount)
	}

	// 3. 模拟 GetAdminDashboard 的查询逻辑
	fmt.Println("\n【3. 模拟前端查询结果（最近7天）】")

	// 3.1 从数据库查询
	type UniqueViewRow struct {
		Day        string `gorm:"column:day"`
		ViewsCount int    `gorm:"column:views_count"`
	}
	var rows []UniqueViewRow
	db.Table("t_unique_view").
		Select(`DATE_FORMAT(create_time, "%Y-%m-%d") as day, views_count`).
		Where("create_time > ? AND create_time <= ?", startTime, time.Now()).
		Order("create_time ASC").
		Find(&rows)

	// 3.2 构建映射
	viewMap := make(map[string]int)
	for _, r := range rows {
		viewMap[r.Day] = r.ViewsCount
	}

	// 3.3 补充今天的实时数据
	today := time.Now().Format("2006-01-02")
	todayCount, _ := rdb.SCard(ctx, fmt.Sprintf("%s:%s", constant.UniqueVisitor, today)).Result()
	if todayCount > 0 {
		viewMap[today] = int(todayCount)
	}

	// 3.4 输出完整数据
	fmt.Println("  日期          | 访客数 | 数据来源")
	fmt.Println("  --------------+--------+----------")
	for i := 6; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		if count, exists := viewMap[date]; exists {
			source := "数据库"
			if date == today && todayCount > 0 {
				source = "Redis实时"
			}
			fmt.Printf("  %s | %6d | %s\n", date, count, source)
		} else {
			fmt.Printf("  %s | %6d | 无数据\n", date, 0)
		}
	}

	// 4. 诊断结论
	fmt.Println("\n【4. 诊断结论】")
	hasTodayData := false
	if _, exists := viewMap[today]; exists && viewMap[today] > 0 {
		hasTodayData = true
		fmt.Println("  ✅ 今天的数据正常显示（来自 Redis 实时统计）")
	} else {
		fmt.Println("  ⚠️  今天暂无访客数据")
	}

	hasHistoryData := false
	for i := 1; i <= 6; i++ {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		if count, exists := viewMap[date]; exists && count > 0 {
			hasHistoryData = true
			break
		}
	}
	if hasHistoryData {
		fmt.Println("  ✅ 历史数据正常（来自数据库归档）")
	} else {
		fmt.Println("  ⚠️  历史数据缺失，请检查 UniqueViewJob 定时任务是否正常执行")
	}

	if hasTodayData && hasHistoryData {
		fmt.Println("\n  🎉 一周访问趋势数据完整！前端图表应该能正常显示。")
	} else {
		fmt.Println("\n  💡 建议：")
		if !hasTodayData {
			fmt.Println("    - 等待用户访问产生数据")
			fmt.Println("    - 或手动触发一次访客上报接口 /api/report")
		}
		if !hasHistoryData {
			fmt.Println("    - 检查定时任务日志: SELECT * FROM t_job_log WHERE job_name LIKE '%UniqueView%' ORDER BY create_time DESC LIMIT 5;")
			fmt.Println("    - 确认 UniqueViewJob 是否每天凌晨0点正常执行")
		}
	}

	fmt.Println("\n【5. 关键说明】")
	fmt.Println("  • 游客访问数据实时更新：每次访问都会写入 Redis Set")
	fmt.Println("  • 一周趋势图数据来源：")
	fmt.Println("    - 今天：从 Redis 实时读取")
	fmt.Println("    - 昨天及之前：从数据库 t_unique_view 读取（由定时任务每天0点归档）")
	fmt.Println("  • 定时任务 UniqueViewJob：")
	fmt.Println("    - Cron表达式: 0 0 0 * * ? （每天凌晨0点执行）")
	fmt.Println("    - 作用：将昨天的 Redis 数据归档到数据库")
	fmt.Println("    - 如果任务未执行，会导致历史数据缺失")
}
