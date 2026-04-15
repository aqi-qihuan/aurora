package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/model"
	"github.com/aurora-go/aurora/internal/util"
	"github.com/aurora-go/aurora/internal/vo"
	"gorm.io/gorm"
)

// CommentService 评论业务逻辑 (对标 Java CommentServiceImpl)
type CommentService struct {
	db           *gorm.DB
	statsService *RedisStatsService // Redis 统计服务
}

func NewCommentService(db *gorm.DB, statsService *RedisStatsService) *CommentService {
	return &CommentService{
		db:           db,
		statsService: statsService,
	}
}

// CreateComment 发表评论 (含IP归属地解析 + 敏感词过滤 + MQ通知)
func (s *CommentService) CreateComment(ctx context.Context, userID uint, vo vo.CommentVO, clientIP string) (*model.Comment, error) {
	var comment model.Comment

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		comment = model.Comment{
			UserID:         userID,
			Type:           vo.Type,
			ParentID:       vo.ParentID,
			CommentContent: vo.Content,
			TopicID:        nil, // 通过 type 区分，统一使用 topic_id
		}
		
		// 设置关联ID (统一使用 topic_id)
		switch vo.Type {
		case 1: // 文章评论
			topicID := vo.ArticleID
			comment.TopicID = &topicID
		case 5: // 说说评论
			topicID := vo.TalkID
			comment.TopicID = &topicID
		case 4: // 友链评论
			topicID := vo.FriendLinkID
			comment.TopicID = &topicID
		case 3: // 关于页评论
			topicID := vo.AboutID
			comment.TopicID = &topicID
		}

		// 回复时记录被回复用户
		if vo.ParentID > 0 && vo.ReplyUserID != nil {
			comment.ReplyUserID = vo.ReplyUserID
		}

		// TODO: P0-8 敏感词过滤
		_ = util.SanitizeHTML(vo.Content)

		if err := tx.Create(&comment).Error; err != nil {
			return fmt.Errorf("创建评论失败: %w", err)
		}

		// 更新关联实体计数(文章/说说/友链的评论数+1)
		s.incrementCommentCount(tx, vo.Type, comment.TopicID)

		// IP归属地解析 (异步不影响主流程)
		// TODO: 添加 IP 和 Location 字段到数据库或去掉此逻辑

		// TODO: P0-7 发布MQ消息 → 邮件通知 + 订阅通知
		// mq.Publish(constant.ExchangeDirect, constant.RoutingKeyEmail, commentMsg)

		slog.Info("评论发表成功",
			"comment_id", comment.ID,
			"type", commentTypeStr(vo.Type),
			"user_id", userID,
		)
		return nil
	})

	return &comment, err
}

// GetCommentsByArticle 获取文章评论列表 (嵌套树形结构)
// 当 articleID=0 时，返回全局最新评论（用于侧边栏）
func (s *CommentService) GetCommentsByArticle(ctx context.Context, articleID uint) ([]dto.CommentTreeDTO, error) {
	var comments []model.Comment

	query := s.db.WithContext(ctx).
		Preload("UserInfo").
		Preload("ReplyUser").
		Where("is_review = 1")

	// 如果指定了 articleID，则只查询该文章的评论
	if articleID > 0 {
		query = query.Where("topic_id = ? AND type = 1", articleID)
	}

	// 按创建时间倒序（最新评论在前）
	query = query.Order("create_time DESC")

	// 如果是全局最新评论，限制返回数量
	if articleID == 0 {
		query = query.Limit(6)
	} else {
		query = query.Order("create_time ASC") // 文章评论按正序
	}

	err := query.Find(&comments).Error
	if err != nil {
		return nil, fmt.Errorf("查询评论失败: %w", err)
	}

	// 确保返回空数组而非 null
	if comments == nil {
		comments = []model.Comment{}
	}

	return s.buildCommentTree(comments), nil
}

