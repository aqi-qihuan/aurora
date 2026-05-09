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

// @Summary 获取分类列表
// @Tags 分类
// @Description 获取前台分类列表（含文章数量）
// @Accept json
// @Produce json
// @Success 200 {object} object "成功返回分类列表"
// @Failure 500 {object} object "服务器内部错误"
// @Router /api/categories/all [get]
func (h *CategoryHandler) ListCategories(c *gin.Context) {
	list, err := h.svc.GetCategories(c.Request.Context())
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, list)
}

// @Summary 获取分类详情
// @Tags 分类
// @Description 根据ID获取分类详情
// @Accept json
// @Produce json
// @Param id path int true "分类ID"
// @Success 200 {object} object "成功返回分类详情"
// @Failure 400 {object} object "请求参数错误"
// @Failure 500 {object} object "服务器内部错误"
// @Router /api/categories/{id} [get]
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

// @Summary 获取分类下拉选项
// @Tags 分类
// @Description 后台获取分类下拉选项列表
// @Accept json
// @Produce json
// @Success 200 {object} object "成功返回分类选项列表"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/categories/options [get]
func (h *CategoryHandler) ListCategoriesOption(c *gin.Context) {
	options, err := h.svc.GetCategoryOptions(c.Request.Context())
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, options)
}

// @Summary 保存或更新分类
// @Tags 分类
// @Description 后台保存或更新分类信息
// @Accept json
// @Produce json
// @Param categoryVO body vo.CategoryVO true "分类信息"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/categories [post]
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

// @Summary 批量删除分类
// @Tags 分类
// @Description 后台批量删除分类
// @Accept json
// @Produce json
// @Param ids body []uint true "分类ID列表"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/categories [delete]
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	// 从请求体接收ID数组（对标Java @RequestBody List<Integer>）
	var ids []uint
	if err := c.ShouldBindJSON(&ids); err != nil || len(ids) == 0 {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("请选择要删除的分类"))
		return
	}

	// 批量删除
	for _, id := range ids {
		if err := h.svc.DeleteCategory(c.Request.Context(), id); err != nil {
			util.ResponseError(c, err)
			return
		}
	}
	util.ResponseSuccess(c, "分类已删除")
}

// @Summary 后台分类管理列表
// @Tags 分类
// @Description 后台获取分类列表（分页，含筛选）
// @Accept json
// @Produce json
// @Param keyword query string false "搜索关键词"
// @Param current query int false "当前页码"
// @Param size query int false "每页数量"
// @Success 200 {object} object "成功返回分类列表"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/categories [get]
func (h *CategoryHandler) ListAdminCategories(c *gin.Context) {
	var condition dto.ConditionVO
	c.ShouldBindQuery(&condition)
	pageNum, pageSize := util.PageQuery(c)
	page := dto.PageVO{PageNum: pageNum, PageSize: pageSize}

	result, err := h.svc.ListAdminCategories(c.Request.Context(), condition, page)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// @Summary 搜索分类
// @Tags 分类
// @Description 搜索分类（用于编辑器下拉）
// @Accept json
// @Produce json
// @Param keywords query string false "搜索关键词"
// @Success 200 {object} object "成功返回分类列表"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/categories/search [get]
func (h *CategoryHandler) SearchCategories(c *gin.Context) {
	keyword := c.DefaultQuery("keywords", "")
	result, err := h.svc.SearchCategories(c.Request.Context(), keyword)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}
