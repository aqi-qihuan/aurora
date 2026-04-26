//go:build ignore

package main

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	dsn := "root:aqi1015@tcp(134.175.206.158:3306)/aurora?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== 评论表 type 分布统计 ===")
	type TypeCount struct {
		Type int8
		Cnt  int64
	}
	var typeCounts []TypeCount
	db.Table("t_comment").Select("type, COUNT(*) as cnt").Group("type").Scan(&typeCounts)
	for _, tc := range typeCounts {
		fmt.Printf("  type=%d: %d 条\n", tc.Type, tc.Cnt)
	}

	fmt.Println("\n=== type IN (1, 2, 5) 的总数 ===")
	var total int64
	db.Table("t_comment").Where("type IN (1, 2, 5)").Count(&total)
	fmt.Printf("  总计: %d 条\n", total)

	fmt.Println("\n=== type=2 (留言板) 的原始数据 ===")
	type MsgRow struct {
		ID      uint
		Content string
		Parent  uint
		Type    int8
	}
	var msgs []MsgRow
	db.Table("t_comment").Select("id, comment_content, parent_id, type").Where("type = 2").Find(&msgs)
	for _, m := range msgs {
		short := m.Content
		if len(short) > 30 {
			short = short[:30] + "..."
		}
		fmt.Printf("  id=%d type=%d parent=%d content=%q\n", m.ID, m.Type, m.Parent, short)
	}

	fmt.Println("\n=== type=1 (文章评论) 的原始数据 ===")
	var arts []MsgRow
	db.Table("t_comment").Select("id, comment_content, parent_id, type").Where("type = 1").Find(&arts)
	for _, m := range arts {
		short := m.Content
		if len(short) > 30 {
			short = short[:30] + "..."
		}
		fmt.Printf("  id=%d type=%d parent=%d content=%q\n", m.ID, m.Type, m.Parent, short)
	}

	fmt.Println("\n=== type=5 (说说评论) 的原始数据 ===")
	var talks []MsgRow
	db.Table("t_comment").Select("id, comment_content, parent_id, type").Where("type = 5").Find(&talks)
	for _, m := range talks {
		short := m.Content
		if len(short) > 30 {
			short = short[:30] + "..."
		}
		fmt.Printf("  id=%d type=%d parent=%d content=%q\n", m.ID, m.Type, m.Parent, short)
	}

	fmt.Println("\n=== 当前 API 返回的 messageCount 值 ===")
	fmt.Printf("  数据库查询结果: %d\n", total)
}
