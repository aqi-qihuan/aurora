package scheduler

import (
	"context"
	"fmt"
	"log/slog"
)

// GitHubSyncFunc 作品集同步函数（由 service 层注入，避免 scheduler→service 循环依赖）
// 装配阶段（main.go）在 Registry 创建后设置：scheduler.GitHubSyncFunc = registry.Portfolio.SyncFromGitHub
var GitHubSyncFunc func(context.Context) error

// GitHubSyncJob GitHub作品集同步任务
// 对标 Java 版 Quartz Job，建议 Cron: 0 0 3 * * ? （每天 03:00 执行一次）
// 业务逻辑：从 GitHub API 拉取用户仓库快照写入 t_portfolio 表
type GitHubSyncJob struct{}

// NewGitHubSyncJob 创建 GitHub 同步任务实例
func NewGitHubSyncJob() *GitHubSyncJob {
	return &GitHubSyncJob{}
}

// Run 执行同步任务（实现 TaskFunc 签名）
func (j *GitHubSyncJob) Run(ctx context.Context, params ...interface{}) error {
	if GitHubSyncFunc == nil {
		slog.Warn("GitHub作品集同步函数未注入，跳过（请检查 portfolio service 是否已初始化）")
		return nil
	}
	if err := GitHubSyncFunc(ctx); err != nil {
		return fmt.Errorf("GitHub作品集同步失败: %w", err)
	}
	return nil
}
