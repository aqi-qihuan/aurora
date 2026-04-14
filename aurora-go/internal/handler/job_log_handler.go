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

// ListJobLogs 获取调度日志列表
// GET /api/admin/jobLogs
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

// DeleteJobLogs 批量删除调度日志（对标Java: deleteJobLogs）
// DELETE /api/admin/jobLogs
// 前端 axios 发送的 body 是原始数组 [id1, id2, ...]
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

// CleanJobLogs 清空所有调度日志（对标Java: cleanJobLogs）
// DELETE /api/admin/jobLogs/clean
func (h *JobLogHandler) CleanJobLogs(c *gin.Context) {
	if err := h.svc.CleanJobLogs(c.Request.Context()); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "日志已清除")
}

// ListJobLogGroups 获取调度日志所有分组名
// GET /api/admin/jobLogs/jobGroups
func (h *JobLogHandler) ListJobLogGroups(c *gin.Context) {
	util.ResponseSuccess(c, []string{"DEFAULT", "SYSTEM"})
}
