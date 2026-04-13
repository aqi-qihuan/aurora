package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/model"
	"github.com/aurora-go/aurora/internal/vo"
	"gorm.io/gorm"
)

// TalkService 说说业务逻辑 (完全对标 Java TalkServiceImpl)
type TalkService struct {
	db *gorm.DB
}

func NewTalkService(db *gorm.DB) *TalkService {
	return &TalkService{db: db}
}

// SaveOrUpdateTalk 保存或更新说说（对标 Java saveOrUpdateTalk）
// 根据ID有无自动判断新增/更新
func (s *TalkService) SaveOrUpdateTalk(ctx context.Context, userID uint, talkVO vo.TalkVO) error {
	talk := model.Talk{
		UserID:  userID,
		Content: talkVO.Content,
		Images:  talkVO.Images,
		IsTop:   talkVO.IsTop,
		Status:  talkVO.Status,
	}

	if talkVO.ID > 0 {
		// 更新：使用Updates()只更新非零值字段，对标Java MyBatis-Plus的saveOrUpdate行为
		// Omit("create_time") 确保不会覆盖原有创建时间
		if err := s.db.WithContext(ctx).Model(&model.Talk{}).Where("id = ?", talkVO.ID).Updates(map[string]interface{}{
			"user_id": userID,
			"content": talkVO.Content,
			"images":  talkVO.Images,
			"is_top":  talkVO.IsTop,
			"status":  talkVO.Status,
		}).Error; err != nil {
			return fmt.Errorf("更新说说失败: %w", err)
		}
	} else {
		// 新增
		if err := s.db.WithContext(ctx).Create(&talk).Error; err != nil {
			return fmt.Errorf("发布说说失败: %w", err)
		}
	}
	return nil
}

// DeleteTalks 批量删除说说（对标 Java deleteTalks）
// 物理删除，需要先删除关联评论
func (s *TalkService) DeleteTalks(ctx context.Context, ids []uint) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先删除关联评论（topic_id + type=5表示说说评论）
		tx.Where("topic_id IN ? AND type = ?", ids, 5).Delete(&model.Comment{})
		
		// 再删除说说本身
		result := tx.Unscoped().Where("id IN ?", ids).Delete(&model.Talk{})
		if result.Error != nil {
			return fmt.Errorf("删除说说失败: %w", result.Error)
		}
		return nil
	})
}

// GetTalks 获取公开说说列表（前台用，对标 Java listTalks）
// 使用JOIN查询用户信息，批量查询评论数，JSON解析images
func (s *TalkService) GetTalks(ctx context.Context, page dto.PageVO) (*dto.PageResultDTO, error) {
	type TalkRow struct {
		ID         uint
		Nickname   string
		Avatar     string
		Content    string
		Images     string
		IsTop      int8
		CreateTime time.Time `gorm:"column:create_time"`  // 直接使用 time.Time，让GORM自动转换
	}

	var talks []TalkRow
	var count int64

	// 统计总数
	s.db.WithContext(ctx).Model(&model.Talk{}).Where("status = 1").Count(&count)
	if count == 0 {
		return &dto.PageResultDTO{
			List:     []dto.TalkDTO{},
			Count:    0,
			PageNum:  page.PageNum,
			PageSize: page.PageSize,
		}, nil
	}

	offset := page.GetOffset()
	err := s.db.WithContext(ctx).
		Table("t_talk t").
		Select("t.id, ui.nickname, ui.avatar, t.content, t.images, t.is_top, t.create_time").
		Joins("JOIN t_user_info ui ON t.user_id = ui.id").
		Where("t.status = 1").
		Order("t.is_top DESC, t.id DESC").
		Limit(page.PageSize).
		Offset(offset).
		Find(&talks).Error

	if err != nil {
		return nil, fmt.Errorf("查询说说列表失败: %w", err)
	}

	list := make([]dto.TalkDTO, len(talks))
	talkIDs := make([]uint, len(talks))
	for i, t := range talks {
		talkIDs[i] = t.ID
		list[i] = dto.TalkDTO{
			ID:         t.ID,
			Nickname:   t.Nickname,
			Avatar:     t.Avatar,
			Content:    t.Content,
			Images:     t.Images,
			IsTop:      t.IsTop,
			CreateTime: t.CreateTime,  // 直接使用，GORM已自动转换
		}
		// 解析images JSON为List<String>
		if t.Images != "" {
			json.Unmarshal([]byte(t.Images), &list[i].Imgs)
		}
	}

	// 批量查询评论数（对标Java: commentMapper.listCommentCountByTypeAndTopicIds）
	if len(talkIDs) > 0 {
		var commentCounts []struct {
			TopicID uint
			Count   int
		}
		s.db.WithContext(ctx).
			Table("t_comment").
			Select("topic_id, COUNT(*) as count").
			Where("topic_id IN ? AND type = 5 AND is_review = 1", talkIDs).
			Group("topic_id").
			Find(&commentCounts)

		countMap := make(map[uint]int)
		for _, cc := range commentCounts {
			countMap[cc.TopicID] = cc.Count
		}
		for i := range list {
			list[i].CommentCount = countMap[list[i].ID]
		}
	}

	return &dto.PageResultDTO{
		List:     list,
		Count:    count,
		PageNum:  page.PageNum,
		PageSize: page.PageSize,
	}, nil
}

