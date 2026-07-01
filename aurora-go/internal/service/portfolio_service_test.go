package service

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/model"
)

// ===== parseExcludeSet 测试 =====

func TestParseExcludeSet(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want map[string]bool
	}{
		{"空字符串", "", map[string]bool{}},
		{"单个仓库", "repo1", map[string]bool{"repo1": true}},
		{"多个仓库", "repo1,repo2,repo3", map[string]bool{"repo1": true, "repo2": true, "repo3": true}},
		{"带空格", " repo1 , repo2 , repo3 ", map[string]bool{"repo1": true, "repo2": true, "repo3": true}},
		{"大写转小写", "Repo1,REPO2", map[string]bool{"repo1": true, "repo2": true}},
		{"混合大小写", "Aurora-Blog,AURORA-GO", map[string]bool{"aurora-blog": true, "aurora-go": true}},
		{"尾部逗号", "repo1,repo2,", map[string]bool{"repo1": true, "repo2": true}},
		{"首部逗号", ",repo1,repo2", map[string]bool{"repo1": true, "repo2": true}},
		{"仅逗号", ",,,,", map[string]bool{}},
		{"仅空格", "   ", map[string]bool{}},
		{"含制表符", "repo1\t,repo2", map[string]bool{"repo1": true, "repo2": true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseExcludeSet(tt.s)
			if len(got) != len(tt.want) {
				t.Errorf("parseExcludeSet(%q) len = %d; want %d (got=%v want=%v)", tt.s, len(got), len(tt.want), got, tt.want)
				return
			}
			for k, v := range tt.want {
				if !got[k] {
					t.Errorf("parseExcludeSet(%q) missing key %q", tt.s, k)
				}
				_ = v // 期望值恒为 true
			}
		})
	}
}

// ===== parseGitHubTime 测试 =====

func TestParseGitHubTime(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		wantNil bool
		wantSec int64 // 期望的 unix 秒（仅 wantNil=false 时校验）
	}{
		{"空字符串", "", true, 0},
		{"无效格式", "not-a-time", true, 0},
		{"无效日期", "2026-13-45T99:99:99Z", true, 0},
		{"标准 RFC3339", "2026-06-01T12:00:00Z", false, 1780315200},
		{"带时区", "2026-06-01T12:00:00+08:00", false, 1780286400},
		{"UTC 零点", "2026-01-01T00:00:00Z", false, 1767225600},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseGitHubTime(tt.s)
			if tt.wantNil {
				if got != nil {
					t.Errorf("parseGitHubTime(%q) = %v; want nil", tt.s, got)
				}
				return
			}
			if got == nil {
				t.Fatalf("parseGitHubTime(%q) = nil; want non-nil", tt.s)
			}
			if got.Unix() != tt.wantSec {
				t.Errorf("parseGitHubTime(%q).Unix() = %d; want %d", tt.s, got.Unix(), tt.wantSec)
			}
		})
	}
}

// ===== parseTopics 测试 =====

func TestParseTopics(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		want    []string
		wantNil bool // 空输入是否返回 nil
	}{
		{"空字符串", "", nil, true},
		{"无效 JSON", "not-json", nil, true},
		{"空数组", "[]", []string{}, false},
		{"单个 topic", `["vue"]`, []string{"vue"}, false},
		{"多个 topics", `["vue","typescript","tailwind"]`, []string{"vue", "typescript", "tailwind"}, false},
		{"含空格的 topic", `["go web"]`, []string{"go web"}, false},
		{"JSON 对象（错误格式）", `{"key":"val"}`, nil, true},
		{"数字数组（错误类型）", `[1,2,3]`, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTopics(tt.s)
			if tt.wantNil {
				if got != nil {
					t.Errorf("parseTopics(%q) = %v; want nil", tt.s, got)
				}
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("parseTopics(%q) len = %d; want %d (got=%v want=%v)", tt.s, len(got), len(tt.want), got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("parseTopics(%q)[%d] = %q; want %q", tt.s, i, got[i], tt.want[i])
				}
			}
		})
	}
}

// ===== toPortfolioDTOs 测试 =====

