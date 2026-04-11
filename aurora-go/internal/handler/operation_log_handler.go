package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/service"
	"github.com/aurora-go/aurora/internal/util"
)

// OperationLogHandler 操作日志处理器（对标 Java OperationLogController）
type OperationLogHandler struct {
	svc *service.OperationLogService
}

func NewOperationLogHandler(svc *service.OperationLogService) *OperationLogHandler {
	return &OperationLogHandler{svc: svc}
}

// ListOperationLogs 获取操作日志列表
// GET /api/admin/operationLogs
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

// DeleteOperationLogs 删除操作日志
// DELETE /api/admin/operationLogs?ids=1,2,3
func (h *OperationLogHandler) DeleteOperationLogs(c *gin.Context) {
	idsStr := c.Query("ids")
	if idsStr != "" {
		parts := strings.Split(idsStr, ",")
		for _, p := range parts {
			id, err := strconv.ParseUint(strings.TrimSpace(p), 10, 64)
			if err != nil {
				continue
			}
			_ = h.svc.DeleteOperationLog(c.Request.Context(), uint(id))
		}
	} else {
		// 清理30天前的日志
		_ = h.svc.ClearOldLogs(c.Request.Context(), 30)
	}
	util.ResponseSuccess(c, "日志已删除")
}
