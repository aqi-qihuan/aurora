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

// ListFriendLinks 获取友链列表（前台，对标Java: listFriendLinks）
func (s *FriendLinkService) ListFriendLinks(ctx context.Context) ([]dto.FriendLinkDTO, error) {
	var links []model.FriendLink

	err := s.db.WithContext(ctx).
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
			CreateTime:  l.CreateTime,
		}
	}
	return list, nil
}

// ListAdminLinks 后台管理分页查询友链(对标Java: listFriendLinksAdmin)
func (s *FriendLinkService) ListAdminLinks(ctx context.Context, cond dto.ConditionVO, page dto.PageVO) (*dto.PageResultDTO, error) {
	var links []model.FriendLink
	var count int64

	baseQuery := s.db.WithContext(ctx).Model(&model.FriendLink{})

	if cond.Keywords != "" {
		baseQuery = baseQuery.Where("link_name LIKE ?", "%"+cond.Keywords+"%")
	}

	if err := baseQuery.Count(&count).Error; err != nil {
		return nil, fmt.Errorf("统计友链数失败: %w", err)
	}

	offset := page.GetOffset()
	if err := baseQuery.
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
			LinkName:    l.LinkName,
			LinkAvatar:  l.LinkAvatar,
			LinkAddress: l.LinkAddress,
			LinkIntro:   l.LinkIntro,
			CreateTime:  l.CreateTime,
		}
	}

	return &dto.PageResultDTO{
		List:     list,
		Count:    count,
		PageNum:  page.PageNum,
		PageSize: page.PageSize,
	}, nil
}

// SaveOrUpdateFriendLink 新增或更新友链 (对标Java: saveOrUpdateFriendLink)
func (s *FriendLinkService) SaveOrUpdateFriendLink(ctx context.Context, linkVO vo.FriendLinkVO) error {
	link := model.FriendLink{
		LinkName:    linkVO.LinkName,
		LinkAvatar:  linkVO.LinkAvatar,
		LinkAddress: linkVO.LinkAddress,
		LinkIntro:   linkVO.LinkIntro,
	}

	if linkVO.ID > 0 {
		link.ID = linkVO.ID
	}

	if err := s.db.WithContext(ctx).Save(&link).Error; err != nil {
		return fmt.Errorf("保存友链失败: %w", err)
	}
	return nil
}

// DeleteFriendLink 删除单个友链
func (s *FriendLinkService) DeleteFriendLink(ctx context.Context, id uint) error {
	result := s.db.WithContext(ctx).Delete(&model.FriendLink{}, id)
	if result.Error != nil {
		return fmt.Errorf("删除友链失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.ErrFriendLinkNotFound
	}
	slog.Info("删除友链", "id", id)
	return nil
}

// DeleteFriendLinks 批量删除友链 (对标Java: removeByIds)
func (s *FriendLinkService) DeleteFriendLinks(ctx context.Context, ids []uint) error {
	if len(ids) == 0 {
		return fmt.Errorf("友链ID列表不能为空")
	}
	result := s.db.WithContext(ctx).Delete(&model.FriendLink{}, ids)
	if result.Error != nil {
		return fmt.Errorf("批量删除友链失败: %w", result.Error)
	}
	slog.Info("批量删除友链", "count", result.RowsAffected)
	return nil
}
