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

// PhotoService 照片业务逻辑 (对标 Java PhotoServiceImpl)
type PhotoService struct {
	db *gorm.DB
}

func NewPhotoService(db *gorm.DB) *PhotoService {
	return &PhotoService{db: db}
}

// UploadPhotos 批量上传照片
func (s *PhotoService) UploadPhotos(ctx context.Context, albumID uint, urls []string) ([]model.Photo, error) {
	if len(urls) == 0 {
		return nil, fmt.Errorf("照片URL列表不能为空")
	}

	var photos []model.Photo
	for _, url := range urls {
		photo := model.Photo{
			AlbumID: albumID,
			URL:     url,
		}
		photos = append(photos, photo)
	}

	if err := s.db.WithContext(ctx).Create(&photos).Error; err != nil {
		return nil, fmt.Errorf("批量保存照片失败: %w", err)
	}

	// 更新相册照片数
	s.db.WithContext(ctx).Model(&model.PhotoAlbum{}).
		Where("id = ?", albumID).
		UpdateColumn("photo_count", gorm.Expr("photo_count + ?", len(urls)))

	slog.Info("批量上传照片完成", "album_id", albumID, "count", len(urls))
	return photos, nil
}

// DeletePhoto 删除照片
func (s *PhotoService) DeletePhoto(ctx context.Context, id uint) error {
	var photo model.Photo

	if err := s.db.WithContext(ctx).First(&photo, id).Error; err != nil {
		return errors.ErrPhotoNotFound
	}

	if err := s.db.WithContext(ctx).Delete(&photo).Error; err != nil {
		return fmt.Errorf("删除照片失败: %w", err)
	}

	// 减少相册计数
	s.db.WithContext(ctx).Model(&model.PhotoAlbum{}).
		Where("id = ?", photo.AlbumID).
		UpdateColumn("photo_count", gorm.Expr("GREATEST(photo_count - 1, 0)"))

	return nil
}

// GetPhotosByAlbum 获取相册下的照片列表
func (s *PhotoService) GetPhotosByAlbum(ctx context.Context, albumID uint) ([]dto.PhotoDTO, error) {
	var photos []model.Photo

	err := s.db.WithContext(ctx).
		Where("album_id = ?", albumID).
		Order("create_time ASC").
		Find(&photos).Error

	if err != nil {
		return nil, fmt.Errorf("查询相册照片失败: %w", err)
	}

	list := make([]dto.PhotoDTO, len(photos))
	for i, p := range photos {
		list[i] = dto.PhotoDTO{ID: p.ID, URL: p.URL}
	}
	return list, nil
}

// BatchDeletePhotos 批量删除照片
func (s *PhotoService) BatchDeletePhotos(ctx context.Context, ids []uint) error {
	result := s.db.WithContext(ctx).
		Delete(&model.Photo{}, "id IN ?", ids)

	if result.Error != nil {
		return fmt.Errorf("批量删除照片失败: %w", result.Error)
	}
	slog.Info("批量删除照片", "count", result.RowsAffected)
	return nil
}
