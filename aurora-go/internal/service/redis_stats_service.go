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
	// 使用 ZSet 存储文章浏览量（对齐 Java 版实现）
	// 1. 增加单篇文章浏览量（ZSet 的 score）
	if err := s.rdb.ZIncrBy(ctx, constant.ArticleViewsRanking, 1, fmt.Sprintf("%d", articleID)).Err(); err != nil {
		return fmt.Errorf("增加文章浏览量失败: %w", err)
	}
	
	// 2. 更新总浏览量
	s.rdb.Incr(ctx, constant.BlogViewsCount)
	
	return nil
}

// GetArticleView 获取文章浏览量
func (s *RedisStatsService) GetArticleView(ctx context.Context, articleID uint) (uint64, error) {
	// 使用 ZSet 获取文章浏览量（对齐 Java 版实现）
	score, err := s.rdb.ZScore(ctx, constant.ArticleViewsRanking, fmt.Sprintf("%d", articleID)).Result()
	if err == redis.Nil {
		return 0, nil // 不存在返回0
	}
	if err != nil {
		return 0, fmt.Errorf("获取文章浏览量失败: %w", err)
	}
	return uint64(score), nil
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
	// 获取 ZSet 中所有文章的浏览量（对齐 Java 版实现）
	items, err := s.rdb.ZRevRangeWithScores(ctx, constant.ArticleViewsRanking, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("获取浏览量失败: %w", err)
	}
	
	for _, item := range items {
		var articleID uint
		fmt.Sscanf(item.Member.(string), "%d", &articleID)
		
		if articleID > 0 && item.Score > 0 {
			if err := updateDB(articleID, uint64(item.Score)); err != nil {
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

// RecordUniqueVisitor 记录独立访客（使用全局 Set，对标Java）
// 对标Java: redisService.sAdd(UNIQUE_VISITOR, md5)
func (s *RedisStatsService) RecordUniqueVisitor(ctx context.Context, ip string) error {
	// Java 版使用全局 Set: unique_visitor
	// 定时任务每天凌晨0点读取后清空
	key := constant.UniqueVisitor
	
	// 使用 Set 去重
	_, err := s.rdb.SAdd(ctx, key, ip).Result()
	if err != nil {
		return err
	}
	
	// 设置过期时间（3 天，防止内存泄漏）
	s.rdb.Expire(ctx, key, 72*time.Hour)
	
	return nil
}

// RecordUniqueVisitorByFingerprint 记录独立访客（使用 MD5 指纹，与检查逻辑一致）
// 对标Java: redisService.sAdd(UNIQUE_VISITOR, md5)
func (s *RedisStatsService) RecordUniqueVisitorByFingerprint(ctx context.Context, fingerprint string) error {
	// Java 版使用全局 Set: unique_visitor（凌晨0点定时任务读取后清空）
	key := constant.UniqueVisitor
	
	// 使用 Set 去重
	_, err := s.rdb.SAdd(ctx, key, fingerprint).Result()
	if err != nil {
		return err
	}
	
	// 设置过期时间（3天，防止内存泄漏）
	s.rdb.Expire(ctx, key, 72*time.Hour)
	
	return nil
}

// IsUniqueVisitor 检查是否为独立访客（根据MD5指纹判断）
// 对标Java: redisService.sIsMember(UNIQUE_VISITOR, md5)
func (s *RedisStatsService) IsUniqueVisitor(ctx context.Context, md5Fingerprint string) (bool, error) {
	// Java 版使用全局 key: unique_visitor
	return s.rdb.SIsMember(ctx, constant.UniqueVisitor, md5Fingerprint).Result()
}

// MarkUniqueVisitor 标记独立访客（将MD5指纹加入Set）
// 对标Java: redisService.sAdd(UNIQUE_VISITOR, md5)
func (s *RedisStatsService) MarkUniqueVisitor(ctx context.Context, md5Fingerprint string) error {
	// Java 版使用全局 key: unique_visitor
	_, err := s.rdb.SAdd(ctx, constant.UniqueVisitor, md5Fingerprint).Result()
	if err != nil {
		return err
	}
	
	// 设置过期时间（3天，防止内存泄漏）
	s.rdb.Expire(ctx, constant.UniqueVisitor, 72*time.Hour)
	return nil
}

// IncrementTotalViews 增加总浏览量（PV）
// 对标Java: redisService.incr(BLOG_VIEWS_COUNT, 1)
func (s *RedisStatsService) IncrementTotalViews(ctx context.Context) error {
	return s.rdb.Incr(ctx, constant.BlogViewsCount).Err()
}

// GetTodayUniqueVisitors 获取今日独立访客数（从全局 Set 读取，对标Java）
func (s *RedisStatsService) GetTodayUniqueVisitors(ctx context.Context) (int64, error) {
	// Java 版使用全局 key: unique_visitor
	return s.rdb.SCard(ctx, constant.UniqueVisitor).Result()
}

// GetUniqueVisitorsByDate 获取指定日期的独立访客数（从全局 Set 读取）
// 对标Java: redisService.sCard(UNIQUE_VISITOR)
func (s *RedisStatsService) GetUniqueVisitorsByDate(ctx context.Context, date string) (int64, error) {
	// Java 版使用全局 key: unique_visitor
	// 定时任务每天凌晨0点归档后清空，所以只有今天的实时数据
	return s.rdb.SCard(ctx, constant.UniqueVisitor).Result()
}

// RecordVisitorArea 记录访客地域（对标Java，使用全局累积key）
func (s *RedisStatsService) RecordVisitorArea(ctx context.Context, area string) error {
	// 对标Java: redisService.hIncr(VISITOR_AREA, ipProvince, 1L)
	// 使用全局累积key（无日期后缀），与读取端保持一致
	key := constant.VisitorArea
	
	// 使用 Hash 统计各地域访问次数
	_, err := s.rdb.HIncrBy(ctx, key, area, 1).Result()
	if err != nil {
		return err
	}
	
	// 设置过期时间（30天，定期清理）
	s.rdb.Expire(ctx, key, 30*24*time.Hour)
	return nil
}

// GetVisitorAreaStats 获取访客地域统计（从全局累积key读取）
func (s *RedisStatsService) GetVisitorAreaStats(ctx context.Context, days int) (map[string]int64, error) {
	// 对标Java: redisService.hGetAll(VISITOR_AREA)
	// 直接从全局key读取，无需按日期聚合
	key := constant.VisitorArea
	
	areaData, err := s.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	
	stats := make(map[string]int64)
	for area, countStr := range areaData {
		var count int64
		fmt.Sscanf(countStr, "%d", &count)
		stats[area] += count
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
