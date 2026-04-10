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

// TalkService 说说业务逻辑 (对标 Java TalkServiceImpl)
type TalkService struct {
	db *gorm.DB
}

func NewTalkService(db *gorm.DB) *TalkService {
	return &TalkService{db: db}
}

// CreateTalk 发布说说
func (s *TalkService) CreateTalk(ctx context.Context, userID uint, vo vo.TalkVO) (*model.Talk, error) {
	talk := model.Talk{
		UserID:  userID,
		Content: vo.Content,
		Status:  1,
	}

	if vo.Status != 0 {
		talk.Status = vo.Status
	}

	if err := s.db.WithContext(ctx).Create(&talk).Error; err != nil {
		return nil, fmt.Errorf("发布说说失败: %w", err)
	}
	return &talk, nil
}

// UpdateTalk 更新说说
func (s *TalkService) UpdateTalk(ctx context.Context, id uint, vo vo.TalkVO) error {
	var talk model.Talk
	if err := s.db.WithContext(ctx).First(&talk, id).Error; err != nil {
		return errors.ErrTalkNotFound
	}

	updates := map[string]interface{}{
		"content": vo.Content,
	}
	if vo.Status != 0 {
		updates["status"] = vo.Status
	}

	return s.db.WithContext(ctx).Model(&talk).Updates(updates).Error
}

// DeleteTalk 删除说说 (事务: 删除关联评论)
func (s *TalkService) DeleteTalk(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var talk model.Talk
		if err := tx.First(&talk, id).Error; err != nil {
			return errors.ErrTalkNotFound
		}

		// 删除说说下的所有评论
		tx.Delete(&model.Comment{}, "talk_id = ? AND type = ?", id, 2)

		if err := tx.Delete(&talk).Error; err != nil {
			return fmt.Errorf("删除说说失败: %w", err)
		}
		return nil
	})
}

// GetTalks 获取公开说说列表(前台用)
func (s *TalkService) GetTalks(ctx context.Context, page dto.PageVO) (*dto.PageResultDTO, error) {
	var talks []model.Talk
	var count int64

	baseQuery := s.db.WithContext(ctx).
		Model(&model.Talk{}).
		Where("status = 1")

	baseQuery.Count(&count)

	offset := page.GetOffset()
	err := baseQuery.
		Preload("UserInfo").
		Preload("Comments", "is_review = 1 AND parent_id = 0"). // 只加载顶级评论
		Order("is_top DESC, create_time DESC").
		Limit(page.PageSize).
		Offset(offset).
		Find(&talks).Error

	if err != nil {
		return nil, fmt.Errorf("查询说说列表失败: %w", err)
	}

	list := make([]dto.TalkDTO, len(talks))
	for i, t := range talks {
		list[i] = dto.TalkDTO{
			ID:         t.ID,
			UserID:     t.UserID,
			Content:    t.Content,
			IsTop:      t.IsTop,
			Status:     t.Status,
			LikeCount:  t.LikeCount,
			CreateTime: t.CreateTime,
		}
		if t.UserInfo != nil {
			list[i].Nickname = t.UserInfo.Nickname
			list[i].Avatar = t.UserInfo.Avatar
		}
		list[i].CommentCount = len(t.Comments)
	}

	return &dto.PageResultDTO{
		List:     list,
		Count:    count,
		PageNum:  page.PageNum,
		PageSize: page.PageSize,
	}, nil
}

// GetTalkByID 根据ID获取说说详情(含评论树)
func (s *TalkService) GetTalkByID(ctx context.Context, id uint) (*dto.TalkDetailDTO, error) {
	var talk model.Talk

	err := s.db.WithContext(ctx).
		Preload("UserInfo").
		First(&talk, id).Error

	if err != nil {
		if errors.IsStd(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrTalkNotFound
		}
		return nil, fmt.Errorf("查询说说详情失败: %w", err)
	}

	dto := &dto.TalkDetailDTO{
		TalkDTO: dto.TalkDTO{
			ID:         talk.ID,
			UserID:     talk.UserID,
			Content:    talk.Content,
			IsTop:      talk.IsTop,
			Status:     talk.Status,
			LikeCount:  talk.LikeCount,
			CreateTime: talk.CreateTime,
		},
	}

	if talk.UserInfo != nil {
		dto.Nickname = talk.UserInfo.Nickname
		dto.Avatar = talk.UserInfo.Avatar
	}

	return dto, nil
}

// LikeTalk 点赞说说
func (s *TalkService) LikeTalk(ctx context.Context, id uint) error {
	result := s.db.WithContext(ctx).
		Model(&model.Talk{}).
		Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("like_count + 1"))

	if result.Error != nil {
		return fmt.Errorf("点赞说说失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.ErrTalkNotFound
	}
	return nil
}

// SetTalkTop 设置说说置顶
func (s *TalkService) SetTalkTop(ctx context.Context, id uint, isTop int8) error {
	result := s.db.WithContext(ctx).
		Model(&model.Talk{}).
		Where("id = ?", id).
		Update("is_top", isTop)

	if result.Error != nil {
		return fmt.Errorf("设置说说置顶失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.ErrTalkNotFound
	}
	return nil
}

// ListAdminTalks 后台管理分页查询说说
func (s *TalkService) ListAdminTalks(ctx context.Context, cond dto.ConditionVO, page dto.PageVO) (*dto.PageResultDTO, error) {
	var talks []model.Talk
	var count int64

	baseQuery := s.db.WithContext(ctx).Model(&model.Talk{})

	if cond.Keywords != "" {
		baseQuery = baseQuery.Where("content LIKE ?", "%"+cond.Keywords+"%")
	}
	if cond.Status != nil && *cond.Status > 0 {
		baseQuery = baseQuery.Where("status = ?", *cond.Status)
	}

	baseQuery.Count(&count)

	offset := page.GetOffset()
	err := baseQuery.
		Preload("UserInfo").
		Order("create_time DESC").
		Limit(page.PageSize).
		Offset(offset).
		Find(&talks).Error

	if err != nil {
		return nil, fmt.Errorf("查询说说列表失败: %w", err)
	}

	list := make([]dto.TalkDTO, len(talks))
	for i, t := range talks {
		list[i] = dto.TalkDTO{
			ID:         t.ID,
			UserID:     t.UserID,
			Content:    t.Content[:MinInt(len(t.Content), 100)] + "...",
			IsTop:      t.IsTop,
			Status:     t.Status,
			LikeCount:  t.LikeCount,
			CreateTime: t.CreateTime,
		}
		if t.UserInfo != nil {
			list[i].Nickname = t.UserInfo.Nickname
		}
	}

	return &dto.PageResultDTO{
		List:     list,
		Count:    count,
		PageNum:  page.PageNum,
		PageSize: page.PageSize,
	}, nil
}

func MinInt(a, b int) int { if a < b { return a }; return b }
