package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/service"
	"github.com/aurora-go/aurora/internal/util"
)

// ExceptionLogHandler 异常日志处理器（对标 Java ExceptionLogController）
type ExceptionLogHandler struct {
	svc *service.ExceptionLogService
}

func NewExceptionLogHandler(svc *service.ExceptionLogService) *ExceptionLogHandler {
	return &ExceptionLogHandler{svc: svc}
}

// ListExceptionLogs 获取异常日志列表
// GET /api/admin/exceptionLogs
func (h *ExceptionLogHandler) ListExceptionLogs(c *gin.Context) {
	var condition dto.ConditionVO
	c.ShouldBindQuery(&condition)
	pageNum, pageSize := util.PageQuery(c)
	page := dto.PageVO{PageNum: pageNum, PageSize: pageSize}

	result, err := h.svc.ListExceptionLogs(c.Request.Context(), condition, page)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// DeleteExceptionLogs 删除异常日志
// DELETE /api/admin/exceptionLogs?ids=1,2,3
func (h *ExceptionLogHandler) DeleteExceptionLogs(c *gin.Context) {
	idsStr := c.Query("ids")
	if idsStr == "" {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("请选择要删除的日志"))
		return
	}
	parts := strings.Split(idsStr, ",")
	for _, p := range parts {
		id, err := strconv.ParseUint(strings.TrimSpace(p), 10, 64)
		if err != nil {
			continue
		}
		_ = h.svc.DeleteExceptionLog(c.Request.Context(), uint(id))
	}
	util.ResponseSuccess(c, "日志已删除")
}