// GetLatestComments 获取最新的评论列表（扁平结构，用于侧边栏）
// 对标 Java CommentMapper.listTopSixComments
func (s *CommentService) GetLatestComments(ctx context.Context, limit int) ([]dto.CommentDTO, error) {
	type CommentWithUser struct {
		ID             uint      `gorm:"column:id"`
		UserID         uint      `gorm:"column:user_id"`
		CommentContent string    `gorm:"column:comment_content"`
		CreateTime     time.Time `gorm:"column:create_time"`
		Nickname       string    `gorm:"column:nickname"`
		Avatar         string    `gorm:"column:avatar"`
	}

	var results []CommentWithUser

	err := s.db.WithContext(ctx).
		Table("t_comment").
		Select("t_comment.id, t_comment.user_id, t_comment.comment_content, t_comment.create_time, t_user_info.nickname, t_user_info.avatar").
		Joins("JOIN t_user_info ON t_comment.user_id = t_user_info.id").
		Where("t_comment.is_review = 1").
		Order("t_comment.id DESC").
		Limit(limit).
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("查询最新评论失败: %w", err)
	}

	// 确保返回空数组而非 null
	if results == nil {
		return []dto.CommentDTO{}, nil
	}

	// 转换为 DTO
	list := make([]dto.CommentDTO, len(results))
	for i, r := range results {
		list[i] = dto.CommentDTO{
			ID:         r.ID,
			UserID:     r.UserID,
			Nickname:   r.Nickname,
			Avatar:     r.Avatar,
			Content:    r.CommentContent,
			CreateTime: r.CreateTime,
		}
	}

	return list, nil
}

// GetCommentsByTalk 获取说说评论列表
func (s *CommentService) GetCommentsByTalk(ctx context.Context, talkID uint) ([]dto.CommentDTO, error) {
	var comments []model.Comment

	err := s.db.WithContext(ctx).
		Where("topic_id = ? AND type = 5 AND is_review = 1", talkID).
		Preload("UserInfo").
		Preload("ReplyUser").
		Order("create_time ASC").
		Find(&comments).Error

	if err != nil {
		return nil, fmt.Errorf("查询说说评论失败: %w", err)
	}

	list := make([]dto.CommentDTO, len(comments))
	for i, c := range comments {
		list[i] = s.toCommentDTO(&c)
	}
	return list, nil
}

// ListComments 前台评论分页查询（对标 Java getComments）
// 支持按 type/topicId 筛选，返回树形结构评论
func (s *CommentService) ListComments(ctx context.Context, commentVO vo.CommentVO) (*dto.PageResultDTO, error) {
	var comments []model.Comment
	var count int64

	baseQuery := s.db.WithContext(ctx).Model(&model.Comment{}).
		Where("is_review = 1")

	// 按类型筛选（type=1文章, type=5说说, type=4友链, type=3关于）
	if commentVO.Type > 0 {
		baseQuery = baseQuery.Where("type = ?", commentVO.Type)
	}
	// 按关联ID筛选
	if commentVO.TopicID != nil && *commentVO.TopicID > 0 {
		baseQuery = baseQuery.Where("topic_id = ?", *commentVO.TopicID)
	}

	// 统计总数
	if err := baseQuery.Count(&count).Error; err != nil {
		return nil, fmt.Errorf("统计评论数失败: %w", err)
	}

	// 分页查询
	page := dto.PageVO{PageNum: commentVO.Current, PageSize: commentVO.Size}
	offset := page.GetOffset()

	if err := baseQuery.
		Preload("UserInfo").
		Preload("ReplyUser").
		Where("parent_id = 0"). // 只查询父评论
		Order("create_time DESC").
		Limit(page.PageSize).
		Offset(offset).
		Find(&comments).Error; err != nil {
		return nil, fmt.Errorf("查询评论列表失败: %w", err)
	}

	// 批量查询子评论
	if len(comments) > 0 {
		parentIDs := make([]uint, len(comments))
		for i, c := range comments {
			parentIDs[i] = c.ID
		}
		var replies []model.Comment
		s.db.WithContext(ctx).
			Where("parent_id IN ? AND is_review = 1", parentIDs).
			Preload("UserInfo").
			Preload("ReplyUser").
			Order("create_time ASC").
			Find(&replies)

		// 构建评论树
		tree := s.buildCommentTreeWithReplies(comments, replies)

		return &dto.PageResultDTO{
			List:     tree,
			Count:    count,
			PageNum:  page.PageNum,
			PageSize: page.PageSize,
		}, nil
	}

	return &dto.PageResultDTO{
		List:     []dto.CommentTreeDTO{},
		Count:    count,
		PageNum:  page.PageNum,
		PageSize: page.PageSize,
	}, nil
}

