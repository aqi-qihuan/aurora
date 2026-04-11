package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/service"
	"github.com/aurora-go/aurora/internal/util"
	"github.com/aurora-go/aurora/internal/vo"
)

// RoleHandler 角色管理处理器（对标 Java RoleController）
type RoleHandler struct {
	svc *service.RoleService
}

func NewRoleHandler(svc *service.RoleService) *RoleHandler {
	return &RoleHandler{svc: svc}
}

// ListRoles 获取角色列表（后台管理）
// GET /api/admin/roles
func (h *RoleHandler) ListRoles(c *gin.Context) {
	list, err := h.svc.ListRoles(c.Request.Context())
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, list)
}

// SaveOrUpdate 保存/更新角色
// POST /api/admin/roles
// PUT /api/admin/roles/:id
func (h *RoleHandler) SaveOrUpdate(c *gin.Context) {
	var roleVO vo.RoleVO
	if err := c.ShouldBindJSON(&roleVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}

	idStr := c.Param("id")
	if idStr != "" {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的角色ID"))
			return
		}
		if err := h.svc.UpdateRole(c.Request.Context(), uint(id), roleVO); err != nil {
			util.ResponseError(c, err)
			return
		}
		util.ResponseSuccess(c, nil)
		return
	}

	result, err := h.svc.CreateRole(c.Request.Context(), roleVO)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// DeleteRoles 批量删除角色
// DELETE /api/admin/roles?ids=1,2,3
func (h *RoleHandler) DeleteRoles(c *gin.Context) {
	idsStr := c.Query("ids")
	if idsStr == "" {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("请选择要删除的角色"))
		return
	}
	parts := strings.Split(idsStr, ",")
	for _, p := range parts {
		id, err := strconv.ParseUint(strings.TrimSpace(p), 10, 64)
		if err != nil {
			continue
		}
		_ = h.svc.DeleteRole(c.Request.Context(), uint(id))
	}
	util.ResponseSuccess(c, "角色已删除")
}

// GetRoleById 获取角色详情（含菜单权限和资源权限）
// GET /api/admin/roles/:id
func (h *RoleHandler) GetRoleById(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的角色ID"))
		return
	}
	result, err := h.svc.GetRoleByID(c.Request.Context(), uint(id))
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// UpdateRoleMenu 更新角色的菜单关联
// PUT /api/admin/roles/:id/menus
func (h *RoleHandler) UpdateRoleMenu(c *gin.Context) {
	roleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的角色ID"))
		return
	}
	var body struct {
		MenuIDs []uint `json:"menuIds"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	// 使用UpdateRole来更新菜单权限
	roleVO := vo.RoleVO{MenuIDs: body.MenuIDs}
	if err := h.svc.UpdateRole(c.Request.Context(), uint(roleID), roleVO); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "菜单权限已更新")
}
