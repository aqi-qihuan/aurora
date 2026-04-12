package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/aurora-go/aurora/internal/constant"
	"github.com/redis/go-redis/v9"
)

// RedisStatsService Redis 统计服务
// 负责管理: 文章浏览量、点赞数、分类/标签文章计数、访客统计等
type RedisStatsService struct {
	rdb *redis.Client
}

func NewRedisStatsService(rdb *redis.Client) *RedisStatsService {
	return &RedisStatsService{rdb: rdb}
}

// ===== 文章浏览量统计 =====

// IncrementArticleView 增加文章浏览量 (原子操作)
func (s *RedisStatsService) IncrementArticleView(ctx context.Context, articleID uint) error {
	key := fmt.Sprintf("%s%d", constant.ArticleViewsCount, articleID)
	
	// 1. 增加单篇文章浏览量
	if err := s.rdb.Incr(ctx, key).Err(); err != nil {
		return fmt.Errorf("增加文章浏览量失败: %w", err)
	}
	
	// 2. 设置过期时间（30天自动清理）
	s.rdb.Expire(ctx, key, 30*24*time.Hour)
	
	// 3. 更新总浏览量
	s.rdb.Incr(ctx, constant.BlogViewsCount)
	
	// 4. 更新浏览排行 ZSet
	s.rdb.ZIncrBy(ctx, constant.ArticleViewsRanking, 1, fmt.Sprintf("%d", articleID))
	
	return nil
}

// GetArticleView 获取文章浏览量
func (s *RedisStatsService) GetArticleView(ctx context.Context, articleID uint) (uint64, error) {
	key := fmt.Sprintf("%s%d", constant.ArticleViewsCount, articleID)
	count, err := s.rdb.Get(ctx, key).Uint64()
	if err == redis.Nil {
		return 0, nil // 不存在返回0
	}
	if err != nil {
		return 0, fmt.Errorf("获取文章浏览量失败: %w", err)
	}
	return count, nil
}

// GetTotalViews 获取总浏览量
func (s *RedisStatsService) GetTotalViews(ctx context.Context) (uint64, error) {
	count, err := s.rdb.Get(ctx, constant.BlogViewsCount).Uint64()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("获取总浏览量失败: %w", err)
	}
	return count, nil
}

// GetTopViewedArticles 获取浏览量最高的文章 (TOP N)
func (s *RedisStatsService) GetTopViewedArticles(ctx context.Context, limit int64) ([]redis.Z, error) {
	items, err := s.rdb.ZRevRangeWithScores(ctx, constant.ArticleViewsRanking, 0, limit-1).Result()
	if err != nil {
		return nil, fmt.Errorf("获取浏览排行失败: %w", err)
	}
	return items, nil
}

// SyncArticleViewsToDB 同步文章浏览量到数据库 (定时任务调用)
func (s *RedisStatsService) SyncArticleViewsToDB(ctx context.Context, updateDB func(articleID uint, views uint64) error) error {
	// 获取所有文章的浏览量
	keys, err := s.rdb.Keys(ctx, constant.ArticleViewsCount+"*").Result()
	if err != nil {
		return fmt.Errorf("获取浏览量keys失败: %w", err)
	}
	
	for _, key := range keys {
		views, err := s.rdb.Get(ctx, key).Uint64()
		if err != nil {
			continue
		}
		
		// 提取 articleID
		var articleID uint
		fmt.Sscanf(key, constant.ArticleViewsCount+"%d", &articleID)
		
		if articleID > 0 && views > 0 {
			if err := updateDB(articleID, views); err != nil {
				slog.Error("同步文章浏览量到DB失败", "article_id", articleID, "error", err)
			}
		}
	}
	
	return nil
}

// ===== 点赞统计 =====

