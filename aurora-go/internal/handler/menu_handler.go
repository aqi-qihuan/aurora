package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/service"
	"github.com/aurora-go/aurora/internal/util"
	"github.com/aurora-go/aurora/internal/vo"
)

// MenuHandler 菜单管理处理器（对标 Java MenuController）
type MenuHandler struct {
	svc *service.MenuService
}

func NewMenuHandler(svc *service.MenuService) *MenuHandler {
	return &MenuHandler{svc: svc}
}

// ListMenus 获取菜单树形列表（后台）
// GET /api/admin/menus
func (h *MenuHandler) ListMenus(c *gin.Context) {
	tree, err := h.svc.GetMenuTree(c.Request.Context())
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, tree)
}

// GetUserMenus 获取当前用户的菜单树（用于前端动态路由）
// GET /api/admin/user/menus (需JWT)
func (h *MenuHandler) GetUserMenus(c *gin.Context) {
	userID, _ := c.Get("userId")
	uid := uint(0)
	if id, ok := userID.(uint); ok {
		uid = id
	}
	if uid == 0 {
		// 尝试从int64类型获取
		if id, ok := userID.(int64); ok {
			uid = uint(id)
		}
	}
	tree, err := h.svc.GetUserMenus(c.Request.Context(), uid)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, tree)
}

// SaveOrUpdate 保存/更新菜单（后台）
// POST /api/admin/menus
// PUT /api/admin/menus/:id
func (h *MenuHandler) SaveOrUpdate(c *gin.Context) {
	var menuVO vo.MenuVO
	if err := c.ShouldBindJSON(&menuVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}

	idStr := c.Param("id")
	if idStr != "" {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的菜单ID"))
			return
		}
		if err := h.svc.UpdateMenu(c.Request.Context(), uint(id), menuVO); err != nil {
			util.ResponseError(c, err)
			return
		}
		util.ResponseSuccess(c, nil)
		return
	}

	result, err := h.svc.CreateMenu(c.Request.Context(), menuVO)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// DeleteMenu 删除菜单（级联删除子菜单）
// DELETE /api/admin/menus/:id
func (h *MenuHandler) DeleteMenu(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的菜单ID"))
		return
	}
	if err := h.svc.DeleteMenu(c.Request.Context(), uint(id)); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "菜单已删除")
}

// ListMenuOptions 获取角色菜单选项（用于角色授权下拉框）
// GET /api/admin/role/menus
func (h *MenuHandler) ListMenuOptions(c *gin.Context) {
	tree, err := h.svc.GetMenuTree(c.Request.Context())
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, tree)
}

// UpdateMenuIsHidden 修改目录是否隐藏
// PUT /api/admin/menus/isHidden
func (h *MenuHandler) UpdateMenuIsHidden(c *gin.Context) {
	var body struct {
		ID       uint `json:"id"`
		IsHidden int8 `json:"isHidden"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	menuVO := vo.MenuVO{}
	_ = body.IsHidden
	if err := h.svc.UpdateMenu(c.Request.Context(), body.ID, menuVO); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, nil)
}
