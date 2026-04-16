package strategy

import (
	"context"
	"log/slog"
	"strings"

	"github.com/aurora-go/aurora/internal/model"

	"gorm.io/gorm"
)

// MySQLSearchStrategy MySQL/LIKE搜索实现 (对标Java MySqlSearchStrategyImpl)
// 作为ES不可用时的降级方案，使用SQL LIKE模糊匹配 + 手动高亮模拟
type MySQLSearchStrategy struct {
	db *gorm.DB
}

// NewMySQLSearchStrategy 创建MySQL搜索策略实例
func NewMySQLSearchStrategy(db *gorm.DB) *MySQLSearchStrategy {
	return &MySQLSearchStrategy{db: db}
}

// SearchArticles 执行MySQL LIKE模糊搜索 (对标Java MySqlSearchStrategyImpl.searchArticle)
//
// 搜索流程:
//  1. 参数校验 → 空关键词返回空列表
//  2. GORM查询 → WHERE is_delete=0 AND status=1 AND (title LIKE OR content LIKE)
//  3. 手动高亮 → 在内容中定位关键词位置，截取上下文，用<mark>包裹
//  4. 大小写不敏感匹配（先尝试lower再尝试upper，对标Java双重查找）
//
// 性能说明:
//  - LIKE '%keyword%' 无法利用索引，适合中小数据量(<10万条)
//  - 大数据量场景建议启用ES搜索引擎
func (s *MySQLSearchStrategy) SearchArticles(ctx context.Context, keywords string, current, size int) ([]map[string]interface{}, int, error) {
	keywords = strings.TrimSpace(keywords)
	if keywords == "" {
		return []map[string]interface{}{}, 0, nil
	}

	lowerKeywords := strings.ToLower(keywords)
	upperKeywords := strings.ToUpper(keywords)

	var articles []model.Article
	var total int64

	// 对标Java MyBatis-Plus查询:
	// articleMapper.selectList(new LambdaQueryWrapper<Article>()
	//     .eq(Article::getIsDelete, FALSE)
	//     .eq(Article::getStatus, PUBLIC.getStatus())
	//     .and(i -> i.like(Article::getArticleTitle, keywords).or().like(Article::getArticleContent, keywords)));
	query := s.db.WithContext(ctx).
		Model(&model.Article{}).
		Where("is_delete = ? AND status = ?", 0, 1).
		Where(
			s.db.Where("article_title LIKE ?", "%"+keywords+"%").
				Or("article_content LIKE ?", "%"+keywords+"%"),
		)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		slog.Error("MySQL search count failed", "error", err)
		return []map[string]interface{}{}, 0, err
	}

	// 分页查询
	offset := (current - 1) * size
	if err := query.Offset(offset).Limit(size).Find(&articles).Error; err != nil {
		slog.Error("MySQL search failed", "error", err)
		return []map[string]interface{}{}, 0, err
	}

	// 转换并手动添加高亮（对标Java stream+map操作）
	results := make([]map[string]interface{}, 0, len(articles))
	for _, item := range articles {
		result := s.highlightArticle(item, lowerKeywords, upperKeywords)
		if result != nil { // 过滤掉无法定位关键词的记录（对标Java filter(Objects::nonNull)）
			results = append(results, *result)
		}
	}

	return results, int(total), nil
}

// highlightArticle 为文章手动添加高亮标记 (对标Java stream.map操作)
//
// 对标Java逻辑:
//  1. 在 articleContent 中定位 keywords 位置（大小写不敏感）
//  2. 截取关键词前后各15~35字符的上下文
//  3. 用 <mark> 包裹关键词
//  4. 同样处理 articleTitle
func (s *MySQLSearchStrategy) highlightArticle(
	item model.Article, lowerKeywords, upperKeywords string,
) *map[string]interface{} {
	content := item.ArticleContent

	// 尝试在内容中定位关键词（大小写不敏感）
	isLowerCase := true
	contentIndex := indexOfIgnoreCase(content, lowerKeywords)
	if contentIndex == -1 {
		contentIndex = indexOfIgnoreCase(content, upperKeywords)
		if contentIndex != -1 {
			isLowerCase = false
		}
	}

	if contentIndex == -1 {
		// 内容中未找到关键词（可能被截断或编码问题），跳过此记录
		// 对标Java: return null (会被filter过滤)
		return nil
	}

	// 截取上下文并添加高亮（对标Java preText/postText截取逻辑）
	highlightedContent := extractAndHighlight(
		content, contentIndex, lowerKeywords, upperKeywords, isLowerCase, 15, 35,
	)

	// 标题高亮处理
	title := item.ArticleTitle
	titleIsLower := true
	titleIndex := indexOfIgnoreCase(title, lowerKeywords)
	if titleIndex == -1 {
		titleIndex = indexOfIgnoreCase(title, upperKeywords)
		if titleIndex != -1 {
			titleIsLower = false
		}
	}

	var highlightedTitle string
	if titleIndex >= 0 {
		targetKw := lowerKeywords
		if !titleIsLower {
			targetKw = upperKeywords
		}
		highlightedTitle = strings.ReplaceAll(title, targetKw, PreTag+targetKw+PostTag)
	} else {
		// 标题中未找到关键词（可能标题不含该词但内容包含），保持原标题
		highlightedTitle = title
	}

	result := map[string]interface{}{
		"id":             item.ID,
		"articleTitle":   highlightedTitle,
		"articleContent": highlightedContent,
	}

	return &result
}

// ==================== 工具函数 ====================

// indexOfIgnoreCase 大小写不敏感的子串定位（对标Java String.indexOf）
func indexOfIgnoreCase(s, substr string) int {
	return strings.Index(strings.ToLower(s), strings.ToLower(substr))
}

// extractAndHighlight 从文本中提取关键词上下文并添加高亮标记
//
// 对标Java:
//   int preIndex = contentIndex > 15 ? contentIndex - 15 : 0;
//   String preText = item.getArticleContent().substring(preIndex, contentIndex);
//   int last = contentIndex + keywords.length();
//   int postLength = item.getArticleContent().length() - last;
//   int postIndex = postLength > 35 ? last + 35 : last + postLength;
//   String postText = item.getArticleContent().substring(contentIndex, postIndex);
//   articleContent = (preText + postText).replaceAll(keywords.toLowerCase(), PRE_TAG + keywords + POST_TAG);
func extractAndHighlight(
	text string, index int,
	lowerKW, upperKW string, isLower bool,
	preLen, postLen int,
) string {
	runes := []rune(text)
	totalLen := len(runes)

	// 计算前文截取范围（对标Java Math.max(0, index-15)）
	preStart := index - preLen
	if preStart < 0 {
		preStart = 0
	}

	// 计算后文截取范围（对标Java Math.min(length, index+kwLen+35)）
	kwLen := len([]rune(lowerKW))
	postEnd := index + kwLen + postLen
	if postEnd > totalLen {
		postEnd = totalLen
	}

	// 截取上下文
	extracted := string(runes[preStart:postEnd])

	// 替换为高亮版本
	targetKW := lowerKW
	if !isLower {
		targetKW = upperKW
	}

	return strings.ReplaceAll(extracted, targetKW, PreTag+targetKW+PostTag)
}