// LikeArticle 点赞文章
func (s *RedisStatsService) LikeArticle(ctx context.Context, articleID uint, userID uint) error {
	key := fmt.Sprintf("article:likes:%d", articleID)
	
	// 使用 Set 存储点赞用户，避免重复点赞
	added, err := s.rdb.SAdd(ctx, key, userID).Result()
	if err != nil {
		return fmt.Errorf("点赞失败: %w", err)
	}
	
	if added == 0 {
		return fmt.Errorf("已经点赞过")
	}
	
	// 设置过期时间
	s.rdb.Expire(ctx, key, 30*24*time.Hour)
	
	return nil
}

// UnlikeArticle 取消点赞
func (s *RedisStatsService) UnlikeArticle(ctx context.Context, articleID uint, userID uint) error {
	key := fmt.Sprintf("article:likes:%d", articleID)
	_, err := s.rdb.SRem(ctx, key, userID).Result()
	if err != nil {
		return fmt.Errorf("取消点赞失败: %w", err)
	}
	return nil
}

// GetArticleLikeCount 获取文章点赞数
func (s *RedisStatsService) GetArticleLikeCount(ctx context.Context, articleID uint) (int64, error) {
	key := fmt.Sprintf("article:likes:%d", articleID)
	count, err := s.rdb.SCard(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("获取点赞数失败: %w", err)
	}
	return count, nil
}

// HasLikedArticle 检查用户是否已点赞
func (s *RedisStatsService) HasLikedArticle(ctx context.Context, articleID uint, userID uint) (bool, error) {
	key := fmt.Sprintf("article:likes:%d", articleID)
	exists, err := s.rdb.SIsMember(ctx, key, userID).Result()
	if err != nil {
		return false, fmt.Errorf("检查点赞状态失败: %w", err)
	}
	return exists, nil
}

// LikeComment 点赞评论
func (s *RedisStatsService) LikeComment(ctx context.Context, commentID uint, userID uint) error {
	key := fmt.Sprintf("comment:likes:%d", commentID)
	added, err := s.rdb.SAdd(ctx, key, userID).Result()
	if err != nil {
		return fmt.Errorf("点赞评论失败: %w", err)
	}
	if added == 0 {
		return fmt.Errorf("已经点赞过")
	}
	s.rdb.Expire(ctx, key, 30*24*time.Hour)
	return nil
}

// GetCommentLikeCount 获取评论点赞数
func (s *RedisStatsService) GetCommentLikeCount(ctx context.Context, commentID uint) (int64, error) {
	key := fmt.Sprintf("comment:likes:%d", commentID)
	count, err := s.rdb.SCard(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("获取评论点赞数失败: %w", err)
	}
	return count, nil
}

// LikeTalk 点赞说说
func (s *RedisStatsService) LikeTalk(ctx context.Context, talkID uint, userID uint) error {
	key := fmt.Sprintf("talk:likes:%d", talkID)
	added, err := s.rdb.SAdd(ctx, key, userID).Result()
	if err != nil {
		return fmt.Errorf("点赞说说失败: %w", err)
	}
	if added == 0 {
		return fmt.Errorf("已经点赞过")
	}
	s.rdb.Expire(ctx, key, 30*24*time.Hour)
	return nil
}

// GetTalkLikeCount 获取说说点赞数
func (s *RedisStatsService) GetTalkLikeCount(ctx context.Context, talkID uint) (int64, error) {
	key := fmt.Sprintf("talk:likes:%d", talkID)
	count, err := s.rdb.SCard(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("获取说说点赞数失败: %w", err)
	}
	return count, nil
}

// ===== 分类/标签文章计数 =====

// IncrementCategoryArticleCount 增加分类文章数
func (s *RedisStatsService) IncrementCategoryArticleCount(ctx context.Context, categoryID uint) error {
	key := fmt.Sprintf("category:article_count:%d", categoryID)
	return s.rdb.Incr(ctx, key).Err()
}

// DecrementCategoryArticleCount 减少分类文章数
func (s *RedisStatsService) DecrementCategoryArticleCount(ctx context.Context, categoryID uint) error {
	key := fmt.Sprintf("category:article_count:%d", categoryID)
	result, err := s.rdb.Decr(ctx, key).Result()
	if err != nil {
		return err
	}
	// 保证不为负数
	if result < 0 {
		s.rdb.Set(ctx, key, 0, 0)
	}
	return nil
}

