package service

import (
	"context"
	"encoding/json" // ✅ 新增导入，用于解析JSON字符串
	"fmt"
	"log/slog"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/infrastructure/database"
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

// GetAboutDTO 获取关于页面DTO（对标Java AuroraInfoServiceImpl.getAbout）
// 数据库存储JSON字符串，解析为AboutDTO对象返回
func (s *AboutService) GetAboutDTO(ctx context.Context) (*dto.AboutDTO, error) {
	var about model.About

	err := s.db.WithContext(ctx).First(&about, 1).Error
	if err != nil && !errors.IsStd(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询关于页面失败: %w", err)
	}

	if about.ID == 0 {
		return &dto.AboutDTO{Content: ""}, nil
	}

	// 对标Java：JSON.parseObject(content, AboutDTO.class)
	// 数据库存储的是JSON格式字符串，解析为DTO对象
	aboutDTO := &dto.AboutDTO{}
	if err := json.Unmarshal([]byte(about.Content), aboutDTO); err != nil {
		// 如果不是JSON格式，直接返回content内容
		aboutDTO.Content = about.Content
	}

	return aboutDTO, nil
}

// GetAbout 获取关于页面内容（保留旧方法，兼容其他调用）
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

// UpdateAbout 更新关于页面内容（对标Java AuroraInfoServiceImpl.updateAbout）
// Java: aboutMapper.updateById(About.builder().id(1).content(JSON.toJSONString(aboutVO)).build());
// Java: redisService.del(ABOUT);
func (s *AboutService) UpdateAbout(ctx context.Context, content string) error {
	// 1. 将内容包装为AboutVO结构，然后序列化为JSON（对标Java JSON.toJSONString(aboutVO)）
	aboutVO := map[string]interface{}{
		"content": content,
	}
	contentJSON, err := json.Marshal(aboutVO)
	if err != nil {
		return fmt.Errorf("序列化关于页面失败: %w", err)
	}

	var about model.About

	// 2. 使用FirstOrCreate确保记录存在
	s.db.WithContext(ctx).FirstOrCreate(&about, model.About{ID: 1})

	// 3. 保存JSON格式的内容（对标Java updateById）
	result := s.db.WithContext(ctx).
		Model(&about).
		Update("content", string(contentJSON))

	if result.Error != nil {
		return fmt.Errorf("更新关于页面失败: %w", result.Error)
	}

	// 4. 删除Redis缓存（对标Java redisService.del(ABOUT)）
	rdb := database.GetRedis()
	if rdb != nil {
		if err := rdb.Del(ctx, "about").Err(); err != nil {
			slog.Warn("删除关于页面Redis缓存失败", "error", err)
		}
		slog.Info("已删除关于页面Redis缓存")
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
			Content:    c.CommentContent, // 修复: Content -> CommentContent
			Type:       4,
			ParentID:   c.ParentID,
			LikeCount:  0,
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
