package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// globalESService 全局 ES 服务实例（单例）
var globalESService *ESService

// SetGlobalESService 设置全局 ES 服务实例
func SetGlobalESService(svc *ESService) {
	globalESService = svc
}

// GetGlobalESService 获取全局 ES 服务实例
func GetGlobalESService() *ESService {
	return globalESService
}

// ESService Elasticsearch 服务
type ESService struct {
	client    *elasticsearch.Client
	indexName string
}

// NewESService 创建 ES 服务实例
func NewESService(urls []string, username, password, indexName string) (*ESService, error) {
	cfg := elasticsearch.Config{
		Addresses: urls,
		Username:  username,
		Password:  password,
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("创建 ES 客户端失败: %w", err)
	}

	// 测试连接
	res, err := client.Info()
	if err != nil {
		return nil, fmt.Errorf("ES 连接测试失败: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("ES 连接错误: %s", res.String())
	}

	log.Printf("✅ Elasticsearch 连接成功: %v", urls)

	return &ESService{
		client:    client,
		indexName: indexName,
	}, nil
}

// GetClient 获取 ES 客户端
func (s *ESService) GetClient() *elasticsearch.Client {
	return s.client
}

// InitArticleIndex 初始化 article 索引（对标 Java ElasticsearchIndexInitializer）
func (s *ESService) InitArticleIndex(ctx context.Context) error {
	// 检查索引是否存在
	exists, err := s.IndexExists(ctx, s.indexName)
	if err != nil {
		return fmt.Errorf("检查索引存在性失败: %w", err)
	}

	if exists {
		log.Printf("📋 article 索引已存在")
		return nil
	}

	log.Printf("📋 article 索引不存在，开始创建...")

	// 定义索引 Mapping（对标 Java 配置）
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type":  "integer",
					"store": true,
				},
				"articleTitle": map[string]interface{}{
					"type":            "text",
					"analyzer":        "ik_max_word",
					"search_analyzer": "ik_smart",
					"store":           true,
				},
				"articleContent": map[string]interface{}{
					"type":            "text",
					"analyzer":        "ik_max_word",
					"search_analyzer": "ik_smart",
					"store":           true,
				},
				"isDelete": map[string]interface{}{
					"type":  "integer",
					"store": true,
				},
				"status": map[string]interface{}{
					"type":  "integer",
					"store": true,
				},
			},
		},
	}

	mappingJSON, err := json.Marshal(mapping)
	if err != nil {
		return fmt.Errorf("序列化 mapping 失败: %w", err)
	}

	// 创建索引
	req := esapi.IndicesCreateRequest{
		Index: s.indexName,
		Body:  bytes.NewReader(mappingJSON),
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("创建索引请求失败: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("创建索引失败: %s", res.String())
	}

	log.Printf("✅ article 索引创建成功，已配置 ik_max_word 分词器")
	log.Printf("========== ES 索引配置信息 ==========")
	log.Printf("articleTitle: analyzer=ik_max_word, searchAnalyzer=ik_smart")
	log.Printf("articleContent: analyzer=ik_max_word, searchAnalyzer=ik_smart")
	log.Printf("====================================")

	return nil
}

