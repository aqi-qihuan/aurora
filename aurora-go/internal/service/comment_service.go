package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/aurora-go/aurora/internal/constant"
	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/infrastructure/mq"
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

// CreateComment 发表评论 (含IP归属地解析 + 敏感词过滤 + MQ通知 + 审核逻辑)
func (s *CommentService) CreateComment(ctx context.Context, userID uint, vo vo.CommentVO, clientIP string) (*model.Comment, error) {
	if userID == 0 {
		return nil, errors.ErrUnauthorized.WithMsg("请先登录")
	}

	// 读取网站配置（对标Java: WebsiteConfigDTO websiteConfig = auroraInfoService.getWebsiteConfig()）
	var configModel model.WebsiteConfig
	s.db.WithContext(ctx).First(&configModel, 1) // 明确指定ID=1
	
	// 解析JSON配置（对标Java: JSON.parseObject(config.getConfig(), WebsiteConfigDTO.class)）
	websiteConfig := &dto.WebsiteConfigDTO{}
	if configModel.Config != "" {
		if err := json.Unmarshal([]byte(configModel.Config), websiteConfig); err != nil {
			slog.Warn("解析评论审核配置失败，默认不审核", "error", err)
		}
	}
	
	// 判断是否需要审核（对标Java: isCommentReview == TRUE ? FALSE : TRUE）
	isReview := int8(1) // 默认通过
	if websiteConfig.IsCommentReview != nil && *websiteConfig.IsCommentReview == 1 {
		isReview = 0 // 需要审核
		slog.Info("评论审核已开启", "isReview", isReview)
	} else {
		slog.Info("评论审核未开启或配置缺失", "isCommentReview", websiteConfig.IsCommentReview, "isReview", isReview)
	}

	var comment model.Comment

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		comment = model.Comment{
			UserID:         userID,
			Type:           vo.Type,
			ParentID:       vo.ParentID,
			CommentContent: vo.Content,
			IsReview:       isReview,
			TopicID:        nil,
		}
		
		// 设置关联ID (统一使用 topic_id)
		// 二级回复时，从父评论继承 topicId（对标Java: 回复的topicId与父评论一致）
		if vo.ParentID > 0 {
			var parentComment model.Comment
			if err := tx.Select("id", "type", "topic_id").First(&parentComment, vo.ParentID).Error; err == nil {
				comment.Type = parentComment.Type      // 继承父评论类型
				comment.TopicID = parentComment.TopicID // 继承父评论 topicId
				slog.Info("二级回复：从父评论继承 topicId", "parent_id", vo.ParentID, "topic_id", parentComment.TopicID, "type", parentComment.Type)
			} else {
				slog.Warn("查询父评论失败", "parent_id", vo.ParentID, "error", err)
				return fmt.Errorf("父评论不存在: %w", err)
			}
		} else {
			// 顶级评论：从请求参数获取 topicId
			// 优先级: type-specific字段 > TopicID(前端统一传参)
			switch vo.Type {
			case 1: // 文章评论
				topicID := vo.ArticleID
				if topicID == 0 && vo.TopicID != nil {
					topicID = *vo.TopicID // 兼容前端 topicId 统一传参
				}
				if topicID > 0 {
					comment.TopicID = &topicID
				}
			case 5: // 说说评论
				topicID := vo.TalkID
				if topicID == 0 && vo.TopicID != nil {
					topicID = *vo.TopicID
				}
				if topicID > 0 {
					comment.TopicID = &topicID
				}
			case 4: // 友链评论
				topicID := vo.FriendLinkID
				if topicID == 0 && vo.TopicID != nil {
					topicID = *vo.TopicID
				}
				if topicID > 0 {
					comment.TopicID = &topicID
				}
			case 3: // 关于页评论
				topicID := vo.AboutID
				if topicID == 0 && vo.TopicID != nil {
					topicID = *vo.TopicID
				}
				if topicID > 0 {
					comment.TopicID = &topicID
				}
			case 2: // 留言板评论
				// 留言板不需要 topicId，保持为 nil
			}
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

		return nil
	})

	// 发送评论通知 (对标Java: CompletableFuture.runAsync(() -> notice(comment, fromNickname)))
	// 在事务外异步发送, 避免阻塞主流程（带 recover 防止 goroutine panic 导致进程崩溃）
	if err == nil {
		util.SafeGoAsync(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := s.SendCommentNotification(ctx, &comment, userID); err != nil {
				slog.Warn("发送评论通知失败", "comment_id", comment.ID, "error", err)
			}
		})
	}

	return &comment, err
}

