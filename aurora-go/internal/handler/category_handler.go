package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/util"
)

// CategoryHandler 分类管理处理器（对标 Java CategoryController）
type CategoryHandler struct {
	// categoryService service.CategoryService
}

func NewCategoryHandler() *CategoryHandler { return &CategoryHandler{} }

// ListCategories 获取分类列表（前台，含文章数量）
// GET /api/categories
func (h *CategoryHandler) ListCategories(c *gin.Context) {
	// TODO: P0-5 查询分类 + 每个分类的文章计数
	util.ResponseSuccess(c, []interface{}{})
}

// GetCategoryById 获取分类详情（含文章列表）
// GET /api/categories/:id
func (h *CategoryHandler) GetCategoryById(c *gin.Context) {
	idStr := c.Param("id")
	_, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的分类ID"))
		return
	}
	util.ResponseSuccess(c, nil)
}

// ListCategoriesOption 后台获取分类下拉选项
// GET /api/admin/categories/options (用于文章编辑器选择)
func (h *CategoryHandler) ListCategoriesOption(c *gin.Context) {
	// 返回精简格式: [{id:1,name:"Go",articleCount:12}, ...]
	util.ResponseSuccess(c, []interface{}{})
}

// SaveOrUpdate 保存或更新分类（后台管理）
// POST /api/admin/categories (新增)
// PUT /api/admin/categories/:id (更新)
func (h *CategoryHandler) SaveOrUpdate(c *gin.Context) {
	var categoryVO dto.CategoryVO
	if err := c.ShouldBindJSON(&categoryVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	_ = categoryVO

	zap.L().Debug("Save category", zap.Any("category", categoryVO))
	util.ResponseSuccess(c, nil)
}

// DeleteCategory 删除分类（后台）
// DELETE /api/admin/categories/:id
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的分类ID"))
		return
	}

	zap.L().Debug("Delete category", zap.Int64("category_id", id))
	util.ResponseSuccess(c, "分类已删除")
}
