package handler

import (
	"io"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/service"
	"github.com/aurora-go/aurora/internal/util"
	"github.com/aurora-go/aurora/internal/vo"
)

// ArticleHandler 文章处理器（对标 Java ArticleController）
type ArticleHandler struct {
	svc  *service.ArticleService
	file *service.FileService
}

func NewArticleHandler(svc *service.ArticleService, file *service.FileService) *ArticleHandler {
	return &ArticleHandler{svc: svc, file: file}
}

// ListArticles 获取文章列表（前台公开）
// GET /api/articles
func (h *ArticleHandler) ListArticles(c *gin.Context) {
	var condition dto.ConditionVO
	c.ShouldBindQuery(&condition)
	pageNum, pageSize := util.PageQuery(c)
	page := dto.PageVO{PageNum: pageNum, PageSize: pageSize}

	result, err := h.svc.ListArticles(c.Request.Context(), condition, page)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// GetArticleById 根据ID获取文章详情
// GET /api/articles/:id
func (h *ArticleHandler) GetArticleById(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的文章ID"))
		return
	}

	articleID := uint(id)

	// 先增加浏览量（异步，不阻塞响应）
	h.svc.IncrementViewCount(c.Request.Context(), articleID)

	// 获取文章详情
	result, err := h.svc.GetArticleByID(c.Request.Context(), articleID)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// SearchArticles 搜索文章
// GET /api/articles/search?keyword=xxx
func (h *ArticleHandler) SearchArticles(c *gin.Context) {
	keyword := c.DefaultQuery("keyword", "")
	if keyword == "" {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("搜索关键词不能为空"))
		return
	}
	pageNum, pageSize := util.PageQuery(c)
	page := dto.PageVO{PageNum: pageNum, PageSize: pageSize}

	result, err := h.svc.SearchArticles(c.Request.Context(), keyword, page)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// TopAndFeaturedArticles 获取置顶和推荐文章
// GET /api/articles/topAndFeatured
func (h *ArticleHandler) TopAndFeaturedArticles(c *gin.Context) {
	list, err := h.svc.GetTopArticles(c.Request.Context(), 10)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	// 分离置顶和推荐
	var topArticles, featuredArticles []dto.ArticleCardDTO
	for _, a := range list {
		if a.IsTop == 1 {
			topArticles = append(topArticles, a)
		}
		if a.IsFeatured == 1 {
			featuredArticles = append(featuredArticles, a)
		}
	}
	util.ResponseSuccess(c, map[string]interface{}{
		"topArticles":      topArticles,
		"featuredArticles": featuredArticles,
	})
}

// GetArchives 获取文章归档列表
// GET /api/articles/archives
func (h *ArticleHandler) GetArchives(c *gin.Context) {
	archives, err := h.svc.GetArchives(c.Request.Context())
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, archives)
}

// ==================== 后台管理端点 ====================

// SaveArticle 新增/更新文章 (对标Java版: 前端统一POST /admin/articles, 通过articleVO.id区分新增/更新)
// POST /api/admin/articles
func (h *ArticleHandler) SaveArticle(c *gin.Context) {
	var articleVO vo.ArticleVO
	if err := c.ShouldBindJSON(&articleVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}

	userID, _ := c.Get("userId")
	uid := uint(0)
	if id, ok := userID.(uint); ok {
		uid = id
	}

	// 通过 articleVO.ID 判断新增还是更新 (前端统一发POST请求)
	if articleVO.ID > 0 {
		// 更新
		result, err := h.svc.UpdateArticle(c.Request.Context(), articleVO)
		if err != nil {
			util.ResponseError(c, err)
			return
		}
		util.ResponseSuccess(c, result)
		return
	}

	// 新增
	result, err := h.svc.CreateArticle(c.Request.Context(), uid, articleVO)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// UpdateArticleStatus 更新文章状态（置顶/推荐）
// PUT /api/admin/articles/:id/status
func (h *ArticleHandler) UpdateArticleStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的文章ID"))
		return
	}
	var statusVO vo.ArticleTopFeaturedVO
	if err := c.ShouldBindJSON(&statusVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	statusVO.ID = uint(id)
	if err := h.svc.UpdateTopFeatured(c.Request.Context(), statusVO); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, nil)
}

// DeleteArticle 彻底删除文章（物理删除 - 回收站使用）
// DELETE /api/admin/articles/delete
// 对标Java: @DeleteMapping("/admin/articles/delete") + articleService.deleteArticles()
// 注意: 这是物理删除，对标Java版的 deleteBatchIds 实现
func (h *ArticleHandler) DeleteArticle(c *gin.Context) {
	// 从请求体接收ID数组（对标Java @RequestBody List<Integer> articleIds）
	var ids []uint
	if err := c.ShouldBindJSON(&ids); err != nil || len(ids) == 0 {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("请选择要删除的文章"))
		return
	}

	// 物理删除（对标Java版 deleteArticles - deleteBatchIds）
	if err := h.svc.DeleteArticlesPermanently(c.Request.Context(), ids); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, "文章已彻底删除")
}

// ImportArticle 导入Markdown文件为文章
// POST /api/admin/articles/import
func (h *ArticleHandler) ImportArticle(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("请选择要导入的Markdown文件"))
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("读取文件失败"))
		return
	}

	userID, _ := c.Get("userId")
	uid := uint(0)
	if id, ok := userID.(uint); ok {
		uid = id
	}

	contents := map[string]string{
		"imported.md": string(content),
	}
	success, failures := h.svc.ImportArticles(c.Request.Context(), uid, contents)
	util.ResponseSuccess(c, map[string]interface{}{
		"success":  success,
		"failures": failures,
	})
}