// buildCommentTreeWithReplies 构建带子评论的树形结构
func (s *CommentService) buildCommentTreeWithReplies(parents []model.Comment, replies []model.Comment) []dto.CommentTreeDTO {
	replyMap := make(map[uint][]model.Comment)
	for _, r := range replies {
		replyMap[r.ParentID] = append(replyMap[r.ParentID], r)
	}

	tree := make([]dto.CommentTreeDTO, len(parents))
	for i, p := range parents {
		tree[i] = dto.CommentTreeDTO{
			CommentDTO: s.toCommentDTO(&p),
			Replies:    s.convertToTreeDTOs(replyMap[p.ID]),
		}
	}
	return tree
}

// convertToTreeDTOs 将 Comment 列表转换为 CommentTreeDTO 列表（一层深度）
func (s *CommentService) convertToTreeDTOs(comments []model.Comment) []dto.CommentTreeDTO {
	if len(comments) == 0 {
		return []dto.CommentTreeDTO{}
	}
	result := make([]dto.CommentTreeDTO, len(comments))
	for i, c := range comments {
		result[i] = dto.CommentTreeDTO{
			CommentDTO: s.toCommentDTO(&c),
			Replies:    []dto.CommentTreeDTO{}, // 子评论不递归，只展示一层
		}
	}
	return result
}

// ListAdminComments 后台管理分页查询评论
func (s *CommentService) ListAdminComments(ctx context.Context, cond dto.ConditionVO, page dto.PageVO) (*dto.PageResultDTO, error) {
	var comments []model.Comment
	var count int64

	baseQuery := s.db.WithContext(ctx).Model(&model.Comment{})

	if cond.Keywords != "" {
		baseQuery = baseQuery.Where("comment_content LIKE ?", "%"+cond.Keywords+"%")
	}
	if cond.Type != nil {
		baseQuery = baseQuery.Where("type = ?", *cond.Type)
	}
	// 处理 isReview 筛选（前端传 isReview=0 表示待审核）
	if cond.IsReview != nil {
		baseQuery = baseQuery.Where("is_review = ?", *cond.IsReview)
	} else if cond.Status != nil {
		// 兼容旧的 status 参数
		baseQuery = baseQuery.Where("is_review = ?", *cond.Status)
	}

	if err := baseQuery.Count(&count).Error; err != nil {
		return nil, fmt.Errorf("统计评论数失败: %w", err)
	}

	offset := page.GetOffset()
	if err := baseQuery.
		Order("create_time DESC").
		Limit(page.PageSize).
		Offset(offset).
		Find(&comments).Error; err != nil {
		return nil, fmt.Errorf("查询评论列表失败: %w", err)
	}

	// 批量预加载 UserInfo 和 ReplyUser（避免 N+1 查询）
	if len(comments) > 0 {
		// 收集所有用户ID
		userIDs := make(map[uint]bool)
		for _, c := range comments {
			userIDs[c.UserID] = true
			if c.ReplyUserID != nil {
				userIDs[*c.ReplyUserID] = true
			}
		}

		// 批量查询用户信息
		ids := make([]uint, 0, len(userIDs))
		for id := range userIDs {
			ids = append(ids, id)
		}

		var users []model.UserInfo
		s.db.WithContext(ctx).Where("id IN ?", ids).Find(&users)

		// 构建用户映射
		userMap := make(map[uint]*model.UserInfo)
		for i := range users {
			userMap[users[i].ID] = &users[i]
		}

		// 填充用户信息
		for i := range comments {
			if u, ok := userMap[comments[i].UserID]; ok {
				comments[i].UserInfo = u
			}
			if comments[i].ReplyUserID != nil {
				if u, ok := userMap[*comments[i].ReplyUserID]; ok {
					comments[i].ReplyUser = u
				}
			}
		}
	}

	list := make([]dto.CommentAdminDTO, len(comments))
	for i, c := range comments {
		list[i] = s.toCommentAdminDTO(&c)
	}

	// 批量查询标题（优化N+1问题）
	if len(comments) > 0 {
		// 收集所有非空的 topicId
		type TopicKey struct {
			Type int8
			ID   uint
		}
		topicMap := make(map[TopicKey]bool)
		for _, c := range comments {
			if c.TopicID != nil && *c.TopicID > 0 {
				topicMap[TopicKey{Type: c.Type, ID: *c.TopicID}] = true
			}
		}

		// 批量查询标题
		type TitleResult struct {
			Type  int8
			ID    uint
			Title string
		}
		var titles []TitleResult

		// 文章标题 (type=1)
		if hasType(comments, 1) {
			var articleTitles []struct {
				ID    uint   `gorm:"column:id"`
				Title string `gorm:"column:article_title"`
			}
			s.db.WithContext(ctx).
				Table("t_article").
				Select("id, article_title").
				Where("id IN (?)", collectTopicIDs(comments, 1)).
				Find(&articleTitles)
			for _, at := range articleTitles {
				titles = append(titles, TitleResult{Type: 1, ID: at.ID, Title: at.Title})
			}
		}

		// 说说内容 (type=5)
		if hasType(comments, 5) {
			var talkTitles []struct {
				ID      uint   `gorm:"column:id"`
				Content string `gorm:"column:content"`
			}
			s.db.WithContext(ctx).
				Table("t_talk").
				Select("id, content").
				Where("id IN (?)", collectTopicIDs(comments, 5)).
				Find(&talkTitles)
			for _, tt := range talkTitles {
				content := tt.Content
				if len(content) > 20 {
					content = content[:20] + "..."
				}
				titles = append(titles, TitleResult{Type: 5, ID: tt.ID, Title: content})
			}
		}

		// 构建映射
		titleMap := make(map[TopicKey]string)
		for _, t := range titles {
			titleMap[TopicKey{Type: t.Type, ID: t.ID}] = t.Title
		}

		// 填充标题
		for i := range list {
			if comments[i].TopicID != nil && *comments[i].TopicID > 0 {
				key := TopicKey{Type: comments[i].Type, ID: *comments[i].TopicID}
				list[i].ArticleTitle = titleMap[key]
			}
		}
	}

	return &dto.PageResultDTO{
		List:     list,
		Count:    count,
		PageNum:  page.PageNum,
		PageSize: page.PageSize,
	}, nil
}

