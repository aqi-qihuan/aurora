package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
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

// GetConfig 获取网站配置 (单行表, ID=1)
func (s *WebsiteConfigService) GetConfig(ctx context.Context) (*dto.WebsiteConfigDTO, error) {
	var config model.WebsiteConfig

	err := s.db.WithContext(ctx).First(&config, 1).Error
	if err != nil && !errors.IsStd(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询网站配置失败: %w", err)
	}

	if config.ID == 0 {
		// 首次访问 → 返回默认配置
		return &dto.WebsiteConfigDTO{
			SiteName:    "Aurora Blog",
			SiteURL:     "https://example.com",
			AuthorName:  "Admin",
			IcpNumber:   "",
			BaiduPushURL: "",
			GAID:         "",
			CommentNotifyEnabled: false,
			RegisterEnabled:        false,
			RewardEnabled:          false,
		}, nil
	}

	return s.toConfigDTO(&config), nil
}

// UpdateConfig 更新网站配置
func (s *WebsiteConfigService) UpdateConfig(ctx context.Context, vo vo.WebsiteConfigVO) error {
	var config model.WebsiteConfig

	// 使用FirstOrCreate确保记录存在
	s.db.WithContext(ctx).FirstOrCreate(&config, model.WebsiteConfig{ID: 1})

	updates := map[string]interface{}{
		"site_name":             vo.SiteName,
		"site_url":              vo.SiteURL,
		"author_name":           vo.AuthorName,
		"author_avatar":         vo.AuthorAvatar,
		"logo":                  vo.Logo,
		"favicon":               vo.Favicon,
		"site_intro":            vo.SiteIntro,
		"notice":                vo.Notice,
		"footer_info":           vo.FooterInfo,
		"icp_number":            vo.IcpNumber,
		"baidu_push_url":        vo.BaiduPushURL,
		"ga_id":                 vo.GAID,
		"comment_notify_enabled": vo.CommentNotifyEnabled,
		"register_enabled":      vo.RegisterEnabled,
		"reward_enabled":        vo.RewardEnabled,
	}

	result := s.db.WithContext(ctx).Model(&config).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("更新网站配置失败: %w", result.Error)
	}
	
	slog.Info("网站配置已更新")
	return nil
}

// UploadConfigImages 上传配置图片 (头像/Logo/Favicon/打赏码)
type ConfigImageVO struct {
	Type string `form:"type" binding:"required"` // avatar/logo/favicon/wechat_pay/alipay
	URL  string `form:"url" binding:"required"`
}

func (s *WebsiteConfigService) UploadConfigImages(ctx context.Context, img ConfigImageVO) error {
	fieldMap := map[string]string{
		"avatar":       "author_avatar",
		"logo":         "logo",
		"favicon":      "favicon",
		"wechat_pay":   "wechat_qrcode",
		"alipay":       "alipay_qrcode",
	}

	field, ok := fieldMap[img.Type]
	if !ok {
		return fmt.Errorf("不支持的图片类型: %s", img.Type)
	}

	result := s.db.WithContext(ctx).
		Model(&model.WebsiteConfig{}).
		Where("id = 1").
		Update(field, img.URL)

	if result.Error != nil {
		return fmt.Errorf("上传配置图片失败: %w", result.Error)
	}
	return nil
}

// ===== DTO转换 =====

func (s *WebsiteConfigService) toConfigDTO(c *model.WebsiteConfig) *dto.WebsiteConfigDTO {
	return &dto.WebsiteConfigDTO{
		ID:                     c.ID,
		SiteName:               c.SiteName,
		SiteURL:                c.SiteURL,
		AuthorName:             c.AuthorName,
		AuthorAvatar:           c.AuthorAvatar,
		Logo:                   c.Logo,
		Favicon:                c.Favicon,
		SiteIntro:              c.SiteIntro,
		Notice:                 c.Notice,
		FooterInfo:             c.FooterInfo,
		IcpNumber:              c.IcpNumber,
		BaiduPushURL:           c.BaiduPushURL,
		GAID:                   c.GAID,
		WechatQRCode:           c.WechatQRCode,
		AlipayQRCode:           c.AlipayQRCode,
		CommentNotifyEnabled:   c.CommentNotifyEnabled == 1,
		RegisterEnabled:        c.RegisterEnabled == 1,
		RewardEnabled:          c.RewardEnabled == 1,
	}
}