// IndexExists 检查索引是否存在
func (s *ESService) IndexExists(ctx context.Context, indexName string) (bool, error) {
	req := esapi.IndicesExistsRequest{
		Index: []string{indexName},
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	return res.StatusCode == 200, nil
}

// IndexArticle 索引文章（新增或更新）
func (s *ESService) IndexArticle(ctx context.Context, articleID uint, title, content string, isDelete, status int8) error {
	doc := map[string]interface{}{
		"id":             articleID,
		"articleTitle":   title,
		"articleContent": content,
		"isDelete":       isDelete,
		"status":         status,
	}

	docJSON, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("序列化文档失败: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      s.indexName,
		DocumentID: fmt.Sprintf("%d", articleID),
		Body:       bytes.NewReader(docJSON),
		Refresh:    "true", // 立即刷新，使文档可搜索
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("索引文章请求失败: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("索引文章失败: %s", res.String())
	}

	return nil
}

// DeleteArticle 删除文章索引
func (s *ESService) DeleteArticle(ctx context.Context, articleID uint) error {
	req := esapi.DeleteRequest{
		Index:      s.indexName,
		DocumentID: fmt.Sprintf("%d", articleID),
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("删除文章索引请求失败: %w", err)
	}
	defer res.Body.Close()

	// 404 表示文档不存在，不算错误
	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("删除文章索引失败: %s", res.String())
	}

	return nil
}

// SearchArticles 搜索文章（对标 Java EsSearchStrategyImpl）
func (s *ESService) SearchArticles(ctx context.Context, keywords string, page, size int) ([]map[string]interface{}, int, error) {
	if keywords == "" {
		return []map[string]interface{}{}, 0, nil
	}

	// 构建查询（对标 Java EsSearchStrategyImpl.buildQuery）
	// 注意：不指定 analyzer，使用 field mapping 的 search_analyzer (ik_smart)
	query := map[string]interface{}{
		"bool": map[string]interface{}{
			"must": []map[string]interface{}{
				{
					"bool": map[string]interface{}{
						"should": []map[string]interface{}{
							{
								"match": map[string]interface{}{
									"articleTitle": map[string]interface{}{
										"query": keywords,
										"boost": 2.0, // 标题权重更高
									},
								},
							},
							{
								"match": map[string]interface{}{
									"articleContent": map[string]interface{}{
										"query": keywords,
									},
								},
							},
						},
					},
				},
				{
					"term": map[string]interface{}{
						"isDelete": 0,
					},
				},
				{
					"term": map[string]interface{}{
						"status": 1, // 只搜索公开文章
					},
				},
			},
		},
	}

	// 构建高亮配置（对标Java: preTags="<mark>", postTags="</mark>"）
	highlight := map[string]interface{}{
		"pre_tags":  []string{"<mark>"},
		"post_tags": []string{"</mark>"},
		"fields": map[string]interface{}{
			"articleTitle": map[string]interface{}{
				"fragment_size": 0, // 不截断，返回完整标题
			},
			"articleContent": map[string]interface{}{
				"fragment_size":         50,
				"number_of_fragments":   3, // 最多返回 3 个片段
			},
		},
	}

	// 构建搜索请求（对标Java: 不指定 from/size，不指定 _source 过滤）
	// Java 版本不限制 _source，ES 默认返回所有字段
	searchBody := map[string]interface{}{
		"query":     query,
		"highlight": highlight,
	}

	bodyJSON, err := json.Marshal(searchBody)
	if err != nil {
		return nil, 0, fmt.Errorf("序列化搜索请求失败: %w", err)
	}

	req := esapi.SearchRequest{
		Index: []string{s.indexName},
		Body:  bytes.NewReader(bodyJSON),
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return nil, 0, fmt.Errorf("搜索请求失败: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, 0, fmt.Errorf("搜索失败: %s", res.String())
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, 0, fmt.Errorf("解析搜索响应失败: %w", err)
	}

	hits := result["hits"].(map[string]interface{})
	total := int(hits["total"].(map[string]interface{})["value"].(float64))

	hitList := hits["hits"].([]interface{})
	articles := make([]map[string]interface{}, 0, len(hitList))

	for _, hit := range hitList {
		h := hit.(map[string]interface{})
		source := h["_source"].(map[string]interface{})
		highlight := h["highlight"]

		article := map[string]interface{}{
			"id":             source["id"],
			"articleTitle":   source["articleTitle"],
			"articleContent": source["articleContent"],
		}

		// 处理高亮
		if highlight != nil {
			hl := highlight.(map[string]interface{})
			if titles, ok := hl["articleTitle"].([]interface{}); ok && len(titles) > 0 {
				article["articleTitle"] = titles[0]
			}
			if contents, ok := hl["articleContent"].([]interface{}); ok && len(contents) > 0 {
				// 取最后一个高亮片段
				article["articleContent"] = contents[len(contents)-1]
			}
		}

		articles = append(articles, article)
	}

	return articles, total, nil
}

// BulkIndexArticles 批量索引文章（用于全量同步）
func (s *ESService) BulkIndexArticles(ctx context.Context, articles []map[string]interface{}) error {
	if len(articles) == 0 {
		return nil
	}

	var buf bytes.Buffer

	for _, article := range articles {
		// 元数据行
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": s.indexName,
				"_id":    fmt.Sprintf("%v", article["id"]),
			},
		}
		metaJSON, _ := json.Marshal(meta)
		buf.Write(metaJSON)
		buf.WriteByte('\n')

		// 数据行
		docJSON, _ := json.Marshal(article)
		buf.Write(docJSON)
		buf.WriteByte('\n')
	}

	req := esapi.BulkRequest{
		Index: s.indexName,
		Body:  &buf,
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("批量索引请求失败: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("批量索引失败: %s", res.String())
	}

	log.Printf("✅ 批量索引 %d 篇文章成功", len(articles))
	return nil
}

// HealthCheck 健康检查
func (s *ESService) HealthCheck(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res, err := s.client.Info(s.client.Info.WithContext(ctx))
	if err != nil {
		return false
	}
	defer res.Body.Close()

	return !res.IsError()
}