// ReviewComment 审核评论 (通过/拒绝)
func (s *CommentService) ReviewComment(ctx context.Context, id uint, isReview int8) error {
	result := s.db.WithContext(ctx).
		Model(&model.Comment{}).
		Where("id = ?", id).
		Update("is_review", isReview)

	if result.Error != nil {
		return fmt.Errorf("审核评论失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.ErrCommentNotFound
	}

	action := "通过"
	if isReview == 0 {
		action = "拒绝"
	}
	slog.Info("评论审核完成", "comment_id", id, "action", action)
	return nil
}

// LikeComment 点赞评论 (使用 Redis 实现)
func (s *CommentService) LikeComment(ctx context.Context, id uint) error {
	// TODO: 实现 Redis 点赞逻辑
	// 临时方案：返回成功
	return nil
}

// DeleteComment 删除评论 (级联删除子评论 + 更新计数)
func (s *CommentService) DeleteComment(ctx context.Context, id uint) error {
	var comment model.Comment

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&comment, id).Error; err != nil {
			return errors.ErrCommentNotFound
		}

		// 级联删除子评论 (parentId=id 的所有回复)
		var childCount int64
		tx.Model(&model.Comment{}).Where("parent_id = ?", id).Count(&childCount)
		
		if childCount > 0 {
			tx.Where("parent_id = ?", id).Delete(&model.Comment{})
		}

		// 删除主评论
		if err := tx.Delete(&comment).Error; err != nil {
			return fmt.Errorf("删除评论失败: %w", err)
		}

		// 更新关联实体评论数(-1 - 子评论数)
		totalCount := int(childCount) + 1
		s.decrementCommentCount(tx, comment.Type, comment.TopicID, totalCount)

		slog.Info("评论已删除",
			"comment_id", id,
			"children_deleted", childCount,
		)
		return nil
	})
}

