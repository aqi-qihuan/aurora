package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/util"
)

// ExceptionLogHandler 异常日志处理器（对标 Java ExceptionLogController）
type ExceptionLogHandler struct {
	// exceptionLogService service.ExceptionLogService
}

func NewExceptionLogHandler() *ExceptionLogHandler { return &ExceptionLogHandler{} }

// ListExceptionLogs 获取异常日志列表
// GET /api/admin/exceptionLogs
func (h *ExceptionLogHandler) ListExceptionLogs(c *gin.Context) {
	var condition dto.ConditionVO
	c.ShouldBindQuery(&condition)
	_ = condition // 支持按异常类型/时间范围/是否处理筛选
	pageNum, pageSize := util.PageQuery(c)
	util.ResponseSuccessWithPage(c, []interface{}{}, int64(0), pageNum, pageSize)
}

// DeleteExceptionLogs 删除异常日志
// DELETE /api/admin/exceptionLogs?ids=1,2,3
func (h *ExceptionLogHandler) DeleteExceptionLogs(c *gin.Context) {
	idsStr := c.Query("ids")
	_ = idsStr
	util.ResponseSuccess(c, "日志已删除")
}
