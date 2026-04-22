package service

import (
	"context"
	"encoding/json"
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
	db                  *gorm.DB
	statsService        *RedisStatsService // Redis 统计服务
	websiteConfigService *WebsiteConfigService // 网站配置服务
}

func NewAuroraInfoService(db *gorm.DB, statsService *RedisStatsService, websiteConfigService *WebsiteConfigService) *AuroraInfoService {
	return &AuroraInfoService{
		db:                  db,
		statsService:        statsService,
		websiteConfigService: websiteConfigService,
	}
}

// GetHomeInfo 获取首页聚合数据 (前台首页, 对标 /api/home/info)
// 对标Java版 getAuroraHomeInfo()，并发查询: 文章列表/置顶推荐/分类列表/标签云/友链/说说/网站配置/统计数据
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

	// 7. 网站配置（对标Java getWebsiteConfig）
	wg.Add(1)
	go func() {
		defer wg.Done()
		config, err := s.getWebsiteConfig(ctx)
		if err != nil {
			slog.Warn("获取网站配置失败", "error", err.Error())
			return
		}
		info.WebsiteConfig = config
	}()

	// 8. 文章总数（is_delete=0，对标Java第97行）
	wg.Add(1)
	go func() {
		defer wg.Done()
		var count int64
		s.db.WithContext(ctx).Model(&model.Article{}).Where("is_delete = 0").Count(&count)
		info.ArticleCount = int(count)
	}()

	// 9. 分类总数（对标Java第104行）
	wg.Add(1)
	go func() {
		defer wg.Done()
		var count int64
		s.db.WithContext(ctx).Model(&model.Category{}).Count(&count)
		info.CategoryCount = int(count)
	}()

	// 10. 标签总数（对标Java第111行）
	wg.Add(1)
	go func() {
		defer wg.Done()
		var count int64
		s.db.WithContext(ctx).Model(&model.Tag{}).Count(&count)
		info.TagCount = int(count)
	}()

	// 11. 说说总数（对标Java第118行）
	wg.Add(1)
	go func() {
		defer wg.Done()
		var count int64
		s.db.WithContext(ctx).Model(&model.Talk{}).Count(&count)
		info.TalkCount = int(count)
	}()

	// 12. 总浏览量（从 Redis 获取，对标Java第132-133行）
	wg.Add(1)
	go func() {
		defer wg.Done()
		if s.statsService != nil {
			views, _ := s.statsService.GetTotalViews(ctx)
			info.ViewCount = int(views)
		}
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

	// 5. 独立访客统计 (最近7天，对标 Java UniqueViewServiceImpl.listUniqueViews)
	// Java: uniqueViewService.listUniqueViews() → UniqueViewMapper.xml 直接查 t_unique_view 表
	// Go增强: 合并数据库历史数据 + 今天Redis实时数据，确保数据及时性
	wg.Add(1)
	go func() {
		defer wg.Done()
		
		// 对标 Java: DateUtil.beginOfDay(DateUtil.offsetDay(new Date(), -7)) ~ DateUtil.endOfDay(new Date())
		// 关键修复: 使用 beginOfDay 和 endOfDay 确保查询完整的日期范围
		now := time.Now()
		startTime := now.AddDate(0, 0, -7).Truncate(24 * time.Hour)                     // 7天前的 00:00:00
		endTime := now.Truncate(24 * time.Hour).Add(24*time.Hour).Add(-time.Second)     // 今天的 23:59:59
		
		type UniqueViewRow struct {
			Day        string `gorm:"column:day"`
			ViewsCount int    `gorm:"column:views_count"`
		}
		var rows []UniqueViewRow
		s.db.WithContext(ctx).
			Table("t_unique_view").
			Select(`DATE_FORMAT(create_time, "%Y-%m-%d") as day, views_count`).
			Where("create_time >= ? AND create_time <= ?", startTime, endTime).  // 修复: 使用 >= 和 endOfDay
			Order("create_time ASC").
			Find(&rows)
		
		// 构建日期 -> 访客数的映射
		viewMap := make(map[string]int)
		for _, r := range rows {
			viewMap[r.Day] = r.ViewsCount
		}
		
		// 补充今天的实时数据（从 Redis 获取）
		today := now.Format("2006-01-02")
		if s.statsService != nil {
			todayCount, err := s.statsService.GetTodayUniqueVisitors(ctx)
			if err == nil && todayCount > 0 {
				viewMap[today] = int(todayCount)
			}
		}
		
		// 生成最近7天的完整数据（缺失日期补0，对齐前端趋势图需求）
		result := make([]dto.UniqueViewDTO, 0, 7)
		for i := 6; i >= 0; i-- {
			date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
			count := viewMap[date] // 没有数据时默认为0
			result = append(result, dto.UniqueViewDTO{
				Day:        date,
				ViewsCount: count,
			})
		}
		
		info.UniqueViewDTOs = result
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

	// 8. 标签列表 (对标Java TagMapper.xml listTags: SQL JOIN统计文章数)
	wg.Add(1)
	go func() {
		defer wg.Done()
		
		// 对标Java: SELECT t.id, tag_name, COUNT(aat.article_id) AS count FROM t_tag t
		// LEFT JOIN (SELECT a.id AS article_id, at.tag_id AS tag_id FROM t_article_tag at
		// LEFT JOIN t_article a ON at.article_id = a.id WHERE a.is_delete = 0 AND a.STATUS in (1, 2)) aat
		// ON t.id = aat.tag_id GROUP BY t.id
		type TagWithCount struct {
			ID           uint   `gorm:"column:id"`
			TagName      string `gorm:"column:tag_name"`
			ArticleCount int    `gorm:"column:count"`
		}
		
		var tags []TagWithCount
		s.db.WithContext(ctx).
			Table("t_tag t").
			Select(`t.id, t.tag_name, COUNT(aat.article_id) AS count`).
			Joins(`LEFT JOIN (
				SELECT a.id AS article_id, at.tag_id AS tag_id
				FROM t_article_tag at
				LEFT JOIN t_article a ON at.article_id = a.id
				WHERE a.is_delete = 0 AND a.status IN (1, 2)
			) aat ON t.id = aat.tag_id`).
			Group("t.id").
			Find(&tags)

		info.TagDTOs = make([]dto.TagDTO, len(tags))
		for i, t := range tags {
			info.TagDTOs[i] = dto.TagDTO{
				ID:           t.ID,
				TagName:      t.TagName,
				ArticleCount: t.ArticleCount,
			}
		}
	}()

	wg.Wait()

	// 9. 文章浏览量排行 (从 Redis ZSet 获取)
	if s.statsService != nil {
		topArticles, err := s.statsService.GetTopViewedArticles(ctx, 5)
		if err == nil && len(topArticles) > 0 {
			// 批量查询文章标题（避免 N+1 查询，过滤已删除的文章）
			articleIDs := make([]uint, len(topArticles))
			for i, item := range topArticles {
				var articleID uint
				fmt.Sscanf(item.Member.(string), "%d", &articleID)
				articleIDs[i] = articleID
			}

			var articles []model.Article
			s.db.WithContext(ctx).
				Select("id, article_title").
				Where("id IN ? AND is_delete = 0", articleIDs).
				Find(&articles)

			// 构建 ID -> Title 映射
			titleMap := make(map[uint]string)
			for _, a := range articles {
				titleMap[a.ID] = a.ArticleTitle
			}

			// 只保留数据库中存在的文章（过滤已删除的）
			validRanks := make([]dto.ArticleRankDTO, 0, len(topArticles))
			for _, item := range topArticles {
				var articleID uint
				fmt.Sscanf(item.Member.(string), "%d", &articleID)
				
				if title, ok := titleMap[articleID]; ok && title != "" {
					validRanks = append(validRanks, dto.ArticleRankDTO{
						ArticleTitle: title,
						ViewsCount:   int(item.Score),
					})
				}
			}
			
			info.ArticleRankDTOs = validRanks
		} else {
			info.ArticleRankDTOs = []dto.ArticleRankDTO{}
		}
	} else {
		info.ArticleRankDTOs = []dto.ArticleRankDTO{}
	}

	return &info, nil
}

// getWebsiteConfig 获取网站配置（对标Java getWebsiteConfig）
func (s *AuroraInfoService) getWebsiteConfig(ctx context.Context) (*dto.WebsiteConfigDTO, error) {
	if s.websiteConfigService == nil {
		return nil, fmt.Errorf("网站配置服务未初始化")
	}
	return s.websiteConfigService.GetConfig(ctx)
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
	type TalkRow struct {
		ID         uint
		Nickname   string
		Avatar     string
		Content    string
		Images     string
		IsTop      int8
		CreateTime string `gorm:"column:create_time"`
	}

	var talks []TalkRow
	err := s.db.WithContext(ctx).
		Table("t_talk t").
		Select("t.id, ui.nickname, ui.avatar, t.content, t.images, t.is_top, t.create_time").
		Joins("JOIN t_user_info ui ON t.user_id = ui.id").
		Where("t.status = 1").
		Order("t.create_time DESC").
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
			Nickname:   t.Nickname,
			Avatar:     t.Avatar,
			Content:    content,
			Images:     t.Images,
			IsTop:      t.IsTop,
			CreateTime: s.parseTime(t.CreateTime),
		}
		// 解析images JSON
		if t.Images != "" {
			json.Unmarshal([]byte(t.Images), &list[i].Imgs)
		}
	}
	return list, err
}

