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
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// 连接 MySQL
	db, err := gorm.Open(mysql.Open("root:password@tcp(localhost:3306)/aurora?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{})
	if err != nil {
		log.Fatal("MySQL连接失败:", err)
	}

	fmt.Println("=== 访客数据统计诊断工具 ===\n")

	// 1. 检查 Redis 中的 key
	fmt.Println("【1. Redis Key 检查】")
	
	// 旧格式 key
	oldKey := constant.UniqueVisitor
	oldCount, _ := rdb.SCard(ctx, oldKey).Result()
	fmt.Printf("  旧格式 key: %s = %d 个IP\n", oldKey, oldCount)
	
	// 新格式 key（最近7天）
	fmt.Println("\n  新格式 key（最近7天）:")
	for i := 6; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		newKey := fmt.Sprintf("%s:%s", constant.UniqueVisitor, date)
		count, _ := rdb.SCard(ctx, newKey).Result()
		status := "✅"
		if count == 0 {
			status = "❌"
		}
		fmt.Printf("    %s %s = %d 个IP\n", status, newKey, count)
	}

	// 2. 检查数据库记录
	fmt.Println("\n【2. 数据库 t_unique_view 记录】")
	var uniqueViews []model.UniqueView
	startTime := time.Now().AddDate(0, 0, -7)
	db.Where("create_time >= ?", startTime).Order("create_time ASC").Find(&uniqueViews)
	
	for _, uv := range uniqueViews {
		date := uv.CreateTime.Format("2006-01-02")
		fmt.Printf("  %s: views_count=%d\n", date, uv.ViewsCount)
	}

	// 3. 诊断结论
	fmt.Println("\n【3. 诊断结论】")
	if oldCount > 0 {
		fmt.Println("  ⚠️  发现旧格式 key 还有数据！这是之前累积的访客数据")
		fmt.Println("  💡 建议：手动执行数据迁移，或清空旧 key 从今天重新开始")
	}
	
	hasError := false
	for _, uv := range uniqueViews {
		if uv.ViewsCount == 0 {
			hasError = true
			break
		}
	}
	if hasError {
		fmt.Println("  ❌ 数据库存在 views_count=0 的错误记录")
		fmt.Println("  💡 原因：之前 UniqueViewJob 使用了错误的 Redis key，导致写入空数据")
		fmt.Println("  💡 建议：执行下面的 SQL 清理错误数据")
	}

	fmt.Println("\n【4. 修复建议】")
	fmt.Println("  方案1: 清理错误数据（推荐）")
	fmt.Println("    DELETE FROM t_unique_view WHERE views_count = 0 AND create_time >= '2026-04-09';")
	fmt.Println("    说明：删除 views_count=0 的错误记录，前端图表会自动隐藏这些天")
	fmt.Println()
	fmt.Println("  方案2: 从今天重新开始")
	fmt.Println("    Redis 中执行: DEL unique_visitor")
	fmt.Println("    说明：清空旧的全局 key，从今天开始使用新的按天 key")
	fmt.Println()
	fmt.Println("  注意：12-14号的历史访客数据已丢失（Redis旧key被删除），无法恢复")
	fmt.Println("        从今天（15号）开始，数据会正确记录和显示")
}
