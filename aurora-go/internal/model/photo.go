package model

import (
	"time"

	"gorm.io/gorm"
)

// Photo 照片实体 (对应 t_photo 表)
type Photo struct {
	ID         uint      `gorm:"primarykey;column:id" json:"id"`
	AlbumID    uint      `gorm:"column:album_id;not null;index" json:"albumId"`
	PhotoName  string    `gorm:"column:photo_name;size:20;not null" json:"photoName"`   // 照片名（Java版：用雪花算法生成ID字符串）
	PhotoDesc  string    `gorm:"column:photo_desc;size:50" json:"photoDesc,omitempty"`    // 照片描述
	PhotoSrc   string    `gorm:"column:photo_src;size:255;not null" json:"photoSrc"`      // 照片URL
	IsDelete   int8      `gorm:"column:is_delete;default:0" json:"isDelete"`              // 是否删除 0否1是
	CreateTime time.Time `gorm:"column:create_time" json:"createTime"`
	UpdateTime *time.Time `gorm:"column:update_time" json:"updateTime,omitempty"`
}

func (Photo) TableName() string { return "t_photo" }

func (p *Photo) BeforeCreate(tx *gorm.DB) error {
	p.CreateTime = time.Now()
	return nil
}

func (p *Photo) BeforeUpdate(tx *gorm.DB) error {
	now := time.Now()
	p.UpdateTime = &now
	return nil
}

// PhotoAlbum 相册实体 (对应 t_photo_album 表)
type PhotoAlbum struct {
	ID         uint      `gorm:"primarykey;column:id" json:"id"`
	AlbumName  string    `gorm:"column:album_name;size:20;not null" json:"albumName"`
	AlbumDesc  string    `gorm:"column:album_desc;size:50;not null" json:"albumDesc"` // 相册描述（Java用albumDesc）
	AlbumCover string    `gorm:"column:album_cover;size:255;not null" json:"albumCover"`
	IsDelete   int8      `gorm:"column:is_delete;default:0" json:"isDelete"`           // 是否删除 0否1是
	Status     int8      `gorm:"column:status;default:1" json:"status"`                // 1公开 2私密
	CreateTime time.Time `gorm:"column:create_time" json:"createTime"`
	UpdateTime *time.Time `gorm:"column:update_time" json:"updateTime,omitempty"`
}

func (PhotoAlbum) TableName() string { return "t_photo_album" }

func (pa *PhotoAlbum) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	pa.CreateTime = now
	pa.UpdateTime = &now
	return nil
}

func (pa *PhotoAlbum) BeforeUpdate(tx *gorm.DB) error {
	now := time.Now()
	pa.UpdateTime = &now
	return nil
}
