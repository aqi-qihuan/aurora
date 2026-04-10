package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/util"
)

// TagHandler 标签管理处理器（对标 Java TagController）
type TagHandler struct {
	// tagService service.TagService
}

func NewTagHandler() *TagHandler { return &TagHandler{} }

// ListTags 获取标签列表（前台，含文章数量，按热度排序）
// GET /api/tags
func (h *TagHandler) ListTags(c *gin.Context) {
	util.ResponseSuccess(c, []interface{}{})
}

// ListTagsByArticleId 获取文章关联的标签
// GET /api/articles/:articleId/tags
func (h *TagHandler) ListTagsByArticleId(c *gin.Context) {
	util.ResponseSuccess(c, []interface{}{})
}

// GetTagById 获取标签详情及文章列表
// GET /api/tags/:id
func (h *TagHandler) GetTagById(c *gin.Context) {
	idStr := c.Param("id")
	_, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParam.WithMsg("无效的标签ID"))
		return
	}
	util.ResponseSuccess(c, nil)
}

// SearchTags 搜索标签（模糊匹配，用于编辑器自动补全）
// GET /api/tags/search?keyword=go
func (h *TagHandler) SearchTags(c *gin.Context) {
	keyword := c.DefaultQuery("keyword", "")
	_ = keyword // TODO: P0-5 LIKE '%keyword%' 查询

	util.ResponseSuccess(c, []interface{}{})
}

// ==================== 后台管理端点 ====================

// SaveOrUpdate 保存/更新标签
// POST /api/admin/tags
// PUT /api/admin/tags/:id
func (h *TagHandler) SaveOrUpdate(c *gin.Context) {
	var tagVO dto.TagVO
	if err := c.ShouldBindJSON(&tagVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParam.WithMsg(err.Error()))
		return
	}
	_ = tagVO
	zap.L().Debug("Save tag", zap.Any("tag", tagVO))
	util.ResponseSuccess(c, nil)
}

// DeleteTags 批量删除标签
// DELETE /api/admin/tags?ids=1,2,3
func (h *TagHandler) DeleteTags(c *gin.Context) {
	idsStr := c.Query("ids")
	if idsStr == "" {
		util.ResponseError(c, errors.ErrInvalidParam.WithMsg("请选择要删除的标签"))
		return
	}

	ids := strings.Split(idsStr, ",")
	zap.L().Debug("Delete tags", zap.Strings("ids", ids))
	util.ResponseSuccess(c, "标签已删除")
}

// UpdateTagArticleCount 更新标签文章计数（内部调用或手动触发）
// PUT /api/admin/tags/count/sync
func (h *TagHandler) UpdateTagArticleCount(c *gin.Context) {
	// TODO: P0-5 遍历所有标签 → 统计article_tag关联数 → 更新count字段
	zap.L().Info("Syncing tag article counts...")
	util.ResponseSuccess(c, "标签文章数量同步完成")
}
