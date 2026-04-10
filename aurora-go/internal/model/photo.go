package model

import (
	"time"

	"gorm.io/gorm"
)

// Photo 照片实体 (对应 t_photo 表)
type Photo struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	AlbumID   uint      `gorm:"index;not null" json:"albumId"`
	URL       string    `gorm:"size:1024;not null" json:"url"` // 照片访问URL
	Sort      int       `gorm:"default:0" json:"sort"`
	CreateTime time.Time `json:"createTime"`
}

func (Photo) TableName() string { return "t_photo" }

// PhotoAlbum 相册实体 (对应 t_photo_album 表)
type PhotoAlbum struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	AlbumName  string    `gorm:"size:50;not null" json:"albumName"`
	AlbumCover string    `gorm:"size:1024" json:"albumCover"`
	Info       string    `gorm:"size:500" json:"info"`       // 相册描述
	Status     int8      `gorm:"default:1" json:"status"`    // 1公开 2私密
	PhotoCount int       `gorm:"default:0" json:"photoCount"`
	CreateTime time.Time `json:"createTime"`
	UpdateTime time.Time `json:"updateTime"`
}

func (PhotoAlbum) TableName() string { return "t_photo_album" }

func (p *Photo) BeforeCreate(tx *gorm.DB) error {
	p.CreateTime = time.Now()
	return nil
}

func (pa *PhotoAlbum) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	pa.CreateTime = now
	pa.UpdateTime = now
	return nil
}

func (pa *PhotoAlbum) BeforeUpdate(tx *gorm.DB) error {
	pa.UpdateTime = time.Now()
	return nil
}
