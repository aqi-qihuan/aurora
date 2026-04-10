package service

import (
	"context"
	stderrors "errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/model"
	"github.com/aurora-go/aurora/internal/vo"
	"gorm.io/gorm"
)

// ArticleService 文章业务逻辑层 (对标 Java ArticleServiceImpl)
// 职责: 文章CRUD / 搜索 / 归档 / 导入导出 / 置顶推荐 / 密码保护 / 缓存策略
type ArticleService struct {
	db *gorm.DB
}

// NewArticleService 创建文章服务实例
func NewArticleService(db *gorm.DB) *ArticleService {
	return &ArticleService{db: db}
}

// CreateArticle 发布文章 (事务操作: 文章+标签关联)
func (s *ArticleService) CreateArticle(ctx context.Context, userID uint, vo vo.ArticleVO) (*model.Article, error) {
	var article model.Article

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		article = model.Article{
			UserID:         userID,
			CategoryID:     vo.CategoryID,
			ArticleTitle:   vo.ArticleTitle,
			ArticleContent: vo.ArticleContent,
			Type:           1, // 默认原创
			Status:         1, // 默认公开
		}
		
		if vo.Type != 0 {
			article.Type = vo.Type
		}
		if vo.Status != nil {
			article.Status = *vo.Status
		}
		if vo.IsTop != 0 {
			article.IsTop = vo.IsTop
		}
		if vo.IsFeatured != 0 {
			article.IsFeatured = vo.IsFeatured
		}
		if vo.Password != "" {
			article.Status = 2 // 密码保护
			article.Password = vo.Password
		}
		if vo.OriginalURL != "" {
			article.OriginalURL = vo.OriginalURL
		}

		if err := tx.Create(&article).Error; err != nil {
			return fmt.Errorf("创建文章失败: %w", err)
		}

		// 处理标签关联 (多对多)
		if len(vo.TagIDs) > 0 {
			var tags []model.Tag
			if err := tx.Find(&tags, vo.TagIDs).Error; err != nil {
				return fmt.Errorf("查询标签失败: %w", err)
			}
			if err := tx.Model(&article).Association("Tags").Replace(tags); err != nil {
				return fmt.Errorf("关联标签失败: %w", err)
			}
		}

		// 更新分类文章计数
		tx.Model(&model.Category{}).Where("id = ?", vo.CategoryID).
			UpdateColumn("article_count", gorm.Expr("article_count + 1"))

		slog.Info("文章发布成功",
			"article_id", article.ID,
			"user_id", userID,
			"title", article.ArticleTitle,
		)
		return nil
	})

	return &article, err
}

// GetArticleByID 根据ID获取文章详情 (含分类+作者+标签)
func (s *ArticleService) GetArticleByID(ctx context.Context, id uint) (*dto.ArticleDTO, error) {
	var article model.Article
	
	query := s.db.WithContext(ctx).
		Preload("Tags").
		Preload("Category").
		Preload("UserInfo")

	if err := query.First(&article, id).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrArticleNotFound
		}
		return nil, fmt.Errorf("查询文章失败: %w", err)
	}

	return s.toArticleDTO(&article), nil
}

// ListArticles 分页查询文章列表 (前台用, 只查已发布非删除)
func (s *ArticleService) ListArticles(ctx context.Context, cond dto.ConditionVO, page dto.PageVO) (*dto.PageResultDTO, error) {
	var articles []model.Article
	var count int64

	baseQuery := s.db.WithContext(ctx).
		Model(&model.Article{}).
		Where("is_delete = 0 AND status = 1") // 前台只展示公开文章

	// 动态条件构建
	if cond.Keywords != "" {
		baseQuery = baseQuery.Where("article_title LIKE ?", "%"+cond.Keywords+"%")
	}
	if cond.CategoryID != nil && *cond.CategoryID > 0 {
		baseQuery = baseQuery.Where("category_id = ?", *cond.CategoryID)
	}
	if cond.Type != nil && *cond.Type > 0 {
		baseQuery = baseQuery.Where("type = ?", *cond.Type)
	}

	// 统计总数
	if err := baseQuery.Count(&count).Error; err != nil {
		return nil, fmt.Errorf("统计文章数失败: %w", err)
	}

	// 排序: 默认按置顶+创建时间
	orderClause := "is_top DESC, create_time DESC"
	if cond.Sort != "" {
		dir := "DESC"
		if strings.EqualFold(cond.Order, "asc") {
			dir = "ASC"
		}
		orderClause = cond.Sort + " " + dir
	}

	// 分页查询 (含预加载)
	offset := page.GetOffset()
	if err := baseQuery.
		Preload("Category").
		Preload("UserInfo").
		Preload("Tags").
		Order(orderClause).
		Limit(page.PageSize).
		Offset(offset).
		Find(&articles).Error; err != nil {
		return nil, fmt.Errorf("查询文章列表失败: %w", err)
	}

	list := make([]dto.ArticleCardDTO, len(articles))
	for i, a := range articles {
		list[i] = s.toArticleCardDTO(&a)
	}

	return &dto.PageResultDTO{
		List:     list,
		Count:    count,
		PageNum:  page.PageNum,
		PageSize: page.PageSize,
	}, nil
}

