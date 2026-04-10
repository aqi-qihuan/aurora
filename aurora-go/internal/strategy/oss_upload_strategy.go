package strategy

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

// OSSConfig OSS配置结构体（对标Java OssConfigProperties）
type OSSConfig struct {
	URL          string // 访问URL前缀 (如 https://bucket.oss-cn-hangzhou.aliyuncs.com/)
	Endpoint     string // OSS endpoint
	AccessKeyId     string
	AccessKeySecret string
	BucketName    string
}

// OSSUploadStrategy 阿里云OSS上传策略 (对标Java OssUploadStrategyImpl)
// 使用阿里云OSS SDK实现文件上传（当前为预留实现，SDK按需引入）
//
// Go增强点:
//  - 支持通过环境变量动态配置（无需硬编码）
//  - 错误日志更详细，便于排查问题
//  - 可通过 go mod 引入 aliyun-oss-go-sdk 启用
type OSSUploadStrategy struct {
	config OSSConfig
}

// NewOSSUploadStrategy 创建OSS上传策略实例
func NewOSSUploadStrategy(cfg OSSConfig) *OSSUploadStrategy {
	return &OSSUploadStrategy{config: cfg}
}

// Exists 检查对象是否存在 (对标Java OssUploadStrategyImpl.exists)
//
// Java: getOssClient().doesObjectExist(bucketName, filePath)
func (s *OSSUploadStrategy) Exists(ctx context.Context, objectName string) (bool, error) {
	// TODO: 引入 github.com/aliyun/aliyun-oss-go-sdk/oss 后实现
	// 示例代码:
	// client, _ := oss.New(s.config.Endpoint, s.config.AccessKeyId, s.config.AccessKeySecret)
	// bucket := client.Bucket(s.config.BucketName)
	// return bucket.IsObjectExist(objectName)

	slog.Warn("OSS Exists not yet implemented (requires aliyun-oss-go-sdk)",
		"bucket", s.config.BucketName,
		"object", objectName,
	)
	return false, fmt.Errorf("OSS upload strategy not fully implemented")
}

// Upload 上传文件到OSS (对标Java OssUploadStrategyImpl.upload)
//
// Java: getOssClient().putObject(bucketName, path + fileName, inputStream)
func (s *OSSUploadStrategy) Upload(ctx context.Context, path, fileName string, data any) error {
	// TODO: 引入 aliyun-oss-go-sdk 后实现
	// 示例代码:
	// reader := bytes.NewReader(data.([]byte))
	// err := bucket.PutObject(path+fileName, reader)

	slog.Warn("OSS Upload not yet implemented (requires aliyun-oss-go-sdk)",
		"bucket", s.config.BucketName,
		"path", path,
		"file", fileName,
	)
	return fmt.Errorf("OSS upload strategy not fully implemented")
}

// GetFileAccessUrl 获取文件的HTTP访问URL (对标Java OssUploadStrategyImpl.getFileAccessUrl)
//
// Java: return ossConfigProperties.getUrl() + filePath
func (s *OSSUploadStrategy) GetFileAccessUrl(objectName string) string {
	return strings.TrimSuffix(s.config.URL, "/") + "/" + objectName
}
