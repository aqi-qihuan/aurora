package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/redis/go-redis/v9"

	"github.com/aurora-go/aurora/internal/constant"
	"github.com/aurora-go/aurora/internal/model"
	"github.com/aurora-go/aurora/internal/util"
	"gorm.io/gorm"
)

// UserAreaDTO 用户地域分布DTO (对标Java com.aurora.model.dto.UserAreaDTO)
type UserAreaDTO struct {
	Name  string `json:"name"`  // 地区名称(省份)
	Value int64  `json:"value"` // 用户数量
}

// UserAreaJob 用户地域分布统计
// 对标 Java AuroraQuartz.statisticalUserArea() (Cron: 0 0,30 * * * ?, ID=81)
//
// 业务逻辑:
// 1. 查询 t_user_auth 表所有用户的 IPSource 字段
// 2. 对每个 IPSource 解析省份名 (IpUtil.getIpProvince)
// 3. 按省份分组计数 (Collectors.groupingBy + Collectors.counting)
// 4. 序列化为 JSON 并存入 Redis (USER_AREA key)
type UserAreaJob struct {
	db  *gorm.DB
	rdb *redis.Client
}

// NewUserAreaJob 创建用户地域分布统计任务实例
func NewUserAreaJob(db *gorm.DB, rdb *redis.Client) *UserAreaJob {
	return &UserAreaJob{db: db, rdb: rdb}
}

// Run 执行地域统计任务
func (j *UserAreaJob) Run(ctx context.Context, params ...interface{}) error {
	// Step 1: 查询所有用户的 IPSource (对标Java userAuthMapper.selectList(...select IpSource))
	var userAuths []model.UserAuth
	if err := j.db.WithContext(ctx).
		Select("ip_source").
		Where("ip_source != ?", "").
		Find(&userAuths).Error; err != nil {
		return fmt.Errorf("failed to query user auth IPs: %w", err)
	}

	// Step 2: 提取省份并分组计数 (对标Java Stream.map→groupingBy→counting)
	areaMap := make(map[string]int64)
	
	// 调试：先输出第一条原始数据，确认数据库格式
	if len(userAuths) > 0 {
		firstIP := userAuths[0].IPSource
		parts := strings.Split(firstIP, "|")
		slog.Info("[DEBUG] 数据库IPSource格式", "raw", firstIP, "parts_count", len(parts), "parts", parts)
	}
	
	for _, ua := range userAuths {
		province := util.GetProvince(ua.IPSource)
		
		if province == "" || province == "未知" {
			province = "未知"
		}
		areaMap[province]++
	}

	// Step 3: 构建DTO列表 (对标Java List<UserAreaDTO>)
	userAreaList := make([]UserAreaDTO, 0, len(areaMap))
	for name, count := range areaMap {
		userAreaList = append(userAreaList, UserAreaDTO{
			Name:  name,
			Value: count,
		})
	}

	// Step 4: 存入Redis (对标Java redisService.set(USER_AREA, JSON.toJSONString(list)))
	jsonBytes, err := json.Marshal(userAreaList)
	if err != nil {
		return fmt.Errorf("failed to marshal user area list: %w", err)
	}

	// 重要：将 []byte 转换为 string，避免 Redis 双重序列化
	if err := j.rdb.Set(ctx, constant.UserArea, string(jsonBytes), 0).Err(); err != nil {
		return fmt.Errorf("failed to save user area to Redis: %w", err)
	}

	slog.Info("用户地域分布统计保存成功", "provinces", len(areaMap), "total_users", len(userAuths))
	return nil
}
