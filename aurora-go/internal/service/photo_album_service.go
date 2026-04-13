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

// PhotoAlbumService 相册业务逻辑 (对标 Java PhotoAlbumServiceImpl)
type PhotoAlbumService struct {
	db *gorm.DB
}

func NewPhotoAlbumService(db *gorm.DB) *PhotoAlbumService {
	return &PhotoAlbumService{db: db}
}

// SaveOrUpdateAlbum 保存或更新相册（对标Java saveOrUpdatePhotoAlbum，根据ID判断新增/更新）
func (s *PhotoAlbumService) SaveOrUpdateAlbum(ctx context.Context, vo vo.PhotoAlbumVO) error {
	// 1. 检查相册名重复（排除当前相册）
	var existing model.PhotoAlbum
	err := s.db.WithContext(ctx).
		Select("id").
		Where("album_name = ? AND is_delete = 0", vo.AlbumName).
		First(&existing).Error

	if err == nil && existing.ID != vo.ID {
		return errors.ErrAlbumNameExists // 相册名已存在
	}

	// 2. 根据ID判断新增或更新
	if vo.ID > 0 {
		return s.UpdateAlbum(ctx, vo.ID, vo)
	}

	// 3. 新增
	_, err = s.CreateAlbum(ctx, vo)
	return err
}

// CreateAlbum 创建相册
func (s *PhotoAlbumService) CreateAlbum(ctx context.Context, vo vo.PhotoAlbumVO) (*model.PhotoAlbum, error) {
	album := model.PhotoAlbum{
		AlbumName:  vo.AlbumName,
		AlbumCover: vo.AlbumCover,
		AlbumDesc:  vo.Info,
		Status:     vo.Status,
		IsDelete:   0,
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
		"album_desc":  vo.Info,
		"status":      vo.Status,
	}
	if vo.AlbumCover != "" {
		updates["album_cover"] = vo.AlbumCover
	}

	return s.db.WithContext(ctx).Model(&album).Updates(updates).Error
}

