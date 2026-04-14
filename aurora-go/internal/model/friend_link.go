package model

import (
	"time"
)

// FriendLink 友链实体 (对应 t_friend_link 表)
// 对标Java：只有 link_name, link_avatar, link_address, link_intro, create_time, update_time
type FriendLink struct {
	ID          uint      `gorm:"primarykey;column:id" json:"id"`
	LinkName    string    `gorm:"column:link_name;size:20;not null" json:"linkName"`
	LinkAvatar  string    `gorm:"column:link_avatar;size:255;not null" json:"linkAvatar"`
	LinkAddress string    `gorm:"column:link_address;size:50;not null" json:"linkAddress"`
	LinkIntro   string    `gorm:"column:link_intro;size:100;not null" json:"linkIntro"`
	CreateTime  time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateTime  time.Time `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
}

func (FriendLink) TableName() string { return "t_friend_link" }
