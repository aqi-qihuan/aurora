package strategy

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/aurora-go/aurora/internal/config"
	"github.com/aurora-go/aurora/internal/infrastructure/storage"
)

// UploadService 文件上传服务（对标Java AbstractUploadStrategyImpl + UploadStrategyContext）
// 核心流程：计算MD5 → 检查是否已存在（去重） → 上传MinIO → 返回访问URL
type UploadService struct {
	minioStrategy *MinIOUploadStrategy
}

// NewUploadService 创建上传服务实例
func NewUploadService(cfg config.MinIOConfig) (*UploadService, error) {
	strategy, err := NewMinIOUploadStrategy(cfg)
	if err != nil {
		return nil, fmt.Errorf("初始化MinIO上传策略失败: %w", err)
	}

	// 设置外部访问URL
	if cfg.URL != "" {
		storage.SetExternalURL(strings.TrimSuffix(cfg.URL, "/"))
	}

	return &UploadService{
		minioStrategy: strategy,
	}, nil
}

// UploadAlbumCover 上传相册封面（对标Java savePhotoAlbumCover）
// 存储路径: /photos/albums/{md5}.{ext}
func (s *UploadService) UploadAlbumCover(ctx context.Context, data []byte, originalFilename string) (string, error) {
	return s.uploadFile(ctx, data, originalFilename, "/photos/albums/")
}

// UploadPhoto 上传照片（对标Java PhotoController.upload）
// 存储路径: /photos/{md5}.{ext}
func (s *UploadService) UploadPhoto(ctx context.Context, data []byte, originalFilename string) (string, error) {
	return s.uploadFile(ctx, data, originalFilename, "/photos/")
}

// UploadTalkImage 上传说说图片
// 存储路径: /talks/{md5}.{ext}
func (s *UploadService) UploadTalkImage(ctx context.Context, data []byte, originalFilename string) (string, error) {
	return s.uploadFile(ctx, data, originalFilename, "/talks/")
}

// uploadFile 核心上传逻辑（对标Java AbstractUploadStrategyImpl.uploadFile）
func (s *UploadService) uploadFile(ctx context.Context, data []byte, originalFilename string, path string) (string, error) {
	// 1. 计算文件MD5（用于去重，对标Java FileUtil.getMd5）
	md5Hash := md5.New()
	md5Hash.Write(data)
	md5 := hex.EncodeToString(md5Hash.Sum(nil))

	// 2. 获取文件扩展名（对标Java FileUtil.getExtName）
	ext := filepath.Ext(originalFilename)
	fileName := md5 + ext

	// 3. 构建完整的对象路径
	objectPath := path + fileName

	// 4. 检查文件是否已存在（MD5去重，对标Java exists()）
	exists, err := s.minioStrategy.Exists(ctx, objectPath)
	if err != nil {
		slog.Warn("MinIO检查文件存在性失败，将尝试上传", "path", objectPath, "error", err)
		// 检查失败不阻断上传，继续执行
	}

	if exists {
		// 文件已存在，直接返回URL（去重命中，无需重复上传）
		slog.Info("文件已存在（MD5去重命中）", "path", objectPath)
		return s.minioStrategy.GetFileAccessUrl(objectPath), nil
	}

	// 5. 上传到MinIO（对标Java upload()）
	err = s.minioStrategy.Upload(ctx, path, fileName, bytes.NewReader(data))
	if err != nil {
		slog.Error("MinIO上传失败", "path", objectPath, "error", err)
		return "", fmt.Errorf("文件上传失败: %w", err)
	}

	slog.Info("文件上传成功", "path", objectPath, "size", len(data))

	// 6. 返回访问URL（对标Java getFileAccessUrl）
	return s.minioStrategy.GetFileAccessUrl(objectPath), nil
}

// UploadFromReader 从io.Reader上传（支持流式上传，对标Java uploadFile(fileName, inputStream, path)）
func (s *UploadService) UploadFromReader(ctx context.Context, fileName string, reader io.Reader, path string) (string, error) {
	// 读取数据到内存
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("读取文件流失败: %w", err)
	}

	// 计算MD5
	md5Hash := md5.New()
	md5Hash.Write(data)
	md5 := hex.EncodeToString(md5Hash.Sum(nil))

	// 获取扩展名
	ext := filepath.Ext(fileName)
	finalFileName := md5 + ext
	objectPath := path + finalFileName

	// 检查是否存在
	exists, err := s.minioStrategy.Exists(ctx, objectPath)
	if err == nil && exists {
		return s.minioStrategy.GetFileAccessUrl(objectPath), nil
	}

	// 上传
	err = s.minioStrategy.Upload(ctx, path, finalFileName, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("文件上传失败: %w", err)
	}

	return s.minioStrategy.GetFileAccessUrl(objectPath), nil
}

// GetMinioStrategy 获取MinIO上传策略（供其他Handler直接使用）
func (s *UploadService) GetMinioStrategy() *MinIOUploadStrategy {
	return s.minioStrategy
}
