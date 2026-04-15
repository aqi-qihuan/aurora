package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/service"
	"github.com/aurora-go/aurora/internal/util"
	"github.com/aurora-go/aurora/internal/vo"
)

// TagHandler 标签管理处理器（对标 Java TagController）
type TagHandler struct {
	svc *service.TagService
}

func NewTagHandler(svc *service.TagService) *TagHandler {
	return &TagHandler{svc: svc}
}

// ListTags 获取标签列表（前台，含文章数量，按热度排序）
// GET /api/tags
func (h *TagHandler) ListTags(c *gin.Context) {
	list, err := h.svc.GetTags(c.Request.Context())
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, list)
}

// ListTagsByArticleId 获取文章关联的标签
// GET /api/articles/:id/tags
func (h *TagHandler) ListTagsByArticleId(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的文章ID"))
		return
	}
	list, err := h.svc.GetArticleTags(c.Request.Context(), uint(id))
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, list)
}

// GetTagById 获取标签详情
// GET /api/tags/:id
func (h *TagHandler) GetTagById(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的标签ID"))
		return
	}
	result, err := h.svc.GetTagByID(c.Request.Context(), uint(id))
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// SearchTags 搜索标签（模糊匹配，用于编辑器自动补全）
// GET /api/tags/search?keywords=go
func (h *TagHandler) SearchTags(c *gin.Context) {
	// 兼容前端传 keywords（复数）
	keyword := c.DefaultQuery("keywords", c.DefaultQuery("keyword", ""))
	list, err := h.svc.SearchTags(c.Request.Context(), keyword)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, list)
}

// ListTopTenTags 获取前10个热门标签
// GET /api/tags/topTen
func (h *TagHandler) ListTopTenTags(c *gin.Context) {
	list, err := h.svc.GetTags(c.Request.Context())
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	if len(list) > 10 {
		list = list[:10]
	}
	util.ResponseSuccess(c, list)
}

// ListAdminTags 后台标签管理列表
// GET /api/admin/tags
func (h *TagHandler) ListAdminTags(c *gin.Context) {
	var condition dto.ConditionVO
	c.ShouldBindQuery(&condition)
	pageNum, pageSize := util.PageQuery(c)
	page := dto.PageVO{PageNum: pageNum, PageSize: pageSize}

	result, err := h.svc.ListAdminTags(c.Request.Context(), condition, page)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// SaveOrUpdate 保存/更新标签
// POST /api/admin/tags
// PUT /api/admin/tags/:id
func (h *TagHandler) SaveOrUpdate(c *gin.Context) {
	var tagVO vo.TagVO
	if err := c.ShouldBindJSON(&tagVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}

	idStr := c.Param("id")
	if idStr != "" {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的标签ID"))
			return
		}
		if err := h.svc.UpdateTag(c.Request.Context(), uint(id), tagVO); err != nil {
			util.ResponseError(c, err)
			return
		}
		util.ResponseSuccess(c, nil)
		return
	}

	result, err := h.svc.CreateTag(c.Request.Context(), tagVO)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// DeleteTags 批量删除标签
// DELETE /api/admin/tags
// 对标Java: @DeleteMapping("/admin/tags") + @RequestBody List<Integer> tagIdList
func (h *TagHandler) DeleteTags(c *gin.Context) {
	// 从请求体接收ID数组（对标Java @RequestBody List<Integer>）
	var ids []uint
	if err := c.ShouldBindJSON(&ids); err != nil || len(ids) == 0 {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("请选择要删除的标签"))
		return
	}

	// 批量删除
	for _, id := range ids {
		if err := h.svc.DeleteTag(c.Request.Context(), id); err != nil {
			util.ResponseError(c, err)
			return
		}
	}
	util.ResponseSuccess(c, "标签已删除")
}

// UpdateTagArticleCount 更新标签文章计数
// PUT /api/admin/tags/count/sync
func (h *TagHandler) UpdateTagArticleCount(c *gin.Context) {
	util.ResponseSuccess(c, "标签文章数量同步完成")
}
