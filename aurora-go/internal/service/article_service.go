package service

import (
	"context"
	stderrors "errors"
	"fmt"
	"log/slog"
	"sort"
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
	fileService  *FileService       // 文件上传服务（用于文章导出）
	searchCtx    interface{}        // 搜索策略上下文 (*strategy.SearchContext)
}

// NewArticleService 创建文章服务实例
func NewArticleService(db *gorm.DB, statsService *RedisStatsService, fileService *FileService) *ArticleService {
	return &ArticleService{
		db:           db,
		statsService: statsService,
		fileService:  fileService,
	}
}

// SetSearchContext 设置搜索策略上下文（延迟注入，避免循环依赖）
func (s *ArticleService) SetSearchContext(searchCtx interface{}) {
	s.searchCtx = searchCtx
}

// CreateArticle 发布文章 (事务操作: 文章+分类+标签关联, 对齐Java版实现)
// 返回完整的 Article 对象，含预加载的关联数据 (对标Java BeanCopyUtil返回ArticleAdminViewDTO)
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
			ArticleAbstract: vo.ArticleAbstract, // 文章摘要
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

		// 4. 重新加载关联数据 (Tags, Category, UserInfo) - 对标Java getArticleByIdAdmin
		if err := tx.Preload("Tags").Preload("Category").Preload("UserInfo").First(&article, article.ID).Error; err != nil {
			return fmt.Errorf("重新加载文章关联数据失败: %w", err)
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

	// 使用 ArticleAdminDTO（对标Java ArticleAdminDTO）
	list := make([]dto.ArticleAdminDTO, len(articles))
	for i, a := range articles {
		list[i] = s.toArticleAdminDTO(&a)
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

		// 4. 重新加载关联数据 (Tags, Category, UserInfo) - 对标Java getArticleByIdAdmin
		if err := tx.Preload("Tags").Preload("Category").Preload("UserInfo").First(&article, article.ID).Error; err != nil {
			return fmt.Errorf("重新加载文章关联数据失败: %w", err)
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

// ExportArticles 批量导出文章为 Markdown 文件（对标Java exportArticles）
// 逻辑: 查询文章 → 生成 Markdown 内容 → 上传到 MinIO → 返回 URL 列表
func (s *ArticleService) ExportArticles(ctx context.Context, ids []uint) ([]string, error) {
	if len(ids) == 0 {
		return nil, errors.ErrInvalidParams.WithMsg("请选择要导出的文章")
	}

	// 查询文章标题和内容（对标Java select(Article::getArticleTitle, Article::getArticleContent)）
	var articles []model.Article
	if err := s.db.WithContext(ctx).
		Select("id, article_title, article_content").
		Where("id IN ?", ids).
		Find(&articles).Error; err != nil {
		return nil, fmt.Errorf("查询文章失败: %w", err)
	}

	if len(articles) == 0 {
		return nil, errors.ErrArticleNotFound
	}

	// 导出每篇文章为 Markdown 文件并上传
	var urls []string
	for _, article := range articles {
		// 构建 Markdown 内容（标题 + 正文）
		mdContent := fmt.Sprintf("# %s\n\n%s", article.ArticleTitle, article.ArticleContent)

		// 上传到 MinIO（对标Java uploadStrategyContext.executeUploadStrategy）
		objectName := fmt.Sprintf("aurora/export/%s.md", article.ArticleTitle)

		// 使用 FileService 的上传逻辑
		if s.fileService != nil {
			url, err := s.fileService.UploadMarkdownContent(ctx, objectName, []byte(mdContent))
			if err != nil {
				slog.Error("导出文章失败", "article_id", article.ID, "title", article.ArticleTitle, "error", err)
				continue
			}
			urls = append(urls, url)
		} else {
			slog.Warn("FileService 未注入，跳过导出", "article_id", article.ID)
		}
	}

	slog.Info("文章批量导出完成", "total", len(ids), "success", len(urls))
	return urls, nil
}
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

// UpdateTopFeatured 设置置顶/推荐（对标Java ArticleTopFeaturedVO）
func (s *ArticleService) UpdateTopFeatured(ctx context.Context, vo vo.ArticleTopFeaturedVO) error {
	// 解包指针类型
	isTop := int8(0)
	isFeatured := int8(0)
	if vo.IsTop != nil {
		isTop = *vo.IsTop
	}
	if vo.IsFeatured != nil {
		isFeatured = *vo.IsFeatured
	}

	result := s.db.WithContext(ctx).
		Model(&model.Article{}).
		Where("id = ?", vo.ID).
		Updates(map[string]interface{}{
			"is_top":      isTop,
			"is_featured": isFeatured,
		})

	if result.Error != nil {
		return fmt.Errorf("设置置顶推荐失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.ErrArticleNotFound
	}

	slog.Info("文章置顶推荐更新", "article_id", vo.ID,
		"is_top", isTop, "is_featured", isFeatured)
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

// GetArchives 获取文章归档 (按年月分组, 对标Java listArchives)
// 返回 PageResultDTO<ArchiveDTO> 结构
func (s *ArticleService) GetArchives(ctx context.Context, current, size int) (*dto.PageResultDTO, error) {
	type Row struct {
		YearMonth time.Time
		Count     int64
	}

	// 统计总数
	var totalCount int64
	if err := s.db.WithContext(ctx).
		Model(&model.Article{}).
		Where("is_delete = 0 AND status = 1").
		Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("统计文章数失败: %w", err)
	}

	// 分页查询文章
	var articles []model.Article
	offset := (current - 1) * size
	if err := s.db.WithContext(ctx).
		Where("is_delete = 0 AND status = 1").
		Preload("Category").
		Preload("UserInfo").
		Preload("Tags").
		Order("create_time DESC").
		Limit(size).
		Offset(offset).
		Find(&articles).Error; err != nil {
		return nil, fmt.Errorf("查询文章列表失败: %w", err)
	}

	// 确保返回空数组而非 null
	if articles == nil {
		articles = []model.Article{}
	}

	// 转换为 ArticleCardDTO
	articleCards := make([]dto.ArticleCardDTO, len(articles))
	for i, a := range articles {
		articleCards[i] = s.toArticleCardDTO(&a)
	}

	// 按年月分组
	type archiveGroup struct {
		Time     string
		Articles []dto.ArticleCardDTO
	}
	groupMap := make(map[string]*archiveGroup)
	var timeOrder []string // 保持插入顺序

	for _, article := range articleCards {
		createTime := article.CreateTime
		key := fmt.Sprintf("%d-%d", createTime.Year(), createTime.Month())

		if _, exists := groupMap[key]; !exists {
			groupMap[key] = &archiveGroup{
				Time:     key,
				Articles: []dto.ArticleCardDTO{},
			}
			timeOrder = append(timeOrder, key)
		}
		groupMap[key].Articles = append(groupMap[key].Articles, article)
	}

	// 按时间倒序排序
	sort.Slice(timeOrder, func(i, j int) bool {
		return timeOrder[i] > timeOrder[j]
	})

	// 构建 ArchiveDTO 列表
	archives := make([]map[string]interface{}, len(timeOrder))
	for i, key := range timeOrder {
		archives[i] = map[string]interface{}{
			"time":     key,
			"articles": groupMap[key].Articles,
		}
	}

	return &dto.PageResultDTO{
		List:     archives,
		Count:    totalCount,
		PageNum:  current,
		PageSize: size,
	}, nil
}

// SearchArticles 搜索文章（对标Java: listArticlesBySearch）
// 优先使用 ES 全文搜索，ES 不可用时降级到 MySQL LIKE 查询
// Java版本返回: List<ArticleSearchDTO> 扁平数组，不分页，不查数据库
func (s *ArticleService) SearchArticles(ctx context.Context, keywords string) ([]dto.ArticleSearchDTO, error) {
	// 尝试使用 ES 搜索策略（如果已注入）
	if s.searchCtx != nil {
		type SearchContextInterface interface {
			ExecuteSearch(ctx context.Context, keywords string, current, size int) ([]map[string]interface{}, int, error)
		}
		if searchCtx, ok := s.searchCtx.(SearchContextInterface); ok {
			// 对标Java: ES查询不指定size，使用默认10条
			esResults, total, err := searchCtx.ExecuteSearch(ctx, keywords, 1, 10)
			if err != nil {
				slog.Warn("ES search failed, falling back to MySQL", "keywords", keywords, "error", err)
			} else if len(esResults) > 0 {
				slog.Info("ES search success", "keywords", keywords, "total", total, "returned", len(esResults))
				return s.convertESResultsToSearchDTO(esResults), nil
			} else {
				slog.Debug("ES search returned no results, falling back to MySQL", "keywords", keywords)
			}
		}
	}

	// 降级方案: MySQL LIKE 查询（对标Java: 不分页，最多10条）
	slog.Info("Using MySQL LIKE search", "keywords", keywords)
	return s.searchArticlesByMySQLDTO(ctx, keywords)
}

// searchArticlesByMySQLDTO MySQL LIKE 搜索（降级方案，返回 ArticleSearchDTO，对标Java）
func (s *ArticleService) searchArticlesByMySQLDTO(ctx context.Context, keywords string) ([]dto.ArticleSearchDTO, error) {
	var articles []model.Article

	searchPattern := "%" + keywords + "%"

	err := s.db.WithContext(ctx).
		Select("id, article_title, article_content").
		Where("is_delete = 0 AND status = 1").
		Where("article_title LIKE ? OR article_content LIKE ?", searchPattern, searchPattern).
		Order("create_time DESC").
		Limit(10).
		Find(&articles).Error
	if err != nil {
		return nil, fmt.Errorf("搜索文章失败: %w", err)
	}

	list := make([]dto.ArticleSearchDTO, len(articles))
	for i, a := range articles {
		list[i] = dto.ArticleSearchDTO{
			ID:             a.ID,
			ArticleTitle:   a.ArticleTitle,
			ArticleContent: a.ArticleContent,
		}
	}
	return list, nil
}

// convertESResultsToSearchDTO 将 ES 结果直接转换为 ArticleSearchDTO（对标Java，不查数据库）
func (s *ArticleService) convertESResultsToSearchDTO(esResults []map[string]interface{}) []dto.ArticleSearchDTO {
	list := make([]dto.ArticleSearchDTO, 0, len(esResults))
	for _, r := range esResults {
		item := dto.ArticleSearchDTO{}
		if id, ok := r["id"].(float64); ok {
			item.ID = uint(id)
		}
		if title, ok := r["articleTitle"].(string); ok {
			item.ArticleTitle = title
		}
		if content, ok := r["articleContent"].(string); ok {
			item.ArticleContent = content // 已包含 <em> 高亮标签
		}
		list = append(list, item)
	}
	return list
}

// searchArticlesByMySQL MySQL LIKE 搜索（降级方案）
func (s *ArticleService) searchArticlesByMySQL(ctx context.Context, keywords string, page dto.PageVO) (*dto.PageResultDTO, error) {
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

// convertESSearchResults 将 ES 搜索结果转换为 PageResultDTO（旧版，保留兼容）
func (s *ArticleService) convertESSearchResults(ctx context.Context, esResults []dto.ArticleSearchDTO, page dto.PageVO) (*dto.PageResultDTO, error) {
	// ES 返回的结果已经是全部匹配数据，需要手动分页
	total := int64(len(esResults))
	offset := page.GetOffset()
	end := offset + page.PageSize

	// 边界检查
	if offset >= len(esResults) {
		return &dto.PageResultDTO{
			List:     []dto.ArticleCardDTO{},
			Count:    total,
			PageNum:  page.PageNum,
			PageSize: page.PageSize,
		}, nil
	}

	if end > len(esResults) {
		end = len(esResults)
	}

	// 提取当前页的 ID 列表
	var articleIDs []uint
	for _, r := range esResults[offset:end] {
		articleIDs = append(articleIDs, r.ID)
	}

	// 从数据库查询完整的文章信息（含分类、作者、标签）
	var articles []model.Article
	if err := s.db.WithContext(ctx).
		Where("id IN ?", articleIDs).
		Preload("Category").
		Preload("UserInfo").
		Preload("Tags").
		Find(&articles).Error; err != nil {
		return nil, fmt.Errorf("查询文章详情失败: %w", err)
	}

	// 构建 ID -> Article 映射
	articleMap := make(map[uint]*model.Article)
	for i := range articles {
		articleMap[articles[i].ID] = &articles[i]
	}

	// 按 ES 搜索结果的顺序构建返回列表
	list := make([]dto.ArticleCardDTO, 0, len(articleIDs))
	for _, r := range esResults[offset:end] {
		if article, ok := articleMap[r.ID]; ok {
			card := s.toArticleCardDTO(article)
			// 用 ES 高亮内容覆盖原始内容
			if r.ArticleTitle != "" {
				card.ArticleTitle = r.ArticleTitle
			}
			if r.ArticleContent != "" {
				card.ArticleContent = r.ArticleContent
			}
			list = append(list, card)
		}
	}

	return &dto.PageResultDTO{
		List:     list,
		Count:    total,
		PageNum:  page.PageNum,
		PageSize: page.PageSize,
	}, nil
}

// convertESSearchResultsFromMap 将 ES 搜索结果（map格式）转换为 PageResultDTO（对标 Java EsSearchStrategyImpl）
func (s *ArticleService) convertESSearchResultsFromMap(ctx context.Context, esResults []map[string]interface{}, total int64, page dto.PageVO) (*dto.PageResultDTO, error) {
	// ES 返回的结果已经是全部匹配数据，需要手动分页
	offset := page.GetOffset()
	end := offset + page.PageSize

	// 边界检查
	if offset >= len(esResults) {
		return &dto.PageResultDTO{
			List:     []dto.ArticleCardDTO{},
			Count:    total,
			PageNum:  page.PageNum,
			PageSize: page.PageSize,
		}, nil
	}

	if end > len(esResults) {
		end = len(esResults)
	}

	// 提取当前页的 ID 列表
	var articleIDs []uint
	for _, r := range esResults[offset:end] {
		if id, ok := r["id"].(float64); ok {
			articleIDs = append(articleIDs, uint(id))
		}
	}

	if len(articleIDs) == 0 {
		return &dto.PageResultDTO{
			List:     []dto.ArticleCardDTO{},
			Count:    total,
			PageNum:  page.PageNum,
			PageSize: page.PageSize,
		}, nil
	}

	// 从数据库查询完整的文章信息（含分类、作者、标签）
	var articles []model.Article
	if err := s.db.WithContext(ctx).
		Where("id IN ?", articleIDs).
		Preload("Category").
		Preload("UserInfo").
		Preload("Tags").
		Find(&articles).Error; err != nil {
		return nil, fmt.Errorf("查询文章详情失败: %w", err)
	}

	// 构建 ID -> (Article, highlight) 映射
	type articleWithHighlight struct {
		article *model.Article
		highlightTitle   string
		highlightContent string
	}
	articleMap := make(map[uint]*articleWithHighlight)
	for i := range articles {
		articleMap[articles[i].ID] = &articleWithHighlight{article: &articles[i]}
	}

	// 将高亮信息填入映射（对标 Java 版处理高亮）
	for _, r := range esResults[offset:end] {
		if id, ok := r["id"].(float64); ok {
			uid := uint(id)
			if entry, exists := articleMap[uid]; exists {
				if title, ok := r["articleTitle"].(string); ok && title != "" {
					entry.highlightTitle = title
				}
				if content, ok := r["articleContent"].(string); ok && content != "" {
					entry.highlightContent = content
				}
			}
		}
	}

	// 按 ES 搜索结果的顺序构建返回列表
	list := make([]dto.ArticleCardDTO, 0, len(articleIDs))
	for _, r := range esResults[offset:end] {
		if id, ok := r["id"].(float64); ok {
			uid := uint(id)
			if entry, exists := articleMap[uid]; exists {
				card := s.toArticleCardDTO(entry.article)
				// 用 ES 高亮内容覆盖原始内容（对标 Java 版）
				if entry.highlightTitle != "" {
					card.ArticleTitle = entry.highlightTitle
				}
				if entry.highlightContent != "" {
					card.ArticleContent = entry.highlightContent
				}
				list = append(list, card)
			}
		}
	}

	return &dto.PageResultDTO{
		List:     list,
		Count:    total,
		PageNum:  page.PageNum,
		PageSize: page.PageSize,
	}, nil
}

// GetTopArticles 获取置顶/推荐文章（对标Java版 listTopAndFeaturedArticles）
func (s *ArticleService) GetTopArticles(ctx context.Context, limit int) ([]dto.ArticleCardDTO, error) {
	var articles []model.Article

	if err := s.db.WithContext(ctx).
		Where("is_delete = 0 AND status IN (1, 2) AND (is_top = 1 OR is_featured = 1)").
		Preload("Tags").
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
	// 构建 Author 嵌套对象（对标Java版 UserInfo author）
	var author *dto.UserInfoInCard
	if a.UserInfo != nil {
		author = &dto.UserInfoInCard{
			Nickname: a.UserInfo.Nickname,
			Website:  a.UserInfo.Website,
			Avatar:   a.UserInfo.Avatar,
		}
	}

	// 构建 Tags 数组（对标Java版 List<Tag> tags）
	tags := make([]dto.TagInCard, len(a.Tags))
	for i, t := range a.Tags {
		tags[i] = dto.TagInCard{
			TagName: t.TagName,
		}
	}

	card := dto.ArticleCardDTO{
		ID:             a.ID,
		ArticleCover:   a.ArticleCover,
		ArticleTitle:   a.ArticleTitle,
		ArticleContent: a.ArticleContent,
		IsTop:          a.IsTop,
		IsFeatured:     a.IsFeatured,
		Author:         author,
		CategoryName:   "",
		Tags:           tags,
		Status:         a.Status,
		CreateTime:     a.CreateTime,
	}

	if a.Category != nil {
		card.CategoryName = a.Category.CategoryName
	}

	// 安全处理 UpdateTime（指针类型转值类型）
	if a.UpdateTime != nil {
		card.UpdateTime = *a.UpdateTime
	}

	return card
}

// toArticleAdminDTO 转换为后台文章管理DTO（完全对标Java ArticleAdminDTO）
func (s *ArticleService) toArticleAdminDTO(a *model.Article) dto.ArticleAdminDTO {
	// 获取浏览量（从 Redis ZSet）
	viewsCount := 0
	if s.statsService != nil {
		if count, err := s.statsService.GetArticleView(context.Background(), a.ID); err == nil {
			viewsCount = int(count)
		}
	}

	// 构建标签列表（tagDTOs）
	tagDTOs := make([]dto.TagDTO, len(a.Tags))
	for i, t := range a.Tags {
		tagDTOs[i] = dto.TagDTO{
			ID:      t.ID,
			TagName: t.TagName,
		}
	}

	// 获取分类名
	categoryName := ""
	if a.Category != nil {
		categoryName = a.Category.CategoryName
	}

	return dto.ArticleAdminDTO{
		ID:           a.ID,
		ArticleCover: a.ArticleCover,
		ArticleTitle: a.ArticleTitle,
		CreateTime:   a.CreateTime,
		ViewsCount:   viewsCount,
		CategoryName: categoryName,
		TagDTOs:      tagDTOs,
		IsTop:        a.IsTop,
		IsFeatured:   a.IsFeatured,
		IsDelete:     a.IsDelete,
		Status:       a.Status,
		Type:         a.Type,
	}
}

// ImportArticles 导入Markdown文章 (对标Java版导入功能)
type ImportError struct {
	Filename string
	Error    error
}

func (s *ArticleService) ImportArticles(ctx context.Context, userID uint, contents map[string]string) (success int, failures []ImportError) {
	for filename, content := range contents {
		// 对标Java NormalArticleImportStrategyImpl: 使用文件名（不含扩展名）作为文章标题
		// String articleTitle = file.getOriginalFilename().split("\\.")[0];
		title := filename
		if idx := strings.LastIndex(title, "."); idx != -1 {
			title = title[:idx]
		}

		// 文章内容为整个文件内容
		body := content

		vo := vo.ArticleVO{
			ArticleTitle:   strings.TrimSpace(title),
			ArticleContent: body,
			CategoryName:   "默认分类",
			Status:         int8Ptr(3), // 草稿状态 (对齐Java DRAFT=3)
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
