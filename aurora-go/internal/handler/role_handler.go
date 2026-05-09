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

// RoleHandler 角色管理处理器（对标 Java RoleController）
type RoleHandler struct {
	svc *service.RoleService
}

func NewRoleHandler(svc *service.RoleService) *RoleHandler {
	return &RoleHandler{svc: svc}
}

// @Summary 获取角色列表
// @Tags 角色
// @Description 获取角色列表（分页）
// @Accept json
// @Produce json
// @Param keyword query string false "搜索关键词"
// @Param current query int false "当前页码"
// @Param size query int false "每页数量"
// @Success 200 {object} object "成功返回角色列表"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/roles [get]
func (h *RoleHandler) ListRoles(c *gin.Context) {
	var condition dto.ConditionVO
	c.ShouldBindQuery(&condition)
	pageNum, pageSize := util.PageQuery(c)
	page := dto.PageVO{PageNum: pageNum, PageSize: pageSize}

	result, err := h.svc.ListRolesPage(c.Request.Context(), condition, page)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// @Summary 获取用户角色选项列表
// @Tags 角色
// @Description 获取用户角色选项列表（用于授权下拉框）
// @Accept json
// @Produce json
// @Success 200 {object} object "成功返回角色选项列表"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/users/role [get]
func (h *RoleHandler) ListUserRoles(c *gin.Context) {
	result, err := h.svc.ListUserRoles(c.Request.Context())
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// @Summary 保存/更新角色
// @Tags 角色
// @Description 后台保存或更新角色
// @Accept json
// @Produce json
// @Param roleVO body vo.RoleVO true "角色信息"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/role [post]
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

// @Summary 批量删除角色
// @Tags 角色
// @Description 后台批量删除角色
// @Accept json
// @Produce json
// @Param ids body []uint true "角色ID列表"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/roles [delete]
func (h *RoleHandler) DeleteRoles(c *gin.Context) {
	var ids []uint
	if err := c.ShouldBindJSON(&ids); err != nil || len(ids) == 0 {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("请选择要删除的角色"))
		return
	}
	for _, id := range ids {
		_ = h.svc.DeleteRole(c.Request.Context(), id)
	}
	util.ResponseSuccess(c, "角色已删除")
}

// @Summary 获取角色详情
// @Tags 角色
// @Description 获取角色详情（含菜单权限和资源权限）
// @Accept json
// @Produce json
// @Param id path int true "角色ID"
// @Success 200 {object} object "成功返回角色详情"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/roles/{id} [get]
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

// @Summary 更新角色的菜单关联
// @Tags 角色
// @Description 为角色分配菜单权限
// @Accept json
// @Produce json
// @Param id path int true "角色ID"
// @Param body body object true "菜单ID列表(menuIds)"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/roles/{id}/menus [put]
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
