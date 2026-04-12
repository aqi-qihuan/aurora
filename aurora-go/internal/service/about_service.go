package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/model"
	"gorm.io/gorm"
)

// AboutService 关于页面业务逻辑 (对标 Java AboutServiceImpl)
type AboutService struct {
	db *gorm.DB
}

func NewAboutService(db *gorm.DB) *AboutService {
	return &AboutService{db: db}
}

// GetAbout 获取关于页面内容
func (s *AboutService) GetAbout(ctx context.Context) (string, error) {
	var about model.About

	err := s.db.WithContext(ctx).First(&about, 1).Error
	if err != nil && !errors.IsStd(err, gorm.ErrRecordNotFound) {
		return "", fmt.Errorf("查询关于页面失败: %w", err)
	}

	if about.ID == 0 {
		return "", nil
	}
	return about.Content, nil
}

// UpdateAbout 更新关于页面内容
func (s *AboutService) UpdateAbout(ctx context.Context, content string) error {
	var about model.About

	// 使用FirstOrCreate确保记录存在
	s.db.WithContext(ctx).FirstOrCreate(&about, model.About{ID: 1})

	result := s.db.WithContext(ctx).
		Model(&about).
		Update("content", content)

	if result.Error != nil {
		return fmt.Errorf("更新关于页面失败: %w", result.Error)
	}

	slog.Info("关于页面已更新")
	return nil
}

// GetAboutComments 获取关于页面的评论列表
func (s *AboutService) GetAboutComments(ctx context.Context) ([]dto.CommentDTO, error) {
	var comments []model.Comment

	err := s.db.WithContext(ctx).
		Where("type = 4 AND is_review = 1").
		Preload("UserInfo").
		Preload("ReplyUser").
		Order("create_time ASC").
		Find(&comments).Error

	if err != nil {
		return nil, fmt.Errorf("查询关于页评论失败: %w", err)
	}

	list := make([]dto.CommentDTO, len(comments))
	for i, c := range comments {
		list[i] = dto.CommentDTO{
			ID:         c.ID,
			UserID:     c.UserID,
			Content:    c.Content,
			Type:       4,
			ParentID:   c.ParentID,
			LikeCount:  0, // t_comment 表没有 like_count 字段，需要从 Redis 获取
			IsReview:   c.IsReview,
			CreateTime: c.CreateTime,
		}
		if c.UserInfo != nil {
			list[i].Nickname = c.UserInfo.Nickname
			list[i].Avatar = c.UserInfo.Avatar
		}
		if c.ReplyUser != nil {
			list[i].ReplyNickname = c.ReplyUser.Nickname
		}
	}
	return list, nil
}
