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

// ListResources 获取资源权限列表
// GET /api/admin/resources
func (h *ResourceHandler) ListResources(c *gin.Context) {
	var condition dto.ConditionVO
	c.ShouldBindQuery(&condition)
	pageNum, pageSize := util.PageQuery(c)
	page := dto.PageVO{PageNum: pageNum, PageSize: pageSize}

	result, err := h.svc.ListResources(c.Request.Context(), condition, page)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// SaveOrUpdate 保存/更新资源
// POST /api/admin/resources
// PUT /api/admin/resources/:id
func (h *ResourceHandler) SaveOrUpdate(c *gin.Context) {
	var resourceVO vo.ResourceVO
	if err := c.ShouldBindJSON(&resourceVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}

	idStr := c.Param("id")
	if idStr != "" {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的资源ID"))
			return
		}
		if err := h.svc.UpdateResource(c.Request.Context(), uint(id), resourceVO); err != nil {
			util.ResponseError(c, err)
			return
		}
		util.ResponseSuccess(c, nil)
		return
	}

	result, err := h.svc.CreateResource(c.Request.Context(), resourceVO)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// DeleteResources 批量删除资源
// DELETE /api/admin/resources?ids=1,2,3
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

// DeleteResource 删除单个资源
// DELETE /api/admin/resources/:id
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

// ListResourceOptions 获取角色资源选项（用于角色授权下拉框）
// GET /api/admin/role/resources
func (h *ResourceHandler) ListResourceOptions(c *gin.Context) {
	var condition dto.ConditionVO
	result, err := h.svc.ListResources(c.Request.Context(), condition, dto.PageVO{PageNum: 1, PageSize: 100})
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// UpdateRoleResource 更新角色的资源关联
// PUT /api/admin/roles/:id/resources
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