// ListAdminArticles 后台管理分页查询 (包含草稿/回收站)
func (s *ArticleService) ListAdminArticles(ctx context.Context, cond dto.ConditionVO, page dto.PageVO) (*dto.PageResultDTO, error) {
	var articles []model.Article
	var count int64

	baseQuery := s.db.WithContext(ctx).Model(&model.Article{})

	if cond.Keywords != "" {
		baseQuery = baseQuery.Where("article_title LIKE ?", "%"+cond.Keywords+"%")
	}
	if cond.Status != nil {
		baseQuery = baseQuery.Where("status = ?", *cond.Status)
	}
	if cond.CategoryID != nil && *cond.CategoryID > 0 {
		baseQuery = baseQuery.Where("category_id = ?", *cond.CategoryID)
	}
	if cond.IsDelete != nil {
		baseQuery = baseQuery.Where("is_delete = ?", *cond.IsDelete)
	} else {
		baseQuery = baseQuery.Where("is_delete = 0") // 默认排除已删除
	}

	if err := baseQuery.Count(&count).Error; err != nil {
		return nil, fmt.Errorf("统计文章数失败: %w", err)
	}

	offset := page.GetOffset()
	if err := baseQuery.
		Preload("Category").
		Order("is_top DESC, create_time DESC").
		Limit(page.PageSize).
		Offset(offset).
		Find(&articles).Error; err != nil {
		return nil, fmt.Errorf("查询文章列表失败: %w", err)
	}

	list := make([]dto.ArticleCardDTO, len(articles))
	for i, a := range articles {
		list[i] = s.toArticleCardDTO(&a)
	}

	return &dto.PageResultDTO{
		List:     list,
		Count:    count,
		PageNum:  page.PageNum,
		PageSize: page.PageSize,
	}, nil
}

// UpdateArticle 更新文章 (事务: 更新内容+替换标签+更新计数)
func (s *ArticleService) UpdateArticle(ctx context.Context, vo vo.ArticleVO) (*model.Article, error) {
	var article model.Article

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 查询原文章
		if err := tx.First(&article, vo.ID).Error; err != nil {
			return errors.ErrArticleNotFound
		}

		// 更新字段
		oldCategoryID := article.CategoryID
		article.ArticleTitle = vo.ArticleTitle
		article.ArticleContent = vo.ArticleContent
		article.CategoryID = vo.CategoryID

		if vo.Type != 0 {
			article.Type = vo.Type
		}
		if vo.Status != nil {
			article.Status = *vo.Status
			if *vo.Status == 2 && vo.Password != "" {
				article.Password = vo.Password
			}
		}
		if vo.IsTop != 0 {
			article.IsTop = vo.IsTop
		}
		if vo.IsFeatured != 0 {
			article.IsFeatured = vo.IsFeatured
		}
		if vo.OriginalURL != "" {
			article.OriginalURL = vo.OriginalURL
		}

		if err := tx.Save(&article).Error; err != nil {
			return fmt.Errorf("更新文章失败: %w", err)
		}

		// 替换标签关联
		if vo.TagIDs != nil {
			if len(vo.TagIDs) > 0 {
				var tags []model.Tag
				tx.Find(&tags, vo.TagIDs)
				tx.Model(&article).Association("Tags").Replace(tags)
			} else {
				tx.Model(&article).Association("Tags").Clear()
			}
		}

		// 分类变更时更新计数
		if oldCategoryID != vo.CategoryID {
			tx.Model(&model.Category{}).Where("id = ?", oldCategoryID).
				UpdateColumn("article_count", gorm.Expr("GREATEST(article_count - 1, 0)"))
			tx.Model(&model.Category{}).Where("id = ?", vo.CategoryID).
				UpdateColumn("article_count", gorm.Expr("article_count + 1"))
		}

		slog.Info("文章更新成功", "article_id", article.ID, "title", article.ArticleTitle)
		return nil
	})

	return &article, err
}

