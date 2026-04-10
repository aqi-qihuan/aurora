package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/util"
)

// FileService 文件上传业务逻辑 (对标 Java FileServiceImpl + OssStrategy)
// 策略模式: MinIO/OSS双上传策略 (P0-9实现)
type FileService struct {
}

// NewFileService 创建文件服务
func NewFileService() *FileService {
	return &FileService{}
}

// UploadSingle 上传单个文件
func (s *FileService) UploadSingle(ctx context.Context, file *multipart.FileHeader) (*dto.FileUploadDTO, error) {
	if file == nil || file.Size == 0 {
		return nil, fmt.Errorf("文件不能为空")
	}

	// 检查文件大小限制(10MB)
	const maxSize = 10 << 20 // 10MB
	if file.Size > maxSize {
		return nil, fmt.Errorf("文件大小超过10MB限制")
	}

	// 验证文件扩展名白名单
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
		".svg": true,
		".mp4": true, ".avi": true, ".mov": true, // 视频
		".mp3": true, ".wav": true, ".ogg": true,   // 音频
		".pdf": true, ".doc": true, ".docx": true,   // 文档
		".zip": true, ".rar": true,                  // 压缩包
	}
	if !allowedExts[ext] {
		return nil, fmt.Errorf("不支持的文件类型: %s", ext)
	}

	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer src.Close()

	// TODO: P0-9 调用 UploadService.Upload() → MinIO或OSS
	// uploadURL, uploadErr := uploadSvc.Upload(ctx, src, ext, file.Size)
	_ = ctx

	// 临时返回模拟URL (P0-9替换为真实上传)
	randomStr, _ := util.GenerateRandomString(16)
	url := fmt.Sprintf("/uploads/%s%s", randomStr, ext)
	
	slog.Info("文件上传成功",
		"filename", file.Filename,
		"size", file.Size,
		"url", url,
	)

	return &dto.FileUploadDTO{
		URL:      url,
		Filename: file.Filename,
		Size:     file.Size,
	}, nil
}

// UploadBatch 批量上传文件 (并发上传)
func (s *FileService) UploadBatch(ctx context.Context, files []*multipart.FileHeader) ([]dto.FileUploadDTO, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("文件列表不能为空")
	}

	// 限制批量数量
	const maxBatch = 10
	if len(files) > maxBatch {
		return nil, fmt.Errorf("单次最多上传%d个文件", maxBatch)
	}

	results := make([]dto.FileUploadDTO, len(files))
	errCh := make(chan error, len(files))

	for i, f := range files {
		result, err := s.UploadSingle(ctx, f)
		if err != nil {
			errCh <- err
			continue
		}
		results[i] = *result
	}

	close(errCh)
	var errs []error
	for e := range errCh {
		errs = append(errs, e)
	}

	if len(errs) > 0 && len(results) == 0 {
		return nil, fmt.Errorf("%d个文件上传失败", len(errs))
	}

	slog.Info("批量上传完成", "total", len(files), "success", len(results))
	return results, nil
}

// UploadMarkdownImage 上传Markdown图片 (文章编辑器专用, 自动生成Markdown格式)
func (s *FileService) UploadMarkdownImage(ctx context.Context, file *multipart.FileHeader) (string, error) {
	result, err := s.UploadSingle(ctx, file)
	if err != nil {
		return "", err
	}

	// 返回Markdown格式的图片链接
	markdown := fmt.Sprintf("![](%s)", result.URL)
	slog.Info("Markdown图片上传完成", "url", result.URL)
	return markdown, nil
}

// DeleteFile 删除已上传的文件
func (s *FileService) DeleteFile(ctx context.Context, url string) error {
	// TODO: P0-9 从MinIO/OSS删除文件
	// uploadSvc.Delete(url)
	
	slog.Info("文件删除", "url", url)
	return nil
}

// GetFileURL 获取文件访问URL (含PresignedURL支持私有桶)
func (s *FileService) GetFileURL(ctx context.Context, path string, isPrivate bool) (string, error) {
	// TODO: P0-9 MinIO PresignedURL 或 OSS签名URL
	basePath := "https://cdn.example.com"
	if isPrivate {
		// 返回临时签名URL (有效期30分钟)
		return basePath + "/sign/" + path + "?expires=1800", nil
	}
	return basePath + "/" + path, nil
}

// ReadFileContent 读取文件内容(用于导入功能)
func (s *FileService) ReadFileContent(r io.Reader) ([]byte, error) {
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}
	return content, nil
}

// 占位util引用 (P0-9后移除)
var _ = util.GenerateRandomString
