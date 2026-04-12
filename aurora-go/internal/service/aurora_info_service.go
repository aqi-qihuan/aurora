package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/model"
	"gorm.io/gorm"
)

// AuroraInfoService 首页信息聚合服务 (对标 Java AuroraInfoController + BlogInfoService)
// 使用 goroutine + sync.WaitGroup 并发查询6大数据模块
type AuroraInfoService struct {
	db           *gorm.DB
	statsService *RedisStatsService // Redis 统计服务
}

func NewAuroraInfoService(db *gorm.DB, statsService *RedisStatsService) *AuroraInfoService {
	return &AuroraInfoService{
		db:           db,
		statsService: statsService,
	}
}

// GetHomeInfo 获取首页聚合数据 (前台首页, 对标 /api/home/info)
// 并发查询: 文章列表/置顶推荐/分类列表/标签云/友链/说说
func (s *AuroraInfoService) GetHomeInfo(ctx context.Context) (*dto.HomeInfoDTO, error) {
	var info dto.HomeInfoDTO
	var wg sync.WaitGroup

	// 1. 置顶/推荐文章 (5篇)
	wg.Add(1)
	go func() {
		defer wg.Done()
		topArticles, err := s.getTopArticles(ctx)
		if err != nil {
			slog.Warn("获取置顶文章失败", "error", err.Error())
			return
		}
		info.TopArticles = topArticles
	}()

	// 2. 最新文章 (10篇)
	wg.Add(1)
	go func() {
		defer wg.Done()
		latestArticles, err := s.getLatestArticles(ctx)
		if err != nil {
			slog.Warn("获取最新文章失败", "error", err.Error())
			return
		}
		info.LatestArticles = latestArticles
	}()

	// 3. 分类列表
	wg.Add(1)
	go func() {
		defer wg.Done()
		categories, err := s.getCategories(ctx)
		if err != nil {
			slog.Warn("获取分类失败", "error", err.Error())
			return
		}
		info.Categories = categories
	}()

	// 4. 标签云(前20个)
	wg.Add(1)
	go func() {
		defer wg.Done()
		tags, err := s.getTags(ctx)
		if err != nil {
			slog.Warn("获取标签失败", "error", err.Error())
			return
		}
		info.Tags = tags
	}()

	// 5. 友链列表
	wg.Add(1)
	go func() {
		defer wg.Done()
		links, err := s.getFriendLinks(ctx)
		if err != nil {
			slog.Warn("获取友链失败", "error", err.Error())
			return
		}
		info.FriendLinks = links
	}()

	// 6. 最新说说(5条)
	wg.Add(1)
	go func() {
		defer wg.Done()
		talks, err := s.getTalks(ctx)
		if err != nil {
			slog.Warn("获取说说失败", "error", err.Error())
			return
		}
		info.Talks = talks
	}()

	wg.Wait()

	return &info, nil
}

