package scheduler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/aurora-go/aurora/internal/model"
	"gorm.io/gorm"
)

// BaiduSeoJob 百度SEO URL推送任务
// 对标 Java AuroraQuartz.baiduSeo() (Cron: 0 0/10 * * * ?, ID=84)
//
// 业务逻辑:
// 1. 查询所有已发布的文章ID列表
// 2. 构建完整URL列表 (siteUrl + /articles/{id})
// 3. POST到百度推送API (http://data.zz.baidu.com/urls)
// 4. 记录推送结果
//
// 注意: 需要在配置中设置 BAIDU_SEO_TOKEN 环境变量
type BaiduSeoJob struct {
	db      *gorm.DB
	siteURL string
}

// NewBaiduSeoJob 创建百度SEO推送任务实例
func NewBaiduSeoJob(db *gorm.DB, siteURL string) *BaiduSeoJob {
	return &BaiduSeoJob{db: db, siteURL: siteURL}
}

// BaiduSEOResponse 百度推送API响应
type BaiduSEOResponse struct {
	Success    int               `json:"success"`            // 成功推送数
	Remain     int               `json:"remain"`             // 当日剩余配额
	NotSame    []string           `json:"not_same,omitempty"` // 不在此站点URL
	NotValid   []string           `json:"not_valid,omitempty"` // 非法URL
	Error      int               `json:"error,omitempty"`      // 错误码
	Message    string            `json:"message,omitempty"`    // 错误信息
}

// Run 执行百度SEO推送
func (j *BaiduSeoJob) Run(ctx context.Context) error {
	// Step 1: 查询所有已发布文章的ID (对标Java articleService.list().stream().map(Article::getId))
	var articles []model.Article
	if err := j.db.WithContext(ctx).
		Select("id").
		Where("is_delete = 0 AND status = 1"). // 未删除且公开
		Find(&articles).Error; err != nil {
		return fmt.Errorf("failed to query articles for SEO: %w", err)
	}

	if len(articles) == 0 {
		slog.Info("没有可推送的文章")
		return nil
	}

	// Step 2: 构建URL列表 (对标Java urlsBuilder.append(siteUrl + "/articles/" + id))
	var urlBuf bytes.Buffer
	for _, a := range articles {
		urlBuf.WriteString(j.siteURL)
		urlBuf.WriteString("/articles/")
		urlBuf.WriteString(fmt.Sprintf("%d", a.ID))
		urlBuf.WriteByte('\n')
	}
	urls := urlBuf.String()

	// Step 3: POST到百度推送API (对标Java restTemplate.postForObject(baiduApiUrl, entity, String.class))
	token := getBaiduSEOToken()
	apiURL := fmt.Sprintf("http://data.zz.baidu.com/urls?site=%s&token=%s", j.siteURL, token)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(urls))
	if err != nil {
		return fmt.Errorf("failed to create baidu request: %w", err)
	}

	// 设置请求头 (对标Java HttpHeaders)
	req.Header.Set("Host", "data.zz.baidu.com")
	req.Header.Set("User-Agent", "curl/7.12.1")
	req.Header.Set("Content-Type", "text/plain")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call baidu API: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// Step 4: 解析响应
	var baiduResp BaiduSEOResponse
	if err := json.Unmarshal(body, &baiduResp); err == nil {
		slog.Info("百度SEO推送结果",
			"success", baiduResp.Success,
			"remain", baiduResp.Remain,
			"total", len(articles),
		)
	} else {
		slog.Info("百度SEO推送原始响应", "status", resp.Status, "body", string(body))
	}

	return nil
}

// getBaiduSEOToken 获取百度推送Token (优先从环境变量读取)
func getBaiduSEOToken() string {
	// 可通过环境变量或后续扩展为配置项
	// 默认返回占位符，实际部署时需替换为真实Token
	token := "" // TODO: 从config或环境变量读取
	if token == "" {
		token = "YOUR_BAIDU_TOKEN" // 占位符
	}
	return token
}
