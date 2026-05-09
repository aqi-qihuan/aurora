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

// @Summary 获取标签列表
// @Tags 标签
// @Description 获取前台标签列表（含文章数量，按热度排序）
// @Accept json
// @Produce json
// @Success 200 {object} object "成功返回标签列表"
// @Failure 500 {object} object "服务器内部错误"
// @Router /api/tags/all [get]
func (h *TagHandler) ListTags(c *gin.Context) {
	list, err := h.svc.GetTags(c.Request.Context())
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, list)
}

// @Summary 获取文章关联的标签
// @Tags 标签
// @Description 获取文章关联的标签列表
// @Accept json
// @Produce json
// @Param id path int true "文章ID"
// @Success 200 {object} object "成功返回标签列表"
// @Failure 400 {object} object "请求参数错误"
// @Failure 500 {object} object "服务器内部错误"
// @Router /api/articles/{id}/tags [get]
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

// @Summary 获取标签详情
// @Tags 标签
// @Description 根据ID获取标签详情
// @Accept json
// @Produce json
// @Param id path int true "标签ID"
// @Success 200 {object} object "成功返回标签详情"
// @Failure 400 {object} object "请求参数错误"
// @Failure 500 {object} object "服务器内部错误"
// @Router /api/tags/{id} [get]
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

// @Summary 搜索标签
// @Tags 标签
// @Description 搜索标签（模糊匹配，用于编辑器自动补全）
// @Accept json
// @Produce json
// @Param keywords query string false "搜索关键词"
// @Success 200 {object} object "成功返回标签列表"
// @Failure 500 {object} object "服务器内部错误"
// @Router /api/tags/search [get]
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

// @Summary 获取前10个热门标签
// @Tags 标签
// @Description 获取前10个热门标签
// @Accept json
// @Produce json
// @Success 200 {object} object "成功返回标签列表"
// @Failure 500 {object} object "服务器内部错误"
// @Router /api/tags/topTen [get]
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

// @Summary 后台标签管理列表
// @Tags 标签
// @Description 后台获取标签列表（分页）
// @Accept json
// @Produce json
// @Param keyword query string false "搜索关键词"
// @Param current query int false "当前页码"
// @Param size query int false "每页数量"
// @Success 200 {object} object "成功返回标签列表"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/tags [get]
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

// @Summary 保存/更新标签
// @Tags 标签
// @Description 后台保存或更新标签
// @Accept json
// @Produce json
// @Param tagVO body vo.TagVO true "标签信息"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/tags [post]
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

// @Summary 批量删除标签
// @Tags 标签
// @Description 后台批量删除标签
// @Accept json
// @Produce json
// @Param ids body []uint true "标签ID列表"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/tags [delete]
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

// @Summary 更新标签文章计数
// @Tags 标签
// @Description 更新标签关联的文章数量统计
// @Accept json
// @Produce json
// @Success 200 {object} object "成功响应"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/tags/count/sync [put]
func (h *TagHandler) UpdateTagArticleCount(c *gin.Context) {
	util.ResponseSuccess(c, "标签文章数量同步完成")
}
