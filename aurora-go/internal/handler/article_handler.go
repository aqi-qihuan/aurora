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
	svc *service.ArticleService
}

func NewArticleHandler(svc *service.ArticleService) *ArticleHandler {
	return &ArticleHandler{svc: svc}
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

// SaveArticle 新增/更新文章
// POST /api/admin/articles
// PUT /api/admin/articles/:id
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

	idStr := c.Param("id")
	if idStr != "" {
		// 更新
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的文章ID"))
			return
		}
		articleVO.ID = uint(id)
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

// DeleteArticle 删除文章（逻辑删除）
// DELETE /api/admin/articles/:ids
func (h *ArticleHandler) DeleteArticle(c *gin.Context) {
	idsStr := c.Param("ids")
	if idsStr == "" {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("请选择要删除的文章"))
		return
	}
	// 支持批量删除 (逗号分隔)
	ids := make([]uint, 0)
	for _, part := range splitIDs(idsStr) {
		id, err := strconv.ParseUint(part, 10, 64)
		if err != nil {
			continue
		}
		ids = append(ids, uint(id))
	}
	if len(ids) == 1 {
		if err := h.svc.DeleteArticle(c.Request.Context(), ids[0]); err != nil {
			util.ResponseError(c, err)
			return
		}
	} else if len(ids) > 1 {
		if err := h.svc.BatchDeleteArticles(c.Request.Context(), ids); err != nil {
			util.ResponseError(c, err)
			return
		}
	}
	util.ResponseSuccess(c, "文章已删除")
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
func (h *ArticleHandler) UpdateArticleDelete(c *gin.Context) {
	var body struct {
		ID      uint `json:"id"`
		IsDelete int8 `json:"isDelete"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	articleVO := vo.ArticleVO{ID: body.ID}
	_, err := h.svc.UpdateArticle(c.Request.Context(), articleVO)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, nil)
}

// UploadArticleImage 上传文章图片
// POST /api/admin/articles/images
func (h *ArticleHandler) UploadArticleImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("请选择要上传的图片"))
		return
	}
	// TODO: 上传到MinIO/OSS
	url := "/uploads/articles/" + file.Filename
	util.ResponseSuccess(c, url)
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
