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
	db           *gorm.DB
	statsService *RedisStatsService // Redis 统计服务
}

// NewArticleService 创建文章服务实例
func NewArticleService(db *gorm.DB, statsService *RedisStatsService) *ArticleService {
	return &ArticleService{
		db:           db,
		statsService: statsService,
	}
}

// CreateArticle 发布文章 (事务操作: 文章+分类+标签关联, 对齐Java版实现)
func (s *ArticleService) CreateArticle(ctx context.Context, userID uint, vo vo.ArticleVO) (*model.Article, error) {
	var article model.Article

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 处理分类 (根据名称查找或创建, 对齐Java saveArticleCategory)
		var categoryID *uint
		if vo.CategoryName != "" {
			var category model.Category
			if err := tx.Where("category_name = ?", vo.CategoryName).First(&category).Error; err != nil {
				// 分类不存在则创建
				if stderrors.Is(err, gorm.ErrRecordNotFound) {
					category = model.Category{CategoryName: vo.CategoryName}
					if err := tx.Create(&category).Error; err != nil {
						return fmt.Errorf("创建分类失败: %w", err)
					}
				} else {
					return fmt.Errorf("查询分类失败: %w", err)
				}
			}
			categoryID = &category.ID
		}

		article = model.Article{
			UserID:         userID,
			CategoryID:     categoryID,
			ArticleTitle:   vo.ArticleTitle,
			ArticleContent: vo.ArticleContent,
			ArticleCover:   vo.ArticleCover,
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

		// 2. 处理标签关联 (根据名称查找或创建, 对齐Java saveArticleTag)
		if len(vo.TagNames) > 0 {
			// 查询已存在的标签
			var existTags []model.Tag
			if err := tx.Where("tag_name IN ?", vo.TagNames).Find(&existTags).Error; err != nil {
				return fmt.Errorf("查询标签失败: %w", err)
			}

			// 构建已存在标签名称集合
			existTagNames := make(map[string]bool)
			var tagIDs []uint
			for _, t := range existTags {
				existTagNames[t.TagName] = true
				tagIDs = append(tagIDs, t.ID)
			}

			// 创建不存在的标签
			for _, tagName := range vo.TagNames {
				if !existTagNames[tagName] {
					newTag := model.Tag{TagName: tagName}
					if err := tx.Create(&newTag).Error; err != nil {
						return fmt.Errorf("创建标签失败: %w", err)
					}
					tagIDs = append(tagIDs, newTag.ID)
				}
			}

			// 关联标签到文章
			if len(tagIDs) > 0 {
				var tags []model.Tag
				if err := tx.Find(&tags, tagIDs).Error; err != nil {
					return fmt.Errorf("查询标签失败: %w", err)
				}
				if err := tx.Model(&article).Association("Tags").Replace(tags); err != nil {
					return fmt.Errorf("关联标签失败: %w", err)
				}
			}
		}

		// 3. 更新分类文章计数（使用 Redis）
		if categoryID != nil && s.statsService != nil {
			s.statsService.IncrementCategoryArticleCount(ctx, *categoryID)
		}

		slog.Info("文章发布成功",
			"article_id", article.ID,
			"user_id", userID,
			"title", article.ArticleTitle,
			"category", vo.CategoryName,
			"tags", vo.TagNames,
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
		Preload("Tags"). // 新增：预加载标签
		Order("is_top DESC, create_time DESC").
		Limit(page.PageSize).
		Offset(offset).
		Find(&articles).Error; err != nil {
		return nil, fmt.Errorf("查询文章列表失败: %w", err)
	}

	slog.Debug("后台文章列表查询完成",
		"total_articles", len(articles),
		"page", page.PageNum,
	)

	list := make([]dto.ArticleCardDTO, len(articles))
	for i, a := range articles {
		slog.Debug("转换文章卡片",
			"article_id", a.ID,
			"title", a.ArticleTitle,
			"tags_loaded", len(a.Tags),
			"category_loaded", a.Category != nil,
		)
		list[i] = s.toArticleCardDTO(&a)
	}

	return &dto.PageResultDTO{
		List:     list,
		Count:    count,
		PageNum:  page.PageNum,
		PageSize: page.PageSize,
	}, nil
}

// UpdateArticle 更新文章 (事务: 更新内容+分类+标签+更新计数, 对齐Java版实现)
func (s *ArticleService) UpdateArticle(ctx context.Context, vo vo.ArticleVO) (*model.Article, error) {
	var article model.Article

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 查询原文章
		if err := tx.First(&article, vo.ID).Error; err != nil {
			return errors.ErrArticleNotFound
		}

		// 1. 处理分类 (根据名称查找或创建)
		oldCategoryID := article.CategoryID
		var newCategoryID *uint
		if vo.CategoryName != "" {
			var category model.Category
			if err := tx.Where("category_name = ?", vo.CategoryName).First(&category).Error; err != nil {
				if stderrors.Is(err, gorm.ErrRecordNotFound) {
					category = model.Category{CategoryName: vo.CategoryName}
					if err := tx.Create(&category).Error; err != nil {
						return fmt.Errorf("创建分类失败: %w", err)
					}
				} else {
					return fmt.Errorf("查询分类失败: %w", err)
				}
			}
			newCategoryID = &category.ID
		}

		// 更新文章字段
		article.ArticleTitle = vo.ArticleTitle
		article.ArticleContent = vo.ArticleContent
		article.ArticleCover = vo.ArticleCover
		article.CategoryID = newCategoryID

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

		// 2. 处理标签关联 (根据名称查找或创建)
		if vo.TagNames != nil {
			if len(vo.TagNames) > 0 {
				// 查询已存在的标签
				var existTags []model.Tag
				if err := tx.Where("tag_name IN ?", vo.TagNames).Find(&existTags).Error; err != nil {
					return fmt.Errorf("查询标签失败: %w", err)
				}

				existTagNames := make(map[string]bool)
				var tagIDs []uint
				for _, t := range existTags {
					existTagNames[t.TagName] = true
					tagIDs = append(tagIDs, t.ID)
				}

				// 创建不存在的标签
				for _, tagName := range vo.TagNames {
					if !existTagNames[tagName] {
						newTag := model.Tag{TagName: tagName}
						if err := tx.Create(&newTag).Error; err != nil {
							return fmt.Errorf("创建标签失败: %w", err)
						}
						tagIDs = append(tagIDs, newTag.ID)
					}
				}

				// 关联标签
				if len(tagIDs) > 0 {
					var tags []model.Tag
					if err := tx.Find(&tags, tagIDs).Error; err != nil {
						return fmt.Errorf("查询标签失败: %w", err)
					}
					tx.Model(&article).Association("Tags").Replace(tags)
				}
			} else {
				tx.Model(&article).Association("Tags").Clear()
			}
		}

		// 3. 分类变更时更新计数（使用 Redis）
		if oldCategoryID != nil && newCategoryID != nil && *oldCategoryID != *newCategoryID {
			if s.statsService != nil {
				s.statsService.DecrementCategoryArticleCount(ctx, *oldCategoryID)
				s.statsService.IncrementCategoryArticleCount(ctx, *newCategoryID)
			}
		} else if oldCategoryID == nil && newCategoryID != nil {
			if s.statsService != nil {
				s.statsService.IncrementCategoryArticleCount(ctx, *newCategoryID)
			}
		} else if oldCategoryID != nil && newCategoryID == nil {
			if s.statsService != nil {
				s.statsService.DecrementCategoryArticleCount(ctx, *oldCategoryID)
			}
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

		// 减少分类计数（使用 Redis）
		if article.CategoryID != nil && s.statsService != nil {
			s.statsService.DecrementCategoryArticleCount(ctx, *article.CategoryID)
		}

		slog.Info("文章已软删除", "article_id", id)
		return nil
	})
}

// BatchDeleteArticles 批量软删除文章（移入回收站）
func (s *ArticleService) BatchDeleteArticles(ctx context.Context, ids []uint) error {
	result := s.db.WithContext(ctx).
		Model(&model.Article{}).
		Where("id IN ? AND is_delete = 0", ids).
		Update("is_delete", 1)

	if result.Error != nil {
		return fmt.Errorf("批量删除文章失败: %w", result.Error)
	}

	slog.Info("批量软删除文章完成", "count", result.RowsAffected, "ids", ids)
	return nil
}

// DeleteArticlesPermanently 彻底删除文章（物理删除 - 回收站使用）
// 对标Java: articleMapper.deleteBatchIds(articleIds)
// 注意: 需要先删除关联的t_article_tag记录
func (s *ArticleService) DeleteArticlesPermanently(ctx context.Context, ids []uint) error {
	if len(ids) == 0 {
		return errors.ErrInvalidParams.WithMsg("请选择要删除的文章")
	}

	// 开启事务（对标Java @Transactional）
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 先删除标签关联表 t_article_tag（对标Java articleTagMapper.delete）
		if err := tx.Exec("DELETE FROM t_article_tag WHERE article_id IN ?", ids).Error; err != nil {
			return fmt.Errorf("删除标签关联失败: %w", err)
		}

		// 2. 物理删除文章（对标Java articleMapper.deleteBatchIds）
		result := tx.Unscoped().Where("id IN ?", ids).Delete(&model.Article{})
		if result.Error != nil {
			return fmt.Errorf("物理删除文章失败: %w", result.Error)
		}

		slog.Info("文章彻底删除成功", "count", result.RowsAffected, "ids", ids)
		return nil
	})
}