// GetTalkByID 根据ID获取说说详情（前台用，对标 Java getTalkById）
func (s *TalkService) GetTalkByID(ctx context.Context, id uint) (*dto.TalkDTO, error) {
	type TalkRow struct {
		ID         uint
		Nickname   string
		Avatar     string
		Content    string
		Images     string
		IsTop      int8
		CreateTime time.Time `gorm:"column:create_time"`  // 直接使用 time.Time，让GORM自动转换
	}

	var talk TalkRow
	err := s.db.WithContext(ctx).
		Table("t_talk t").
		Select("t.id, ui.nickname, ui.avatar, t.content, t.images, t.is_top, t.create_time").
		Joins("JOIN t_user_info ui ON t.user_id = ui.id").
		Where("t.id = ? AND t.status = 1", id).
		First(&talk).Error

	if err != nil {
		if errors.IsStd(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrTalkNotFound
		}
		return nil, fmt.Errorf("查询说说详情失败: %w", err)
	}

	dto := &dto.TalkDTO{
		ID:         talk.ID,
		Nickname:   talk.Nickname,
		Avatar:     talk.Avatar,
		Content:    talk.Content,
		Images:     talk.Images,
		IsTop:      talk.IsTop,
		CreateTime: talk.CreateTime,  // 直接使用，GORM已自动转换
	}

	// 解析images JSON
	if talk.Images != "" {
		json.Unmarshal([]byte(talk.Images), &dto.Imgs)
	}

	// 查询评论数（对标Java: commentMapper.listCommentCountByTypeAndTopicId）
	var commentCount struct {
		TopicID uint
		Count   int
	}
	s.db.WithContext(ctx).
		Table("t_comment").
		Select("topic_id, COUNT(*) as count").
		Where("topic_id = ? AND type = 5 AND is_review = 1", id).
		Group("topic_id").
		First(&commentCount)

	dto.CommentCount = commentCount.Count

	return dto, nil
}

