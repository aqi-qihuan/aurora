package service

import (
	"testing"

	"github.com/aurora-go/aurora/internal/model"
)

// ===== getCommentPath 测试 =====

func TestGetCommentPath(t *testing.T) {
	tests := []struct {
		name        string
		commentType int8
		want        string
	}{
		{"文章评论 type=1", 1, "articles"},
		{"说说评论 type=5", 5, "talks"},
		{"友链评论 type=4", 4, "links"},
		{"关于页评论 type=3", 3, "about"},
		{"留言板 type=2 无路径", 2, ""},
		{"未知类型 type=99", 99, ""},
		{"零值 type=0", 0, ""},
		{"负值 type=-1", -1, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getCommentPath(tt.commentType)
			if got != tt.want {
				t.Errorf("getCommentPath(%d) = %q; want %q", tt.commentType, got, tt.want)
			}
		})
	}
}

// ===== commentTypeStr 测试 =====

func TestCommentTypeStr(t *testing.T) {
	tests := []struct {
		name        string
		commentType int8
		want        string
	}{
		{"文章", 1, "文章"},
		{"留言板", 2, "留言板"},
		{"关于页", 3, "关于页"},
		{"友链", 4, "友链"},
		{"说说", 5, "说说"},
		{"未知类型", 0, "其他"},
		{"超出范围", 99, "其他"},
		{"负值", -1, "其他"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := commentTypeStr(tt.commentType)
			if got != tt.want {
				t.Errorf("commentTypeStr(%d) = %q; want %q", tt.commentType, got, tt.want)
			}
		})
	}
}

// ===== hasType 测试 =====

func TestHasType(t *testing.T) {
	topicID1 := uint(100)
	topicID2 := uint(200)
	comments := []model.Comment{
		{ID: 1, Type: 1, TopicID: &topicID1},
		{ID: 2, Type: 2},
		{ID: 3, Type: 5, TopicID: &topicID2},
	}

	tests := []struct {
		name     string
		comments []model.Comment
		t        int8
		want     bool
	}{
		{"包含文章评论", comments, 1, true},
		{"包含留言板", comments, 2, true},
		{"包含说说", comments, 5, true},
		{"不包含友链", comments, 4, false},
		{"不包含关于页", comments, 3, false},
		{"空列表", []model.Comment{}, 1, false},
		{"nil 列表", nil, 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasType(tt.comments, tt.t)
			if got != tt.want {
				t.Errorf("hasType(..., %d) = %v; want %v", tt.t, got, tt.want)
			}
		})
	}
}

// ===== collectTopicIDs 测试 =====

func TestCollectTopicIDs(t *testing.T) {
	topicID1 := uint(100)
	topicID2 := uint(200)
	topicID3 := uint(300)
	zero := uint(0)

	tests := []struct {
		name     string
		comments []model.Comment
		t        int8
		want     []uint
	}{
		{
			"收集文章评论的 topicID",
			[]model.Comment{
				{ID: 1, Type: 1, TopicID: &topicID1},
				{ID: 2, Type: 2, TopicID: &topicID2}, // 类型不匹配，跳过
				{ID: 3, Type: 1, TopicID: &topicID3},
			},
			1,
			[]uint{100, 300},
		},
		{
			"收集说说评论的 topicID",
			[]model.Comment{
				{ID: 1, Type: 5, TopicID: &topicID1},
				{ID: 2, Type: 5, TopicID: &topicID2},
			},
			5,
			[]uint{100, 200},
		},
		{
			"过滤 nil TopicID",
			[]model.Comment{
				{ID: 1, Type: 1, TopicID: nil},
				{ID: 2, Type: 1, TopicID: &topicID1},
			},
			1,
			[]uint{100},
		},
		{
			"过滤零值 TopicID",
			[]model.Comment{
				{ID: 1, Type: 1, TopicID: &zero},
				{ID: 2, Type: 1, TopicID: &topicID1},
			},
			1,
			[]uint{100},
		},
		{
			"无匹配类型返回空切片",
			[]model.Comment{
				{ID: 1, Type: 2, TopicID: &topicID1},
			},
			1,
			[]uint{},
		},
		{
			"空列表返回空切片",
			[]model.Comment{},
			1,
			[]uint{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := collectTopicIDs(tt.comments, tt.t)
			if len(got) != len(tt.want) {
				t.Errorf("collectTopicIDs len = %d; want %d (got=%v want=%v)", len(got), len(tt.want), got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("collectTopicIDs[%d] = %d; want %d", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// ===== uintPtr 测试 =====

func TestUintPtr(t *testing.T) {
	v := uint(42)
	p := uintPtr(v)
	if p == nil {
		t.Fatal("uintPtr returned nil")
	}
	if *p != v {
		t.Errorf("uintPtr(%d) = %d; want %d", v, *p, v)
	}
}

// ===== getSiteURL 测试 =====

func TestGetSiteURL(t *testing.T) {
	url := getSiteURL()
	if url == "" {
		t.Error("getSiteURL() should not return empty string")
	}
	// 当前实现返回固定值，验证不为空即可（未来若改为配置读取，此测试仍有效）
}