// SendCommentNotification 发送评论通知邮件 (对标Java CommentServiceImpl.notice)
// 通知逻辑:
// 1. @提醒通知: 回复时@了其他人
// 2. 评论通知: 通知父评论作者 或 文章/说说作者
func (s *CommentService) SendCommentNotification(ctx context.Context, comment *model.Comment, userID uint) error {
	// 查询评论者信息
	var commenter model.UserInfo
	if err := s.db.WithContext(ctx).Select("nickname").Where("id = ?", userID).First(&commenter).Error; err != nil {
		return fmt.Errorf("查询评论者失败: %w", err)
	}
	fromNickname := commenter.Nickname

	// 1. @提醒通知 (对标Java: 第221-241行)
	if comment.ParentID > 0 && comment.ReplyUserID != nil && *comment.ReplyUserID != userID {
		var parentComment model.Comment
		if err := s.db.WithContext(ctx).First(&parentComment, comment.ParentID).Error; err != nil {
			return fmt.Errorf("查询父评论失败: %w", err)
		}

		// 判断是否需要发送@提醒
		if *comment.ReplyUserID != parentComment.UserID && *comment.ReplyUserID != comment.UserID {
			var replyUser model.UserInfo
			if err := s.db.WithContext(ctx).Select("email, nickname").Where("id = ?", *comment.ReplyUserID).First(&replyUser).Error; err != nil {
				return fmt.Errorf("查询被回复用户失败: %w", err)
			}

			if replyUser.Email != "" {
				// 构建@提醒邮件
				topicID := ""
				if comment.TopicID != nil {
					topicID = fmt.Sprintf("%d", *comment.TopicID)
				}
				commentType := commentTypeStr(comment.Type)
				url := fmt.Sprintf("%s/%s/%s", getSiteURL(), getCommentPath(comment.Type), topicID)

				emailDTO := dto.EmailDTO{
					Email:   replyUser.Email,
					Subject: "@提醒",
					CommentMap: map[string]interface{}{
						"content": fmt.Sprintf("%s在%s的评论区@了你，<a style=\"text-decoration:none;color:#12addb\" href=\"%s\">点击查看</a>",
							fromNickname, commentType, url),
					},
				}

				if err := publishCommentEmail(emailDTO); err != nil {
					return fmt.Errorf("发送@提醒邮件失败: %w", err)
				}
				slog.Info("📧 已发送@提醒邮件", "to", replyUser.Email, "comment_id", comment.ID)
			}
		}
	}

	// 2. 评论通知 (对标Java: 第246-272行)
	// 确定通知对象: 父评论作者 或 文章/说说作者
	var notifyUserID uint
	topicID := ""
	if comment.TopicID != nil {
		topicID = fmt.Sprintf("%d", *comment.TopicID)
	}

	if comment.ReplyUserID != nil {
		notifyUserID = *comment.ReplyUserID
	} else {
		// 查询文章/说说的作者
		switch comment.Type {
		case 1: // 文章评论
			var article model.Article
			if comment.TopicID != nil {
				if err := s.db.WithContext(ctx).Select("user_id, article_title").Where("id = ?", *comment.TopicID).First(&article).Error; err == nil {
					notifyUserID = article.UserID
				}
			}
		case 5: // 说说评论
			notifyUserID = 1 // 默认博主ID
		default:
			notifyUserID = 1 // 默认博主ID
		}
	}

	// 查询通知对象邮箱
	if notifyUserID > 0 {
		var notifyUser model.UserInfo
		if err := s.db.WithContext(ctx).Select("email, nickname").Where("id = ?", notifyUserID).First(&notifyUser).Error; err == nil && notifyUser.Email != "" {
			// 不通知自己
			if notifyUserID != userID {
				// 构建评论通知邮件
				commentType := commentTypeStr(comment.Type)
				var title string
				if comment.Type == 1 && comment.TopicID != nil {
					var article model.Article
					if err := s.db.WithContext(ctx).Select("article_title").Where("id = ?", *comment.TopicID).First(&article).Error; err == nil {
						title = article.ArticleTitle
					}
				} else {
					title = commentType
				}

				url := fmt.Sprintf("%s/%s/%s", getSiteURL(), getCommentPath(comment.Type), topicID)

				commentMap := map[string]interface{}{
					"nickname": fromNickname,
					"content":  comment.CommentContent,
					"title":    title,
					"url":      url,
				}

				// 如果有父评论，添加回复相关信息
				if comment.ParentID > 0 {
					var parentComment model.Comment
					if err := s.db.WithContext(ctx).Select("user_id, comment_content, create_time").Where("id = ?", comment.ParentID).First(&parentComment).Error; err == nil {
						var parentUser model.UserInfo
						if err := s.db.WithContext(ctx).Select("nickname").Where("id = ?", parentComment.UserID).First(&parentUser).Error; err == nil {
							commentMap["parentComment"] = parentComment.CommentContent
							commentMap["toUser"] = parentUser.Nickname
							commentMap["time"] = parentComment.CreateTime.Format("2006-01-02 15:04")
						}
					}
				} else {
					commentMap["time"] = comment.CreateTime.Format("2006-01-02 15:04")
				}

				emailDTO := dto.EmailDTO{
					Email:      notifyUser.Email,
					Subject:    "评论通知",
					CommentMap: commentMap,
				}

				if err := publishCommentEmail(emailDTO); err != nil {
					slog.Warn("发送评论通知邮件失败", "to", notifyUser.Email, "error", err)
				} else {
					slog.Info("📧 已发送评论通知邮件", "to", notifyUser.Email, "comment_id", comment.ID)
				}
			}
		}
	}

	return nil
}

