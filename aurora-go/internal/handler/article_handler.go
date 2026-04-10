package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/util"
)

// ArticleHandler 文章处理器（对标 Java ArticleController）
// 端点: 15个 (CRUD + 搜索 + 归档 + 导出 + 推荐 + 置顶)
type ArticleHandler struct {
	// articleService service.ArticleService // P0-5 注入
}

// NewArticleHandler 创建文章Handler
func NewArticleHandler() *ArticleHandler {
	return &ArticleHandler{}
}

// ListArticles 获取文章列表（前台公开）
// GET /api/articles
// 对标 ArticleController.getArticles() + 前台文章列表
func (h *ArticleHandler) ListArticles(c *gin.Context) {
	pageNum, pageSize := util.PageQuery(c)

	// TODO: P0-5 注入 Service 后调用
	// result, total, err := h.articleService.ListArticles(c.Request.Context(), pageNum, pageSize)
	// if err != nil {
	//     util.ResponseError(c, errors.ErrInternalServerError)
	//     return
	// }
	// pageResult := util.BuildPageResult(result, total, pageNum, pageSize)
	util.ResponseSuccessWithPage(c, []interface{}{}, int64(0), pageNum, pageSize)
}

// GetArticleById 根据ID获取文章详情
// GET /api/articles/:id
// 对标 ArticleController.getArticleById()
func (h *ArticleHandler) GetArticleById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParam.WithMsg("无效的文章ID"))
		return
	}

	// TODO: P0-5
	// articleDTO, err := h.articleService.GetArticleById(id)
	// if err != nil {
	//     if errors.IsNotFound(err) {
	//         util.ResponseError(c, errors.ErrArticleNotFound)
	//         return
	//     }
	//     util.ResponseError(c, err)
	//     return
	// }
	// util.ResponseSuccess(c, articleDTO)
	util.ResponseSuccess(c, nil)
}

// SearchArticles 搜索文章（ES全文搜索 或 MySQL模糊搜索）
// GET /api/articles/search?keyword=xxx
// 对标 ArticleController.searchArticles()
func (h *ArticleHandler) SearchArticles(c *gin.Context) {
	keyword := c.DefaultQuery("keyword", "")
	if keyword == "" {
		util.ResponseError(c, errors.ErrInvalidParam.WithMsg("搜索关键词不能为空"))
		return
	}
	pageNum, pageSize := util.PageQuery(c)

	// TODO: P0-5
	// result, total, err := h.searchService.Search(c.Request.Context(), keyword, pageNum, pageSize)
	// if err != nil {
	//     util.ResponseError(c, err)
	//     return
	// }
	util.ResponseSuccessWithPage(c, []interface{}{}, int64(0), pageNum, pageSize)
}

// TopAndFeaturedArticles 获取置顶和推荐文章
// GET /api/articles/topAndFeatured
// 对标 ArticleController.topAndFeaturedArticles()
func (h *ArticleHandler) TopAndFeaturedArticles(c *gin.Context) {
	// TODO: P0-5
	util.ResponseSuccess(c, map[string]interface{}{
		"topArticles":      []interface{}{},
		"featuredArticles": []interface{}{},
	})
}

// GetArchives 获取文章归档列表（按年月分组）
// GET /api/articles/archives
// 对标 ArticleController.listArchives()
func (h *ArticleHandler) GetArchives(c *gin.Context) {
	// TODO: P0-5
	util.ResponseSuccess(c, []interface{}{})
}

// ==================== 后台管理端点（需JWT认证）====================

// SaveArticle 新增/更新文章
// POST /api/admin/articles
// PUT /api/admin/articles/:id
// 对标 ArticleController.saveArticle()
func (h *ArticleHandler) SaveArticle(c *gin.Context) {
	var articleVO dto.ArticleVO
	if err := c.ShouldBindJSON(&articleVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParam.WithMsg(err.Error()))
		return
	}
	_ = articleVO // TODO: P0-5 调用Service保存

	util.ResponseSuccess(c, nil)
}

