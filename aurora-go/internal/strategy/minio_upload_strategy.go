package strategy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/aurora-go/aurora/internal/config"
	"github.com/aurora-go/aurora/internal/infrastructure/storage"
)

// MinIOUploadStrategy MinIO对象存储上传策略 (对标Java MinioUploadStrategyImpl)
type MinIOUploadStrategy struct {
	client *minio.Client
	bucket string
	url    string
}

// NewMinIOUploadStrategy 创建MinIO上传策略实例
func NewMinIOUploadStrategy(cfg config.MinIOConfig) (*MinIOUploadStrategy, error) {
	endpoint := cfg.Endpoint
	useSSL := cfg.UseSSL

	if strings.HasPrefix(endpoint, "https://") {
		useSSL = true
		endpoint = strings.TrimPrefix(endpoint, "https://")
	} else if strings.HasPrefix(endpoint, "http://") {
		endpoint = strings.TrimPrefix(endpoint, "http://")
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	return &MinIOUploadStrategy{
		client: client,
		bucket: cfg.Bucket,
		url:    strings.TrimSuffix(cfg.Endpoint, "/"),
	}, nil
}

// Exists 检查对象是否存在
func (s *MinIOUploadStrategy) Exists(ctx context.Context, objectName string) (bool, error) {
	_, err := s.client.StatObject(ctx, s.bucket, objectName, minio.StatObjectOptions{})
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" || errResp.StatusCode == 404 {
			return false, nil
		}
		slog.Warn("MinIO stat object error", "bucket", s.bucket, "object", objectName, "error", err)
		return false, err
	}
	return true, nil
}

// Upload 上传文件到MinIO (实现 UploadStrategy 接口)
func (s *MinIOUploadStrategy) Upload(ctx context.Context, path, fileName string, reader interface{}) error {
	var data []byte
	switch v := reader.(type) {
	case []byte:
		data = v
	case io.Reader:
		var err error
		data, err = io.ReadAll(v)
		if err != nil {
			return fmt.Errorf("failed to read upload data: %w", err)
		}
	default:
		return fmt.Errorf("expected []byte or io.Reader data for MinIO upload")
	}

	_, err := s.client.PutObject(ctx, s.bucket, path+fileName,
		bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
			ContentType: guessContentType(fileName),
		})

	if err != nil {
		slog.Error("MinIO upload failed",
			"bucket", s.bucket,
			"path", path,
			"file", fileName,
			"error", err,
		)
		return fmt.Errorf("minio put object failed: %w", err)
	}
	return nil
}

// GetFileAccessUrl 获取文件的HTTP访问URL
func (s *MinIOUploadStrategy) GetFileAccessUrl(objectName string) string {
	prefix := storage.GetExternalURL()
	if prefix == "" {
		prefix = s.url
	}
	return strings.TrimSuffix(prefix, "/") + "/" + objectName
}

// guessContentType 根据文件扩展名猜测MIME类型
func guessContentType(filename string) string {
	idx := strings.LastIndex(filename, ".")
	ext := ""
	if idx >= 0 {
		ext = strings.ToLower(filename[idx+1:])
	}
	mimeTypes := map[string]string{
		"jpg":  "image/jpeg",
		"jpeg": "image/jpeg",
		"png":  "image/png",
		"gif":  "image/gif",
		"webp": "image/webp",
		"svg":  "image/svg+xml",
		"mp4":  "video/mp4",
		"mp3":  "audio/mpeg",
		"pdf":  "application/pdf",
		"md":   "text/markdown",
		"txt":  "text/plain",
		"html": "text/html",
		"css":  "text/css",
		"js":   "application/javascript",
	}
	if ct, ok := mimeTypes[ext]; ok {
		return ct
	}
	return "application/octet-stream"
}