// BatchReviewComments 批量审核评论
func (s *CommentService) BatchReviewComments(ctx context.Context, ids []uint, isReview int8) error {
	result := s.db.WithContext(ctx).
		Model(&model.Comment{}).
		Where("id IN ?", ids).
		Update("is_review", isReview)

	if result.Error != nil {
		return fmt.Errorf("批量审核评论失败: %w", result.Error)
	}

	slog.Info("批量审核评论", "count", result.RowsAffected, "status", isReview)
	return nil
}

// BatchDeleteComments 批量删除评论
func (s *CommentService) BatchDeleteComments(ctx context.Context, ids []uint) error {
	for _, id := range ids {
		if err := s.DeleteComment(ctx, id); err != nil {
			slog.Warn("批量删除评论失败", "comment_id", id, "error", err.Error())
		}
	}
	slog.Info("批量删除评论完成", "total_requested", len(ids))
	return nil
}

// GetCommentStats 获取各类型评论统计
type CommentStats struct {
	Total    int64 `json:"total"`
	Article  int64 `json:"article"`
	Talk     int64 `json:"talk"`
	Link     int64 `json:"link"`
	Pending  int64 `json:"pending"`  // 待审核
	Approved int64 `json:"approved"` // 已通过
}

func (s *CommentService) GetCommentStats(ctx context.Context) (*CommentStats, error) {
	stats := &CommentStats{}

	// 使用goroutine并发查询4个统计 (errgroup模式替代CompletableFuture)
	ch := make(chan struct {
		key string
		val int64
	}, 5)

	go func() { ch <- struct{ key string; val int64 }{"total", s.countAll(ctx)} }()
	go func() { ch <- struct{ key string; val int64 }{"article", s.countByType(ctx, 1)} }()
	go func() { ch <- struct{ key string; val int64 }{"talk", s.countByType(ctx, 2)} }()
	go func() { ch <- struct{ key string; val int64 }{"pending", s.countByReview(ctx, 0)} }()
	go func() { ch <- struct{ key string; val int64 }{"approved", s.countByReview(ctx, 1)} }()

	for i := 0; i < 5; i++ {
		r := <-ch
		switch r.key {
		case "total": stats.Total = r.val
		case "article": stats.Article = r.val
		case "talk": stats.Talk = r.val
		case "pending": stats.Pending = r.val
		case "approved": stats.Approved = r.val
		}
	}

	stats.Link = stats.Total - stats.Article - stats.Talk
	return stats, nil
}

// ===== 内部方法 =====

func (s *CommentService) incrementCommentCount(tx *gorm.DB, commentType int8, topicID *uint) {
	if topicID == nil || *topicID == 0 {
		return
	}
	switch commentType {
	case 1: // 文章
		tx.Exec("UPDATE t_article SET comment_count = COALESCE(comment_count, 0) + 1 WHERE id = ?", *topicID)
	case 5: // 说说
		tx.Exec("UPDATE t_talk SET comment_count = COALESCE(comment_count, 0) + 1 WHERE id = ?", *topicID)
	case 4: // 友链
		tx.Exec("UPDATE t_friend_link SET comment_count = COALESCE(comment_count, 0) + 1 WHERE id = ?", *topicID)
	case 3: // 关于
		// t_about 可能没有 comment_count 字段，视情况处理
	}
}

func (s *CommentService) decrementCommentCount(tx *gorm.DB, commentType int8, topicID *uint, count int) {
	if topicID == nil || *topicID == 0 {
		return
	}
	switch commentType {
	case 1: // 文章
		tx.Exec("UPDATE t_article SET comment_count = GREATEST(COALESCE(comment_count, 0) - ?, 0) WHERE id = ?", count, *topicID)
	case 5: // 说说
		tx.Exec("UPDATE t_talk SET comment_count = GREATEST(COALESCE(comment_count, 0) - ?, 0) WHERE id = ?", count, *topicID)
	case 4: // 友链
		tx.Exec("UPDATE t_friend_link SET comment_count = GREATEST(COALESCE(comment_count, 0) - ?, 0) WHERE id = ?", count, *topicID)
	}
}