// 辅助转换函数 (避免循环引用) - 对标Java版 ArticleCardDTO结构
func toSimpleArticleCard(a *model.Article) dto.ArticleCardDTO {
	card := dto.ArticleCardDTO{
		ID:             a.ID,
		ArticleCover:   a.ArticleCover,
		ArticleTitle:   a.ArticleTitle,
		ArticleContent: a.ArticleContent,
		IsTop:          a.IsTop,
		IsFeatured:     a.IsFeatured,
		Status:         a.Status,
		CreateTime:     a.CreateTime,
	}
	if a.Category != nil {
		card.CategoryName = a.Category.CategoryName
	}
	// 构建 Author 嵌套对象（对标Java UserInfo author）
	if a.UserInfo != nil {
		card.Author = &dto.UserInfoInCard{
			Nickname: a.UserInfo.Nickname,
			Website:  a.UserInfo.Website,
			Avatar:   a.UserInfo.Avatar,
		}
	}
	// 构建 Tags 数组
	if len(a.Tags) > 0 {
		card.Tags = make([]dto.TagInCard, len(a.Tags))
		for i, t := range a.Tags {
			card.Tags[i] = dto.TagInCard{
				TagName: t.TagName,
			}
		}
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

// parseTime 解析时间字符串（对标Java的LocalDateTime）
func (s *AuroraInfoService) parseTime(timeStr string) time.Time {
	t, err := time.Parse("2006-01-02 15:04:05", timeStr)
	if err != nil {
		return time.Time{}
	}
	return t
}
