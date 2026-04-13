package dto

import "time"

// TalkDTO 说说前台展示DTO（对标 Java TalkDTO）
type TalkDTO struct {
	ID           uint     `json:"id"`
	Nickname     string   `json:"nickname"`
	Avatar       string   `json:"avatar"`
	Content      string   `json:"content"`
	Images       string   `json:"images"`
	Imgs         []string `json:"imgs"`         // 解析后的图片URL列表
	IsTop        int8     `json:"isTop"`
	CommentCount int      `json:"commentCount"`
	CreateTime   time.Time `json:"createTime"`
}

// TalkAdminDTO 说说后台管理DTO（对标 Java TalkAdminDTO）
type TalkAdminDTO struct {
	ID         uint     `json:"id"`
	Nickname   string   `json:"nickname"`
	Avatar     string   `json:"avatar"`
	Content    string   `json:"content"`
	Images     string   `json:"images"`
	Imgs       []string `json:"imgs"` // 解析后的图片URL列表
	IsTop      int8     `json:"isTop"`
	Status     int8     `json:"status"`
	CreateTime time.Time `json:"createTime"`
}
