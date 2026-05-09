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

// @Summary 获取定时任务列表
// @Tags 定时任务
// @Description 获取定时任务列表（分页）
// @Accept json
// @Produce json
// @Param keyword query string false "搜索关键词"
// @Param current query int false "当前页码"
// @Param size query int false "每页数量"
// @Success 200 {object} object "成功返回任务列表"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/jobs [get]
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

// @Summary 保存/更新定时任务
// @Tags 定时任务
// @Description 后台保存或更新定时任务
// @Accept json
// @Produce json
// @Param jobVO body vo.JobVO true "任务信息"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/jobs [post]
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

// @Summary 批量删除定时任务
// @Tags 定时任务
// @Description 后台批量删除定时任务
// @Accept json
// @Produce json
// @Param ids body []uint true "任务ID列表"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/jobs [delete]
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

// @Summary 启用/禁用定时任务
// @Tags 定时任务
// @Description 启用或禁用定时任务
// @Accept json
// @Produce json
// @Param body body object true "任务ID和状态(id, status)"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/jobs/status [put]
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

// @Summary 立即执行一次定时任务
// @Tags 定时任务
// @Description 立即手动执行一次定时任务
// @Accept json
// @Produce json
// @Param body body object true "任务ID和分组(id, jobGroup)"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/jobs/run [put]
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

// @Summary 新增定时任务
// @Tags 定时任务
// @Description 新增定时任务
// @Accept json
// @Produce json
// @Param jobVO body vo.JobVO true "任务信息"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/jobs [post]
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

// @Summary 修改定时任务
// @Tags 定时任务
// @Description 修改定时任务信息
// @Accept json
// @Produce json
// @Param body body object true "任务ID和任务信息(id, jobVO)"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/jobs [put]
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

// @Summary 获取任务详情
// @Tags 定时任务
// @Description 根据ID获取定时任务详情
// @Accept json
// @Produce json
// @Param id path int true "任务ID"
// @Success 200 {object} object "成功返回任务详情"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/jobs/{id} [get]
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

// @Summary 获取所有任务分组
// @Tags 定时任务
// @Description 获取所有定时任务分组列表
// @Accept json
// @Produce json
// @Success 200 {object} object "成功返回任务分组列表"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/jobs/jobGroups [get]
func (h *JobHandler) ListJobGroups(c *gin.Context) {
	util.ResponseSuccess(c, []string{"DEFAULT", "SYSTEM"})
}
