package storage

import (
	"bytes"
	"context"
	"log/slog"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/aurora-go/aurora/internal/config"
)

var MinIOClient *minio.Client
var minIOEndpoint string // 内部endpoint（用于SDK连接）
var externalURL   string // 外部访问端点（用于生成公开访问URL）

// InitMinIO 初始化 MinIO 客户端连接
func InitMinIO(cfg *config.MinIOConfig) {
	var err error
	useSSL := false

	endpoint := cfg.Endpoint
	if strings.HasPrefix(endpoint, "https://") {
		useSSL = true
		endpoint = strings.TrimPrefix(endpoint, "https://")
	} else if strings.HasPrefix(endpoint, "http://") {
		endpoint = strings.TrimPrefix(endpoint, "http://")
	}

	MinIOClient, err = minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		slog.Error("Failed to connect to MinIO", "error", err)
		panic("Failed to connect to MinIO: " + err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	bucketExists, err := MinIOClient.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		slog.Error("Failed to check bucket existence", "error", err)
		panic(err.Error())
	}

	if !bucketExists {
		err = MinIOClient.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{})
		if err != nil {
			slog.Error("Failed to create bucket", "error", err)
			panic(err.Error())
		}
		slog.Info("MinIO bucket created", "bucket", cfg.Bucket)
	}

	minIOEndpoint = endpoint
	// 对外访问URL默认与内部endpoint相同（可通过配置覆盖）
	externalURL = endpoint

	slog.Info("MinIO connected successfully",
		"endpoint", endpoint,
		"bucket", cfg.Bucket,
	)
}

// UploadFile 上传文件到 MinIO（从本地路径上传）
func UploadFile(ctx context.Context, objectName string, filePath string, contentType string, bucketName string) (string, error) {
	_, err := MinIOClient.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}
	return GetObjectURL(bucketName, objectName), nil
}

// UploadBytes 上传字节数据到 MinIO（内存中数据）
func UploadBytes(ctx context.Context, objectName string, data []byte, size int64, contentType string, bucketName string) (string, error) {
	reader := bytes.NewReader(data)
	_, err := MinIOClient.PutObject(ctx, bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}
	return GetObjectURL(bucketName, objectName), nil
}

// DeleteFile 从 MinIO 删除文件
func DeleteFile(ctx context.Context, objectName string, bucketName string) error {
	return MinIOClient.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
}

// GetObjectURL 获取文件的完整访问URL
func GetObjectURL(bucketName string, objectName string) string {
	u := url.URL{
		Scheme: "http",
		Host:   minIOEndpoint,
		Path:   filepath.Join(bucketName, objectName),
	}
	return u.String()
}

// GetPresignedURL 获取临时预签名URL（私有桶/鉴权场景，7天有效期）
func GetPresignedURL(ctx context.Context, bucketName string, objectName string, expiry time.Duration) (string, error) {
	presignedURL, err := MinIOClient.PresignedGetObject(ctx, bucketName, objectName, expiry, nil)
	if err != nil {
		return "", err
	}
	return presignedURL.String(), nil
}

// SetExternalURL 设置外部访问URL（用于CDN/域名代理场景）
func SetExternalURL(url string) {
	externalURL = url
}

// GetExternalURL 获取外部访问URL（用于生成公开访问链接）
func GetExternalURL() string {
	return externalURL
}
