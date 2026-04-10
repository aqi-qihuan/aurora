package model

import (
	"time"

	"gorm.io/gorm"
)

// WebsiteConfig 网站配置实体 (对应 t_website_config 表)
// 使用单行模式: ID=1 始终为唯一配置行
type WebsiteConfig struct {
	ID                   uint  `gorm:"primarykey" json:"id"`
	SiteName             string `gorm:"size:50" json:"siteName"`                     // 站点名称
	SiteURL              string `gorm:"size:255" json:"siteUrl"`                      // 站点URL
	AuthorName           string `gorm:"size:30" json:"authorName"`                   // 作者名称
	AuthorAvatar         string `gorm:"size:1024" json:"authorAvatar"`               // 作者头像
	Logo                 string `gorm:"size:1024" json:"logo"`                       // 站点Logo
	Favicon              string `gorm:"size:1024" json:"favicon"`                    // 站点Favicon
	SiteIntro            string `gorm:"size:500" json:"siteIntro"`                    // 站点简介
	Notice               string `gorm:"size:500" json:"notice"`                      // 公告信息
	FooterInfo           string `gorm:"size:500" json:"footerInfo"`                  // 页脚信息
	IcpNumber            string `gorm:"size:50" json:"icpNumber"`                   // ICP备案号
	BaiduPushURL         string `gorm:"size:255" json:"baiduPushUrl"`               // 百度SEO推送URL
	GAID                 string `gorm:"size:100" json:"gaId"`                       // Google Analytics ID
	WechatQRCode         string `gorm:"size:1024" json:"wechatQrcode"`               // 微信收款码
	AlipayQRCode         string `gorm:"size:1024" json:"alipayQrcode"`               // 支付宝收款码
	CommentNotifyEnabled int8  `gorm:"default:0" json:"commentNotifyEnabled"`        // 评论邮件通知开关(0关1开)
	RegisterEnabled      int8  `gorm:"default:0" json:"registerEnabled"`            // 注册功能开关(0关1开)
	RewardEnabled         int8  `gorm:"default:0" json:"rewardEnabled"`              // 打赏功能开关(0关1开)
	CreateTime           time.Time `json:"createTime"`
	UpdateTime           time.Time `json:"updateTime"`
}

func (WebsiteConfig) TableName() string { return "t_website_config" }

func (w *WebsiteConfig) BeforeCreate(tx *gorm.DB) error { now := time.Now(); w.CreateTime = now; w.UpdateTime = now; return nil }
func (w *WebsiteConfig) BeforeUpdate(tx *gorm.DB) error { w.UpdateTime = time.Now(); return nil }
