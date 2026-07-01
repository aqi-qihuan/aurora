package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/aurora-go/aurora/internal/config"
	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/model"
	"github.com/aurora-go/aurora/internal/vo"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const portfolioFeaturedKey = "portfolio:featured"

// PortfolioService 作品集业务逻辑（GitHub 仓库同步 + 前后台查询）
type PortfolioService struct {
	db        *gorm.DB
	rdb       *redis.Client
	githubCfg config.GitHubConfig
	httpCli   *http.Client
}

func NewPortfolioService(db *gorm.DB, rdb *redis.Client, githubCfg config.GitHubConfig) *PortfolioService {
	return &PortfolioService{
		db:        db,
		rdb:       rdb,
		githubCfg: githubCfg,
		httpCli:   &http.Client{Timeout: 15 * time.Second},
	}
}

// ListFeatured 首页置顶作品（置顶优先 + 排序 + star），最多 limit 条
func (s *PortfolioService) ListFeatured(ctx context.Context, limit int) ([]dto.PortfolioDTO, error) {
	if limit <= 0 || limit > 20 {
		limit = 6
	}
	var list []model.Portfolio
	err := s.db.WithContext(ctx).
		Where("is_visible = 1").
		Order("is_featured DESC, sort DESC, stargazers_count DESC, repo_updated_at DESC").
		Limit(limit).
		Find(&list).Error
	if err != nil {
		return nil, fmt.Errorf("查询首页作品集失败: %w", err)
	}
	return toPortfolioDTOs(list), nil
}

// ListAll 前台分页查询（仅可见）
func (s *PortfolioService) ListAll(ctx context.Context, page dto.PageVO) (*dto.PageResultDTO, error) {
	var list []model.Portfolio
	var count int64
	q := s.db.WithContext(ctx).Model(&model.Portfolio{}).Where("is_visible = 1")
	if err := q.Count(&count).Error; err != nil {
		return nil, fmt.Errorf("统计作品集失败: %w", err)
	}
	if err := q.Order("is_featured DESC, sort DESC, stargazers_count DESC, repo_updated_at DESC").
		Limit(page.PageSize).Offset(page.GetOffset()).Find(&list).Error; err != nil {
		return nil, fmt.Errorf("查询作品集列表失败: %w", err)
	}
	return &dto.PageResultDTO{
		List:     toPortfolioDTOs(list),
		Count:    count,
		PageNum:  page.PageNum,
		PageSize: page.PageSize,
	}, nil
}

// ListAdmin 后台分页查询（含隐藏项）
func (s *PortfolioService) ListAdmin(ctx context.Context, cond dto.ConditionVO, page dto.PageVO) (*dto.PageResultDTO, error) {
	var list []model.Portfolio
	var count int64
	q := s.db.WithContext(ctx).Model(&model.Portfolio{})
	if cond.Keywords != "" {
		q = q.Where("name LIKE ? OR full_name LIKE ? OR description LIKE ?",
			"%"+cond.Keywords+"%", "%"+cond.Keywords+"%", "%"+cond.Keywords+"%")
	}
	if err := q.Count(&count).Error; err != nil {
		return nil, fmt.Errorf("统计作品集失败: %w", err)
	}
	if err := q.Order("is_featured DESC, sort DESC, stargazers_count DESC, repo_updated_at DESC").
		Limit(page.PageSize).Offset(page.GetOffset()).Find(&list).Error; err != nil {
		return nil, fmt.Errorf("查询作品集列表失败: %w", err)
	}
	return &dto.PageResultDTO{
		List:     toPortfolioAdminDTOs(list),
		Count:    count,
		PageNum:  page.PageNum,
		PageSize: page.PageSize,
	}, nil
}

// UpdatePortfolio 后台编辑（仅覆盖人工配置项：封面/排序/置顶/可见性）
func (s *PortfolioService) UpdatePortfolio(ctx context.Context, v vo.PortfolioVO) error {
	updates := map[string]interface{}{}
	if v.Cover != "" {
		updates["cover"] = v.Cover
	}
	updates["sort"] = v.Sort
	if v.IsFeatured != nil {
		updates["is_featured"] = *v.IsFeatured
	}
	if v.IsVisible != nil {
		updates["is_visible"] = *v.IsVisible
	}
	if len(updates) == 0 {
		return nil
	}
	result := s.db.WithContext(ctx).Model(&model.Portfolio{}).Where("id = ?", v.ID).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("更新作品集失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.ErrPortfolioNotFound
	}
	s.invalidateCache(ctx)
	return nil
}