// ListAdminTalks 后台管理分页查询说说（对标 Java listBackTalks）
// 返回TalkAdminDTO，包含status字段
func (s *TalkService) ListAdminTalks(ctx context.Context, cond dto.ConditionVO, page dto.PageVO) (*dto.PageResultDTO, error) {
	type TalkRow struct {
		ID         uint
		Nickname   string
		Avatar     string
		Content    string
		Images     string
		IsTop      int8
		Status     int8
		CreateTime time.Time `gorm:"column:create_time"`  // 直接使用 time.Time，让GORM自动转换
	}

	var talks []TalkRow
	var count int64

	baseQuery := s.db.WithContext(ctx).Table("t_talk t").
		Select("t.id, ui.nickname, ui.avatar, t.content, t.images, t.is_top, t.status, t.create_time").
		Joins("JOIN t_user_info ui ON t.user_id = ui.id")

	// 条件过滤
	if cond.Status != nil && *cond.Status > 0 {
		baseQuery = baseQuery.Where("t.status = ?", *cond.Status)
	}

	baseQuery.Count(&count)
	if count == 0 {
		return &dto.PageResultDTO{
			List:     []dto.TalkAdminDTO{},
			Count:    0,
			PageNum:  page.PageNum,
			PageSize: page.PageSize,
		}, nil
	}

	offset := page.GetOffset()
	err := baseQuery.
		Order("t.is_top DESC, t.id DESC").
		Limit(page.PageSize).
		Offset(offset).
		Find(&talks).Error

	if err != nil {
		return nil, fmt.Errorf("查询说说列表失败: %w", err)
	}

	list := make([]dto.TalkAdminDTO, len(talks))
	for i, t := range talks {
		list[i] = dto.TalkAdminDTO{
			ID:         t.ID,
			Nickname:   t.Nickname,
			Avatar:     t.Avatar,
			Content:    t.Content,
			Images:     t.Images,
			IsTop:      t.IsTop,
			Status:     t.Status,
			CreateTime: t.CreateTime,  // 直接使用，GORM已自动转换
		}
		// 解析images JSON
		if t.Images != "" {
			json.Unmarshal([]byte(t.Images), &list[i].Imgs)
		}
	}

	return &dto.PageResultDTO{
		List:     list,
		Count:    count,
		PageNum:  page.PageNum,
		PageSize: page.PageSize,
	}, nil
}

// GetAdminTalkByID 后台获取说说详情（对标 Java getBackTalkById）
func (s *TalkService) GetAdminTalkByID(ctx context.Context, id uint) (*dto.TalkAdminDTO, error) {
	type TalkRow struct {
		ID         uint
		Nickname   string
		Avatar     string
		Content    string
		Images     string
		IsTop      int8
		Status     int8
		CreateTime time.Time `gorm:"column:create_time"`  // 直接使用 time.Time，让GORM自动转换
	}

	var talk TalkRow
	err := s.db.WithContext(ctx).
		Table("t_talk t").
		Select("t.id, ui.nickname, ui.avatar, t.content, t.images, t.is_top, t.status, t.create_time").
		Joins("JOIN t_user_info ui ON t.user_id = ui.id").
		Where("t.id = ?", id).
		First(&talk).Error

	if err != nil {
		if errors.IsStd(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrTalkNotFound
		}
		return nil, fmt.Errorf("查询说说详情失败: %w", err)
	}

	dto := &dto.TalkAdminDTO{
		ID:         talk.ID,
		Nickname:   talk.Nickname,
		Avatar:     talk.Avatar,
		Content:    talk.Content,
		Images:     talk.Images,
		IsTop:      talk.IsTop,
		Status:     talk.Status,
		CreateTime: talk.CreateTime,  // 直接使用，GORM已自动转换
	}

	// 解析images JSON
	if talk.Images != "" {
		json.Unmarshal([]byte(talk.Images), &dto.Imgs)
	}

	return dto, nil
}

// parseTime 解析时间字符串（支持多种格式，对标Java的LocalDateTime）
func (s *TalkService) parseTime(timeStr string) time.Time {
	if timeStr == "" {
		return time.Time{}
	}
	
	// 调试日志：查看原始时间字符串
	log.Printf("[parseTime] 原始时间字符串: '%s'", timeStr)
	
	// 尝试多种时间格式（优先匹配带毫秒的格式）
	formats := []string{
		"2006-01-02 15:04:05.000",  // MySQL datetime with milliseconds (优先)
		"2006-01-02 15:04:05",      // MySQL datetime without milliseconds
		"2006-01-02T15:04:05Z",     // ISO8601
		"2006-01-02T15:04:05.000Z", // ISO8601 with milliseconds
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			log.Printf("[parseTime] 成功解析，使用格式: %s", format)
			return t
		}
	}
	
	// 如果都失败，尝试使用 time.ParseInLocation
	if t, err := time.ParseInLocation("2006-01-02 15:04:05", timeStr, time.Local); err == nil {
		log.Printf("[parseTime] 使用 ParseInLocation 成功")
		return t
	}
	
	// 如果还是失败，记录日志并返回零值
	log.Printf("[parseTime] 解析失败! 原始字符串: '%s'", timeStr)
	return time.Time{}
}