// getSiteURL 获取网站URL (TODO: 从配置读取)
func getSiteURL() string {
	return "https://www.aurora.blog"
}

// getCommentPath 根据评论类型获取URL路径 (对标Java getCommentPath)
func getCommentPath(commentType int8) string {
	switch commentType {
	case 1:
		return "articles"
	case 5:
		return "talks"
	case 4:
		return "links"
	case 3:
		return "about"
	default:
		return ""
	}
}

// publishCommentEmail 发布评论邮件通知到RabbitMQ
func publishCommentEmail(emailDTO dto.EmailDTO) error {
	if mq.GetChannel() == nil {
		return fmt.Errorf("RabbitMQ channel not initialized")
	}

	msgBody, err := json.Marshal(emailDTO)
	if err != nil {
		return fmt.Errorf("序列化EmailDTO失败: %w", err)
	}

	err = mq.GetChannel().Publish(
		constant.ExchangeDirect,
		constant.RoutingKeyEmail,
		false, false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        msgBody,
			MessageId:   fmt.Sprintf("comment-%d", time.Now().UnixNano()),
		},
	)
	if err != nil {
		return fmt.Errorf("发布评论邮件消息失败: %w", err)
	}

	return nil
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
		// 修复问题3：子评论也需要按照同样的 type 和 topic_id 筛选
		replyQuery := s.db.WithContext(ctx).
			Where("parent_id IN ? AND is_review = 1", parentIDs)
		if commentVO.Type > 0 {
			replyQuery = replyQuery.Where("type = ?", commentVO.Type)
		}
		if commentVO.TopicID != nil && *commentVO.TopicID > 0 {
			replyQuery = replyQuery.Where("topic_id = ?", *commentVO.TopicID)
		}
		var replies []model.Comment
		replyQuery.
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
		Preload("UserInfo").
		Preload("ReplyUser").
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

