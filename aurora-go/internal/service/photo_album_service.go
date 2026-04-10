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

// PhotoAlbumService 相册业务逻辑 (对标 Java AlbumServiceImpl)
type PhotoAlbumService struct {
	db *gorm.DB
}

func NewPhotoAlbumService(db *gorm.DB) *PhotoAlbumService {
	return &PhotoAlbumService{db: db}
}

// CreateAlbum 创建相册
func (s *PhotoAlbumService) CreateAlbum(ctx context.Context, vo vo.PhotoAlbumVO) (*model.PhotoAlbum, error) {
	album := model.PhotoAlbum{
		AlbumName:  vo.AlbumName,
		AlbumCover: vo.AlbumCover,
		Info:       vo.Info,
		Status:     vo.Status,
	}

	if err := s.db.WithContext(ctx).Create(&album).Error; err != nil {
		return nil, fmt.Errorf("创建相册失败: %w", err)
	}
	return &album, nil
}

// UpdateAlbum 更新相册
func (s *PhotoAlbumService) UpdateAlbum(ctx context.Context, id uint, vo vo.PhotoAlbumVO) error {
	var album model.PhotoAlbum
	if err := s.db.WithContext(ctx).First(&album, id).Error; err != nil {
		return errors.ErrAlbumNotFound
	}

	updates := map[string]interface{}{
		"album_name":  vo.AlbumName,
		"info":        vo.Info,
		"status":      vo.Status,
	}
	if vo.AlbumCover != "" {
		updates["album_cover"] = vo.AlbumCover
	}

	return s.db.WithContext(ctx).Model(&album).Updates(updates).Error
}

// DeleteAlbum 删除相册 (事务: 清除照片)
func (s *PhotoAlbumService) DeleteAlbum(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var album model.PhotoAlbum
		if err := tx.First(&album, id).Error; err != nil {
			return errors.ErrAlbumNotFound
		}

		// 删除相册下所有照片
		tx.Where("album_id = ?", id).Delete(&model.Photo{})

		if err := tx.Delete(&album).Error; err != nil {
			return fmt.Errorf("删除相册失败: %w", err)
		}
		return nil
	})
}

// GetAlbums 获取相册列表(前台用, 只显示公开相册)
func (s *PhotoAlbumService) GetAlbums(ctx context.Context) ([]dto.AlbumDTO, error) {
	var albums []model.PhotoAlbum

	err := s.db.WithContext(ctx).
		Where("status = 1").
		Order("create_time DESC").
		Find(&albums).Error

	if err != nil {
		return nil, fmt.Errorf("查询相册列表失败: %w", err)
	}

	list := make([]dto.AlbumDTO, len(albums))
	for i, a := range albums {
		list[i] = dto.AlbumDTO{
			ID:          a.ID,
			AlbumName:   a.AlbumName,
			AlbumCover:  a.AlbumCover,
			Info:         a.Info,
			PhotoCount:  a.PhotoCount,
			IsPrivate:   a.Status == 2,
			CreateTime:   a.CreateTime,
		}
	}
	return list, nil
}

// GetAdminAlbums 后台管理分页查询相册(含私密)
func (s *PhotoAlbumService) GetAdminAlbums(ctx context.Context, cond dto.ConditionVO, page dto.PageVO) (*dto.PageResultDTO, error) {
	var albums []model.PhotoAlbum
	var count int64

	baseQuery := s.db.WithContext(ctx).Model(&model.PhotoAlbum{})

	if cond.Keywords != "" {
		baseQuery = baseQuery.Where("album_name LIKE ?", "%"+cond.Keywords+"%")
	}

	baseQuery.Count(&count)

	offset := page.GetOffset()
	err := baseQuery.
		Order("create_time DESC").
		Limit(page.PageSize).
		Offset(offset).
		Find(&albums).Error

	if err != nil {
		return nil, fmt.Errorf("查询相册列表失败: %w", err)
	}

	list := make([]dto.AlbumDTO, len(albums))
	for i, a := range albums {
		list[i] = dto.AlbumDTO{
			ID:          a.ID,
			AlbumName:   a.AlbumName,
			AlbumCover:  a.AlbumCover,
			Info:         a.Info,
			PhotoCount:  a.PhotoCount,
			IsPrivate:   a.Status == 2,
			CreateTime:   a.CreateTime,
		}
	}

	return &dto.PageResultDTO{
		List:     list,
		Count:    count,
		PageNum:  page.PageNum,
		PageSize: page.PageSize,
	}, nil
}

// UploadAlbumCover 上传相册封面
func (s *PhotoAlbumService) UploadAlbumCover(ctx context.Context, albumID uint, coverURL string) error {
	result := s.db.WithContext(ctx).
		Model(&model.PhotoAlbum{}).
		Where("id = ?", albumID).
		Update("album_cover", coverURL)

	if result.Error != nil {
		return fmt.Errorf("更新相册封面失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.ErrAlbumNotFound
	}
	return nil
}

// SetAlbumPrivacy 设置相册隐私状态
func (s *PhotoAlbumService) SetAlbumPrivacy(ctx context.Context, albumID uint, isPrivate bool) error {
	status := int8(1) // 公开
	if isPrivate {
		status = 2 // 私密
	}

	result := s.db.WithContext(ctx).
		Model(&model.PhotoAlbum{}).
		Where("id = ?", albumID).
		Update("status", status)

	if result.Error != nil {
		return fmt.Errorf("设置相册隐私失败: %w", result.Error)
	}
	return nil
}
