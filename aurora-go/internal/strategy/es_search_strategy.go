package strategy

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/infrastructure/search"
)

// ESSearchStrategy Elasticsearch搜索实现 (对标Java EsSearchStrategyImpl)
// 使用 ik_max_word 分词器进行中文全文搜索，支持标题/内容双字段匹配和高亮显示
type ESSearchStrategy struct {
	client *search.ESClient
}

// NewESSearchStrategy 创建ES搜索策略实例
func NewESSearchStrategy(client *search.ESClient) *ESSearchStrategy {
	return &ESSearchStrategy{client: client}
}

// SearchArticle 执行ES全文搜索 (对标Java EsSearchStrategyImpl.searchArticle)
//
// 搜索流程:
//  1. 参数校验 → 空关键词返回空列表
//  2. 构建BoolQuery → MUST(标题OR内容) + MUST(未删除) + MUST(已发布)
//  3. 高亮配置 → 标题完整返回(mark标签) + 内容片段化(50字符×3段)
//  4. 执行查询 → 解析JSON响应
//  5. 错误降级 → ES异常时返回空列表(不抛出异常，对标Java降级机制)
func (s *ESSearchStrategy) SearchArticle(ctx context.Context, keywords string) ([]dto.ArticleSearchDTO, error) {
	if strings.TrimSpace(keywords) == "" {
		return []dto.ArticleSearchDTO{}, nil
	}

	query := s.buildQuery(keywords)
	resp, err := s.client.Search(ctx, "article", query)
	if err != nil {
		slog.Error("Elasticsearch search failed",
			"keywords", keywords,
			"error", err,
			"hint", "请检查: 1)IK分词器是否安装; 2)article索引是否存在; 3)ES集群状态",
		)
		// ★ 降级: 异常时返回空列表，不抛出（对标Java catch→return new ArrayList<>()）
		return []dto.ArticleSearchDTO{}, nil
	}

	return s.parseResponse(resp), nil
}

// buildQuery 构建ES BoolQuery搜索请求 (对标Java buildQuery方法)
//
// 对标Java代码:
//   BoolQuery boolQuery = BoolQuery.of(b -> b
//       .must(m -> m.bool(bb -> bb
//           .should(s -> s.match(t -> t.field("articleTitle").query(keywords)))
//           .should(s -> s.match(t -> t.field("articleContent").query(keywords)))
//       ))
//       .must(m -> m.term(t -> t.field("isDelete").value(FALSE)))
//       .must(m -> m.term(t -> t.field("status").value(PUBLIC.getStatus())))
//   );
func (s *ESSearchStrategy) buildQuery(keywords string) *search.ESSearchRequest {
	return &search.ESSearchRequest{
		Query: map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					map[string]interface{}{
						"bool": map[string]interface{}{
							"should": []interface{}{
								map[string]interface{}{"match": map[string]string{"articleTitle": keywords}},
								map[string]interface{}{"match": map[string]string{"articleContent": keywords}},
							},
							"minimum_should_match": 1,
						},
					},
					// 过滤条件: 未删除 + 已发布
					map[string]interface{}{"term": map[string]interface{}{"isDelete": 0}},
					map[string]interface{}{"term": map[string]interface{}{"status": 1}},
				},
			},
		},
		Highlight: &search.ESSearchHighlight{
			PreTags:  []string{PreTag},  // <mark>
			PostTags: []string{PostTag}, // </mark>
			Fields: map[string]interface{}{
				// 标题字段: fragmentSize=0 返回完整标题
				"articleTitle": map[string]interface{}{
					"fragment_size": 0,
				},
				// 内容字段: 最多返回3个片段,每段50字符
				"articleContent": map[string]interface{}{
					"fragment_size":        50,
					"number_of_fragments": 3,
				},
			},
		},
	}
}

// parseResponse 解析ES搜索响应为DTO列表 (对标Java search方法中的stream解析)
//
// 处理逻辑:
//  1. 遍历 hits.hits[]
//  2. 解析 _source JSON → ArticleSearchDTO
//  3. 用 highlight 字段覆盖原始标题/内容（添加 <mark> 高亮标记）
//  4. 过滤掉解析失败的记录
func (s *ESSearchStrategy) parseResponse(resp *search.ESSearchResponse) []dto.ArticleSearchDTO {
	if resp == nil || len(resp.Hits.Hits) == 0 {
		return []dto.ArticleSearchDTO{}
	}

	results := make([]dto.ArticleSearchDTO, 0, len(resp.Hits.Hits))
	for _, hit := range resp.Hits.Hits {
		article, err := parseHitToDTO(hit)
		if err != nil {
			slog.Debug("Failed to parse ES hit",
				"id", hit.ID,
				"error", err,
			)
			continue
		}
		results = append(results, *article)
	}
	return results
}

// parseHitToDTO 将单条ES命中记录解析为DTO
func parseHitToDTO(hit search.ESSearchHit) (*dto.ArticleSearchDTO, error) {
	// 解析_source字段
	var source struct {
		ID             uint    `json:"id"`
		ArticleTitle   string  `json:"articleTitle"`
		ArticleContent string  `json:"articleContent"`
		IsDelete       int     `json:"isDelete"`
		Status         int     `json:"status"`
	}

	if err := json.Unmarshal(hit.Source, &source); err != nil {
		return nil, err
	}

	dto := &dto.ArticleSearchDTO{
		ID:             source.ID,
		ArticleTitle:   source.ArticleTitle,
		ArticleContent: source.ArticleContent,
		Score:          hit.Score,
	}

	// 处理高亮字段覆盖（对标Java hit.highlight()处理）
	if len(hit.Highlights) > 0 {
		// 标题高亮: 取第一个高亮结果
		if titleHL, ok := hit.Highlights["articleTitle"]; ok && len(titleHL) > 0 {
			dto.ArticleTitle = titleHL[0]
		}
		// 内容高亮: 取最后一个片段（对标Java contentHighlights.get(size-1)）
		if contentHL, ok := hit.Highlights["articleContent"]; ok && len(contentHL) > 0 {
			dto.ArticleContent = contentHL[len(contentHL)-1]
			// 提取所有高亮片段
			for _, hl := range contentHL {
				dto.Highlight = append(dto.Highlight, hl)
			}
		}
	}

	return dto, nil
}