// DeleteArticle 删除文章 (软删除, 事务处理)
func (s *ArticleService) DeleteArticle(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var article model.Article
		if err := tx.First(&article, id).Error; err != nil {
			return errors.ErrArticleNotFound
		}

		// 软删除
		article.IsDelete = 1
		if err := tx.Save(&article).Error; err != nil {
			return fmt.Errorf("删除文章失败: %w", err)
		}

		// 清除标签关联
		tx.Model(&article).Association("Tags").Clear()

		// 减少分类计数
		tx.Model(&model.Category{}).Where("id = ?", article.CategoryID).
			UpdateColumn("article_count", gorm.Expr("GREATEST(article_count - 1, 0)"))

		slog.Info("文章已软删除", "article_id", id)
		return nil
	})
}

// BatchDeleteArticles 批量软删除文章
func (s *ArticleService) BatchDeleteArticles(ctx context.Context, ids []uint) error {
	result := s.db.WithContext(ctx).
		Model(&model.Article{}).
		Where("id IN ? AND is_delete = 0", ids).
		Update("is_delete", 1)

	if result.Error != nil {
		return fmt.Errorf("批量删除文章失败: %w", result.Error)
	}
	
	slog.Info("批量删除文章完成", "count", result.RowsAffected, "ids", ids)
	return nil
}

