package handler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/util"
)

// WebsiteConfigHandler 网站配置处理器（对标 Java WebsiteConfigController）
type WebsiteConfigHandler struct {
	// websiteConfigService service.WebsiteConfigService
}

func NewWebsiteConfigHandler() *WebsiteConfigHandler { return &WebsiteConfigHandler{} }

// GetWebsiteConfig 获取网站前台配置（公开）
// GET /api/website/config
// 对标 WebsiteConfigController.get() - 返回前台展示用配置
func (h *WebsiteConfigHandler) GetWebsiteConfig(c *gin.Context) {
	util.ResponseSuccess(c, map[string]interface{}{
		"siteName":        "",
		"siteUrl":         "",
		"authorAvatar":    "",
		"logo":            "",
		"favicon":         "",
		"socialLinks":     map[string]interface{}{},
		"isCommentNotice": false,
		"isEmailNotice":   false,
	})
}

// UpdateWebsiteConfig 更新网站配置（后台）
// PUT /api/admin/website/config
// 对标 WebsiteConfigController.update()
func (h *WebsiteConfigHandler) UpdateWebsiteConfig(c *gin.Context) {
	var configVO dto.WebsiteConfigVO
	if err := c.ShouldBindJSON(&configVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}

	zap.L().Debug("Update website config", zap.Any("config", configVO))
	_ = configVO // TODO: P0-5 保存配置 → 清除Redis缓存

	util.ResponseSuccess(c, "网站配置已更新")
}

// UploadConfigImage 上传网站图片(Logo/Favicon/头像等)
// POST /api/admin/website/config/images
func (h *WebsiteConfigHandler) UploadConfigImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("请选择要上传的文件"))
		return
	}
	defer file.Close()

	zap.L().Info("Upload config image", zap.String("filename", header.Filename))
	// TODO: P0-5 上传MinIO → 返回URL

	util.ResponseSuccess(c, map[string]interface{}{
		"url": "",
	})
}
