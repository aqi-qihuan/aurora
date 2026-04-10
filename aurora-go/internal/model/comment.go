package model

import (
	"time"

	"gorm.io/gorm"
)

// Comment 评论实体 (对应 t_comment 表)
type Comment struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	UserID      uint      `gorm:"index" json:"userId"`
	ArticleID   *uint     `gorm:"index;comment:文章ID,可为空(说说/友链/关于等)" json:"articleId"`
	TalkID      *uint     `json:"talkId,omitempty"`                      // 说说ID(说说评论时)
	FriendLinkID *uint    `json:"friendLinkId,omitempty"`                 // 友链ID(友链评论时)
	AboutID     *uint     `json:"aboutId,omitempty"`                     // 关于页ID(关于评论时)
	Type        int8      `gorm:"not null;index" json:"type"`             // 1文章 2说说 3友链 4关于 5留言
	ParentID    uint      `gorm:"default:0;index" json:"parentId"`        // 0=顶级评论
	ReplyUserID *uint     `json:"replyUserId,omitempty"`                 // 被回复用户ID(回复时)
	IsReview    int8      `gorm:"default:0;index" json:"isReview"`         // 是否审核通过(0待审核 1已通过)
	Content     string    `gorm:"size:2000;not null" json:"content"`
	IP          string    `gorm:"size:64" json:"ip"`
	Location    string    `gorm:"size:50" json:"location"`
	LikeCount   int64     `gorm:"default:0" json:"likeCount"`
	CreateTime  time.Time `json:"createTime"`

	// 关联
	UserInfo   *UserInfo `gorm:"foreignKey:UserID" json:"userInfo,omitempty"`
	ReplyUser  *UserInfo `gorm:"foreignKey:ReplyUserID" json:"replyUser,omitempty"`
}

func (Comment) TableName() string { return "t_comment" }

func (c *Comment) BeforeCreate(tx *gorm.DB) error {
	c.CreateTime = time.Now()
	return nil
}
