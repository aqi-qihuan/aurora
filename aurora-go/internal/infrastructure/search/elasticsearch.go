package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/aurora-go/aurora/internal/config"
	"github.com/aurora-go/aurora/internal/model"
)

// ESClient Elasticsearch 客户端封装
// 支持ES 7.x/8.x双版本兼容（对标Java版olivere/elastic + ES 8.13.4）
type ESClient struct {
	baseURL     string
	username    string
	password    string
	httpClient  *http.Client
	indexPrefix string
	mu          sync.RWMutex // 保护并发写操作
}

var Client *ESClient

// InitElasticsearch 初始化 Elasticsearch 连接
func InitElasticsearch(cfg *config.ESConfig) error {
	// 从URLs列表提取主节点地址
	host := cfg.GetPrimaryURL()
	if host == "" {
		return fmt.Errorf("elasticsearch URL is not configured")
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 10 // 默认10秒
	}

	indexPrefix := cfg.IndexPrefix

	Client = &ESClient{
		baseURL:     strings.TrimSuffix(host, "/"),
		username:    cfg.Username,
		password:    cfg.Password,
		indexPrefix: indexPrefix,
		httpClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}

	// 验证连接（调用ES _cluster/health）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	health, err := Client.Health(ctx)
	if err != nil {
		slog.Error("Failed to connect to Elasticsearch", "error", err)
		Client = nil // 确保降级
		return err
	}

	slog.Info("Elasticsearch connected successfully",
		"cluster_name", health.ClusterName,
		"status", health.Status,
	)
	return nil
}

// Health 检查ES集群健康状态
func (c *ESClient) Health(ctx context.Context) (*ESHealthResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "_cluster/health", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var health ESHealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return nil, fmt.Errorf("failed to decode health response: %w", err)
	}
	return &health, nil
}

// CreateIndex 创建索引（如果不存在）
// 对标 Java 版 EsUtils.createIndex() + ik_max_word 分词器配置
func (c *ESClient) CreateIndex(ctx context.Context, indexName string, mapping map[string]interface{}) error {
	fullPath := indexName

	// 先检查索引是否存在
	exists, err := c.IndexExists(ctx, indexName)
	if err != nil {
		return err
	}
	if exists {
		slog.Debug("Index already exists, skipping creation", "index", indexName)
		return nil
	}

	body, _ := json.Marshal(mapping)
	resp, err := c.doRequest(ctx, http.MethodPut, fullPath, bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create index %s: status=%d body=%s",
			indexName, resp.StatusCode, string(bodyBytes))
	}

	slog.Info("ES index created successfully", "index", indexName)
	return nil
}

// IndexExists 检查索引是否存在
func (c *ESClient) IndexExists(ctx context.Context, indexName string) (bool, error) {
	resp, err := c.doRequest(ctx, http.MethodHead, indexName, nil)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		return false, fmt.Errorf("unexpected status checking index existence: %d", resp.StatusCode)
	}
}

// IndexDocument 索引单个文档（新增/更新）
func (c *ESClient) IndexDocument(ctx context.Context, indexName string, docID string, doc interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	path := fmt.Sprintf("%s/_doc/%s", indexName, docID)
	body, _ := json.Marshal(doc)
	resp, err := c.doRequest(ctx, http.MethodPost, path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to index document %s in %s: status=%d body=%s",
			docID, indexName, resp.StatusCode, string(bodyBytes))
	}
	return nil
}

// DeleteDocument 删除文档
func (c *ESClient) DeleteDocument(ctx context.Context, indexName string, docID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	path := fmt.Sprintf("%s/_doc/%s", indexName, docID)
	resp, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil // 文档不存在也算成功（幂等）
	}
	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete document %s from %s: status=%d body=%s",
			docID, indexName, resp.StatusCode, string(bodyBytes))
	}
	return nil
}

// DeleteIndex 删除整个索引（用于ES全量同步前清理）
func (c *ESClient) DeleteIndex(ctx context.Context, indexName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	resp, err := c.doRequest(ctx, http.MethodDelete, indexName, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 索引不存在时返回404，视为成功
	if resp.StatusCode == http.StatusNotFound {
		slog.Debug("Index does not exist, skip delete", "index", indexName)
		return nil
	}
	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete index %s: status=%d body=%s",
			indexName, resp.StatusCode, string(bodyBytes))
	}

	slog.Info("ES index deleted successfully", "index", indexName)
	return nil
}

