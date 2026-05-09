package handler

import (
	"io"
	"log/slog"
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

// @Summary 获取文章列表
// @Tags 文章
// @Description 获取前台公开的文章列表（分页）
// @Accept json
// @Produce json
// @Param keyword query string false "搜索关键词"
// @Param categoryId query int false "分类ID"
// @Param tagId query int false "标签ID"
// @Param current query int false "当前页码"
// @Param size query int false "每页数量"
// @Success 200 {object} object "成功返回文章列表"
// @Failure 500 {object} object "服务器内部错误"
// @Router /api/articles/all [get]
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

// @Summary 获取文章详情
// @Tags 文章
// @Description 根据ID获取文章详情并增加浏览量
// @Accept json
// @Produce json
// @Param id path int true "文章ID"
// @Success 200 {object} object "成功返回文章详情"
// @Failure 400 {object} object "请求参数错误"
// @Failure 500 {object} object "服务器内部错误"
// @Router /api/articles/{id} [get]
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

// @Summary 搜索文章
// @Tags 文章
// @Description 根据关键词搜索文章
// @Accept json
// @Produce json
// @Param keywords query string true "搜索关键词"
// @Success 200 {object} object "成功返回搜索结果"
// @Failure 400 {object} object "请求参数错误"
// @Failure 500 {object} object "服务器内部错误"
// @Router /api/articles/search [get]
func (h *ArticleHandler) SearchArticles(c *gin.Context) {
	// 兼容 keywords 和 keyword 两种参数名
	keyword := c.DefaultQuery("keywords", c.DefaultQuery("keyword", ""))
	if keyword == "" {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("搜索关键词不能为空"))
		return
	}

	// 对标Java: 不分页，最多返回10条
	result, err := h.svc.SearchArticles(c.Request.Context(), keyword)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// @Summary 获取置顶和推荐文章
// @Tags 文章
// @Description 获取置顶文章和推荐文章列表
// @Accept json
// @Produce json
// @Success 200 {object} object "成功返回置顶和推荐文章"
// @Failure 500 {object} object "服务器内部错误"
// @Router /api/articles/topAndFeatured [get]
func (h *ArticleHandler) TopAndFeaturedArticles(c *gin.Context) {
	list, err := h.svc.GetTopArticles(c.Request.Context(), 10)
	if err != nil {
		util.ResponseError(c, err)
		return
	}

	// 对标Java版：如果没有文章，返回空结构
	if len(list) == 0 {
		util.ResponseSuccess(c, map[string]interface{}{
			"topArticle":       nil,
			"featuredArticles": []dto.ArticleCardDTO{},
		})
		return
	}

	// 对标Java版：最多取3篇文章
	if len(list) > 3 {
		list = list[:3]
	}

	// 第一篇文章作为置顶文章（单数对象）
	topArticle := list[0]
	// 剩余文章作为推荐文章
	var featuredArticles []dto.ArticleCardDTO
	if len(list) > 1 {
		featuredArticles = list[1:]
	}

	util.ResponseSuccess(c, map[string]interface{}{
		"topArticle":       topArticle,
		"featuredArticles": featuredArticles,
	})
}

// @Summary 获取文章归档列表
// @Tags 文章
// @Description 获取按月分组的文章归档列表
// @Accept json
// @Produce json
// @Param current query int false "当前页码"
// @Param size query int false "每页数量"
// @Success 200 {object} object "成功返回归档列表"
// @Failure 500 {object} object "服务器内部错误"
// @Router /api/archives/all [get]
func (h *ArticleHandler) GetArchives(c *gin.Context) {
	// 获取分页参数
	current, _ := util.ParseInt(c.DefaultQuery("current", "1"), 1)
	size, _ := util.ParseInt(c.DefaultQuery("size", "12"), 12)

	archives, err := h.svc.GetArchives(c.Request.Context(), current, size)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, archives)
}

// ==================== 后台管理端点 ====================

