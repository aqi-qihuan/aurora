package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/service"
	"github.com/aurora-go/aurora/internal/util"
	"github.com/aurora-go/aurora/internal/vo"
)

// PortfolioHandler 作品集管理处理器（对标 Java PortfolioController）
type PortfolioHandler struct {
	svc *service.PortfolioService
}

func NewPortfolioHandler(svc *service.PortfolioService) *PortfolioHandler {
	return &PortfolioHandler{svc: svc}
}

// ListFeatured 前台首页置顶作品集（6 条）
// @Tags 作品集
// @Router /api/portfolios/featured [get]
func (h *PortfolioHandler) ListFeatured(c *gin.Context) {
	list, err := h.svc.ListFeatured(c.Request.Context(), 6)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, list)
}

// ListAll 前台作品集分页
// @Tags 作品集
// @Router /api/portfolios [get]
func (h *PortfolioHandler) ListAll(c *gin.Context) {
	pageNum, pageSize := util.PageQuery(c)
	page := dto.PageVO{PageNum: pageNum, PageSize: pageSize}
	result, err := h.svc.ListAll(c.Request.Context(), page)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// ListAdmin 后台作品集分页（含隐藏项）
// @Tags 作品集
// @Router /api/admin/portfolios [get]
func (h *PortfolioHandler) ListAdmin(c *gin.Context) {
	var cond dto.ConditionVO
	c.ShouldBindQuery(&cond)
	pageNum, pageSize := util.PageQuery(c)
	page := dto.PageVO{PageNum: pageNum, PageSize: pageSize}
	result, err := h.svc.ListAdmin(c.Request.Context(), cond, page)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// Update 后台编辑作品（封面/排序/置顶/可见性）
// @Tags 作品集
// @Router /api/admin/portfolios [put]
func (h *PortfolioHandler) Update(c *gin.Context) {
	var v vo.PortfolioVO
	if err := c.ShouldBindJSON(&v); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	if err := h.svc.UpdatePortfolio(c.Request.Context(), v); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, nil)
}

// Delete 后台批量删除作品
// @Tags 作品集
// @Router /api/admin/portfolios [delete]
func (h *PortfolioHandler) Delete(c *gin.Context) {
	var ids []uint
	if err := c.ShouldBindJSON(&ids); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("请提供要删除的作品ID列表"))
		return
	}
	if len(ids) == 0 {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("作品ID列表不能为空"))
		return
	}
	if err := h.svc.DeletePortfolios(c.Request.Context(), ids); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "作品已删除")
}

// Sync 手动触发 GitHub 同步
// @Tags 作品集
// @Router /api/admin/portfolios/sync [post]
func (h *PortfolioHandler) Sync(c *gin.Context) {
	if err := h.svc.SyncFromGitHub(c.Request.Context()); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "同步完成")
}