func (s *CommentService) buildCommentTree(comments []model.Comment) []dto.CommentTreeDTO {
	commentMap := make(map[uint]*dto.CommentTreeDTO)
	var roots []dto.CommentTreeDTO

	// 先转为DTO并建立映射
	for _, c := range comments {
		dto := s.toCommentTreeDTO(&c)
		commentMap[c.ID] = &dto
	}

	// 构建树形结构
	for _, c := range comments {
		node := commentMap[c.ID]
		if c.ParentID == 0 || commentMap[c.ParentID] == nil {
			roots = append(roots, *node)
		} else {
			parent := commentMap[c.ParentID]
			parent.Replies = append(parent.Replies, *node)
		}
	}

	return roots
}

func (s *CommentService) toCommentDTO(c *model.Comment) dto.CommentDTO {
	dto := dto.CommentDTO{
		ID:         c.ID,
		UserID:     c.UserID,
		Content:    c.CommentContent,
		Type:       c.Type,
		ParentID:   c.ParentID,
		LikeCount:  s.getCommentLikeCount(c.ID),
		IsReview:   c.IsReview,
		CreateTime: c.CreateTime,
	}
	if c.UserInfo != nil {
		dto.Nickname = c.UserInfo.Nickname
		dto.Avatar = c.UserInfo.Avatar
	}
	if c.ReplyUser != nil {
		dto.ReplyNickname = c.ReplyUser.Nickname
	}
	return dto
}

func (s *CommentService) toCommentTreeDTO(c *model.Comment) dto.CommentTreeDTO {
	return dto.CommentTreeDTO{
		CommentDTO: s.toCommentDTO(c),
		Replies:    []dto.CommentTreeDTO{},
	}
}

func (s *CommentService) toCommentAdminDTO(c *model.Comment) dto.CommentAdminDTO {
	dto := dto.CommentAdminDTO{
		ID:             c.ID,
		UserID:         c.UserID,
		CommentContent: c.CommentContent,
		Type:           c.Type,
		TopicID:        c.TopicID,
		ReplyUserID:    c.ReplyUserID,
		ParentID:       c.ParentID,
		IsReview:       c.IsReview,
		LikeCount:      s.getCommentLikeCount(c.ID),
		CreateTime:     c.CreateTime,
	}
	// 评论人信息
	if c.UserInfo != nil {
		dto.Nickname = c.UserInfo.Nickname
		dto.Avatar = c.UserInfo.Avatar
	}
	// 回复人信息
	if c.ReplyUser != nil {
		dto.ReplyNickname = c.ReplyUser.Nickname
	}
	return dto
}

// 并发查询辅助方法
func (s *CommentService) countAll(ctx context.Context) int64 {
	var count int64
	s.db.WithContext(ctx).Model(&model.Comment{}).Count(&count)
	return count
}

// uintPtr 辅助函数: uint → *uint
func uintPtr(v uint) *uint {
	return &v
}

func (s *CommentService) countByType(ctx context.Context, t int8) int64 {
	var count int64
	s.db.WithContext(ctx).Model(&model.Comment{}).Where("type = ?", t).Count(&count)
	return count
}

func (s *CommentService) countByReview(ctx context.Context, review int8) int64 {
	var count int64
	s.db.WithContext(ctx).Model(&model.Comment{}).Where("is_review = ?", review).Count(&count)
	return count
}

func commentTypeStr(t int8) string {
	switch t {
	case 1: return "文章"
	case 2: return "说说"
	case 3: return "友链"
	case 4: return "关于"
	default: return "其他"
	}
}

// getCommentLikeCount 获取评论点赞数（从 Redis）
func (s *CommentService) getCommentLikeCount(commentID uint) int64 {
	if s.statsService == nil {
		return 0
	}
	count, _ := s.statsService.GetCommentLikeCount(context.Background(), commentID)
	return count
}

// hasType 检查评论列表中是否包含指定类型
func hasType(comments []model.Comment, t int8) bool {
	for _, c := range comments {
		if c.Type == t {
			return true
		}
	}
	return false
}

// collectTopicIDs 收集指定类型的topic_id列表
func collectTopicIDs(comments []model.Comment, t int8) []uint {
	ids := make([]uint, 0)
	for _, c := range comments {
		if c.Type == t && c.TopicID != nil && *c.TopicID > 0 {
			ids = append(ids, *c.TopicID)
		}
	}
	return ids
}