// UpdateArticleStatus 更新文章状态（发布/下架/置顶/推荐）
// PUT /api/admin/articles/:id/status
// 对标 ArticleController.updateArticleTopOrStatus()
func (h *ArticleHandler) UpdateArticleStatus(c *gin.Context) {
	var statusVO dto.ArticleStatusUpdateVO
	if err := c.ShouldBindJSON(&statusVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParam.WithMsg(err.Error()))
		return
	}
	_ = statusVO // TODO: P0-5

	util.ResponseSuccess(c, nil)
}

// DeleteArticle 删除文章（逻辑删除）
// DELETE /api/admin/articles/:id
// 对标 ArticleController.deleteArticles()
func (h *ArticleHandler) DeleteArticle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParam.WithMsg("无效的文章ID"))
		return
	}

	zap.L().Debug("Delete article requested", zap.Int64("article_id", id))
	// TODO: P0-5 调用Service软删除

	util.ResponseSuccess(c, "文章已删除")
}

// ImportArticle 导入Markdown文件为文章
// POST /api/admin/articles/import
// 对标 ArticleController.importArticle() - MD导入功能
func (h *ArticleHandler) ImportArticle(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParam.WithMsg("请选择要导入的Markdown文件"))
		return
	}
	defer file.Close()

	// TODO: P0-5 解析MD → 提取front-matter(标题/分类/标签/日期) → 保存到DB

	util.ResponseSuccess(c, "文章导入成功")
}

// ExportArticle 导出文章为Markdown
// GET /api/admin/articles/:id/export
// 对标 ArticleController.exportArticle()
func (h *ArticleHandler) ExportArticle(c *gin.Context) {
	idStr := c.Param("id")
	_, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParam.WithMsg("无效的文章ID"))
		return
	}

	// TODO: P0-5 查询文章 → 格式化MD → 返回文件下载
	c.Header("Content-Disposition", "attachment; filename=article.md")
	c.Header("Content-Type", "text/markdown; charset=utf-8")

	// c.Data(http.StatusOK, "text/markdown", mdBytes)
	c.String(200, "# 导出的文章内容\n\n待实现...")
}

// ListAdminArticles 后台文章管理列表（含分页+条件筛选）
// GET /api/admin/articles
// 对标 ArticleController.listArticles() - 后台管理版
func (h *ArticleHandler) ListAdminArticles(c *gin.Context) {
	var condition dto.ConditionVO
	if err := c.ShouldBindQuery(&condition); err != nil {
		util.ResponseError(c, errors.ErrInvalidParam.WithMsg(err.Error()))
		return
	}
	_ = condition // TODO: P0-5 按条件查询(状态/关键词/分类/时间范围)

	pageNum, pageSize := util.PageQuery(c)
	util.ResponseSuccessWithPage(c, []interface{}{}, int64(0), pageNum, pageSize)
}

// UpdateArticlePassword 设置/取消文章密码访问
// PUT /api/admin/articles/:id/password
// 密码保护文章功能
func (h *ArticleHandler) UpdateArticlePassword(c *gin.Context) {
	var pwdVO dto.ArticlePasswordVO
	if err := c.ShouldBindJSON(&pwdVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParam.WithMsg(err.Error()))
		return
	}
	_ = pwdVO // TODO: P0-5

	util.ResponseSuccess(c, nil)
}

// VerifyArticlePassword 验证文章访问密码
// POST /api/articles/:id/password/verify
func (h *ArticleHandler) VerifyArticlePassword(c *gin.Context) {
	var pwdVO dto.ArticlePasswordVerifyVO
	if err := c.ShouldBindJSON(&pwdVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParam.WithMsg(err.Error()))
		return
	}
	_ = pwdVO // TODO: P0-5 验证密码 → 返回JWT token或cookie

	util.ResponseSuccess(c, map[string]interface{}{
		"verified": true,
		"token":    "",
	})
}
