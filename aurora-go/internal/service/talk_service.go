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
	db           *gorm.DB
	statsService *RedisStatsService // Redis 统计服务
}

func NewTalkService(db *gorm.DB, statsService *RedisStatsService) *TalkService {
	return &TalkService{
		db:           db,
		statsService: statsService,
	}
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

		// 删除说说下的所有评论（对标Java: topic_id + type=5）
		tx.Where("topic_id = ? AND type = ?", id, 5).Delete(&model.Comment{})

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
			LikeCount:  s.getTalkLikeCount(t.ID),
			CreateTime: t.CreateTime,
		}
		if t.UserInfo != nil {
			list[i].Nickname = t.UserInfo.Nickname
			list[i].Avatar = t.UserInfo.Avatar
		}
	}

	// 批量查询评论数（一次SQL替代N次循环查询，优化N+1问题）
	if len(talks) > 0 {
		talkIDs := make([]uint, len(talks))
		for i, t := range talks {
			talkIDs[i] = t.ID
		}

		var commentCounts []struct {
			TopicID uint
			Count   int
		}
		s.db.WithContext(ctx).
			Table("t_comment").
			Select("topic_id, COUNT(*) as count").
			Where("topic_id IN ? AND type = 5 AND is_review = 1", talkIDs).
			Group("topic_id").
			Find(&commentCounts)

		// 构建 ID -> Count 映射
		countMap := make(map[uint]int, len(commentCounts))
		for _, cc := range commentCounts {
			countMap[cc.TopicID] = cc.Count
		}

		// 填充评论数
		for i := range list {
			list[i].CommentCount = countMap[list[i].ID]
		}
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
			LikeCount:  s.getTalkLikeCount(talk.ID),
			CreateTime: talk.CreateTime,
		},
	}

	if talk.UserInfo != nil {
		dto.Nickname = talk.UserInfo.Nickname
		dto.Avatar = talk.UserInfo.Avatar
	}

	return dto, nil
}

// LikeTalk 点赞说说 (使用 Redis)
func (s *TalkService) LikeTalk(ctx context.Context, id uint) error {
	// TODO: 从 Context 中获取 userID
	// 临时方案：直接返回成功
	if s.statsService != nil {
		// 假设 userID = 1，实际需要从中获取
		return s.statsService.LikeTalk(ctx, id, 1)
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
			LikeCount:  s.getTalkLikeCount(t.ID),
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

// getTalkLikeCount 获取说说点赞数（从 Redis）
func (s *TalkService) getTalkLikeCount(talkID uint) int64 {
	if s.statsService == nil {
		return 0
	}
	count, _ := s.statsService.GetTalkLikeCount(context.Background(), talkID)
	return count
}
