package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/util"
)

// JobHandler 定时任务管理处理器（对标 Java QuartzController）
type JobHandler struct {
	// jobService service.JobService
}

func NewJobHandler() *JobHandler { return &JobHandler{} }

// ListJobs 获取定时任务列表
// GET /api/admin/jobs
func (h *JobHandler) ListJobs(c *gin.Context) {
	pageNum, pageSize := util.PageQuery(c)
	util.ResponseSuccessWithPage(c, []interface{}{}, int64(0), pageNum, pageSize)
}

// SaveOrUpdate 保存/更新定时任务
// POST /api/admin/jobs
// PUT /api/admin/jobs/:id
func (h *JobHandler) SaveOrUpdate(c *gin.Context) {
	var jobVO dto.JobVO
	if err := c.ShouldBindJSON(&jobVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParam.WithMsg(err.Error()))
		return
	}
	_ = jobVO

	zap.L().Debug("Save job", zap.Any("job", jobVO))
	// TODO: P0-5 保存任务 → 更新cron调度器

	util.ResponseSuccess(c, nil)
}

// DeleteJob 删除定时任务
// DELETE /api/admin/jobs/:id
func (h *JobHandler) DeleteJob(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	zap.L().Debug("Delete job", zap.Int64("id", id))
	util.ResponseSuccess(c, "任务已删除")
}

// UpdateJobStatus 启用/禁用定时任务
// PUT /api/admin/jobs/:id/status
func (h *JobHandler) UpdateJobStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	var statusVO dto.StatusUpdateVO
	if err := c.ShouldBindJSON(&statusVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParam.WithMsg(err.Error()))
		return
	}

	zap.L().Debug("Update job status", "id", id, "status", statusVO.Status)
	util.ResponseSuccess(c, "任务状态已更新")
}

// RunJobOnce 立即执行一次定时任务
// POST /api/admin/jobs/:id/run
// 对标 QuartzController.triggerJob()
func (h *JobHandler) RunJobOnce(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	zap.L().Info("Manual trigger job execution", zap.Int64("id", id))
	// TODO: P0-5 手动触发一次执行

	util.ResponseSuccess(c, "任务已触发执行")
}
