package handler

import (
	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/service"
	"github.com/aurora-go/aurora/internal/util"
	"github.com/gin-gonic/gin"
)

// ExceptionLogHandler 异常日志处理器（对标 Java ExceptionLogController）
type ExceptionLogHandler struct {
	svc *service.ExceptionLogService
}

func NewExceptionLogHandler(svc *service.ExceptionLogService) *ExceptionLogHandler {
	return &ExceptionLogHandler{svc: svc}
}

// ListExceptionLogs 获取异常日志列表（对标Java ExceptionLogController.listExceptionLogs）
// GET /api/admin/exception/logs
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

// DeleteExceptionLogs 批量删除异常日志（对标Java: @DeleteMapping + @RequestBody List<Integer>）
// DELETE /api/admin/exception/logs
// 前端 axios 发送的 body 是原始数组 [id1, id2, ...]
func (h *ExceptionLogHandler) DeleteExceptionLogs(c *gin.Context) {
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
	if err := h.svc.DeleteExceptionLogs(c.Request.Context(), ids); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, nil)
}
