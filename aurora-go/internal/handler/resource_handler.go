package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/service"
	"github.com/aurora-go/aurora/internal/util"
	"github.com/aurora-go/aurora/internal/vo"
)

// ResourceHandler 资源权限管理处理器（对标 Java ResourceController）
type ResourceHandler struct {
	svc *service.ResourceService
}

func NewResourceHandler(svc *service.ResourceService) *ResourceHandler {
	return &ResourceHandler{svc: svc}
}

// @Summary 获取资源权限列表
// @Tags 资源
// @Description 获取资源权限列表（树形结构）
// @Accept json
// @Produce json
// @Param keyword query string false "搜索关键词"
// @Success 200 {object} object "成功返回资源树"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/resources [get]
func (h *ResourceHandler) ListResources(c *gin.Context) {
	var condition dto.ConditionVO
	c.ShouldBindQuery(&condition)

	result, err := h.svc.ListResourcesTree(c.Request.Context(), condition)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// @Summary 保存/更新资源
// @Tags 资源
// @Description 后台保存或更新资源权限
// @Accept json
// @Produce json
// @Param resourceVO body vo.ResourceVO true "资源信息"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/resources [post]
func (h *ResourceHandler) SaveOrUpdate(c *gin.Context) {
	var resourceVO vo.ResourceVO
	if err := c.ShouldBindJSON(&resourceVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}

	// 根据ID判断新增或更新（对标Java版 MyBatis-Plus saveOrUpdate）
	if resourceVO.ID > 0 {
		// 更新
		if err := h.svc.UpdateResource(c.Request.Context(), resourceVO.ID, resourceVO); err != nil {
			util.ResponseError(c, err)
			return
		}
		util.ResponseSuccess(c, nil)
		return
	}

	// 新增
	result, err := h.svc.CreateResource(c.Request.Context(), resourceVO)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// @Summary 批量删除资源
// @Tags 资源
// @Description 批量删除资源权限
// @Accept json
// @Produce json
// @Param ids query string true "资源ID列表，逗号分隔"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/resources [delete]
func (h *ResourceHandler) DeleteResources(c *gin.Context) {
	idsStr := c.Query("ids")
	if idsStr == "" {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("请选择要删除的资源"))
		return
	}
	parts := strings.Split(idsStr, ",")
	for _, p := range parts {
		id, err := strconv.ParseUint(strings.TrimSpace(p), 10, 64)
		if err != nil {
			continue
		}
		_ = h.svc.DeleteResource(c.Request.Context(), uint(id))
	}
	util.ResponseSuccess(c, "资源已删除")
}

// @Summary 删除单个资源
// @Tags 资源
// @Description 删除单个资源权限
// @Accept json
// @Produce json
// @Param id path int true "资源ID"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/resources/{id} [delete]
func (h *ResourceHandler) DeleteResource(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的资源ID"))
		return
	}
	if err := h.svc.DeleteResource(c.Request.Context(), uint(id)); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "资源已删除")
}

// @Summary 获取角色资源选项
// @Tags 资源
// @Description 获取角色授权资源选项（用于下拉框，树形结构）
// @Accept json
// @Produce json
// @Success 200 {object} object "成功返回资源树"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/role/resources [get]
func (h *ResourceHandler) ListResourceOptions(c *gin.Context) {
	result, err := h.svc.ListResourceOptions(c.Request.Context())
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// @Summary 更新角色的资源关联
// @Tags 资源
// @Description 为角色分配资源权限
// @Accept json
// @Produce json
// @Param id path int true "角色ID"
// @Param body body object true "资源ID列表(resourceIds)"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/roles/{id}/resources [put]
func (h *ResourceHandler) UpdateRoleResource(c *gin.Context) {
	roleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的角色ID"))
		return
	}
	var body struct {
		ResourceIDs []uint `json:"resourceIds"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	if err := h.svc.AssignResourceToRole(c.Request.Context(), uint(roleID), body.ResourceIDs); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "资源权限已更新")
}
