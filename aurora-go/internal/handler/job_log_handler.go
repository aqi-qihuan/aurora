package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/util"
)

// JobLogHandler 调度日志处理器（对标 Java JobLogController）
type JobLogHandler struct {
	// jobLogService service.JobLogService
}

func NewJobLogHandler() *JobLogHandler { return &JobLogHandler{} }

// ListJobLogs 获取调度日志列表
// GET /api/admin/jobLogs
func (h *JobLogHandler) ListJobLogs(c *gin.Context) {
	var condition dto.ConditionVO
	c.ShouldBindQuery(&condition)
	_ = condition
	pageNum, pageSize := util.PageQuery(c)
	util.ResponseSuccessWithPage(c, []interface{}{}, int64(0), pageNum, pageSize)
}

// DeleteJobLogs 清理调度日志（按时间范围或ID）
// DELETE /api/admin/jobLogs
func (h *JobLogHandler) DeleteJobLogs(c *gin.Context) {
	// TODO: P0-5 支持按时间范围清理或按ID批量删除
	util.ResponseSuccess(c, "日志已清理")
}