// GetAdminDashboard 后台管理首页数据 (对标 Java AuroraInfoServiceImpl.getAuroraAdminInfo)
func (s *AuroraInfoService) GetAdminDashboard(ctx context.Context) (*dto.AuroraAdminInfoDTO, error) {
	var info dto.AuroraAdminInfoDTO
	var wg sync.WaitGroup

	// 1. 总浏览量 (从 Redis 获取)
	if s.statsService != nil {
		views, _ := s.statsService.GetTotalViews(ctx)
		info.ViewsCount = int(views)
	} else {
		info.ViewsCount = 0
	}

	// 2. 留言数 (type=2 的评论，对标 Java 第155行)
	wg.Add(1)
	go func() {
		defer wg.Done()
		var count int64
		s.db.WithContext(ctx).Model(&model.Comment{}).
			Where("type = 2").
			Count(&count)
		info.MessageCount = int(count)
	}()

	// 3. 用户数 (对标 Java 第156行)
	wg.Add(1)
	go func() {
		defer wg.Done()
		var count int64
		s.db.WithContext(ctx).Model(&model.UserInfo{}).Count(&count)
		info.UserCount = int(count)
	}()

	// 4. 文章数 (is_delete=0，对标 Java 第157-158行)
	wg.Add(1)
	go func() {
		defer wg.Done()
		var count int64
		s.db.WithContext(ctx).Model(&model.Article{}).
			Where("is_delete = 0").
			Count(&count)
		info.ArticleCount = int(count)
	}()

	// 5. 独立访客统计 (最近7天，从数据库 t_unique_view 获取)
	wg.Add(1)
	go func() {
		defer wg.Done()
		// 对标 Java UniqueViewServiceImpl.listUniqueViews()
		// 查询最近7天的数据
		startTime := time.Now().AddDate(0, 0, -7)
		endTime := time.Now()
		
		var uniqueViews []model.UniqueView
		s.db.WithContext(ctx).
			Where("create_time >= ? AND create_time <= ?", startTime, endTime).
			Order("create_time ASC").
			Find(&uniqueViews)
		
		info.UniqueViewDTOs = make([]dto.UniqueViewDTO, len(uniqueViews))
		for i, uv := range uniqueViews {
			info.UniqueViewDTOs[i] = dto.UniqueViewDTO{
				Day:        uv.CreateTime.Format("2006-01-02"),
				ViewsCount: uv.ViewsCount,
			}
		}
	}()

	// 6. 文章统计 (按日期分组，对标 Java 第160行)
	wg.Add(1)
	go func() {
		defer wg.Done()
		type Result struct {
			Date  string
			Count int
		}
		var results []Result
		s.db.WithContext(ctx).
			Model(&model.Article{}).
			Select("DATE(create_time) as date, COUNT(*) as count").
			Where("is_delete = 0").
			Group("DATE(create_time)").
			Order("date DESC").
			Limit(7).
			Find(&results)

		info.ArticleStatistics = make([]dto.ArticleStatisticsDTO, len(results))
		for i, r := range results {
			info.ArticleStatistics[i] = dto.ArticleStatisticsDTO{
				Date:  r.Date,
				Count: r.Count,
			}
		}
	}()

	// 7. 分类列表 (对标Java CategoryMapper.xml listCategories: SQL JOIN统计)
	wg.Add(1)
	go func() {
		defer wg.Done()
		
		// 使用 SQL 直接统计每个分类的文章数量（对标 Java Mapper XML）
		type CategoryWithCount struct {
			ID           uint
			CategoryName string
			ArticleCount int
		}
		
		var categories []CategoryWithCount
		s.db.WithContext(ctx).
			Table("t_category c").
			Select("c.id, c.category_name, COUNT(a.id) as article_count").
			Joins("LEFT JOIN t_article a ON c.id = a.category_id AND a.is_delete = 0 AND a.status IN (1, 2)").
			Group("c.id").
			Find(&categories)

		info.CategoryDTOs = make([]dto.CategoryDTO, len(categories))
		for i, c := range categories {
			info.CategoryDTOs[i] = dto.CategoryDTO{
				ID:           c.ID,
				CategoryName: c.CategoryName,
				ArticleCount: c.ArticleCount,
			}
		}
	}()

	// 8. 标签列表 (从 Redis 获取文章计数)
	wg.Add(1)
	go func() {
		defer wg.Done()
		var tags []model.Tag
		s.db.WithContext(ctx).
			Select("id, tag_name").
			Order("create_time DESC").
			Limit(20).
			Find(&tags)

		info.TagDTOs = make([]dto.TagDTO, len(tags))
		for i, t := range tags {
			// TODO: 如果需要标签文章计数，可以从 Redis 获取
			// if s.statsService != nil {
			// 	articleCount, _ = s.statsService.GetTagArticleCount(ctx, t.ID)
			// }
			info.TagDTOs[i] = dto.TagDTO{
				ID:      t.ID,
				TagName: t.TagName,
			}
		}
	}()

	wg.Wait()

	// 9. 文章浏览量排行 (从 Redis ZSet 获取)
	if s.statsService != nil {
		topArticles, err := s.statsService.GetTopViewedArticles(ctx, 5)
		if err == nil && len(topArticles) > 0 {
			info.ArticleRankDTOs = make([]dto.ArticleRankDTO, len(topArticles))
			for i, item := range topArticles {
				var articleID uint
				fmt.Sscanf(item.Member.(string), "%d", &articleID)
				
				// 查询文章标题
				var article model.Article
				s.db.WithContext(ctx).Select("id, article_title").First(&article, articleID)
				
				info.ArticleRankDTOs[i] = dto.ArticleRankDTO{
					ArticleTitle: article.ArticleTitle,
					ViewsCount:   int(item.Score),
				}
			}
		} else {
			info.ArticleRankDTOs = []dto.ArticleRankDTO{}
		}
	} else {
		info.ArticleRankDTOs = []dto.ArticleRankDTO{}
	}

	return &info, nil
}

// ===== 私有并发查询方法 =====

