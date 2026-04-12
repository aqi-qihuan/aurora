package util

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileType 文件类型枚举
type FileType string

const (
	FileTypeImage  FileType = "image"
	FileTypeVideo  FileType = "video"
	FileTypeAudio  FileType = "audio"
	FileTypeDoc    FileType = "document"
	FileTypeOther  FileType = "other"
)

// FileMeta 文件元信息
type FileMeta struct {
	Filename     string   // 原始文件名
	Extension    string   // 扩展名
	ContentType  string   // MIME类型
	Size         int64    // 字节大小
	Type         FileType // 文件类型类别
	IsImage      bool     // 是否图片
	IsAllowedExt bool     // 是否允许的扩展名
}

// AnalyzeFile 分析上传文件的元信息
// 对标 Java 版 FileUtil + 文件类型判断逻辑
func AnalyzeFile(file *multipart.FileHeader) (*FileMeta, error) {
	f, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// 读取前512字节用于检测真实MIME类型（防止伪造扩展名）
	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	if n > 0 {
		buf = buf[:n]
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != "" {
		ext = ext[1:] // 去掉 "."
	}
	contentType := http.DetectContentType(buf)

	meta := &FileMeta{
	 Filename:     file.Filename,
	 Extension:     ext,
	 ContentType:   contentType,
	 Size:          file.Size,
	 Type:          classifyFileType(contentType, ext),
	 IsImage:       strings.HasPrefix(contentType, "image/"),
	 IsAllowedExt: isAllowedExtension(ext),
	}
	return meta, nil
}

// GetMd5 计算文件输入流的MD5值 (对标 Java FileUtil.getMd5)
func GetMd5(reader io.Reader) (string, error) {
	hash := md5.New()
	buf := make([]byte, 8192)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			hash.Write(buf[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("计算MD5失败: %w", err)
		}
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// GetExtName 获取文件扩展名 (对标 Java FileUtil.getExtName)
func GetExtName(fileName string) string {
	if fileName == "" {
		return ""
	}
	return filepath.Ext(fileName)
}

// classifyFileType 根据MIME类型和扩展名分类
func classifyFileType(contentType, ext string) FileType {
	switch {
	case strings.HasPrefix(contentType, "image/"):
		return FileTypeImage
	case strings.HasPrefix(contentType, "video/"):
		return FileTypeVideo
	case strings.HasPrefix(contentType, "audio/"):
		return FileTypeAudio
	case containsString(AllowedDocExtensions, ext):
		return FileTypeDoc
	default:
		return FileTypeOther
	}
}

// isAllowedExtension 检查是否为允许的上传扩展名
func isAllowedExtension(ext string) bool {
	return containsString(AllAllowedExtensions, ext)
}

// GenerateObjectName 生成存储对象名（含日期路径）
// 格式: {year}/{month}/{day}/{uuid}.{ext}
// 对标 Java 版 MinIOUtils.generateObjectName()
func GenerateObjectName(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	uuid := generateUUID()
	now := time.Now()
	return fmt.Sprintf("%04d/%02d/%02d/%s%s",
		now.Year(), now.Month(), now.Day(), uuid, ext)
}

// SaveTempFile 将 multipart 文件保存到临时目录（用于本地处理后再上传MinIO）
func SaveTempFile(file *multipart.FileHeader, tempDir string) (string, error) {
	os.MkdirAll(tempDir, 0755)
	tempPath := filepath.Join(tempDir, file.Filename)

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	dst, err := os.Create(tempPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		os.Remove(tempPath)
		return "", err
	}
	return tempPath, nil
}

// RemoveTempFile 清理临时文件
func RemoveTempFile(path string) error {
	if path == "" {
		return nil
	}
	return os.Remove(path)
}

// GetFileSize 获取文件大小的人类可读格式
func GetFileSize(bytes int64) string {
	const unit = 1024
	switch {
	case bytes < unit:
		return fmt.Sprintf("%dB", bytes)
	case bytes < unit*unit:
		return fmt.Sprintf("%.1fKB", float64(bytes)/float64(unit))
	case bytes < unit*unit*unit:
		return fmt.Sprintf("%.1fMB", float64(bytes)/float64(unit*unit))
	default:
		return fmt.Sprintf("%.1fGB", float64(bytes)/float64(unit*unit*unit))
	}
}

// ==================== 允许的文件扩展名白名单 ====================
// 对标 Java 版 FileUploadConfig.allowedExtensions

var (
	// AllowedImageExtensions 允许的图片格式
	AllowedImageExtensions = []string{"jpg", "jpeg", "png", "gif", "webp", "bmp", "svg", "ico"}

	// AllowedDocExtensions 允许的文档格式
	AllowedDocExtensions = []string{"pdf", "doc", "docx", "xls", "xlsx", "ppt", "pptx", "txt", "md"}

	// AllAllowedExtensions 全部允许的上传扩展名
	AllAllowedExtensions = mergeExtensions(AllowedImageExtensions, AllowedDocExtensions,
		[]string{"mp4", "mp3", "avi", "mov", "zip", "rar"},
	)
)

func mergeExtensions(slices ...[]string) []string {
	var result []string
	for _, s := range slices {
		result = append(result, s...)
	}
	return result
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// generateUUID 简易UUID生成器
func generateUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