// UpdateArticleDelete 批量逻辑删除/恢复文章 (对标Java版 updateArticleDelete)
// 对标Java: List<Article> articles = deleteVO.getIds().stream()
//                 .map(id -> Article.builder().id(id).isDelete(deleteVO.getIsDelete()).build())
//                 .collect(Collectors.toList());
//            this.updateBatchById(articles);
func (s *ArticleService) UpdateArticleDelete(ctx context.Context, ids []uint, isDelete int8) error {
	if len(ids) == 0 {
		return errors.ErrInvalidParams.WithMsg("请选择要操作的文章")
	}

	// 批量更新is_delete字段 (对标Java updateBatchById)
	result := s.db.WithContext(ctx).
		Model(&model.Article{}).
		Where("id IN ?", ids).
		Update("is_delete", isDelete)

	if result.Error != nil {
		return fmt.Errorf("更新文章状态失败: %w", result.Error)
	}

	action := "恢复"
	if isDelete == 1 {
		action = "删除"
	}
	slog.Info("文章批量"+action+"成功", "count", result.RowsAffected, "is_delete", isDelete, "ids", ids)
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
	YearMonth string               `json:"yearMonth"` // "2024-01"
	Count     int64                `json:"count"`
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

// IncrementViewCount 增加浏览量 (使用 Redis)
func (s *ArticleService) IncrementViewCount(ctx context.Context, articleID uint) {
	if s.statsService != nil {
		// 异步执行，不阻塞主流程
		go func() {
			if err := s.statsService.IncrementArticleView(context.Background(), articleID); err != nil {
				slog.Error("增加文章浏览量失败", "article_id", articleID, "error", err)
			}
		}()
	}
}

// ===== DTO转换方法 =====

func (s *ArticleService) toArticleDTO(a *model.Article) *dto.ArticleDTO {
	tags := make([]dto.TagDTO, len(a.Tags))
	for i, t := range a.Tags {
		tags[i] = dto.TagDTO{ID: t.ID, TagName: t.TagName}
	}

	// 从 Redis 获取统计信息
	var viewCount uint64
	var likeCount int64
	if s.statsService != nil {
		ctx := context.Background()
		viewCount, _ = s.statsService.GetArticleView(ctx, a.ID)
		likeCount, _ = s.statsService.GetArticleLikeCount(ctx, a.ID)
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
		ViewCount:      viewCount,
		LikeCount:      likeCount,
		CategoryID:     getCategoryID(a.CategoryID),
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
		tags[i] = dto.TagDTO{
			ID:         t.ID,
			TagName:    t.TagName,
			CreateTime: t.CreateTime,
		}
	}

	// 从 Redis 获取浏览量
	var viewCount uint64
	if s.statsService != nil {
		ctx := context.Background()
		var err error
		viewCount, err = s.statsService.GetArticleView(ctx, a.ID)
		if err != nil {
			slog.Warn("获取文章浏览量失败", "article_id", a.ID, "error", err)
		}
		slog.Info("文章浏览量查询结果", "article_id", a.ID, "view_count", viewCount, "title", a.ArticleTitle)
	} else {
		slog.Warn("statsService 为 nil，无法获取浏览量", "article_id", a.ID)
	}

	card := dto.ArticleCardDTO{
		ID:           a.ID,
		ArticleTitle: a.ArticleTitle,
		ArticleCover: a.ArticleCover,
		IsTop:        a.IsTop,
		IsFeatured:   a.IsFeatured,
		IsDelete:     a.IsDelete,
		Status:       a.Status,
		Type:         a.Type,
		ViewCount:    viewCount,
		TagDTOs:      tags,
		CreateTime:   a.CreateTime,
	}

	if a.Category != nil {
		card.CategoryName = a.Category.CategoryName
	}
	if a.UserInfo != nil {
		card.Nickname = a.UserInfo.Nickname
	}

	slog.Debug("文章卡片转换",
		"article_id", a.ID,
		"tags_count", len(tags),
		"view_count", viewCount,
		"category_name", card.CategoryName,
	)

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
			CategoryName:   "默认分类",
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

// getCategoryID 安全获取分类ID（处理指针类型）
func getCategoryID(categoryID *uint) uint {
	if categoryID == nil {
		return 0
	}
	return *categoryID
}
