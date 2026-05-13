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

// @Summary 获取网站前台配置
// @Tags 网站配置
// @Description 获取网站前台公开配置
// @Accept json
// @Produce json
// @Success 200 {object} object "成功返回网站配置"
// @Failure 500 {object} object "服务器内部错误"
// @Router /api/admin/website/config [get]
func (h *WebsiteConfigHandler) GetWebsiteConfig(c *gin.Context) {
	// 获取首页聚合数据（包含统计数据和网站配置）
	info, err := h.registry.AuroraInfo.GetHomeInfo(c.Request.Context())
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	
	// 构建前端期望的数据结构
	result := map[string]interface{}{
		"viewCount":         info.ViewCount,
		"articleCount":      info.ArticleCount,
		"categoryCount":     info.CategoryCount,
		"tagCount":          info.TagCount,
		"talkCount":         info.TalkCount,
		"websiteConfigDTO":  info.WebsiteConfig,
	}
	
	util.ResponseSuccess(c, result)
}

// @Summary 更新网站配置
// @Tags 网站配置
// @Description 后台更新网站配置
// @Accept json
// @Produce json
// @Param configVO body vo.WebsiteConfigVO true "网站配置信息"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/website/config [put]
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

// @Summary 上传网站图片
// @Tags 网站配置
// @Description 上传网站配置图片（Logo/Favicon/头像等）
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "图片文件"
// @Param type formData string false "图片类型"
// @Success 200 {object} object "成功返回图片URL"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/website/config/images [post]
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
	if err := h.registry.WebsiteConfig.UploadConfigImage(c.Request.Context(), imgType, result.URL); err != nil {
		util.ResponseError(c, err)
		return
	}

	util.ResponseSuccess(c, map[string]interface{}{
		"url": result.URL,
	})
}

// ensure dto import is used
var _ dto.ConditionVO