// ListAdminArticles 后台文章管理列表
// GET /api/admin/articles
func (h *ArticleHandler) ListAdminArticles(c *gin.Context) {
	var condition dto.ConditionVO
	if err := c.ShouldBindQuery(&condition); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	pageNum, pageSize := util.PageQuery(c)
	page := dto.PageVO{PageNum: pageNum, PageSize: pageSize}

	result, err := h.svc.ListAdminArticles(c.Request.Context(), condition, page)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// UpdateArticlePassword 设置/取消文章密码访问
// PUT /api/admin/articles/:id/password
func (h *ArticleHandler) UpdateArticlePassword(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的文章ID"))
		return
	}
	var body struct {
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	// 更新文章密码
	articleVO := vo.ArticleVO{ID: uint(id), Password: body.Password}
	_, err = h.svc.UpdateArticle(c.Request.Context(), articleVO)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, nil)
}

// VerifyArticlePassword 验证文章访问密码
// POST /api/articles/:id/password/verify
func (h *ArticleHandler) VerifyArticlePassword(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的文章ID"))
		return
	}
	var body struct {
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	verified := h.svc.VerifyPassword(c.Request.Context(), uint(id), body.Password)
	util.ResponseSuccess(c, map[string]interface{}{
		"verified": verified,
	})
}

// ListArticlesByCategoryId 根据分类ID获取文章列表
// GET /api/articles/categoryId?categoryId=1
func (h *ArticleHandler) ListArticlesByCategoryId(c *gin.Context) {
	categoryID, err := strconv.ParseUint(c.DefaultQuery("categoryId", "0"), 10, 64)
	if err != nil || categoryID == 0 {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的分类ID"))
		return
	}
	pageNum, pageSize := util.PageQuery(c)
	page := dto.PageVO{PageNum: pageNum, PageSize: pageSize}
	condition := dto.ConditionVO{CategoryID: ptrUint(uint(categoryID))}
	result, err := h.svc.ListArticles(c.Request.Context(), condition, page)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// ListArticlesByTagId 根据标签ID获取文章列表
// GET /api/articles/tagId?tagId=1
func (h *ArticleHandler) ListArticlesByTagId(c *gin.Context) {
	tagID, err := strconv.ParseUint(c.DefaultQuery("tagId", "0"), 10, 64)
	if err != nil || tagID == 0 {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的标签ID"))
		return
	}
	pageNum, pageSize := util.PageQuery(c)
	page := dto.PageVO{PageNum: pageNum, PageSize: pageSize}
	condition := dto.ConditionVO{Keywords: ""}
	result, err := h.svc.ListArticles(c.Request.Context(), condition, page)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// UpdateArticleTopAndFeatured 修改文章置顶/推荐状态
// PUT /api/admin/articles/topAndFeatured
func (h *ArticleHandler) UpdateArticleTopAndFeatured(c *gin.Context) {
	var statusVO vo.ArticleTopFeaturedVO
	if err := c.ShouldBindJSON(&statusVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	if err := h.svc.UpdateTopFeatured(c.Request.Context(), statusVO); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, nil)
}

// UpdateArticleDelete 逻辑删除/恢复文章
// PUT /api/admin/articles
// 对标Java: @PutMapping("/admin/articles") + @RequestBody DeleteVO (ids数组 + isDelete)
func (h *ArticleHandler) UpdateArticleDelete(c *gin.Context) {
	var body struct {
		Ids      []uint `json:"ids"`
		IsDelete int8   `json:"isDelete"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || len(body.Ids) == 0 {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("请选择要操作的文章"))
		return
	}
	if err := h.svc.UpdateArticleDelete(c.Request.Context(), body.Ids, body.IsDelete); err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, nil)
}

// UploadArticleImage 上传文章图片/封面
// POST /api/admin/articles/images
// 对标Java: uploadStrategyContext.executeUploadStrategy(file, FilePathEnum.ARTICLE.getPath())
func (h *ArticleHandler) UploadArticleImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("请选择要上传的图片"))
		return
	}

	// 调用Service上传 (返回完整URL)
	result, err := h.file.UploadArticleImage(c.Request.Context(), file)
	if err != nil {
		util.ResponseError(c, errors.ErrFileUploadFailed.WithMsg(err.Error()))
		return
	}

	// 返回完整URL (前端需要完整URL才能加载图片)
	util.ResponseSuccess(c, result)
}

// GetAdminArticleById 后台根据ID获取文章详情
// GET /api/admin/articles/:id
func (h *ArticleHandler) GetAdminArticleById(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的文章ID"))
		return
	}
	result, err := h.svc.GetArticleByID(c.Request.Context(), uint(id))
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// ExportArticle 批量导出文章
// POST /api/admin/articles/export
func (h *ArticleHandler) ExportArticle(c *gin.Context) {
	var ids []uint
	if err := c.ShouldBindJSON(&ids); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	// TODO: 批量导出Markdown
	util.ResponseSuccess(c, ids)
}

// ptrUint 辅助函数: 返回uint指针
func ptrUint(v uint) *uint { return &v }

// splitIDs 辅助函数: 拆分逗号分隔的ID字符串
func splitIDs(s string) []string {
	if s == "" {
		return nil
	}
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			part := s[start:i]
			if part != "" {
				result = append(result, part)
			}
			start = i + 1
		}
	}
	if start < len(s) {
		result = append(result, s[start:])
	}
	return result
}
