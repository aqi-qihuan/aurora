package model

import (
	"time"

	"gorm.io/gorm"
)

// Comment 评论实体 (对标 Java Comment.java)
// 数据库实际字段: topic_id (统一引用文章/说说/友链/关于的ID), 通过 type 区分类型
type Comment struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	UserID         uint      `gorm:"index" json:"userId"`
	TopicID        *uint     `gorm:"column:topic_id;index" json:"topicId"`          // 对标Java topicId
	CommentContent string    `gorm:"column:comment_content;type:text;not null" json:"commentContent"` // 对标Java
	ReplyUserID    *uint     `gorm:"column:reply_user_id" json:"replyUserId"`
	ParentID       uint      `gorm:"default:0;index" json:"parentId"`
	Type           int8      `gorm:"not null;index" json:"type"` // 1文章 2留言 3关于我 4友链 5说说
	IsDelete       int8      `gorm:"default:0" json:"isDelete"`
	IsReview       int8      `gorm:"default:1;index" json:"isReview"`
	CreateTime     time.Time `json:"createTime"`
	UpdateTime     *time.Time `json:"updateTime"`

	// 关联
	UserInfo  *UserInfo `gorm:"foreignKey:UserID" json:"userInfo,omitempty"`
	ReplyUser *UserInfo `gorm:"foreignKey:ReplyUserID" json:"replyUser,omitempty"`
}

func (Comment) TableName() string { return "t_comment" }

func (c *Comment) BeforeCreate(tx *gorm.DB) error {
	c.CreateTime = time.Now()
	return nil
}

func (c *Comment) BeforeUpdate(tx *gorm.DB) error {
	now := time.Now()
	c.UpdateTime = &now
	return nil
}
