package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

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

// ListAdminPhotos 后台照片管理分页查询（对标Java listPhotos）
func (s *PhotoService) ListAdminPhotos(ctx context.Context, cond dto.ConditionVO, page dto.PageVO) (*dto.PageResultDTO, error) {
	var photos []model.Photo
	var count int64

	baseQuery := s.db.WithContext(ctx).Model(&model.Photo{})

	// 按相册ID过滤
	if cond.AlbumID > 0 {
		baseQuery = baseQuery.Where("album_id = ?", cond.AlbumID)
	}

	// 按删除状态过滤
	if cond.IsDelete != nil {
		baseQuery = baseQuery.Where("is_delete = ?", *cond.IsDelete)
	} else {
		baseQuery = baseQuery.Where("is_delete = 0") // 默认只显示未删除的
	}

	// 获取总数
	baseQuery.Count(&count)
	if count == 0 {
		return &dto.PageResultDTO{
			List:     []dto.PhotoAdminDTO{},
			Count:    0,
			PageNum:  page.PageNum,
			PageSize: page.PageSize,
		}, nil
	}

	offset := page.GetOffset()
	err := baseQuery.
		Order("id DESC, update_time DESC").
		Limit(page.PageSize).
		Offset(offset).
		Find(&photos).Error

	if err != nil {
		return nil, fmt.Errorf("查询照片列表失败: %w", err)
	}

	list := make([]dto.PhotoAdminDTO, len(photos))
	for i, p := range photos {
		list[i] = dto.PhotoAdminDTO{
			ID:        p.ID,
			AlbumID:   p.AlbumID,
			PhotoName: p.PhotoName,
			PhotoDesc: p.PhotoDesc,
			PhotoSrc:  p.PhotoSrc,
			IsDelete:  p.IsDelete,
		}
	}

	return &dto.PageResultDTO{
		List:     list,
		Count:    count,
		PageNum:  page.PageNum,
		PageSize: page.PageSize,
	}, nil
}

// UpdatePhoto 更新照片信息（对标Java updatePhoto）
func (s *PhotoService) UpdatePhoto(ctx context.Context, id uint, photoName string) error {
	var photo model.Photo
	if err := s.db.WithContext(ctx).First(&photo, id).Error; err != nil {
		return errors.ErrPhotoNotFound
	}

	result := s.db.WithContext(ctx).Model(&photo).Update("photo_name", photoName)
	if result.Error != nil {
		return fmt.Errorf("更新照片失败: %w", result.Error)
	}
	return nil
}

// SavePhotos 保存照片（对标Java savePhotos）
func (s *PhotoService) SavePhotos(ctx context.Context, albumID uint, photoURLs []string) error {
	if len(photoURLs) == 0 {
		return fmt.Errorf("照片URL列表不能为空")
	}

	// 检查相册是否存在且未删除
	var album model.PhotoAlbum
	if err := s.db.WithContext(ctx).Where("id = ? AND is_delete = 0", albumID).First(&album).Error; err != nil {
		return errors.ErrAlbumNotFound
	}

	var photos []model.Photo
	for _, url := range photoURLs {
		photos = append(photos, model.Photo{
			AlbumID:   albumID,
			PhotoName: generatePhotoID(), // Java用雪花算法生成ID
			PhotoSrc:  url,
			IsDelete:  0,
		})
	}

	return s.db.WithContext(ctx).Create(&photos).Error
}

// UpdatePhotosAlbum 移动照片到其他相册（对标Java updatePhotosAlbum）
func (s *PhotoService) UpdatePhotosAlbum(ctx context.Context, albumID uint, photoIDs []uint) error {
	if len(photoIDs) == 0 {
		return fmt.Errorf("请选择要移动的照片")
	}

	// 检查目标相册是否存在
	var album model.PhotoAlbum
	if err := s.db.WithContext(ctx).Where("id = ? AND is_delete = 0", albumID).First(&album).Error; err != nil {
		return errors.ErrAlbumNotFound
	}

	// 批量更新照片的相册ID
	result := s.db.WithContext(ctx).
		Model(&model.Photo{}).
		Where("id IN ? AND is_delete = 0", photoIDs).
		Update("album_id", albumID)

	if result.Error != nil {
		return fmt.Errorf("移动照片失败: %w", result.Error)
	}
	return nil
}

// UpdatePhotoDelete 更新照片删除状态（逻辑删除/恢复，对标Java updatePhotoDelete）
func (s *PhotoService) UpdatePhotoDelete(ctx context.Context, ids []uint, isDelete int8) error {
	if len(ids) == 0 {
		return fmt.Errorf("请选择要操作的照片")
	}

	// 批量更新照片删除状态
	result := s.db.WithContext(ctx).
		Model(&model.Photo{}).
		Where("id IN ?", ids).
		Update("is_delete", isDelete)

	if result.Error != nil {
		return fmt.Errorf("更新照片状态失败: %w", result.Error)
	}

	// 如果是恢复操作（isDelete=0），需要恢复相册的删除状态
	if isDelete == 0 {
		// 查询受影响照片所属的相册ID
		var albumIDs []uint
		s.db.WithContext(ctx).Model(&model.Photo{}).
			Select("album_id").
			Where("id IN ?", ids).
			Group("album_id").
			Find(&albumIDs)

		// 恢复这些相册的删除状态
		s.db.WithContext(ctx).
			Model(&model.PhotoAlbum{}).
			Where("id IN ?", albumIDs).
			Update("is_delete", 0)
	}
	return nil
}

// DeletePhotos 批量删除照片（物理删除，对标Java deletePhotos）
func (s *PhotoService) DeletePhotos(ctx context.Context, ids []uint) error {
	if len(ids) == 0 {
		return fmt.Errorf("请选择要删除的照片")
	}

	result := s.db.WithContext(ctx).Delete(&model.Photo{}, ids)
	if result.Error != nil {
		return fmt.Errorf("删除照片失败: %w", result.Error)
	}
	slog.Info("批量删除照片", "count", result.RowsAffected)
	return nil
}

// ListPhotosByAlbumId 根据相册ID查看照片列表（前台，对标Java listPhotosByAlbumId）
func (s *PhotoService) ListPhotosByAlbumId(ctx context.Context, albumID uint) (*dto.PhotoAlbumInfoDTO, error) {
	// 1. 检查相册是否存在、未删除、公开
	var album model.PhotoAlbum
	if err := s.db.WithContext(ctx).
		Where("id = ? AND is_delete = 0 AND status = 1", albumID).
		First(&album).Error; err != nil {
		return nil, errors.ErrAlbumNotFound
	}

	// 2. 获取照片URL列表（分页）
	var photoURLs []string
	s.db.WithContext(ctx).Model(&model.Photo{}).
		Select("photo_src").
		Where("album_id = ? AND is_delete = 0", albumID).
		Order("id DESC").
		Pluck("photo_src", &photoURLs)

	return &dto.PhotoAlbumInfoDTO{
		PhotoAlbumCover: album.AlbumCover,
		PhotoAlbumName:  album.AlbumName,
		Photos:          photoURLs,
	}, nil
}

// generatePhotoID 生成照片名（模拟Java雪花算法）
func generatePhotoID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
