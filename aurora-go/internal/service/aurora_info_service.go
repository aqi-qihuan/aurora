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
	db *gorm.DB
}

func NewAuroraInfoService(db *gorm.DB) *AuroraInfoService {
	return &AuroraInfoService{db: db}
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

// GetAdminDashboard 后台管理仪表盘数据 (对标 Java AdminController.getAdminInfo)
func (s *AuroraInfoService) GetAdminDashboard(ctx context.Context) (*dto.AdminDashboardDTO, error) {
	var dashboard dto.AdminDashboardDTO
	var wg sync.WaitGroup

	// 并发统计各维度数据
	wg.Add(7)

	go func() { // 总文章数
		defer wg.Done()
		s.db.WithContext(ctx).Model(&model.Article{}).Where("is_delete = 0").Count(&dashboard.TotalArticles)
	}()

	go func() { // 总用户数
		defer wg.Done()
		s.db.WithContext(ctx).Model(&model.UserInfo{}).Count(&dashboard.TotalUsers)
	}()

	go func() { // 总评论数
		defer wg.Done()
		s.db.WithContext(ctx).Model(&model.Comment{}).Where("is_review = 1").Count(&dashboard.TotalComments)
	}()

	go func() { // 总浏览量
		defer wg.Done()
		s.db.WithContext(ctx).Model(&model.Article{}).
			Select("COALESCE(SUM(view_count), 0)").
			Scan(&dashboard.TotalViews)
	}()

	go func() { // 待审核评论
		defer wg.Done()
		s.db.WithContext(ctx).Model(&model.Comment{}).Where("is_review = 0").Count(&dashboard.PendingReviews)
	}()

	go func() { // 今日新增文章
		defer wg.Done()
		today := fmt.Sprintf("%d-%02d-%02d", time.Now().Year(), time.Now().Month(), time.Now().Day())
		s.db.WithContext(ctx).Model(&model.Article{}).
			Where("DATE(create_time) = ?", today).
			Count(&dashboard.TodayArticles)
	}()

	go func() { // 今日访问量
		defer wg.Done()
		// TODO: P0-7 Redis HyperLogLog 统计今日独立访客
		dashboard.UniqueVisitors = 0
	}()

	wg.Wait()
	return &dashboard, nil
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
		Select("id, category_name, article_count").
		Order("sort ASC").
		Find(&categories).Error

	list := make([]dto.CategoryDTO, len(categories))
	for i, c := range categories {
		list[i] = dto.CategoryDTO{ID: c.ID, CategoryName: c.CategoryName, ArticleCount: c.ArticleCount}
	}
	return list, err
}

func (s *AuroraInfoService) getTags(ctx context.Context) ([]dto.TagDTO, error) {
	var tags []model.Tag
	err := s.db.WithContext(ctx).
		Select("id, tag_name, article_count").
		Where("article_count > 0").
		Order("article_count DESC").
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
			LikeCount:  t.LikeCount,
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
		ViewCount:    a.ViewCount,
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