// @Summary 新增/更新文章
// @Tags 文章
// @Description 后台新增或更新文章，通过articleVO.id区分新增/更新
// @Accept json
// @Produce json
// @Param articleVO body vo.ArticleVO true "文章信息"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/articles [post]
func (h *ArticleHandler) SaveArticle(c *gin.Context) {
	var articleVO vo.ArticleVO
	if err := c.ShouldBindJSON(&articleVO); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}

	// 使用 user_info_id（对标Java UserUtil.getUserDetailsDTO().getUserInfoId()）
	userInfoID, _ := c.Get("user_info_id")
	uid := uint(0)
	if id, ok := userInfoID.(uint); ok {
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

// @Summary 更新文章状态（置顶/推荐）
// @Tags 文章
// @Description 更新文章的置顶或推荐状态
// @Accept json
// @Produce json
// @Param id path int true "文章ID"
// @Param statusVO body vo.ArticleTopFeaturedVO true "状态信息"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/articles/{id}/status [put]
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

// @Summary 彻底删除文章
// @Tags 文章
// @Description 物理删除文章（回收站使用）
// @Accept json
// @Produce json
// @Param ids body []uint true "文章ID列表"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/articles/delete [delete]
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

// @Summary 批量导入Markdown文件为文章
// @Tags 文章
// @Description 批量导入Markdown文件创建文章
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Markdown文件"
// @Success 200 {object} object "成功返回导入结果"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/articles/import [post]
func (h *ArticleHandler) ImportArticle(c *gin.Context) {
	// 获取所有上传的文件（支持批量）
	form, err := c.MultipartForm()
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("文件上传失败"))
		return
	}
	files := form.File["file"]
	if len(files) == 0 {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("请选择要导入的Markdown文件"))
		return
	}

	// 使用 user_info_id（对标Java UserUtil.getUserDetailsDTO().getUserInfoId()）
	userInfoIDD, _ := c.Get("user_info_id")
	uid := uint(0)
	if id, ok := userInfoIDD.(uint); ok {
		uid = id
	}

	// 读取所有文件内容
	contents := make(map[string]string)
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			slog.Warn("打开文件失败", "filename", fileHeader.Filename, "error", err)
			continue
		}
		content, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			slog.Warn("读取文件失败", "filename", fileHeader.Filename, "error", err)
			continue
		}
		contents[fileHeader.Filename] = string(content)
	}

	if len(contents) == 0 {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("没有成功读取到文件内容"))
		return
	}

	success, failures := h.svc.ImportArticles(c.Request.Context(), uid, contents)
	util.ResponseSuccess(c, map[string]interface{}{
		"success":  success,
		"failures": failures,
	})
}

// @Summary 后台文章管理列表
// @Tags 文章
// @Description 获取后台文章列表（分页，含状态筛选）
// @Accept json
// @Produce json
// @Param keyword query string false "搜索关键词"
// @Param current query int false "当前页码"
// @Param size query int false "每页数量"
// @Success 200 {object} object "成功返回文章列表"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/articles [get]
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

// @Summary 设置/取消文章密码访问
// @Tags 文章
// @Description 为文章设置或取消密码访问
// @Accept json
// @Produce json
// @Param id path int true "文章ID"
// @Param password body string true "密码，为空则取消密码访问"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/articles/{id}/password [put]
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

