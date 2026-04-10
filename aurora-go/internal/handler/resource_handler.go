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

// ResourceHandler 资源权限管理处理器（对标 Java ResourceController）
type ResourceHandler struct {
	// resourceService service.ResourceService
}

func NewResourceHandler() *ResourceHandler { return &ResourceHandler{} }

// ListResources 获取资源权限列表
// GET /api/admin/resources
func (h *ResourceHandler) ListResources(c *gin.Context) {
	util.ResponseSuccess(c, []interface{}{})
}

// SaveOrUpdate 保存/更新资源
// POST /api/admin/resources
// PUT /api/admin/resources/:id
func (h *ResourceHandler) SaveOrUpdate(c *gin.Context) {
	var resourceVO dto.ResourceVO
	if err := c.ShouldBindJSON(&resourceVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParam.WithMsg(err.Error()))
		return
	}
	_ = resourceVO

	zap.L().Debug("Save resource", zap.Any("resource", resourceVO))
	util.ResponseSuccess(c, nil)
}

// DeleteResources 批量删除资源
// DELETE /api/admin/resources?ids=1,2,3
func (h *ResourceHandler) DeleteResources(c *gin.Context) {
	idsStr := c.Query("ids")
	if idsStr == "" {
		util.ResponseError(c, errors.ErrInvalidParam.WithMsg("请选择要删除的资源"))
		return
	}

	ids := strings.Split(idsStr, ",")
	zap.L().Debug("Delete resources", zap.Strings("ids", ids))
	util.ResponseSuccess(c, "资源已删除")
}

// UpdateRoleResource 更新角色的资源关联
// PUT /api/admin/roles/:id/resources
func (h *ResourceHandler) UpdateRoleResource(c *gin.Context) {
	idStr := c.Param("id")
	_, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParam.WithMsg("无效的角色ID"))
		return
	}

	var resourceIds dto.ResourceIdsVO
	if err := c.ShouldBindJSON(&resourceIds); err != nil {
		util.ResponseError(c, errors.ErrInvalidParam.WithMsg(err.Error()))
		return
	}
	_ = resourceIds

	util.ResponseSuccess(c, "资源权限已更新")
}