// UpdateTopFeatured 设置置顶/推荐
func (s *ArticleService) UpdateTopFeatured(ctx context.Context, vo vo.ArticleTopFeaturedVO) error {
	result := s.db.WithContext(ctx).
		Model(&model.Article{}).
		Where("id = ?", vo.ID).
		Updates(map[string]interface{}{
			"is_top":      vo.IsTop,
			"is_featured": vo.IsFeatured,
		})

	if result.Error != nil {
		return fmt.Errorf("设置置顶推荐失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.ErrArticleNotFound
	}

	slog.Info("文章置顶推荐更新", "article_id", vo.ID,
		"is_top", vo.IsTop, "is_featured", vo.IsFeatured)
	return nil
}

// VerifyPassword 验证密码文章访问权限
func (s *ArticleService) VerifyPassword(ctx context.Context, id uint, password string) bool {
	var article model.Article
	err := s.db.WithContext(ctx).
		Select("id, status, password").
		First(&article, id).Error

	if err != nil || article.Status != 2 { // 非密码保护文章直接返回false
		return false
	}
	return article.Password == password
}

// GetArchives 获取文章归档 (按年月分组)
type ArchiveDTO struct {
	YearMonth string            `json:"yearMonth"` // "2024-01"
	Count     int64             `json:"count"`
	Articles  []dto.ArticleCardDTO `json:"articles"`
}

func (s *ArticleService) GetArchives(ctx context.Context) ([]ArchiveDTO, error) {
	type Row struct {
		YearMonth time.Time
		Count     int64
	}

	var rows []Row
	err := s.db.WithContext(ctx).
		Model(&model.Article{}).
		Select("DATE_FORMAT(create_time, '%Y-%m-01') as year_month, COUNT(*) as count").
		Where("is_delete = 0 AND status = 1").
		Group("year_month").
		Order("year_month DESC").
		Find(&rows).Error

	if err != nil {
		return nil, fmt.Errorf("查询归档失败: %w", err)
	}

	archives := make([]ArchiveDTO, len(rows))
	for i, row := range rows {
		archives[i] = ArchiveDTO{
			YearMonth: row.YearMonth.Format("2006-01"),
			Count:     row.Count,
		}
	}

	return archives, nil
}

// SearchArticles 搜索文章 (关键词搜索, 对标 EsSearchStrategyImpl)
func (s *ArticleService) SearchArticles(ctx context.Context, keywords string, page dto.PageVO) (*dto.PageResultDTO, error) {
	var articles []model.Article
	var count int64

	searchPattern := "%" + keywords + "%"
	
	baseQuery := s.db.WithContext(ctx).
		Model(&model.Article{}).
		Where("is_delete = 0 AND status = 1").
		Where("article_title LIKE ? OR article_content LIKE ?", searchPattern, searchPattern)

	if err := baseQuery.Count(&count).Error; err != nil {
		return nil, fmt.Errorf("统计搜索结果失败: %w", err)
	}

	offset := page.GetOffset()
	if err := baseQuery.
		Preload("Category").
		Preload("UserInfo").
		Order("create_time DESC").
		Limit(page.PageSize).
		Offset(offset).
		Find(&articles).Error; err != nil {
		return nil, fmt.Errorf("搜索文章失败: %w", err)
	}

	list := make([]dto.ArticleCardDTO, len(articles))
	for i, a := range articles {
		list[i] = s.toArticleCardDTO(&a)
	}

	return &dto.PageResultDTO{
		List:     list,
		Count:    count,
		PageNum:  page.PageNum,
		PageSize: page.PageSize,
	}, nil
}

// GetTopArticles 获取置顶/推荐文章
func (s *ArticleService) GetTopArticles(ctx context.Context, limit int) ([]dto.ArticleCardDTO, error) {
	var articles []model.Article

	if err := s.db.WithContext(ctx).
		Where("is_delete = 0 AND status = 1 AND (is_top = 1 OR is_featured = 1)").
		Preload("Category").
		Preload("UserInfo").
		Order("is_top DESC, is_featured DESC, create_time DESC").
		Limit(limit).
		Find(&articles).Error; err != nil {
		return nil, fmt.Errorf("查询置顶推荐文章失败: %w", err)
	}

	list := make([]dto.ArticleCardDTO, len(articles))
	for i, a := range articles {
		list[i] = s.toArticleCardDTO(&a)
	}
	return list, nil
}

// IncrementViewCount 增加浏览量 (Redis缓存 + 异步刷写DB)
func (s *ArticleService) IncrementViewCount(ctx context.Context, articleID uint) {
	// TODO: P0-7 实现Redis ZSet浏览量排行后, 改为 Redis.Incr + 定时批量写入DB
	s.db.WithContext(ctx).
		Model(&model.Article{}).
		Where("id = ?", articleID).
		UpdateColumn("view_count", gorm.Expr("view_count + 1"))
}

// ===== DTO转换方法 =====

func (s *ArticleService) toArticleDTO(a *model.Article) *dto.ArticleDTO {
	tags := make([]dto.TagDTO, len(a.Tags))
	for i, t := range a.Tags {
		tags[i] = dto.TagDTO{ID: t.ID, TagName: t.TagName}
	}

	dto := &dto.ArticleDTO{
		ID:             a.ID,
		UserID:         a.UserID,
		ArticleCover:   a.ArticleCover,
		ArticleTitle:   a.ArticleTitle,
		ArticleContent: a.ArticleContent,
		IsTop:          a.IsTop,
		IsFeatured:     a.IsFeatured,
		Status:         a.Status,
		Type:           a.Type,
		ViewCount:      a.ViewCount,
		LikeCount:      a.LikeCount,
		CategoryID:     a.CategoryID,
		Tags:           tags,
		CreateTime:     a.CreateTime,
	}

	if a.Category != nil {
		dto.CategoryName = a.Category.CategoryName
	}
	if a.UserInfo != nil {
		dto.Nickname = a.UserInfo.Nickname
		dto.Avatar = a.UserInfo.Avatar
	}
	return dto
}

func (s *ArticleService) toArticleCardDTO(a *model.Article) dto.ArticleCardDTO {
	tags := make([]dto.TagDTO, len(a.Tags))
	for i, t := range a.Tags {
		tags[i] = dto.TagDTO{ID: t.ID, TagName: t.TagName}
	}

	card := dto.ArticleCardDTO{
		ID:           a.ID,
		ArticleTitle: a.ArticleTitle,
		ArticleCover: a.ArticleCover,
		IsTop:        a.IsTop,
		IsFeatured:   a.IsFeatured,
		Status:       a.Status,
		ViewCount:    a.ViewCount,
		Tags:         tags,
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

// ImportArticles 导入Markdown文章 (对标Java版导入功能)
type ImportError struct {
	Filename string
	Error    error
}

func (s *ArticleService) ImportArticles(ctx context.Context, userID uint, contents map[string]string) (success int, failures []ImportError) {
	for filename, content := range contents {
		// 简单解析: 第一行作为标题(去掉#), 其余作为内容
		lines := strings.SplitN(content, "\n", 2)
		title := filename
		body := content
		
		if len(lines) >= 2 {
			title = strings.TrimPrefix(lines[0], "# ")
			title = strings.TrimSpace(title)
			body = lines[1]
		}

		vo := vo.ArticleVO{
			ArticleTitle:   title,
			ArticleContent: body,
			CategoryID:     1, // 默认分类
			Status:         int8Ptr(0), // 草稿状态
		}

		if _, err := s.CreateArticle(ctx, userID, vo); err != nil {
			failures = append(failures, ImportError{Filename: filename, Error: err})
		} else {
			success++
		}
	}
	return
}

func int8Ptr(v int8) *int8 { return &v }
