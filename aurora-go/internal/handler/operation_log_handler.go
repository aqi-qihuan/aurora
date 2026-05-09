package handler

import (
	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/service"
	"github.com/aurora-go/aurora/internal/util"
	"github.com/gin-gonic/gin"
)

// OperationLogHandler 操作日志处理器（对标 Java OperationLogController）
type OperationLogHandler struct {
	svc *service.OperationLogService
}

func NewOperationLogHandler(svc *service.OperationLogService) *OperationLogHandler {
	return &OperationLogHandler{svc: svc}
}

// @Summary 获取操作日志列表
// @Tags 操作日志
// @Description 获取操作日志列表（分页）
// @Accept json
// @Produce json
// @Param keyword query string false "搜索关键词"
// @Param current query int false "当前页码"
// @Param size query int false "每页数量"
// @Success 200 {object} object "成功返回操作日志列表"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/operation/logs [get]
func (h *OperationLogHandler) ListOperationLogs(c *gin.Context) {
	var condition dto.ConditionVO
	c.ShouldBindQuery(&condition)
	pageNum, pageSize := util.PageQuery(c)
	page := dto.PageVO{PageNum: pageNum, PageSize: pageSize}

	result, err := h.svc.ListOperationLogs(c.Request.Context(), condition, page)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// @Summary 批量删除操作日志
// @Tags 操作日志
// @Description 批量删除操作日志
// @Accept json
// @Produce json
// @Param ids body []uint true "日志ID列表"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/operation/logs [delete]
func (h *OperationLogHandler) DeleteOperationLogs(c *gin.Context) {
	var ids []uint
	if err := c.ShouldBindJSON(&ids); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("请提供要删除的日志ID列表"))
		return
	}
	if len(ids) == 0 {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("日志ID列表不能为空"))
		return
	}

	// 批量删除（对标Java removeByIds）
	if err := h.svc.DeleteOperationLogs(c.Request.Context(), ids); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, nil)
}