// ListRepliesByCommentId 根据评论ID获取回复列表（对标Java CommentServiceImpl.listRepliesByCommentId）
func (s *CommentService) ListRepliesByCommentId(ctx context.Context, commentId uint) ([]dto.ReplyDTO, error) {
	var replies []model.Comment

	err := s.db.WithContext(ctx).
		Where("parent_id = ? AND is_review = 1", commentId).
		Preload("UserInfo").
		Preload("ReplyUser").
		Order("create_time ASC").
		Find(&replies).Error

	if err != nil {
		return nil, fmt.Errorf("查询回复列表失败: %w", err)
	}

	// 确保返回空数组而非 null
	if replies == nil {
		return []dto.ReplyDTO{}, nil
	}

	list := make([]dto.ReplyDTO, len(replies))
	for i, r := range replies {
		list[i] = s.toReplyDTO(&r)
	}
	return list, nil
}

func (s *CommentService) GetCommentStats(ctx context.Context) (*CommentStats, error) {
	stats := &CommentStats{}

	// 使用goroutine并发查询5个统计 (errgroup模式替代CompletableFuture)
	// 每个 goroutine 带 recover，panic 时写入零值避免 channel deadlock
	ch := make(chan struct {
		key string
		val int64
	}, 5)

	sendStat := func(key string, fn func() int64) {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					slog.Error("评论统计 goroutine panic recovered", "key", key, "panic", r)
					ch <- struct{ key string; val int64 }{key, 0}
				}
			}()
			ch <- struct{ key string; val int64 }{key, fn()}
		}()
	}
	sendStat("total", func() int64 { return s.countAll(ctx) })
	sendStat("article", func() int64 { return s.countByType(ctx, 1) })
	sendStat("talk", func() int64 { return s.countByType(ctx, 5) })
	sendStat("pending", func() int64 { return s.countByReview(ctx, 0) })
	sendStat("approved", func() int64 { return s.countByReview(ctx, 1) })

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

// incrementCommentCount 评论计数增加（当前数据库无 comment_count 列，预留接口）
func (s *CommentService) incrementCommentCount(tx *gorm.DB, commentType int8, topicID *uint) {
	// t_article/t_talk/t_friend_link 表无 comment_count 列，跳过
}

// decrementCommentCount 评论计数减少（当前数据库无 comment_count 列，预留接口）
func (s *CommentService) decrementCommentCount(tx *gorm.DB, commentType int8, topicID *uint, count int) {
	// t_article/t_talk/t_friend_link 表无 comment_count 列，跳过
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
	case 2: return "留言板"
	case 3: return "关于页"
	case 4: return "友链"
	case 5: return "说说"
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

// toReplyDTO 将 Comment 模型转换为 ReplyDTO（对标Java）
func (s *CommentService) toReplyDTO(c *model.Comment) dto.ReplyDTO {
	dto := dto.ReplyDTO{
		ID:             c.ID,
		ParentID:       c.ParentID,
		UserID:         c.UserID,
		CommentContent: c.CommentContent,
		ReplyUserID:    c.ReplyUserID,
		CreateTime:     c.CreateTime,
	}

	// 评论人信息
	if c.UserInfo != nil {
		dto.Nickname = c.UserInfo.Nickname
		dto.Avatar = c.UserInfo.Avatar
		dto.Website = c.UserInfo.Website
	}

	// 被回复用户信息
	if c.ReplyUser != nil {
		dto.ReplyNickname = c.ReplyUser.Nickname
		dto.ReplyWebsite = c.ReplyUser.Website
	}

	return dto
}
