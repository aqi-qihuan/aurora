package handler

import (
	"strconv"
	"strings"

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
	idsStr := c.Query("ids")
	if idsStr != "" {
		// 按ID批量删除
		parts := strings.Split(idsStr, ",")
		for _, p := range parts {
			id, err := strconv.ParseUint(strings.TrimSpace(p), 10, 64)
			if err != nil {
				continue
			}
			_ = h.svc.ClearJobLogs(c.Request.Context(), int(id)) // 复用清理方法
		}
	} else {
		// 清理30天前的日志
		if err := h.svc.ClearJobLogs(c.Request.Context(), 30); err != nil {
			util.ResponseError(c, err)
			return
		}
	}
	util.ResponseSuccess(c, "日志已清理")
}
