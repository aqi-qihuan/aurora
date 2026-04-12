package service

import (
	"context"
	"fmt"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/model"
	"github.com/aurora-go/aurora/internal/vo"
	"gorm.io/gorm"
)

// CategoryService 分类业务逻辑 (对标 Java CategoryServiceImpl)
type CategoryService struct {
	db *gorm.DB
}

func NewCategoryService(db *gorm.DB) *CategoryService {
	return &CategoryService{db: db}
}

// CreateCategory 创建分类
func (s *CategoryService) CreateCategory(ctx context.Context, vo vo.CategoryVO) (*model.Category, error) {
	category := model.Category{
		CategoryName: vo.CategoryName,
	}

	if err := s.db.WithContext(ctx).Create(&category).Error; err != nil {
		if errors.IsStd(err, gorm.ErrDuplicatedKey) {
			return nil, errors.ErrCategoryNameExists
		}
		return nil, fmt.Errorf("创建分类失败: %w", err)
	}

	return &category, nil
}

// UpdateCategory 更新分类
func (s *CategoryService) UpdateCategory(ctx context.Context, id uint, vo vo.CategoryVO) error {
	var category model.Category
	if err := s.db.WithContext(ctx).First(&category, id).Error; err != nil {
		return errors.ErrCategoryNotFound
	}

	updates := map[string]interface{}{
		"category_name": vo.CategoryName,
	}

	result := s.db.WithContext(ctx).Model(&category).Updates(updates)
	if result.Error != nil {
		if errors.IsStd(result.Error, gorm.ErrDuplicatedKey) {
			return errors.ErrCategoryNameExists
		}
		return fmt.Errorf("更新分类失败: %w", result.Error)
	}

	return nil
}

// DeleteCategory 删除分类 (事务: 检查关联文章 + 清理标签 + 递归删除子分类)
func (s *CategoryService) DeleteCategory(ctx context.Context, id uint) error {
	var category model.Category

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&category, id).Error; err != nil {
			return errors.ErrCategoryNotFound
		}

		// 检查是否有关联文章
		var articleCount int64
		tx.Model(&model.Article{}).Where("category_id = ? AND is_delete = 0", id).Count(&articleCount)
		if articleCount > 0 {
			return errors.ErrCategoryHasArticles
		}

		// 删除子分类 (递归)
		tx.Delete(&model.Category{}, "parent_id = ?", id)

		// 删除自身
		if err := tx.Delete(&category).Error; err != nil {
			return fmt.Errorf("删除分类失败: %w", err)
		}

		return nil
	})
}

// GetCategories 获取所有分类列表
func (s *CategoryService) GetCategories(ctx context.Context) ([]dto.CategoryDTO, error) {
	var categories []model.Category

	err := s.db.WithContext(ctx).
		Order("create_time ASC").
		Find(&categories).Error

	if err != nil {
		return nil, fmt.Errorf("查询分类失败: %w", err)
	}

	list := make([]dto.CategoryDTO, len(categories))
	for i, c := range categories {
		list[i] = dto.CategoryDTO{
			ID:            c.ID,
			CategoryName:  c.CategoryName,
			ArticleCount:  0, // t_category 表没有 article_count 字段，需要动态统计
			CreateTime:    c.CreateTime,
		}
	}
	return list, nil
}

// ListAdminCategories 后台管理分页查询分类（对标 Java listCategoriesAdmin）
func (s *CategoryService) ListAdminCategories(ctx context.Context, cond dto.ConditionVO, page dto.PageVO) (*dto.PageResultDTO, error) {
	var categories []model.Category
	var count int64

	baseQuery := s.db.WithContext(ctx).Model(&model.Category{})

	if cond.Keywords != "" {
		baseQuery = baseQuery.Where("category_name LIKE ?", "%"+cond.Keywords+"%")
	}

	baseQuery.Count(&count)

	offset := page.GetOffset()
	err := baseQuery.
		Order("create_time ASC").
		Limit(page.PageSize).
		Offset(offset).
		Find(&categories).Error

	if err != nil {
		return nil, fmt.Errorf("查询分类列表失败: %w", err)
	}

	list := make([]dto.CategoryDTO, len(categories))
	for i, c := range categories {
		list[i] = dto.CategoryDTO{
			ID:           c.ID,
			CategoryName: c.CategoryName,
			ArticleCount: 0,
			CreateTime:   c.CreateTime,
		}
	}

	return &dto.PageResultDTO{
		List:     list,
		Count:    count,
		PageNum:  page.PageNum,
		PageSize: page.PageSize,
	}, nil
}

// GetCategoryOptions 获取分类下拉选项(用于文章编辑器)
func (s *CategoryService) GetCategoryOptions(ctx context.Context) ([]dto.OptionDTO, error) {
	var categories []model.Category

	err := s.db.WithContext(ctx).
		Select("id, category_name").
		Order("create_time ASC").
		Find(&categories).Error

	if err != nil {
		return nil, fmt.Errorf("查询分类选项失败: %w", err)
	}

	options := make([]dto.OptionDTO, len(categories))
	for i, c := range categories {
		options[i] = dto.OptionDTO{Label: c.CategoryName, Value: c.ID}
	}
	return options, nil
}

// GetCategoryByID 根据ID获取分类详情
func (s *CategoryService) GetCategoryByID(ctx context.Context, id uint) (*dto.CategoryDTO, error) {
	var category model.Category
	
	err := s.db.WithContext(ctx).First(&category, id).Error
	if err != nil {
		if errors.IsStd(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrCategoryNotFound
		}
		return nil, fmt.Errorf("查询分类失败: %w", err)
	}

	return &dto.CategoryDTO{
		ID:            category.ID,
		CategoryName:  category.CategoryName,
		ArticleCount:  0, // t_category 表没有 article_count 字段，需要动态统计
	}, nil
}
