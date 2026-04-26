//go:build ignore
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aurora-go/aurora/internal/config"
	"github.com/aurora-go/aurora/internal/constant"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"github.com/redis/go-redis/v9"
)

func main() {
	ctx := context.Background()
	
	// 加载配置
	cfg, err := config.Load("configs/config-local.yaml")
	if err != nil {
		log.Printf("Warning: load config failed: %v", err)
		// 继续执行，使用默认配置
	}

	// 连接数据库
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.MySQL.User,
		cfg.MySQL.Password,
		cfg.MySQL.Host,
		cfg.MySQL.Port,
		cfg.MySQL.DBName,
		"utf8mb4",
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	// 连接 Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Redis connection failed: %v", err)
	}

	fmt.Println("=== 一周访问趋势诊断 ===\n")
	
	// 1. 检查数据库查询条件
	now := time.Now()
	startTime := now.AddDate(0, 0, -7)
	endTime := now
	
	fmt.Printf("当前时间: %s\n", now.Format("2006-01-02 15:04:05"))
	fmt.Printf("查询范围: %s ~ %s\n\n", startTime.Format("2006-01-02 15:04:05"), endTime.Format("2006-01-02 15:04:05"))

	// 2. 查询 t_unique_view 表的所有记录
	fmt.Println("【1】t_unique_view 表所有记录:")
	type UniqueViewRow struct {
		ID         uint
		ViewsCount int
		CreateTime time.Time
	}
	var allRows []UniqueViewRow
	db.Table("t_unique_view").Select("id, views_count, create_time").Order("create_time DESC").Limit(20).Find(&allRows)
	
	for _, row := range allRows {
		fmt.Printf("  ID=%d, views_count=%d, create_time=%s\n", 
			row.ID, row.ViewsCount, row.CreateTime.Format("2006-01-02 15:04:05"))
	}
	
	// 3. 模拟实际查询（带 > 和 <= 条件）
	fmt.Println("\n【2】实际查询结果（当前SQL条件）:")
	type QueryRow struct {
		Day        string
		ViewsCount int
	}
	var queryRows []QueryRow
	db.Table("t_unique_view").
		Select(`DATE_FORMAT(create_time, "%Y-%m-%d") as day, views_count`).
		Where("create_time > ? AND create_time <= ?", startTime, endTime).
		Order("create_time ASC").
		Find(&queryRows)
	
	for _, row := range queryRows {
		fmt.Printf("  day=%s, views_count=%d\n", row.Day, row.ViewsCount)
	}
	
	// 4. 修复后的查询（使用 beginOfDay 和 endOfDay）
	fmt.Println("\n【3】修复后的查询结果（使用完整的日期范围）:")
	fixedStartTime := startTime.Truncate(24 * time.Hour) // 2026-04-14 00:00:00
	fixedEndTime := endTime.Truncate(24 * time.Hour).Add(24 * time.Hour).Add(-time.Second) // 2026-04-21 23:59:59
	
	var fixedRows []QueryRow
	db.Table("t_unique_view").
		Select(`DATE_FORMAT(create_time, "%Y-%m-%d") as day, views_count`).
		Where("create_time >= ? AND create_time <= ?", fixedStartTime, fixedEndTime).
		Order("create_time ASC").
		Find(&fixedRows)
	
	for _, row := range fixedRows {
		fmt.Printf("  day=%s, views_count=%d\n", row.Day, row.ViewsCount)
	}
	
	// 5. 检查 Redis 访客数据（全局 Set，对标Java）
	fmt.Println("\n【4】Redis 访客数据:")
	globalKey := constant.UniqueVisitor
	globalCount, err := rdb.SCard(ctx, globalKey).Result()
	if err != nil {
		fmt.Printf("  全局key(%s)获取失败: %v\n", globalKey, err)
	} else {
		fmt.Printf("  全局key(%s)访客数: %d\n", globalKey, globalCount)
		
		// 查看 Set 中的成员（仅显示前10个）
		if globalCount > 0 {
			members, _ := rdb.SMembers(ctx, globalKey).Result()
			if len(members) > 10 {
				fmt.Printf("  Set 成员（前10个）: %v...\n", members[:10])
			} else {
				fmt.Printf("  Set 成员: %v\n", members)
			}
		}
	}
	
	// 检查是否有遗留的按天 key（错误实现留下的）
	fmt.Println("\n  检查遗留的按天 key:")
	for i := 6; i >= 0; i-- {
		date := now.AddDate(0, 0, -i).Format("2006-01-02")
		dailyKey := fmt.Sprintf("%s:%s", constant.UniqueVisitor, date)
		count, _ := rdb.SCard(ctx, dailyKey).Result()
		if count > 0 {
			fmt.Printf("  ⚠️  %s = %d (遗留数据，建议手动 DEL)\n", dailyKey, count)
		}
	}
	
	// 6. 生成完整7天数据
	fmt.Println("\n【5】生成的7天完整数据:")
	viewMap := make(map[string]int)
	for _, row := range fixedRows {
		viewMap[row.Day] = row.ViewsCount
	}
	
	// 补充今天的实时数据（从全局 Set 读取）
	today := now.Format("2006-01-02")
	if globalCount > 0 {
		viewMap[today] = int(globalCount)
	}
	
	for i := 6; i >= 0; i-- {
		date := now.AddDate(0, 0, -i).Format("2006-01-02")
		count := viewMap[date]
		fmt.Printf("  %s -> %d\n", date, count)
	}
	
	fmt.Println("\n=== 诊断建议 ===")
	fmt.Println("如果【2】和【3】的结果不一致，说明查询条件有问题，需要修复为 >= 和 endOfDay")
	fmt.Println("如果 Redis 数据为 0，说明前端访客上报没有记录到 Redis，需要检查 aurora_info_handler.go")
}