// @Summary 验证文章访问密码
// @Tags 文章
// @Description 验证文章访问密码是否正确
// @Accept json
// @Produce json
// @Param body body object true "验证参数"
// @Success 200 {object} object "成功返回验证结果"
// @Failure 400 {object} object "请求参数错误"
// @Failure 500 {object} object "服务器内部错误"
// @Router /api/articles/access [post]
func (h *ArticleHandler) VerifyArticlePassword(c *gin.Context) {
	var body struct {
		ArticleID      uint   `json:"articleId"`
		ArticlePassword string `json:"articlePassword"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	if body.ArticleID == 0 {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的文章ID"))
		return
	}
	verified := h.svc.VerifyPassword(c.Request.Context(), body.ArticleID, body.ArticlePassword)
	util.ResponseSuccess(c, map[string]interface{}{
		"verified": verified,
	})
}

// @Summary 根据分类ID获取文章列表
// @Tags 文章
// @Description 根据分类ID分页获取文章列表
// @Accept json
// @Produce json
// @Param categoryId query int true "分类ID"
// @Param current query int false "当前页码"
// @Param size query int false "每页数量"
// @Success 200 {object} object "成功返回文章列表"
// @Failure 400 {object} object "请求参数错误"
// @Failure 500 {object} object "服务器内部错误"
// @Router /api/articles/categoryId [get]
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

// @Summary 根据标签ID获取文章列表
// @Tags 文章
// @Description 根据标签ID分页获取文章列表
// @Accept json
// @Produce json
// @Param tagId query int true "标签ID"
// @Param current query int false "当前页码"
// @Param size query int false "每页数量"
// @Success 200 {object} object "成功返回文章列表"
// @Failure 400 {object} object "请求参数错误"
// @Failure 500 {object} object "服务器内部错误"
// @Router /api/articles/tagId [get]
func (h *ArticleHandler) ListArticlesByTagId(c *gin.Context) {
	tagID, err := strconv.ParseUint(c.DefaultQuery("tagId", "0"), 10, 64)
	if err != nil || tagID == 0 {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的标签ID"))
		return
	}
	pageNum, pageSize := util.PageQuery(c)
	page := dto.PageVO{PageNum: pageNum, PageSize: pageSize}
	condition := dto.ConditionVO{TagID: ptrUint(uint(tagID))}
	result, err := h.svc.ListArticles(c.Request.Context(), condition, page)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// @Summary 修改文章置顶/推荐状态
// @Tags 文章
// @Description 修改文章的置顶或推荐状态
// @Accept json
// @Produce json
// @Param statusVO body vo.ArticleTopFeaturedVO true "置顶/推荐状态"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/articles/topAndFeatured [put]
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

// @Summary 逻辑删除/恢复文章
// @Tags 文章
// @Description 批量逻辑删除或恢复文章
// @Accept json
// @Produce json
// @Param ids body []uint true "文章ID列表"
// @Param isDelete body int8 true "是否删除(0恢复/1删除)"
// @Success 200 {object} object "成功响应"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/articles [put]
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

// @Summary 上传文章图片/封面
// @Tags 文章
// @Description 上传文章图片或封面图片
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "图片文件"
// @Success 200 {object} object "成功返回图片URL"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/articles/images [post]
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

// @Summary 后台获取文章详情
// @Tags 文章
// @Description 后台根据ID获取文章详情（含编辑信息）
// @Accept json
// @Produce json
// @Param id path int true "文章ID"
// @Success 200 {object} object "成功返回文章详情"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/articles/{id} [get]
func (h *ArticleHandler) GetAdminArticleById(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg("无效的文章ID"))
		return
	}
	result, err := h.svc.GetArticleByIDAdmin(c.Request.Context(), uint(id))
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, result)
}

// @Summary 批量导出文章
// @Tags 文章
// @Description 批量导出文章为Markdown文件
// @Accept json
// @Produce json
// @Param ids body []uint true "文章ID列表"
// @Success 200 {object} object "成功返回导出URL列表"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权/Token无效"
// @Failure 500 {object} object "服务器内部错误"
// @Security BearerAuth
// @Router /api/admin/articles/export [post]
func (h *ArticleHandler) ExportArticle(c *gin.Context) {
	var ids []uint
	if err := c.ShouldBindJSON(&ids); err != nil {
		util.ResponseError(c, errors.ErrInvalidParams.WithMsg(err.Error()))
		return
	}
	
	// 调用Service层导出Markdown文件
	urls, err := h.svc.ExportArticles(c.Request.Context(), ids)
	if err != nil {
		util.ResponseError(c, err)
		return
	}
	util.ResponseSuccess(c, urls)
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
