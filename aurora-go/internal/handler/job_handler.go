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

// DeleteJob 批量删除定时任务
// DELETE /api/admin/jobs
// 前端 axios 发送的 body 是原始数组 [id1, id2, ...]
func (h *JobHandler) DeleteJob(c *gin.Context) {
	var ids []uint
	if err := c.ShouldBindJSON(&ids); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("请提供要删除的任务ID列表"))
		return
	}
	if len(ids) == 0 {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("任务ID列表不能为空"))
		return
	}
	for _, id := range ids {
		if err := h.svc.DeleteJob(c.Request.Context(), id); err != nil {
			util.ResponseError(c, err)
			return
		}
	}
	util.ResponseSuccess(c, "任务已删除")
}

// UpdateJobStatus 启用/禁用定时任务
// PUT /api/admin/jobs/status
// Java: @PutMapping("/jobs/status")
func (h *JobHandler) UpdateJobStatus(c *gin.Context) {
	var body struct {
		ID     uint `json:"id"`
		Status int8 `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	if err := h.svc.ChangeJobStatus(c.Request.Context(), body.ID, body.Status); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "任务状态已更新")
}

// RunJobOnce 立即执行一次定时任务
// PUT /api/admin/jobs/run
// Java: @PutMapping("/jobs/run")
func (h *JobHandler) RunJobOnce(c *gin.Context) {
	var body struct {
		ID       uint   `json:"id"`
		JobGroup string `json:"jobGroup"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	log, err := h.svc.RunJobNow(c.Request.Context(), body.ID)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, log)
}

// SaveJob 新增定时任务
// POST /api/admin/jobs
func (h *JobHandler) SaveJob(c *gin.Context) {
	var jobVO vo.JobVO
	if err := c.ShouldBindJSON(&jobVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	result, err := h.svc.CreateJob(c.Request.Context(), jobVO)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// UpdateJob 修改定时任务
// PUT /api/admin/jobs
func (h *JobHandler) UpdateJob(c *gin.Context) {
	var body struct {
		ID    uint        `json:"id"`
		JobVO vo.JobVO    `json:"jobVO"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	if body.ID > 0 {
		if err := h.svc.UpdateJob(c.Request.Context(), body.ID, body.JobVO); err != nil {
			util.ResponseError(c, err)
			return
		}
	}
	util.ResponseSuccess(c, nil)
}

// GetJobById 根据ID获取任务详情
// GET /api/admin/jobs/:id
func (h *JobHandler) GetJobById(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的任务ID"))
		return
	}
	job, err := h.svc.GetJobByID(c.Request.Context(), uint(id))
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, job)
}

// ListJobGroups 获取所有任务分组
// GET /api/admin/jobs/jobGroups
func (h *JobHandler) ListJobGroups(c *gin.Context) {
	util.ResponseSuccess(c, []string{"DEFAULT", "SYSTEM"})
}
