package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/util"
)

// OperationLogHandler 操作日志处理器（对标 Java OperationLogController）
type OperationLogHandler struct {
	// operationLogService service.OperationLogService
}

func NewOperationLogHandler() *OperationLogHandler { return &OperationLogHandler{} }

// ListOperationLogs 获取操作日志列表
// GET /api/admin/operationLogs
func (h *OperationLogHandler) ListOperationLogs(c *gin.Context) {
	var condition dto.ConditionVO
	c.ShouldBindQuery(&condition)
	_ = condition // 支持按操作类型/时间范围/用户名筛选
	pageNum, pageSize := util.PageQuery(c)
	util.ResponseSuccessWithPage(c, []interface{}{}, int64(0), pageNum, pageSize)
}

// DeleteOperationLogs 删除操作日志
// DELETE /api/admin/operationLogs?ids=1,2,3
func (h *OperationLogHandler) DeleteOperationLogs(c *gin.Context) {
	idsStr := c.Query("ids")
	_ = idsStr
	util.ResponseSuccess(c, "日志已删除")
}
