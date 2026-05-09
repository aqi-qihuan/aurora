package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/service"
	"github.com/aurora-go/aurora/internal/util"
)

// JobLogHandler 调度日志处理器（对标 Java JobLogController）
type JobLogHandler struct {
	svc *service.JobLogService
}

func NewJobLogHandler(svc *service.JobLogService) *JobLogHandler {
	return &JobLogHandler{svc: svc}
}

// @Summary 获取调度日志列表
// @Tags 任务日志
// @Description 获取定时任务调度日志列表（分页）
// @Accept json
// @Produce json
// @Param keyword query string false "搜索关键词"
// @Param current query int false "当前页码"
// @Param size query int false "每页数量"
// @Success 200 {object} object "成功返回调度日志列表"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/jobLogs [get]
func (h *JobLogHandler) ListJobLogs(c *gin.Context) {
	var condition dto.ConditionVO
	c.ShouldBindQuery(&condition)
	pageNum, pageSize := util.PageQuery(c)
	page := dto.PageVO{PageNum: pageNum, PageSize: pageSize}

	result, err := h.svc.ListJobLogs(c.Request.Context(), condition, page)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// @Summary 批量删除调度日志
// @Tags 任务日志
// @Description 批量删除定时任务调度日志
// @Accept json
// @Produce json
// @Param ids body []uint true "日志ID列表"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/jobLogs [delete]
func (h *JobLogHandler) DeleteJobLogs(c *gin.Context) {
	var ids []uint
	if err := c.ShouldBindJSON(&ids); err != nil {
		util.ResponseError(c, err)
		return
	}
	if len(ids) == 0 {
		util.ResponseError(c, fmt.Errorf("请提供要删除的日志ID列表"))
		return
	}

	// 批量删除（对标Java deleteJobLogs）
	result := h.svc.DeleteJobLogs(c.Request.Context(), ids)
	if result != nil {
		util.ResponseError(c, result)
		return
	}
	util.ResponseSuccess(c, "日志已删除")
}

// @Summary 清空所有调度日志
// @Tags 任务日志
// @Description 清空所有定时任务调度日志
// @Accept json
// @Produce json
// @Success 200 {object} object "成功响应"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/jobLogs/clean [delete]
func (h *JobLogHandler) CleanJobLogs(c *gin.Context) {
	if err := h.svc.CleanJobLogs(c.Request.Context()); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "日志已清除")
}

// @Summary 获取调度日志所有分组名
// @Tags 任务日志
// @Description 获取调度日志所有分组名称
// @Accept json
// @Produce json
// @Success 200 {object} object "成功返回分组列表"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/jobLogs/jobGroups [get]
func (h *JobLogHandler) ListJobLogGroups(c *gin.Context) {
	util.ResponseSuccess(c, []string{"DEFAULT", "SYSTEM"})
}