func TestToPortfolioDTOs(t *testing.T) {
	repoUpdated := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	list := []model.Portfolio{
		{
			ID:              1,
			Name:            "aurora-blog",
			FullName:        "aqi/aurora-blog",
			Description:     "博客前端",
			HtmlURL:         "https://github.com/aqi/aurora-blog",
			Homepage:        "https://www.aqi125.cn",
			Language:        "Vue",
			StargazersCount: 10,
			ForksCount:      2,
			Topics:          `["vue","typescript"]`,
			RepoUpdatedAt:   &repoUpdated,
			Cover:           "https://ws.aqi125.cn/cover.png",
			Sort:            100,
			IsFeatured:      1,
			IsVisible:       1,
		},
		{
			ID:              2,
			Name:            "aurora-go",
			Topics:          "[]",
			IsVisible:       0,
		},
	}

	got := toPortfolioDTOs(list)
	if len(got) != 2 {
		t.Fatalf("toPortfolioDTOs len = %d; want 2", len(got))
	}

	// 验证第一条：完整字段
	first := got[0]
	if first.ID != 1 || first.Name != "aurora-blog" || first.HtmlURL != "https://github.com/aqi/aurora-blog" {
		t.Errorf("first DTO mismatch: %+v", first)
	}
	if first.StargazersCount != 10 || first.ForksCount != 2 {
		t.Errorf("first counts mismatch: stars=%d forks=%d", first.StargazersCount, first.ForksCount)
	}
	if len(first.Topics) != 2 || first.Topics[0] != "vue" || first.Topics[1] != "typescript" {
		t.Errorf("first topics mismatch: %v", first.Topics)
	}
	if first.Cover != "https://ws.aqi125.cn/cover.png" || first.Sort != 100 || first.IsFeatured != 1 {
		t.Errorf("first cover/sort/featured mismatch: %+v", first)
	}

	// 验证第二条：topics 应解析为空切片（parseTopics("[]") 返回非 nil 空切片）
	second := got[1]
	if second.Topics != nil && len(second.Topics) != 0 {
		t.Errorf("second topics should be empty, got %v", second.Topics)
	}
	// is_visible 不在 PortfolioDTO 中，不应影响 DTO 内容
	if second.Name != "aurora-go" {
		t.Errorf("second name mismatch: %s", second.Name)
	}
}

func TestToPortfolioDTOs_EmptyInput(t *testing.T) {
	got := toPortfolioDTOs(nil)
	if len(got) != 0 {
		t.Errorf("toPortfolioDTOs(nil) len = %d; want 0", len(got))
	}
	// 应返回非 nil 空切片（make 保证）
	if got == nil {
		t.Error("toPortfolioDTOs(nil) should return non-nil empty slice")
	}
}

// ===== toPortfolioAdminDTOs 测试 =====

func TestToPortfolioAdminDTOs(t *testing.T) {
	repoCreated := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	repoUpdated := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	createTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	list := []model.Portfolio{
		{
			ID:              1,
			RepoID:          123456,
			Name:            "aurora-go",
			FullName:        "aqi/aurora-go",
			Description:     "后端",
			HtmlURL:         "https://github.com/aqi/aurora-go",
			Homepage:        "",
			Language:        "Go",
			StargazersCount: 5,
			ForksCount:      1,
			Topics:          `["go","gin"]`,
			RepoCreatedAt:   &repoCreated,
			RepoUpdatedAt:   &repoUpdated,
			Cover:           "",
			Sort:            50,
			IsFeatured:      0,
			IsVisible:       0, // 隐藏项，AdminDTO 应包含
			CreateTime:      createTime,
		},
	}

	got := toPortfolioAdminDTOs(list)
	if len(got) != 1 {
		t.Fatalf("toPortfolioAdminDTOs len = %d; want 1", len(got))
	}

	d := got[0]
	// AdminDTO 应包含 RepoID（前台 DTO 不含）
	if d.RepoID != 123456 {
		t.Errorf("RepoID = %d; want 123456", d.RepoID)
	}
	// AdminDTO 应包含 IsVisible 和 CreateTime
	if d.IsVisible != 0 {
		t.Errorf("IsVisible = %d; want 0 (隐藏项)", d.IsVisible)
	}
	if !d.CreateTime.Equal(createTime) {
		t.Errorf("CreateTime = %v; want %v", d.CreateTime, createTime)
	}
	if !d.RepoCreatedAt.Equal(repoCreated) {
		t.Errorf("RepoCreatedAt = %v; want %v", d.RepoCreatedAt, repoCreated)
	}
	if !d.RepoUpdatedAt.Equal(repoUpdated) {
		t.Errorf("RepoUpdatedAt = %v; want %v", d.RepoUpdatedAt, repoUpdated)
	}
	if len(d.Topics) != 2 || d.Topics[0] != "go" {
		t.Errorf("Topics = %v; want [go gin]", d.Topics)
	}
}

// ===== DTO 与 JSON 序列化兼容性 =====

func TestPortfolioDTO_JSONSerialization(t *testing.T) {
	d := dto.PortfolioDTO{
		ID:              1,
		Name:            "test",
		StargazersCount: 10,
		Topics:          []string{"go"},
	}

	data, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	// 验证 JSON 包含 camelCase 字段名（对标 Java 前端兼容性）
	jsonStr := string(data)
	expectedFields := []string{`"id"`, `"name"`, `"stargazersCount"`, `"topics"`}
	for _, f := range expectedFields {
		if !contains(jsonStr, f) {
			t.Errorf("JSON output missing %s: %s", f, jsonStr)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
