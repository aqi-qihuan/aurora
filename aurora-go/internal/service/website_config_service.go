package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/infrastructure/database"
	"github.com/aurora-go/aurora/internal/model"
	"github.com/aurora-go/aurora/internal/vo"
	"gorm.io/gorm"
)

// WebsiteConfigService 网站配置业务逻辑 (对标 Java WebsiteConfigServiceImpl)
type WebsiteConfigService struct {
	db *gorm.DB
}

func NewWebsiteConfigService(db *gorm.DB) *WebsiteConfigService {
	return &WebsiteConfigService{db: db}
}

// GetConfig 获取网站配置 (对标Java: 从config JSON字段解析)
// Java: WebsiteConfigDTO websiteConfigDTO = JSON.parseObject(config.getConfig(), WebsiteConfigDTO.class);
func (s *WebsiteConfigService) GetConfig(ctx context.Context) (*dto.WebsiteConfigDTO, error) {
	var config model.WebsiteConfig

	err := s.db.WithContext(ctx).First(&config, 1).Error
	if err != nil && !errors.IsStd(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询网站配置失败: %w", err)
	}

	if config.ID == 0 || config.Config == "" {
		// 首次访问 → 返回空配置
		slog.Warn("网站配置记录不存在或为空，返回空配置", "id", config.ID)
		return &dto.WebsiteConfigDTO{}, nil
	}

	// 对标Java: JSON.parseObject(config.getConfig(), WebsiteConfigDTO.class)
	// 数据库存储的是JSON格式字符串，解析为DTO对象
	websiteConfigDTO := &dto.WebsiteConfigDTO{}
	if err := json.Unmarshal([]byte(config.Config), websiteConfigDTO); err != nil {
		preview := config.Config
		if len(preview) > 100 {
			preview = preview[:100]
		}
		slog.Warn("解析网站配置JSON失败，返回空配置", "error", err, "config_preview", preview)
		return &dto.WebsiteConfigDTO{}, nil
	}

	slog.Info("网站配置加载成功", "has_name", websiteConfigDTO.Name != "", "has_author", websiteConfigDTO.Author != "")
	return websiteConfigDTO, nil
}

// UpdateConfig 更新网站配置 (对标Java: 将VO直接转JSON存储到config字段)
// Java: websiteConfigMapper.updateById(WebsiteConfig.builder().id(1).config(JSON.toJSONString(websiteConfigVO)).build());
// Java: redisService.del(WEBSITE_CONFIG);  // 删除Redis缓存
func (s *WebsiteConfigService) UpdateConfig(ctx context.Context, configVO vo.WebsiteConfigVO) error {
	// 1. 直接将VO序列化为JSON（对标Java JSON.toJSONString(websiteConfigVO)）
	configJSON, err := json.Marshal(configVO)
	if err != nil {
		return fmt.Errorf("序列化网站配置失败: %w", err)
	}

	// 2. 保存到数据库（对标Java updateById）
	result := s.db.WithContext(ctx).
		Model(&model.WebsiteConfig{}).
		Where("id = ?", 1).
		Update("config", string(configJSON))

	if result.Error != nil {
		return fmt.Errorf("更新网站配置失败: %w", result.Error)
	}

	// 3. 删除Redis缓存（对标Java redisService.del(WEBSITE_CONFIG)）
	rdb := database.GetRedis()
	if rdb != nil {
		if err := rdb.Del(ctx, "website_config").Err(); err != nil {
			slog.Warn("删除网站配置Redis缓存失败", "error", err)
		}
		slog.Info("已删除网站配置Redis缓存")
	}

	slog.Info("网站配置已更新", "json_length", len(string(configJSON)))
	return nil
}

// UploadConfigImage 上传配置图片 (头像/Logo/Favicon/打赏码)
// 对标Java: 读取现有config JSON → 修改对应字段 → 写回JSON
func (s *WebsiteConfigService) UploadConfigImage(ctx context.Context, imgType string, url string) error {
	// 1. 读取现有配置
	var config model.WebsiteConfig
	err := s.db.WithContext(ctx).First(&config, 1).Error
	if err != nil && !errors.IsStd(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("查询网站配置失败: %w", err)
	}

	// 2. 解析JSON
	websiteConfigDTO := &dto.WebsiteConfigDTO{}
	if config.ID > 0 && config.Config != "" {
		if err := json.Unmarshal([]byte(config.Config), websiteConfigDTO); err != nil {
			return fmt.Errorf("解析网站配置JSON失败: %w", err)
		}
	}

	// 3. 根据类型更新对应字段
	switch imgType {
	case "avatar":
		websiteConfigDTO.AuthorAvatar = url
	case "logo":
		websiteConfigDTO.Logo = url
	case "favicon":
		websiteConfigDTO.Favicon = url
	case "wechat_pay":
		websiteConfigDTO.WeiXinQRCode = url
	case "alipay":
		websiteConfigDTO.AlipayQRCode = url
	case "userAvatar":
		websiteConfigDTO.UserAvatar = url
	case "touristAvatar":
		websiteConfigDTO.TouristAvatar = url
	default:
		return fmt.Errorf("不支持的图片类型: %s", imgType)
	}

	// 4. 序列化并保存
	configJSON, err := json.Marshal(websiteConfigDTO)
	if err != nil {
		return fmt.Errorf("序列化网站配置失败: %w", err)
	}

	// 使用FirstOrCreate确保记录存在
	s.db.WithContext(ctx).FirstOrCreate(&config, model.WebsiteConfig{ID: 1})

	result := s.db.WithContext(ctx).
		Model(&config).
		Update("config", string(configJSON))

	if result.Error != nil {
		return fmt.Errorf("上传配置图片失败: %w", result.Error)
	}

	slog.Info("配置图片已更新", "type", imgType, "url", url)
	return nil
}