// GetCategoryArticleCount 获取分类文章数
func (s *RedisStatsService) GetCategoryArticleCount(ctx context.Context, categoryID uint) (int64, error) {
	key := fmt.Sprintf("category:article_count:%d", categoryID)
	count, err := s.rdb.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return count, nil
}

// IncrementTagArticleCount 增加标签文章数
func (s *RedisStatsService) IncrementTagArticleCount(ctx context.Context, tagID uint) error {
	key := fmt.Sprintf("tag:article_count:%d", tagID)
	return s.rdb.Incr(ctx, key).Err()
}

// DecrementTagArticleCount 减少标签文章数
func (s *RedisStatsService) DecrementTagArticleCount(ctx context.Context, tagID uint) error {
	key := fmt.Sprintf("tag:article_count:%d", tagID)
	result, err := s.rdb.Decr(ctx, key).Result()
	if err != nil {
		return err
	}
	if result < 0 {
		s.rdb.Set(ctx, key, 0, 0)
	}
	return nil
}

// GetTagArticleCount 获取标签文章数
func (s *RedisStatsService) GetTagArticleCount(ctx context.Context, tagID uint) (int64, error) {
	key := fmt.Sprintf("tag:article_count:%d", tagID)
	count, err := s.rdb.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return count, nil
}

// ===== 访客统计 =====

// RecordUniqueVisitor 记录独立访客
func (s *RedisStatsService) RecordUniqueVisitor(ctx context.Context, ip string) error {
	today := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("%s:%s", constant.UniqueVisitor, today)
	
	// 使用 Set 去重
	_, err := s.rdb.SAdd(ctx, key, ip).Result()
	if err != nil {
		return err
	}
	
	// 设置过期时间（7天）
	s.rdb.Expire(ctx, key, 7*24*time.Hour)
	
	return nil
}

// GetTodayUniqueVisitors 获取今日独立访客数
func (s *RedisStatsService) GetTodayUniqueVisitors(ctx context.Context) (int64, error) {
	today := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("%s:%s", constant.UniqueVisitor, today)
	count, err := s.rdb.SCard(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return count, nil
}

// RecordVisitorArea 记录访客地域
func (s *RedisStatsService) RecordVisitorArea(ctx context.Context, area string) error {
	today := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("%s:%s", constant.VisitorArea, today)
	
	// 使用 Hash 统计各地域访问次数
	_, err := s.rdb.HIncrBy(ctx, key, area, 1).Result()
	if err != nil {
		return err
	}
	
	s.rdb.Expire(ctx, key, 7*24*time.Hour)
	return nil
}

// GetVisitorAreaStats 获取访客地域统计
func (s *RedisStatsService) GetVisitorAreaStats(ctx context.Context, days int) (map[string]int64, error) {
	stats := make(map[string]int64)
	
	for i := 0; i < days; i++ {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		key := fmt.Sprintf("%s:%s", constant.VisitorArea, date)
		
		areaData, err := s.rdb.HGetAll(ctx, key).Result()
		if err != nil {
			continue
		}
		
		for area, countStr := range areaData {
			var count int64
			fmt.Sscanf(countStr, "%d", &count)
			stats[area] += count
		}
	}
	
	return stats, nil
}

// ===== 缓存辅助方法 =====

// CacheSet 设置缓存
func (s *RedisStatsService) CacheSet(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, key, data, expiration).Err()
}

// CacheGet 获取缓存
func (s *RedisStatsService) CacheGet(ctx context.Context, key string, dest interface{}) error {
	data, err := s.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// CacheDelete 删除缓存
func (s *RedisStatsService) CacheDelete(ctx context.Context, pattern string) error {
	keys, err := s.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return s.rdb.Del(ctx, keys...).Err()
	}
	return nil
}
