package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/service"
	"github.com/aurora-go/aurora/internal/util"
	"github.com/aurora-go/aurora/internal/vo"
)

// JobHandler 定时任务管理处理器（对标 Java QuartzController）
type JobHandler struct {
	svc *service.JobService
}

func NewJobHandler(svc *service.JobService) *JobHandler {
	return &JobHandler{svc: svc}
}

// ListJobs 获取定时任务列表
// GET /api/admin/jobs
func (h *JobHandler) ListJobs(c *gin.Context) {
	var condition dto.ConditionVO
	c.ShouldBindQuery(&condition)
	pageNum, pageSize := util.PageQuery(c)
	page := dto.PageVO{PageNum: pageNum, PageSize: pageSize}

	result, err := h.svc.ListJobs(c.Request.Context(), condition, page)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// SaveOrUpdate 保存/更新定时任务
// POST /api/admin/jobs
// PUT /api/admin/jobs/:id
func (h *JobHandler) SaveOrUpdate(c *gin.Context) {
	var jobVO vo.JobVO
	if err := c.ShouldBindJSON(&jobVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}

	idStr := c.Param("id")
	if idStr != "" {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的任务ID"))
			return
		}
		if err := h.svc.UpdateJob(c.Request.Context(), uint(id), jobVO); err != nil {
			util.ResponseError(c, err)
			return
		}
		util.ResponseSuccess(c, nil)
		return
	}

	result, err := h.svc.CreateJob(c.Request.Context(), jobVO)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// DeleteJob 删除定时任务
// DELETE /api/admin/jobs/:id
func (h *JobHandler) DeleteJob(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的任务ID"))
		return
	}
	if err := h.svc.DeleteJob(c.Request.Context(), uint(id)); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "任务已删除")
}

// UpdateJobStatus 启用/禁用定时任务
// PUT /api/admin/jobs/:id/status
func (h *JobHandler) UpdateJobStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的任务ID"))
		return
	}
	var statusVO struct {
		Status int8 `json:"status"`
	}
	if err := c.ShouldBindJSON(&statusVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	if err := h.svc.ChangeJobStatus(c.Request.Context(), uint(id), statusVO.Status); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "任务状态已更新")
}

// RunJobOnce 立即执行一次定时任务
// POST /api/admin/jobs/:id/run
func (h *JobHandler) RunJobOnce(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的任务ID"))
		return
	}
	log, err := h.svc.RunJobNow(c.Request.Context(), uint(id))
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, log)
}
