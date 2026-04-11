package handler

import (
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

// DeleteJobLogs 清理调度日志
// DELETE /api/admin/jobLogs
func (h *JobLogHandler) DeleteJobLogs(c *gin.Context) {
	var ids []int
	c.ShouldBindJSON(&ids)
	if len(ids) > 0 {
		for _, id := range ids {
			_ = h.svc.ClearJobLogs(c.Request.Context(), id)
		}
	} else {
		if err := h.svc.ClearJobLogs(c.Request.Context(), 30); err != nil {
			util.ResponseError(c, err)
			return
		}
	}
	util.ResponseSuccess(c, "日志已清理")
}

// CleanJobLogs 清除所有调度日志
// DELETE /api/admin/jobLogs/clean
func (h *JobLogHandler) CleanJobLogs(c *gin.Context) {
	if err := h.svc.ClearJobLogs(c.Request.Context(), 0); err != nil {
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
