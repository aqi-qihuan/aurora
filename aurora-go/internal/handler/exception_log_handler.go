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

// @Summary 获取异常日志列表
// @Tags 异常日志
// @Description 获取异常日志列表（分页）
// @Accept json
// @Produce json
// @Param keyword query string false "搜索关键词"
// @Param current query int false "当前页码"
// @Param size query int false "每页数量"
// @Success 200 {object} object "成功返回异常日志列表"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/exception/logs [get]
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

// @Summary 批量删除异常日志
// @Tags 异常日志
// @Description 批量删除异常日志
// @Accept json
// @Produce json
// @Param ids body []uint true "异常日志ID列表"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/exception/logs [delete]
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