// DeletePortfolios 批量删除
func (s *PortfolioService) DeletePortfolios(ctx context.Context, ids []uint) error {
	if len(ids) == 0 {
		return fmt.Errorf("作品ID列表不能为空")
	}
	result := s.db.WithContext(ctx).Delete(&model.Portfolio{}, ids)
	if result.Error != nil {
		return fmt.Errorf("批量删除作品失败: %w", result.Error)
	}
	s.invalidateCache(ctx)
	slog.Info("批量删除作品集", "count", result.RowsAffected)
	return nil
}

// SyncFromGitHub 从 GitHub API 同步仓库快照（定时任务/手动触发调用）
// upsert 策略：按 repo_id 唯一键更新动态字段，保留 cover/sort/is_featured/is_visible
func (s *PortfolioService) SyncFromGitHub(ctx context.Context) error {
	if !s.githubCfg.Enabled || s.githubCfg.Username == "" {
		slog.Warn("GitHub作品集同步未启用（github.enabled=false 或 username 为空），跳过")
		return nil
	}

	repos, err := s.fetchGitHubRepos(ctx, s.githubCfg.Username, s.githubCfg.Token)
	if err != nil {
		slog.Error("拉取GitHub仓库失败", "username", s.githubCfg.Username, "error", err)
		return errors.ErrPortfolioSyncFailed.WithMsg(err.Error())
	}

	exclude := parseExcludeSet(s.githubCfg.Exclude)
	syncedRepoIDs := make([]int64, 0, len(repos))
	synced := 0
	for _, r := range repos {
		// 仅过滤 archived 仓库；fork 仓库由 exclude 配置决定是否排除
		if r.Archived {
			continue
		}
		if _, ok := exclude[strings.ToLower(r.Name)]; ok {
			continue
		}
		if err := s.upsertPortfolio(ctx, r); err != nil {
			slog.Warn("upsert作品失败", "repo", r.FullName, "error", err)
			continue
		}
		syncedRepoIDs = append(syncedRepoIDs, r.ID)
		synced++
	}

	// 删除数据库中已不在 GitHub 仓库列表里的陈旧记录（仓库被删除/改名/exclude 后自动清理）
	if len(syncedRepoIDs) > 0 {
		if err := s.db.WithContext(ctx).Where("repo_id NOT IN ?", syncedRepoIDs).Delete(&model.Portfolio{}).Error; err != nil {
			slog.Warn("清理陈旧作品记录失败", "error", err)
		}
	}

	s.invalidateCache(ctx)
	slog.Info("GitHub作品集同步完成", "username", s.githubCfg.Username, "fetched", len(repos), "synced", synced, "kept_ids", len(syncedRepoIDs))
	return nil
}

// fetchGitHubRepos 调用 GitHub REST API 拉取用户仓库列表
func (s *PortfolioService) fetchGitHubRepos(ctx context.Context, username, token string) ([]githubRepo, error) {
	url := fmt.Sprintf("https://api.github.com/users/%s/repos?sort=updated&per_page=100&type=owner", username)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	// mercy-preview Accept 头使响应包含 topics 字段
	req.Header.Set("Accept", "application/vnd.github.mercy-preview+json")
	req.Header.Set("User-Agent", "aurora-blog")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := s.httpCli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求GitHub API失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("GitHub API返回 %d: %s", resp.StatusCode, string(body))
	}
	var repos []githubRepo
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, fmt.Errorf("解析GitHub响应失败: %w", err)
	}
	return repos, nil
}

