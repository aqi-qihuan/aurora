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

// CategoryHandler 分类管理处理器（对标 Java CategoryController）
type CategoryHandler struct {
	svc *service.CategoryService
}

func NewCategoryHandler(svc *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

// ListCategories 获取分类列表（前台，含文章数量）
// GET /api/categories
func (h *CategoryHandler) ListCategories(c *gin.Context) {
	list, err := h.svc.GetCategories(c.Request.Context())
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, list)
}

// GetCategoryById 获取分类详情
// GET /api/categories/:id
func (h *CategoryHandler) GetCategoryById(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的分类ID"))
		return
	}
	result, err := h.svc.GetCategoryByID(c.Request.Context(), uint(id))
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// ListCategoriesOption 后台获取分类下拉选项
// GET /api/admin/categories/options
func (h *CategoryHandler) ListCategoriesOption(c *gin.Context) {
	options, err := h.svc.GetCategoryOptions(c.Request.Context())
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, options)
}

// SaveOrUpdate 保存或更新分类（后台管理）
// POST /api/admin/categories (新增)
// PUT /api/admin/categories/:id (更新)
func (h *CategoryHandler) SaveOrUpdate(c *gin.Context) {
	var categoryVO vo.CategoryVO
	if err := c.ShouldBindJSON(&categoryVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}

	idStr := c.Param("id")
	if idStr != "" {
		// 更新
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的分类ID"))
			return
		}
		if err := h.svc.UpdateCategory(c.Request.Context(), uint(id), categoryVO); err != nil {
			util.ResponseError(c, err)
			return
		}
		util.ResponseSuccess(c, nil)
		return
	}

	// 新增
	result, err := h.svc.CreateCategory(c.Request.Context(), categoryVO)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// DeleteCategory 删除分类（后台）
// DELETE /api/admin/categories/:id
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的分类ID"))
		return
	}
	if err := h.svc.DeleteCategory(c.Request.Context(), uint(id)); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "分类已删除")
}

// ensure CategoryVO is used via vo package
var _ dto.CategoryVO
