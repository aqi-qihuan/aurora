package handler

import (
	"strconv"

	"github.com/aurora-go/aurora/internal/dto"
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

// @Summary 获取菜单列表
// @Tags 菜单
// @Description 后台获取菜单列表（包含隐藏菜单）
// @Accept json
// @Produce json
// @Success 200 {object} object "成功返回菜单列表"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/menus [get]
func (h *MenuHandler) ListMenus(c *gin.Context) {
	// 后台管理应显示所有菜单（包含隐藏），对标Java listMenus无is_hidden过滤
	result, err := h.svc.ListMenus(c.Request.Context())
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// @Summary 获取当前用户的菜单树
// @Tags 菜单
// @Description 获取当前登录用户的菜单树（用于前端动态路由）
// @Accept json
// @Produce json
// @Success 200 {object} object "成功返回菜单树"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/user/menus [get]
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

	// 临时方案：如果无法获取用户ID（使用临时Token而非JWT），则返回所有可见菜单
	var tree []dto.MenuTreeDTO
	var err error
	if uid > 0 {
		tree, err = h.svc.GetUserMenus(c.Request.Context(), uid)
	} else {
		// 返回所有未隐藏的菜单
		tree, err = h.svc.GetMenuTree(c.Request.Context())
	}

	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, tree)
}

// @Summary 保存/更新菜单
// @Tags 菜单
// @Description 后台保存或更新菜单
// @Accept json
// @Produce json
// @Param menuVO body vo.MenuVO true "菜单信息"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/menus [post]
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

// @Summary 删除菜单
// @Tags 菜单
// @Description 删除菜单（级联删除子菜单）
// @Accept json
// @Produce json
// @Param id path int true "菜单ID"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/menus/{id} [delete]
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

// @Summary 获取角色菜单选项
// @Tags 菜单
// @Description 获取角色授权菜单选项（用于下拉框）
// @Accept json
// @Produce json
// @Success 200 {object} object "成功返回菜单选项列表"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/role/menus [get]
func (h *MenuHandler) ListMenuOptions(c *gin.Context) {
	tree, err := h.svc.GetMenuTree(c.Request.Context())
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, tree)
}

// @Summary 修改目录是否隐藏
// @Tags 菜单
// @Description 修改菜单目录的隐藏状态
// @Accept json
// @Produce json
// @Param body body object true "菜单ID和隐藏状态(id, isHidden)"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/menus/isHidden [put]
func (h *MenuHandler) UpdateMenuIsHidden(c *gin.Context) {
	var body struct {
		ID       uint `json:"id"`
		IsHidden int8 `json:"isHidden"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	// 修复：将IsHidden赋值到menuVO中
	isHidden := body.IsHidden
	menuVO := vo.MenuVO{IsHidden: &isHidden}
	if err := h.svc.UpdateMenu(c.Request.Context(), body.ID, menuVO); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, nil)
}
