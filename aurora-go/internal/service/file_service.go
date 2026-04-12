package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/model"
	"github.com/aurora-go/aurora/internal/util"
	"gorm.io/gorm"
)

// FileService 文件上传业务逻辑 (对标 Java FileServiceImpl + UploadStrategyContext)
// 策略模式: MinIO/OSS双上传策略 (P0-9实现)
type FileService struct {
	db *gorm.DB
}

// NewFileService 创建文件服务
func NewFileService(db *gorm.DB) *FileService {
	return &FileService{db: db}
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
		".mp3": true, ".wav": true, ".ogg": true, // 音频
		".pdf": true, ".doc": true, ".docx": true, // 文档
		".zip": true, ".rar": true, // 压缩包
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

// UploadArticleImage 上传文章图片/封面 (完全对标 Java 版 ArticleController 的文章图片上传)
// 对标Java: uploadStrategyContext.executeUploadStrategy(file, FilePathEnum.ARTICLE.getPath())
// FilePathEnum.ARTICLE = "aurora/articles/"
func (s *FileService) UploadArticleImage(ctx context.Context, file *multipart.FileHeader) (string, error) {
	if file == nil || file.Size == 0 {
		return "", fmt.Errorf("图片文件不能为空")
	}

	// 检查文件大小限制(10MB)
	const maxSize = 10 << 20 // 10MB
	if file.Size > maxSize {
		return "", fmt.Errorf("图片大小超过10MB限制")
	}

	// 验证图片扩展名白名单
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
		".svg": true,
	}
	if !allowedExts[ext] {
		return "", fmt.Errorf("不支持的图片类型: %s", ext)
	}

	// 生成文件名: 时间戳 + 随机字符串 + 扩展名
	randomStr, _ := util.GenerateRandomString(16)
	timestamp := time.Now().Unix()
	fileName := fmt.Sprintf("%d_%s%s", timestamp, randomStr, ext)

	// 对标Java FilePathEnum.ARTICLE = "aurora/articles/"
	articlePath := "aurora/articles/" + fileName

	// 上传文件
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer src.Close()

	if err := s.uploadFile(ctx, src, articlePath); err != nil {
		return "", fmt.Errorf("上传图片失败: %w", err)
	}

	// 获取完整URL (对标Java getFileAccessUrl)
	fullURL := s.getFileAccessURL(articlePath)

	slog.Info("文章图片上传成功", "url", fullURL)
	return fullURL, nil
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

// UploadAvatar 上传用户头像 (完全对标 Java 版 UserInfoServiceImpl.updateUserAvatar)
// 流程: 1)计算文件MD5去重 2)上传到对象存储 3)更新数据库avatar字段 4)返回完整URL
func (s *FileService) UploadAvatar(ctx context.Context, file *multipart.FileHeader, userID uint) (string, error) {
	if file == nil || file.Size == 0 {
		return "", fmt.Errorf("头像文件不能为空")
	}

	// 对标 Java FileUtil.getMd5(file.getInputStream()) - 计算文件MD5用于去重
	// 注意: Java 版会调用两次 getInputStream()，Go 版需要重新打开文件
	src1, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	md5Hash, err := util.GetMd5(src1)
	src1.Close()
	if err != nil {
		return "", fmt.Errorf("计算文件MD5失败: %w", err)
	}

	// 对标 Java FileUtil.getExtName() - 获取文件扩展名
	ext := util.GetExtName(file.Filename)

	// 对标 Java 版: 文件名 = MD5 + 扩展名 (实现去重)
	fileName := md5Hash + ext
	avatarPath := "aurora/avatar/" + fileName

	// 对标 Java AbstractUploadStrategyImpl.exists() - 检查文件是否已存在
	if !s.exists(avatarPath) {
		// 对标 Java upload(path, fileName, file.getInputStream()) - 重新打开文件流上传
		src2, err := file.Open()
		if err != nil {
			return "", fmt.Errorf("重新打开文件失败: %w", err)
		}
		defer src2.Close()

		if err := s.uploadFile(ctx, src2, avatarPath); err != nil {
			return "", fmt.Errorf("上传文件失败: %w", err)
		}
	}

	// 对标 Java getFileAccessUrl() - 返回完整URL
	fullURL := s.getFileAccessURL(avatarPath)

	// 对标 Java userInfoMapper.updateById(userInfo) - 更新数据库avatar字段
	if s.db != nil {
		if err := s.db.Model(&model.UserInfo{}).Where("id = ?", userID).Update("avatar", fullURL).Error; err != nil {
			return "", fmt.Errorf("更新用户头像失败: %w", err)
		}
	}

	slog.Info("用户头像上传成功", "userID", userID, "url", fullURL, "md5", md5Hash)
	return fullURL, nil
}

// exists 检查文件是否已存在 (对标 Java AbstractUploadStrategyImpl.exists)
func (s *FileService) exists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

// uploadFile 上传文件到本地存储 (对标 Java AbstractUploadStrategyImpl.upload)
func (s *FileService) uploadFile(ctx context.Context, src io.Reader, filePath string) error {
	// 创建目录
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 读取全部内容
	data, err := io.ReadAll(src)
	if err != nil {
		return fmt.Errorf("读取文件内容失败: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

// getFileAccessURL 获取文件访问URL (对标 Java MinioUploadStrategyImpl.getFileAccessUrl)
func (s *FileService) getFileAccessURL(filePath string) string {
	// 对标 Java: minioProperties.getUrl() + filePath
	// TODO: P0-9 接入MinIO/OSS后改为真实URL
	baseURL := "http://localhost:8080"
	return fmt.Sprintf("%s/%s", baseURL, filePath)
}

// 占位util引用 (P0-9后移除)
var _ = util.GenerateRandomString
var _ = time.Now
