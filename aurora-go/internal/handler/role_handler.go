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

// RoleHandler 角色管理处理器（对标 Java RoleController）
type RoleHandler struct {
	// roleService service.RoleService
}

func NewRoleHandler() *RoleHandler { return &RoleHandler{} }

// ListRoles 获取角色列表（后台管理）
// GET /api/admin/roles
func (h *RoleHandler) ListRoles(c *gin.Context) {
	util.ResponseSuccess(c, []interface{}{})
}

// SaveOrUpdate 保存/更新角色
// POST /api/admin/roles
// PUT /api/admin/roles/:id
func (h *RoleHandler) SaveOrUpdate(c *gin.Context) {
	var roleVO dto.RoleVO
	if err := c.ShouldBindJSON(&roleVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParam.WithMsg(err.Error()))
		return
	}
	_ = roleVO

	zap.L().Debug("Save role", zap.Any("role", roleVO))
	util.ResponseSuccess(c, nil)
}

// DeleteRoles 批量删除角色
// DELETE /api/admin/roles?ids=1,2,3
func (h *RoleHandler) DeleteRoles(c *gin.Context) {
	idsStr := c.Query("ids")
	if idsStr == "" {
		util.ResponseError(c, errors.ErrInvalidParam.WithMsg("请选择要删除的角色"))
		return
	}

	ids := strings.Split(idsStr, ",")
	zap.L().Debug("Delete roles", zap.Strings("ids", ids))
	util.ResponseSuccess(c, "角色已删除")
}

// GetRoleById 获取角色详情（含菜单权限和资源权限）
// GET /api/admin/roles/:id
func (h *RoleHandler) GetRoleById(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	// TODO: P0-5 查询角色 + 关联的菜单列表 + 资源列表
	util.ResponseSuccess(c, map[string]interface{}{
		"id":       id,
		"name":     "",
		"menus":    []interface{}{},
		"resources": []interface{}{},
	})
}

// UpdateRoleMenu 更新角色的菜单关联
// PUT /api/admin/roles/:id/menus
func (h *RoleHandler) UpdateRoleMenu(c *gin.Context) {
	var menuIds dto.MenuIdsVO
	if err := c.ShouldBindJSON(&menuIds); err != nil {
		util.ResponseError(c, errors.ErrInvalidParam.WithMsg(err.Error()))
		return
	}
	_ = menuIds

	util.ResponseSuccess(c, "菜单权限已更新")
}
