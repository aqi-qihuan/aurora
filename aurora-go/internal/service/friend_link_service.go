package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/model"
	"github.com/aurora-go/aurora/internal/vo"
	"gorm.io/gorm"
)

// FriendLinkService 友链业务逻辑 (对标 Java FriendLinkServiceImpl)
type FriendLinkService struct {
	db *gorm.DB
}

func NewFriendLinkService(db *gorm.DB) *FriendLinkService {
	return &FriendLinkService{db: db}
}

// ApplyFriendLink 申请友链
func (s *FriendLinkService) ApplyFriendLink(ctx context.Context, userID uint, vo vo.FriendLinkVO) (*model.FriendLink, error) {
	link := model.FriendLink{
		UserID:      &userID,
		LinkName:    vo.LinkName,
		LinkAvatar:  vo.LinkAvatar,
		LinkAddress: vo.LinkAddress,
		LinkIntro:   vo.LinkIntro,
		Status:      0, // 待审核
	}

	if err := s.db.WithContext(ctx).Create(&link).Error; err != nil {
		return nil, fmt.Errorf("申请友链失败: %w", err)
	}
	return &link, nil
}

// ReviewFriendLink 审核友链申请 (通过/拒绝)
func (s *FriendLinkService) ReviewFriendLink(ctx context.Context, id uint, status int8) error {
	result := s.db.WithContext(ctx).
		Model(&model.FriendLink{}).
		Where("id = ?", id).
		Update("status", status)

	if result.Error != nil {
		return fmt.Errorf("审核友链失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.ErrFriendLinkNotFound
	}

	statusText := "通过"
	if status == -1 {
		statusText = "拒绝"
	}
	slog.Info("友链审核完成", "id", id, "status", statusText)
	return nil
}

// GetApprovedLinks 获取已通过审核的友链列表(前台用)
func (s *FriendLinkService) GetApprovedLinks(ctx context.Context) ([]dto.FriendLinkDTO, error) {
	var links []model.FriendLink

	err := s.db.WithContext(ctx).
		Preload("UserInfo").
		Where("status = 1").
		Order("create_time DESC").
		Find(&links).Error

	if err != nil {
		return nil, fmt.Errorf("查询友链列表失败: %w", err)
	}

	list := make([]dto.FriendLinkDTO, len(links))
	for i, l := range links {
		list[i] = dto.FriendLinkDTO{
			ID:          l.ID,
			LinkName:    l.LinkName,
			LinkAvatar:  l.LinkAvatar,
			LinkAddress: l.LinkAddress,
			LinkIntro:   l.LinkIntro,
			CreateTime:   l.CreateTime,
		}
		if l.UserInfo != nil {
			list[i].Nickname = l.UserInfo.Nickname
		}
	}
	return list, nil
}

// ListAdminLinks 后台管理分页查询友链(含待审核/已拒绝)
func (s *FriendLinkService) ListAdminLinks(ctx context.Context, cond dto.ConditionVO, page dto.PageVO) (*dto.PageResultDTO, error) {
	var links []model.FriendLink
	var count int64

	baseQuery := s.db.WithContext(ctx).Model(&model.FriendLink{})

	if cond.Keywords != "" {
		baseQuery = baseQuery.Where("link_name LIKE ? OR link_address LIKE ?", "%"+cond.Keywords+"%", "%"+cond.Keywords+"%")
	}
	if cond.Status != nil && *cond.Status >= -1 {
		baseQuery = baseQuery.Where("status = ?", *cond.Status)
	}

	if err := baseQuery.Count(&count).Error; err != nil {
		return nil, fmt.Errorf("统计友链数失败: %w", err)
	}

	offset := page.GetOffset()
	if err := baseQuery.
		Preload("UserInfo").
		Order("create_time DESC").
		Limit(page.PageSize).
		Offset(offset).
		Find(&links).Error; err != nil {
		return nil, fmt.Errorf("查询友链列表失败: %w", err)
	}

	list := make([]dto.FriendLinkAdminDTO, len(links))
	for i, l := range links {
		list[i] = dto.FriendLinkAdminDTO{
			ID:          l.ID,
			UserID:      l.UserID,
			LinkName:    l.LinkName,
			LinkAvatar:  l.LinkAvatar,
			LinkAddress: l.LinkAddress,
			LinkIntro:   l.LinkIntro,
			Status:      l.Status,
			CreateTime:   l.CreateTime,
		}
		if l.UserInfo != nil {
			list[i].Nickname = l.UserInfo.Nickname
		}
	}

	return &dto.PageResultDTO{
		List:     list,
		Count:    count,
		PageNum:  page.PageNum,
		PageSize: page.PageSize,
	}, nil
}

// UpdateFriendLink 更新友链信息
func (s *FriendLinkService) UpdateFriendLink(ctx context.Context, id uint, vo vo.FriendLinkVO) error {
	var link model.FriendLink
	if err := s.db.WithContext(ctx).First(&link, id).Error; err != nil {
		return errors.ErrFriendLinkNotFound
	}

	updates := map[string]interface{}{
		"link_name":     vo.LinkName,
		"link_avatar":   vo.LinkAvatar,
		"link_address":  vo.LinkAddress,
		"link_intro":    vo.LinkIntro,
	}

	return s.db.WithContext(ctx).Model(&link).Updates(updates).Error
}

// DeleteFriendLink 删除友链
func (s *FriendLinkService) DeleteFriendLink(ctx context.Context, id uint) error {
	result := s.db.WithContext(ctx).Delete(&model.FriendLink{}, id)
	if result.Error != nil {
		return fmt.Errorf("删除友链失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.ErrFriendLinkNotFound
	}
	return nil
}

// SetFriendLinkOnline 设置友链上线/下线
func (s *FriendLinkService) SetFriendLinkOnline(ctx context.Context, id uint, online bool) error {
	status := int8(1)
	if !online {
		status = -1 // 下线
	}

	result := s.db.WithContext(ctx).
		Model(&model.FriendLink{}).
		Where("id = ?", id).
		Where("status = 1").
		Update("status", status)

	if result.Error != nil {
		return fmt.Errorf("设置友链状态失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.ErrFriendLinkNotFound
	}
	return nil
}