// BulkIndex 批量索引文档（用于全量同步）
func (c *ESClient) BulkIndex(ctx context.Context, indexName string, docs []model.Article) error {
	if len(docs) == 0 {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	var buf bytes.Buffer
	for _, article := range docs {
		// 构建bulk操作格式
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": indexName,
				"_id":    fmt.Sprintf("%d", article.ID),
			},
		}
		metaJSON, _ := json.Marshal(meta)
		buf.Write(metaJSON)
		buf.WriteByte('\n')
		docJSON, _ := json.Marshal(article)
		buf.Write(docJSON)
		buf.WriteByte('\n')
	}

	resp, err := c.doRequest(ctx, http.MethodPost, "_bulk", &buf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bulk index failed: status=%d body=%s", resp.StatusCode, string(bodyBytes))
	}

	slog.Info("Bulk indexed documents", "count", len(docs), "index", indexName)
	return nil
}

// Search 执行搜索查询（对标 Java 版 EsSearchStrategyImpl）
// 支持: 全文搜索(ik_max_word)/高亮显示/分页/排序/过滤条件
func (c *ESClient) Search(ctx context.Context, indexName string, query *ESSearchRequest) (*ESSearchResponse, error) {
	path := indexName + "/_search"
	body, _ := json.Marshal(query)

	resp, err := c.doRequest(ctx, http.MethodPost, path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ESSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	return &result, nil
}

// Count 统计文档数量
func (c *ESClient) Count(ctx context.Context, indexName string, query interface{}) (int64, error) {
	path := indexName + "/_count"
	body, _ := json.Marshal(query)

	resp, err := c.doRequest(ctx, http.MethodPost, path, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var countResp struct {
		Count int64 `json:"count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&countResp); err != nil {
		return 0, fmt.Errorf("failed to decode count response: %w", err)
	}
	return countResp.Count, nil
}

// doRequest 发送HTTP请求到ES
func (c *ESClient) doRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	url := c.baseURL + "/" + path
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	// 认证头（Basic Auth）
	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return c.httpClient.Do(req)
}

// GetFullIndexName 获取完整索引名（含前缀）
func (c *ESClient) GetFullIndexName(name string) string {
	if c.indexPrefix == "" {
		return name
	}
	return c.indexPrefix + "_" + name
}

// ==================== 数据结构定义 ====================

// ESSearchRequest ES搜索请求体
type ESSearchRequest struct {
	From           int                    `json:"from,omitempty"`
	Size           int                    `json:"size,omitempty"`
	Sort           []interface{}          `json:"sort,omitempty"`
	Query          map[string]interface{} `json:"query,omitempty"`
	Highlight      *ESSearchHighlight     `json:"highlight,omitempty"`
	Source         *ESSourceFilter        `json:"_source,omitempty"`
	Aggregations   map[string]interface{} `json:"aggs,omitempty"`
	Suggest        map[string]interface{} `json:"suggest,omitempty"`
}

// ESSearchHighlight 高亮配置
type ESSearchHighlight struct {
	PreTags       []string             `json:"pre_tags"`
	PostTags      []string             `json:"post_tags"`
	Fields        map[string]interface{} `json:"fields"`
	FragmentSize  int                  `json:"fragment_size,omitempty"`   // 默认100
	NumberOfFragments int               `json:"number_of_fragments"`      // 默认1
}

// ESSourceFilter 返回字段过滤
type ESSourceFilter struct {
	Includes []string `json:"includes"`
	Excludes []string `json:"excludes"`
}

// ESSearchResponse ES搜索响应
type ESSearchResponse struct {
	Took     int                      `json:"took"`
	TimedOut bool                     `json:"timed_out"`
	Hits     ESSearchHits             `json:"hits"`
}

// ESSearchHits 搜索命中结果
type ESSearchHits struct {
	Total ESSearchTotal `json:"total"`
	MaxScore float64    `json:"max_score"`
	Hits   []ESSearchHit `json:"hits"`
}

// ESSearchTotal 总数信息
type ESSearchTotal struct {
	Value    int64 `json:"value"`
	Relation string `json:"relation"` // "eq" = exact, "gte" = approximate
}

// ESSearchHit 单条命中记录
type ESSearchHit struct {
	ID         string                 `json:"_id"`
	Score      float64                `json:"_score"`
	Source     json.RawMessage        `json:"_source"`
	Highlights map[string][]string    `json:"highlight"`
}

// ESHealthResponse 集群健康状态响应
type ESHealthResponse struct {
	ClusterName string `json:"cluster_name"`
	Status      string `json:"status"`      // green/yellow/red
	NumberOfNodes int  `json:"number_of_nodes"`
	NumberOfDataNodes int `json:"number_of_data_nodes"`
}