func (s *AuroraInfoService) getTopArticles(ctx context.Context) ([]dto.ArticleCardDTO, error) {
	var articles []model.Article
	err := s.db.WithContext(ctx).
		Where("is_delete = 0 AND status = 1 AND (is_top = 1 OR is_featured = 1)").
		Preload("Category").
		Preload("UserInfo").
		Order("is_top DESC, create_time DESC").
		Limit(5).
		Find(&articles).Error
	
	list := make([]dto.ArticleCardDTO, len(articles))
	for i, a := range articles {
		list[i] = toSimpleArticleCard(&a)
	}
	return list, err
}

func (s *AuroraInfoService) getLatestArticles(ctx context.Context) ([]dto.ArticleCardDTO, error) {
	var articles []model.Article
	err := s.db.WithContext(ctx).
		Where("is_delete = 0 AND status = 1").
		Preload("Category").
		Preload("UserInfo").
		Order("create_time DESC").
		Limit(10).
		Find(&articles).Error

	list := make([]dto.ArticleCardDTO, len(articles))
	for i, a := range articles {
		list[i] = toSimpleArticleCard(&a)
	}
	return list, err
}

func (s *AuroraInfoService) getCategories(ctx context.Context) ([]dto.CategoryDTO, error) {
	var categories []model.Category
	err := s.db.WithContext(ctx).
		Select("id, category_name").
		Find(&categories).Error

	list := make([]dto.CategoryDTO, len(categories))
	for i, c := range categories {
		list[i] = dto.CategoryDTO{ID: c.ID, CategoryName: c.CategoryName, ArticleCount: 0}
	}
	return list, err
}

func (s *AuroraInfoService) getTags(ctx context.Context) ([]dto.TagDTO, error) {
	var tags []model.Tag
	err := s.db.WithContext(ctx).
		Select("id, tag_name").
		Order("create_time DESC").
		Limit(20).
		Find(&tags).Error

	list := make([]dto.TagDTO, len(tags))
	for i, t := range tags {
		list[i] = dto.TagDTO{ID: t.ID, TagName: t.TagName}
	}
	return list, err
}

func (s *AuroraInfoService) getFriendLinks(ctx context.Context) ([]dto.FriendLinkDTO, error) {
	var links []model.FriendLink
	err := s.db.WithContext(ctx).
		Where("status = 1").
		Order("create_time ASC").
		Find(&links).Error

	list := make([]dto.FriendLinkDTO, len(links))
	for i, l := range links {
		list[i] = dto.FriendLinkDTO{
			ID:          l.ID,
			LinkName:    l.LinkName,
			LinkAvatar:  l.LinkAvatar,
			LinkAddress: l.LinkAddress,
			LinkIntro:   l.LinkIntro,
		}
	}
	return list, err
}

func (s *AuroraInfoService) getTalks(ctx context.Context) ([]dto.TalkDTO, error) {
	var talks []model.Talk
	err := s.db.WithContext(ctx).
		Where("status = 1").
		Preload("UserInfo").
		Order("create_time DESC").
		Limit(5).
		Find(&talks).Error

	list := make([]dto.TalkDTO, len(talks))
	for i, t := range talks {
		content := t.Content
		if len(content) > 100 {
			content = content[:100] + "..."
		}
		list[i] = dto.TalkDTO{
			ID:         t.ID,
			UserID:     t.UserID,
			Content:    content,
			LikeCount:  s.getTalkLikeCount(t.ID),
			CreateTime: t.CreateTime,
		}
		if t.UserInfo != nil {
			list[i].Nickname = t.UserInfo.Nickname
			list[i].Avatar = t.UserInfo.Avatar
		}
	}
	return list, err
}

// 辅助转换函数 (避免循环引用)
func toSimpleArticleCard(a *model.Article) dto.ArticleCardDTO {
	card := dto.ArticleCardDTO{
		ID:           a.ID,
		ArticleTitle: a.ArticleTitle,
		ArticleCover: a.ArticleCover,
		IsTop:        a.IsTop,
		IsFeatured:   a.IsFeatured,
		Status:       a.Status,
		ViewCount:    0, // TODO: 需要从外部传入 statsService 或改为方法
		CreateTime:   a.CreateTime,
	}
	if a.Category != nil {
		card.CategoryName = a.Category.CategoryName
	}
	if a.UserInfo != nil {
		card.Nickname = a.UserInfo.Nickname
	}
	return card
}

// getTalkLikeCount 获取说说点赞数（从 Redis）
func (s *AuroraInfoService) getTalkLikeCount(talkID uint) int64 {
	if s.statsService == nil {
		return 0
	}
	count, _ := s.statsService.GetTalkLikeCount(context.Background(), talkID)
	return count
}
