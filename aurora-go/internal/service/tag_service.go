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

// TagService 标签业务逻辑 (对标 Java TagServiceImpl)
type TagService struct {
	db *gorm.DB
}

func NewTagService(db *gorm.DB) *TagService {
	return &TagService{db: db}
}

// CreateTag 创建标签
func (s *TagService) CreateTag(ctx context.Context, vo vo.TagVO) (*model.Tag, error) {
	tag := model.Tag{
		TagName: vo.TagName,
	}

	if err := s.db.WithContext(ctx).Create(&tag).Error; err != nil {
		if errors.IsStd(err, gorm.ErrDuplicatedKey) {
			return nil, errors.ErrTagNameExists
		}
		return nil, fmt.Errorf("创建标签失败: %w", err)
	}

	return &tag, nil
}

// UpdateTag 更新标签
func (s *TagService) UpdateTag(ctx context.Context, id uint, vo vo.TagVO) error {
	var tag model.Tag
	if err := s.db.WithContext(ctx).First(&tag, id).Error; err != nil {
		return errors.ErrTagNotFound
	}

	result := s.db.WithContext(ctx).
		Model(&tag).
		Update("tag_name", vo.TagName)

	if result.Error != nil {
		if errors.IsStd(result.Error, gorm.ErrDuplicatedKey) {
			return errors.ErrTagNameExists
		}
		return fmt.Errorf("更新标签失败: %w", result.Error)
	}
	return nil
}

// DeleteTag 删除标签 (清理文章关联)
func (s *TagService) DeleteTag(ctx context.Context, id uint) error {
	var tag model.Tag

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&tag, id).Error; err != nil {
			return errors.ErrTagNotFound
		}

		// 清理文章-标签关联表
		tx.Exec("DELETE FROM t_article_tag WHERE tag_id = ?", id)

		if err := tx.Delete(&tag).Error; err != nil {
			return fmt.Errorf("删除标签失败: %w", err)
		}
		return nil
	})
}

// GetTags 获取所有标签列表 (含文章计数)
func (s *TagService) GetTags(ctx context.Context) ([]dto.TagDTO, error) {
	var tags []model.Tag

	err := s.db.WithContext(ctx).
		Order("article_count DESC, create_time DESC").
		Find(&tags).Error

	if err != nil {
		return nil, fmt.Errorf("查询标签失败: %w", err)
	}

	list := make([]dto.TagDTO, len(tags))
	for i, t := range tags {
		list[i] = dto.TagDTO{ID: t.ID, TagName: t.TagName}
	}
	return list, nil
}

// SearchTags 模糊搜索标签 (用于文章编辑器标签选择器)
func (s *TagService) SearchTags(ctx context.Context, keyword string) ([]dto.TagDTO, error) {
	var tags []model.Tag

	err := s.db.WithContext(ctx).
		Where("tag_name LIKE ?", "%"+keyword+"%").
		Limit(20).
		Find(&tags).Error

	if err != nil {
		return nil, fmt.Errorf("搜索标签失败: %w", err)
	}

	list := make([]dto.TagDTO, len(tags))
	for i, t := range tags {
		list[i] = dto.TagDTO{ID: t.ID, TagName: t.TagName}
	}
	return list, nil
}

// GetTagByID 根据ID获取标签详情
func (s *TagService) GetTagByID(ctx context.Context, id uint) (*dto.TagDetailDTO, error) {
	var tag model.Tag

	err := s.db.WithContext(ctx).First(&tag, id).Error
	if err != nil {
		if errors.IsStd(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrTagNotFound
		}
		return nil, fmt.Errorf("查询标签详情失败: %w", err)
	}

	result := &dto.TagDetailDTO{
		ID:           tag.ID,
		TagName:      tag.TagName,
		ArticleCount: tag.ArticleCount,
	}
	return result, nil
}

// BatchSaveOrGetTags 批量保存或获取标签 (文章发布时使用)
// 输入: ["Go", "Gin"] → 返回: [1, 2] (已存在则返回ID, 不存在则新建)
func (s *TagService) BatchSaveOrGetTags(ctx context.Context, tagNames []string) ([]uint, error) {
	if len(tagNames) == 0 {
		return []uint{}, nil
	}

	var tagIDs []uint

	for _, name := range tagNames {
		var tag model.Tag
		err := s.db.WithContext(ctx).
			Where("tag_name = ?", name).
			First(&tag).Error

		if err == nil {
			// 标签已存在 → 使用已有ID
			tagIDs = append(tagIDs, tag.ID)
		} else if errors.IsStd(err, gorm.ErrRecordNotFound) {
			// 标签不存在 → 新建
			newTag := model.Tag{TagName: name}
			if createErr := s.db.Create(&newTag).Error; createErr != nil {
				return nil, fmt.Errorf("创建标签[%s]失败: %w", name, createErr)
			}
			tagIDs = append(tagIDs, newTag.ID)
		} else {
			return nil, fmt.Errorf("查询标签[%s]失败: %w", name, err)
		}
	}

	return tagIDs, nil
}

// GetArticleTags 获取文章的标签列表
func (s *TagService) GetArticleTags(ctx context.Context, articleID uint) ([]dto.TagDTO, error) {
	var tags []model.Tag

	err := s.db.WithContext(ctx).
		Joins("JOIN t_article_tag ON t_article_tag.tag_id = t_tag.id").
		Where("t_article_tag.article_id = ?", articleID).
		Find(&tags).Error

	if err != nil {
		return nil, fmt.Errorf("查询文章标签失败: %w", err)
	}

	list := make([]dto.TagDTO, len(tags))
	for i, t := range tags {
		list[i] = dto.TagDTO{ID: t.ID, TagName: t.TagName}
	}
	return list, nil
}
