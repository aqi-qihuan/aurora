package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/service"
	"github.com/aurora-go/aurora/internal/util"
	"github.com/aurora-go/aurora/internal/vo"
)

// WebsiteConfigHandler 网站配置处理器（对标 Java WebsiteConfigController）
type WebsiteConfigHandler struct {
	registry *service.Registry
}

func NewWebsiteConfigHandler(registry *service.Registry) *WebsiteConfigHandler {
	return &WebsiteConfigHandler{registry: registry}
}

// GetWebsiteConfig 获取网站前台配置（公开）
// GET /api/website/config
func (h *WebsiteConfigHandler) GetWebsiteConfig(c *gin.Context) {
	config, err := h.registry.WebsiteConfig.GetConfig(c.Request.Context())
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, config)
}

// UpdateWebsiteConfig 更新网站配置（后台）
// PUT /api/admin/website/config
func (h *WebsiteConfigHandler) UpdateWebsiteConfig(c *gin.Context) {
	var configVO vo.WebsiteConfigVO
	if err := c.ShouldBindJSON(&configVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	if err := h.registry.WebsiteConfig.UpdateConfig(c.Request.Context(), configVO); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "网站配置已更新")
}

// UploadConfigImage 上传网站图片(Logo/Favicon/头像等)
// POST /api/admin/website/config/images
func (h *WebsiteConfigHandler) UploadConfigImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("请选择要上传的文件"))
		return
	}

	// 使用Registry中的FileService上传
	result, err := h.registry.File.UploadSingle(c.Request.Context(), file)
	if err != nil {
		util.ResponseError(c, err)
		return
	}

	// 更新配置中的图片URL
	imgType := c.DefaultPostForm("type", "avatar")
	imgVO := service.ConfigImageVO{
		Type: imgType,
		URL:  result.URL,
	}
	if err := h.registry.WebsiteConfig.UploadConfigImages(c.Request.Context(), imgVO); err != nil {
		util.ResponseError(c, err)
		return
	}

	util.ResponseSuccess(c, map[string]interface{}{
		"url": result.URL,
	})
}

// ensure dto import is used
var _ dto.ConditionVO