// DeleteAlbum 删除相册（对标Java deletePhotoAlbumById）
// 有照片→软删除；无照片→硬删除
func (s *PhotoAlbumService) DeleteAlbum(ctx context.Context, id uint) error {
	// 1. 统计相册下照片数（含已删除的）
	var photoCount int64
	s.db.WithContext(ctx).Model(&model.Photo{}).Where("album_id = ?", id).Count(&photoCount)

	if photoCount > 0 {
		// 有照片：软删除相册 + 逻辑删除照片
		s.db.WithContext(ctx).Model(&model.PhotoAlbum{}).Where("id = ?", id).Update("is_delete", 1)
		s.db.WithContext(ctx).Model(&model.Photo{}).Where("album_id = ?", id).Update("is_delete", 1)
	} else {
		// 无照片：硬删除
		result := s.db.WithContext(ctx).Delete(&model.PhotoAlbum{}, id)
		if result.Error != nil {
			return fmt.Errorf("删除相册失败: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return errors.ErrAlbumNotFound
		}
	}
	return nil
}

// GetAlbums 获取相册列表（前台用，只显示公开+未删除的相册）
func (s *PhotoAlbumService) GetAlbums(ctx context.Context) ([]dto.AlbumDTO, error) {
	var albums []model.PhotoAlbum

	err := s.db.WithContext(ctx).
		Where("status = 1 AND is_delete = 0").
		Order("id DESC").
		Find(&albums).Error

	if err != nil {
		return nil, fmt.Errorf("查询相册列表失败: %w", err)
	}

	list := make([]dto.AlbumDTO, len(albums))
	for i, a := range albums {
		// 动态统计照片数
		var photoCount int64
		s.db.WithContext(ctx).Model(&model.Photo{}).
			Where("album_id = ? AND is_delete = 0", a.ID).
			Count(&photoCount)

		list[i] = dto.AlbumDTO{
			ID:         a.ID,
			AlbumName:  a.AlbumName,
			AlbumDesc:  a.AlbumDesc, // 前端用 albumDesc 显示
			AlbumCover: a.AlbumCover,
			Info:       a.AlbumDesc, // 兼容旧版
			PhotoCount: int(photoCount),
			Status:     a.Status,    // 前端用 status 计算公开/私密数量
			IsPrivate:  a.Status == 2,
			CreateTime: a.CreateTime,
		}
	}
	return list, nil
}

// GetAlbumInfos 获取后台相册列表信息（用于下拉选择/移动照片，对标Java listPhotoAlbumInfosAdmin）
// 返回所有未删除的相册（含私密），不含photoCount（对标Java PhotoAlbumDTO无此字段）
func (s *PhotoAlbumService) GetAlbumInfos(ctx context.Context) ([]dto.AlbumDTO, error) {
	var albums []model.PhotoAlbum

	err := s.db.WithContext(ctx).
		Where("is_delete = 0").
		Order("id DESC").
		Find(&albums).Error

	if err != nil {
		return nil, fmt.Errorf("查询相册列表失败: %w", err)
	}

	list := make([]dto.AlbumDTO, len(albums))
	for i, a := range albums {
		// 动态统计照片数
		var photoCount int64
		s.db.WithContext(ctx).Model(&model.Photo{}).
			Where("album_id = ? AND is_delete = 0", a.ID).
			Count(&photoCount)

		list[i] = dto.AlbumDTO{
			ID:         a.ID,
			AlbumName:  a.AlbumName,
			AlbumDesc:  a.AlbumDesc,
			AlbumCover: a.AlbumCover,
			PhotoCount: int(photoCount),
		}
	}
	return list, nil
}

// GetAdminAlbums 后台管理分页查询相册（含私密+软删除）
// 对标Java listPhotoAlbumsAdmin
func (s *PhotoAlbumService) GetAdminAlbums(ctx context.Context, cond dto.ConditionVO, page dto.PageVO) (*dto.PageResultDTO, error) {
	var albums []model.PhotoAlbum
	var count int64

	baseQuery := s.db.WithContext(ctx).Model(&model.PhotoAlbum{}).Where("is_delete = 0")

	if cond.Keywords != "" {
		baseQuery = baseQuery.Where("album_name LIKE ?", "%"+cond.Keywords+"%")
	}

	baseQuery.Count(&count)
	if count == 0 {
		return &dto.PageResultDTO{
			List:     []dto.AlbumDTO{},
			Count:    0,
			PageNum:  page.PageNum,
			PageSize: page.PageSize,
		}, nil
	}

	offset := page.GetOffset()
	err := baseQuery.
		Order("id DESC").
		Limit(page.PageSize).
		Offset(offset).
		Find(&albums).Error

	if err != nil {
		return nil, fmt.Errorf("查询相册列表失败: %w", err)
	}

	list := make([]dto.AlbumDTO, len(albums))
	for i, a := range albums {
		// 动态统计照片数
		var photoCount int64
		s.db.WithContext(ctx).Model(&model.Photo{}).
			Where("album_id = ? AND is_delete = 0", a.ID).
			Count(&photoCount)

		list[i] = dto.AlbumDTO{
			ID:         a.ID,
			AlbumName:  a.AlbumName,
			AlbumDesc:  a.AlbumDesc, // 前端用 albumDesc 显示
			AlbumCover: a.AlbumCover,
			Info:       a.AlbumDesc, // 兼容旧版
			PhotoCount: int(photoCount),
			Status:     a.Status,    // 前端用 status 计算公开/私密数量
			IsPrivate:  a.Status == 2,
			CreateTime: a.CreateTime,
		}
	}

	return &dto.PageResultDTO{
		List:     list,
		Count:    count,
		PageNum:  page.PageNum,
		PageSize: page.PageSize,
	}, nil
}

// GetAlbumById 根据ID获取相册详情（后台）
// 对标Java getPhotoAlbumByIdAdmin
func (s *PhotoAlbumService) GetAlbumById(ctx context.Context, id uint) (*dto.AlbumDTO, error) {
	var album model.PhotoAlbum
	if err := s.db.WithContext(ctx).Where("is_delete = 0").First(&album, id).Error; err != nil {
		return nil, errors.ErrAlbumNotFound
	}

	var photoCount int64
	s.db.WithContext(ctx).Model(&model.Photo{}).
		Where("album_id = ? AND is_delete = 0", id).
		Count(&photoCount)

	return &dto.AlbumDTO{
		ID:         album.ID,
		AlbumName:  album.AlbumName,
		AlbumDesc:  album.AlbumDesc, // 前端用 albumDesc 显示
		AlbumCover: album.AlbumCover,
		Info:       album.AlbumDesc, // 兼容旧版
		PhotoCount: int(photoCount),
		Status:     album.Status,    // 前端用 status 计算公开/私密数量
		IsPrivate:  album.Status == 2,
		CreateTime: album.CreateTime,
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
