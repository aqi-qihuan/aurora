package strategy

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/aurora-go/aurora/internal/errors"
)

// UploadStrategy 文件上传策略接口 (对标Java com.aurora.strategy.UploadStrategy)
// 定义文件上传的统一行为契约，支持多种存储后端（MinIO/OSS/本地FS）
type UploadStrategy interface {
	// Exists 检查文件是否已存在（用于去重上传）
	Exists(ctx context.Context, objectName string) (bool, error)
	// Upload 上传文件到指定路径 (data 支持 []byte 或 io.Reader)
	Upload(ctx context.Context, path, fileName string, data any) error
	// GetFileAccessUrl 获取文件的完整访问URL
	GetFileAccessUrl(objectName string) string
}

// UploadConfig 上传策略通用配置
type UploadConfig struct {
	BucketName string // 存储桶名称
	URLPrefix  string // 访问URL前缀（如 https://cdn.example.com/）
}

// BaseUploadStrategy 上传策略抽象基类 (对标Java AbstractUploadStrategyImpl)
// 提供 MD5 去重、文件名生成等公共逻辑，具体存储实现由子类完成
type BaseUploadStrategy struct {
	config UploadConfig
}

// UploadFileFromStream 从 io.Reader 上传文件（带MD5去重）
//
// 对标Java AbstractUploadStrategyImpl.uploadFile(MultipartFile):
//   1. 计算文件MD5作为唯一标识
//   2. 提取文件扩展名
//   3. 生成去重文件名: MD5 + extName
//   4. 检查文件是否存在（避免重复上传）
//   5. 不存在则调用子类Upload方法
//   6. 返回完整访问URL
func (b *BaseUploadStrategy) UploadFileFromStream(
	ctx context.Context,
	strategy UploadStrategy,
	fileName string,
	reader io.Reader,
	path string,
) (string, error) {
	// Step 1: 读取全部数据计算MD5（对标Java FileUtil.getMd5(inputStream)）
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("%w: %v", errors.ErrFileUploadFailed, err)
	}
	defer func() { _ = reader.(io.Closer).Close() }()

	// Step 2: 计算MD5哈希（对标Java FileUtil.getMd5）
	hash := md5.Sum(data)
	md5Str := fmt.Sprintf("%x", hash)

	// Step 3: 提取扩展名（对标Java FileUtil.getExtName(originalFilename)）
	ext := filepath.Ext(fileName)

	// Step 4: 生成去重文件名
	objectName := md5Str + ext
	fullPath := strings.TrimSuffix(path, "/") + "/" + objectName

	// Step 5: 去重检查（对标Java exists(path + fileName)判断）
	exists, err := strategy.Exists(ctx, fullPath)
	if err != nil {
		slog.Warn("Check file existence failed, will re-upload", "path", fullPath, "error", err)
		// 继续执行上传，不因检查失败而中断
	}

	// Step 6: 上传或跳过（已存在则不重复上传）
	if !exists {
		if err := strategy.Upload(ctx, path, objectName, data); err != nil {
			return "", fmt.Errorf("%w: %v", errors.ErrFileUploadFailed, err)
		}
		slog.Info("File uploaded successfully", "objectName", objectName, "size", len(data))
	} else {
		slog.Info("File already exists, skip upload", "objectName", objectName)
	}

	// Step 7: 返回访问URL
	url := strategy.GetFileAccessUrl(fullPath)
	return url, nil
}

// Upload 上传字节数据（由子类实现的具体上传方法）
// 这是BaseUploadStrategy的默认实现，实际应被子类重写
func (b *BaseUploadStrategy) Upload(ctx context.Context, path, fileName string, data any) error {
	// 子类应重写此方法
	return fmt.Errorf("Upload method must be implemented by concrete strategy")
}
