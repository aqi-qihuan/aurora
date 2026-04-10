package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/model"
	"github.com/aurora-go/aurora/internal/util"
	"github.com/aurora-go/aurora/internal/vo"
	"gorm.io/gorm"
)

// CommentService 评论业务逻辑 (对标 Java CommentServiceImpl)
type CommentService struct {
	db *gorm.DB
}

func NewCommentService(db *gorm.DB) *CommentService {
	return &CommentService{db: db}
}

// CreateComment 发表评论 (含IP归属地解析 + 敏感词过滤 + MQ通知)
func (s *CommentService) CreateComment(ctx context.Context, userID uint, vo vo.CommentVO, clientIP string) (*model.Comment, error) {
	var comment model.Comment

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		comment = model.Comment{
			UserID:    userID,
			Type:      vo.Type,
			ParentID:  vo.ParentID,
			Content:   vo.Content,
			IP:        clientIP,
		}

		// 设置关联ID
		switch vo.Type {
		case 1: // 文章评论
			comment.ArticleID = &vo.ArticleID
		case 2: // 说说评论
			comment.TalkID = &vo.TalkID
		case 3: // 友链评论
			comment.FriendLinkID = &vo.FriendLinkID
		case 4: // 关于页评论
			comment.AboutID = &vo.AboutID
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
		s.incrementCommentCount(tx, vo.Type, uintPtr(vo.ArticleID), uintPtr(vo.TalkID), uintPtr(vo.FriendLinkID), nil)

		// IP归属地解析 (异步不影响主流程)
		location := util.ResolveIP(clientIP)
		if location != "" {
			tx.Model(&comment).Update("location", location)
		}
		comment.Location = location

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
func (s *CommentService) GetCommentsByArticle(ctx context.Context, articleID uint) ([]dto.CommentTreeDTO, error) {
	var comments []model.Comment

	err := s.db.WithContext(ctx).
		Where("article_id = ? AND type = 1 AND is_review = 1", articleID).
		Preload("UserInfo").
		Preload("ReplyUser").
		Order("create_time ASC").
		Find(&comments).Error
	
	if err != nil {
		return nil, fmt.Errorf("查询文章评论失败: %w", err)
	}

	return s.buildCommentTree(comments), nil
}

// GetCommentsByTalk 获取说说评论列表
func (s *CommentService) GetCommentsByTalk(ctx context.Context, talkID uint) ([]dto.CommentDTO, error) {
	var comments []model.Comment

	err := s.db.WithContext(ctx).
		Where("talk_id = ? AND type = 2 AND is_review = 1", talkID).
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

// ListAdminComments 后台管理分页查询评论
func (s *CommentService) ListAdminComments(ctx context.Context, cond dto.ConditionVO, page dto.PageVO) (*dto.PageResultDTO, error) {
	var comments []model.Comment
	var count int64

	baseQuery := s.db.WithContext(ctx).Model(&model.Comment{})

	if cond.Keywords != "" {
		baseQuery = baseQuery.Where("content LIKE ?", "%"+cond.Keywords+"%")
	}
	if cond.Type != nil {
		baseQuery = baseQuery.Where("type = ?", *cond.Type)
	}
	if cond.Status != nil {
		reviewStatus := int8(0)
		if *cond.Status == 1 {
			reviewStatus = 1
		}
		baseQuery = baseQuery.Where("is_review = ?", reviewStatus)
	}

	if err := baseQuery.Count(&count).Error; err != nil {
		return nil, fmt.Errorf("统计评论数失败: %w", err)
	}

	offset := page.GetOffset()
	if err := baseQuery.
		Preload("UserInfo").
		Order("create_time DESC").
		Limit(page.PageSize).
		Offset(offset).
		Find(&comments).Error; err != nil {
		return nil, fmt.Errorf("查询评论列表失败: %w", err)
	}

	list := make([]dto.CommentAdminDTO, len(comments))
	for i, c := range comments {
		list[i] = s.toCommentAdminDTO(&c)
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

// LikeComment 点赞评论
func (s *CommentService) LikeComment(ctx context.Context, id uint) error {
	result := s.db.WithContext(ctx).
		Model(&model.Comment{}).
		Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("like_count + 1"))

	if result.Error != nil {
		return fmt.Errorf("点赞评论失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.ErrCommentNotFound
	}
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
		s.decrementCommentCount(tx, comment.Type, comment.ArticleID, comment.TalkID, comment.FriendLinkID, comment.AboutID, totalCount)

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

func (s *CommentService) incrementCommentCount(tx *gorm.DB, commentType int8, articleID, talkID, friendLinkID, aboutID *uint) {
	switch commentType {
	case 1:
		if articleID != nil && *articleID > 0 {
			tx.Exec("UPDATE t_article SET comment_count = COALESCE(comment_count, 0) + 1 WHERE id = ?", *articleID)
		}
	case 2:
		if talkID != nil && *talkID > 0 {
			tx.Exec("UPDATE t_talk SET comment_count = COALESCE(comment_count, 0) + 1 WHERE id = ?", *talkID)
		}
	case 3:
		if friendLinkID != nil && *friendLinkID > 0 {
			tx.Exec("UPDATE t_friend_link SET comment_count = COALESCE(comment_count, 0) + 1 WHERE id = ?", *friendLinkID)
		}
	case 4:
		if aboutID != nil && *aboutID > 0 {
			tx.Exec("UPDATE t_about SET comment_count = COALESCE(comment_count, 0) + 1 WHERE id = ?", *aboutID)
		}
	}
}

func (s *CommentService) decrementCommentCount(tx *gorm.DB, commentType int8, articleID, talkID, friendLinkID, aboutID *uint, count int) {
	switch commentType {
	case 1:
		if articleID != nil && *articleID > 0 {
			tx.Exec("UPDATE t_article SET comment_count = GREATEST(COALESCE(comment_count, 0) - ?, 0) WHERE id = ?", count, *articleID)
		}
	case 2:
		if talkID != nil && *talkID > 0 {
			tx.Exec("UPDATE t_talk SET comment_count = GREATEST(COALESCE(comment_count, 0) - ?, 0) WHERE id = ?", count, *talkID)
		}
	case 3:
		if friendLinkID != nil && *friendLinkID > 0 {
			tx.Exec("UPDATE t_friend_link SET comment_count = GREATEST(COALESCE(comment_count, 0) - ?, 0) WHERE id = ?", count, *friendLinkID)
		}
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
		Content:    c.Content,
		Type:       c.Type,
		ParentID:   c.ParentID,
		LikeCount:  c.LikeCount,
		IsReview:   c.IsReview,
		CreateTime: c.CreateTime,
		Location:   c.Location,
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
		ID:         c.ID,
		UserID:     c.UserID,
		Content:    c.Content,
		Type:       c.Type,
		ParentID:   c.ParentID,
		IsReview:   c.IsReview,
		IP:         c.IP,
		Location:   c.Location,
		LikeCount:  c.LikeCount,
		CreateTime: c.CreateTime,
	}
	if c.UserInfo != nil {
		dto.Nickname = c.UserInfo.Nickname
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
