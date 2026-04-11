package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/util"
)

// MenuHandler 菜单管理处理器（对标 Java MenuController）
// 菜单用于后台动态路由 + RBAC权限控制
type MenuHandler struct {
	// menuService service.MenuService
}

func NewMenuHandler() *MenuHandler { return &MenuHandler{} }

// ListMenus 获取菜单树形列表（后台）
// GET /api/admin/menus
func (h *MenuHandler) ListMenus(c *gin.Context) {
	// TODO: P0-5 查询全部菜单 → 构建树形结构(递归parent-child)
	util.ResponseSuccess(c, []interface{}{})
}

// GetUserMenus 获取当前用户的菜单树（用于前端动态路由）
// GET /api/admin/user/menus (需JWT)
// 对标 PermissionController.getUserMenu()
func (h *MenuHandler) GetUserMenus(c *gin.Context) {
	userID := c.GetInt64("userId")
	_ = userID

	// TODO: P0-5 根据用户角色 → 查询角色关联菜单 → 过滤 → 返回树形结构

	zap.L().Debug("Get user menus", zap.Int64("userId", userID))
	util.ResponseSuccess(c, []interface{}{})
}

// SaveOrUpdate 保存/更新菜单（后台）
// POST /api/admin/menus
// PUT /api/admin/menus/:id
func (h *MenuHandler) SaveOrUpdate(c *gin.Context) {
	var menuVO dto.MenuVO
	if err := c.ShouldBindJSON(&menuVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	_ = menuVO

	zap.L().Debug("Save menu", zap.Any("menu", menuVO))
	util.ResponseSuccess(c, nil)
}

// DeleteMenu 删除菜单（级联删除子菜单）
// DELETE /api/admin/menus/:id
func (h *MenuHandler) DeleteMenu(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	zap.L().Debug("Delete menu", zap.Int64("id", id))
	// TODO: P0-5 删除菜单 + 级联删除所有子菜单 + 清除角色关联

	util.ResponseSuccess(c, "菜单已删除")
}