// upsertPortfolio 按 repo_id upsert：存在则更新动态字段，不存在则插入
func (s *PortfolioService) upsertPortfolio(ctx context.Context, r githubRepo) error {
	topicsJSON := "[]"
	if len(r.Topics) > 0 {
		if b, err := json.Marshal(r.Topics); err == nil {
			topicsJSON = string(b)
		}
	}

	var existing model.Portfolio
	err := s.db.WithContext(ctx).Where("repo_id = ?", r.ID).First(&existing).Error
	if err != nil && !errors.IsStd(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("查询作品失败: %w", err)
	}

	if errors.IsStd(err, gorm.ErrRecordNotFound) {
		// 新增：默认可见、不置顶
		p := model.Portfolio{
			RepoID:          r.ID,
			Name:            r.Name,
			FullName:        r.FullName,
			Description:     r.Description,
			HtmlURL:         r.HTMLURL,
			Homepage:        r.Homepage,
			Language:        r.Language,
			StargazersCount: r.StargazersCount,
			ForksCount:      r.ForksCount,
			Topics:          topicsJSON,
			RepoCreatedAt:   parseGitHubTime(r.CreatedAt),
			RepoUpdatedAt:   parseGitHubTime(r.UpdatedAt),
			IsVisible:       1,
		}
		return s.db.WithContext(ctx).Create(&p).Error
	}

	// 已存在：仅更新动态字段，保留 cover/sort/is_featured/is_visible
	return s.db.WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
		"name":             r.Name,
		"full_name":        r.FullName,
		"description":      r.Description,
		"html_url":         r.HTMLURL,
		"homepage":         r.Homepage,
		"language":         r.Language,
		"stargazers_count": r.StargazersCount,
		"forks_count":      r.ForksCount,
		"topics":           topicsJSON,
		"repo_created_at":  parseGitHubTime(r.CreatedAt),
		"repo_updated_at":  parseGitHubTime(r.UpdatedAt),
	}).Error
}

func (s *PortfolioService) invalidateCache(ctx context.Context) {
	if s.rdb != nil {
		if err := s.rdb.Del(ctx, portfolioFeaturedKey).Err(); err != nil {
			slog.Warn("删除作品集Redis缓存失败", "error", err)
		}
	}
}

// ===== 内部类型与工具 =====

// githubRepo GitHub REST API 仓库响应字段（仅取需要的子集）
type githubRepo struct {
	ID              int64    `json:"id"`
	Name            string   `json:"name"`
	FullName        string   `json:"full_name"`
	Description     string   `json:"description"`
	HTMLURL         string   `json:"html_url"`
	Homepage        string   `json:"homepage"`
	Language        string   `json:"language"`
	StargazersCount int      `json:"stargazers_count"`
	ForksCount      int      `json:"forks_count"`
	Topics          []string `json:"topics"`
	Fork            bool     `json:"fork"`
	Archived        bool     `json:"archived"`
	CreatedAt       string   `json:"created_at"` // GitHub 返回 ISO8601 字符串
	UpdatedAt       string   `json:"updated_at"`
}

func parseGitHubTime(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil
	}
	return &t
}

func parseExcludeSet(s string) map[string]bool {
	m := map[string]bool{}
	if s == "" {
		return m
	}
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(strings.ToLower(p))
		if p != "" {
			m[p] = true
		}
	}
	return m
}

func toPortfolioDTOs(list []model.Portfolio) []dto.PortfolioDTO {
	out := make([]dto.PortfolioDTO, 0, len(list))
	for _, p := range list {
		out = append(out, dto.PortfolioDTO{
			ID:              p.ID,
			Name:            p.Name,
			FullName:        p.FullName,
			Description:     p.Description,
			HtmlURL:         p.HtmlURL,
			Homepage:        p.Homepage,
			Language:        p.Language,
			StargazersCount: p.StargazersCount,
			ForksCount:      p.ForksCount,
			Topics:          parseTopics(p.Topics),
			RepoUpdatedAt:   p.RepoUpdatedAt,
			Cover:           p.Cover,
			Sort:            p.Sort,
			IsFeatured:      p.IsFeatured,
		})
	}
	return out
}

func toPortfolioAdminDTOs(list []model.Portfolio) []dto.PortfolioAdminDTO {
	out := make([]dto.PortfolioAdminDTO, 0, len(list))
	for _, p := range list {
		out = append(out, dto.PortfolioAdminDTO{
			ID:              p.ID,
			RepoID:          p.RepoID,
			Name:            p.Name,
			FullName:        p.FullName,
			Description:     p.Description,
			HtmlURL:         p.HtmlURL,
			Homepage:        p.Homepage,
			Language:        p.Language,
			StargazersCount: p.StargazersCount,
			ForksCount:      p.ForksCount,
			Topics:          parseTopics(p.Topics),
			RepoCreatedAt:   p.RepoCreatedAt,
			RepoUpdatedAt:   p.RepoUpdatedAt,
			Cover:           p.Cover,
			Sort:            p.Sort,
			IsFeatured:      p.IsFeatured,
			IsVisible:       p.IsVisible,
			CreateTime:      p.CreateTime,
		})
	}
	return out
}

func parseTopics(s string) []string {
	if s == "" {
		return nil
	}
	var arr []string
	if err := json.Unmarshal([]byte(s), &arr); err != nil {
		return nil
	}
	return arr
}
