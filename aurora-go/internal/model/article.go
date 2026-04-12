package model

import (
	"time"

	"gorm.io/gorm"
)

// Article 文章实体 (对应 t_article 表)
type Article struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	UserID         uint           `gorm:"index;not null" json:"userId"`
	CategoryID     *uint          `gorm:"index" json:"categoryId"`
	ArticleCover   string         `gorm:"size:1024" json:"articleCover"`
	ArticleTitle   string         `gorm:"size:50;not null;index" json:"articleTitle"`
	ArticleContent string         `gorm:"type:longtext;not null" json:"articleContent"`
	IsTop          int8           `gorm:"default:0;index" json:"isTop"`       // 0否 1是
	IsFeatured     int8           `gorm:"default:0;index" json:"isFeatured"`   // 是否推荐
	IsDelete       int8           `gorm:"default:0;index" json:"isDelete"`     // 软删除标记
	Status         int8           `gorm:"default:1;index" json:"status"`      // 1公开 2私密 3草稿
	Type           int8           `gorm:"default:1" json:"type"`              // 1原创 2转载 3翻译
	Password       string         `gorm:"size:255" json:"password"`            // 访问密码
	OriginalURL    string         `gorm:"size:255" json:"originalUrl"`        // 原文链接
	CreateTime     time.Time      `json:"createTime"`
	UpdateTime     *time.Time     `json:"updateTime,omitempty"`

	// 关联 (不存入数据库)
	Tags       []Tag       `gorm:"many2many:t_article_tag;" json:"tags,omitempty"`
	Category   *Category   `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	UserInfo   *UserInfo   `gorm:"foreignKey:UserID" json:"userInfo,omitempty"`
}

func (Article) TableName() string {
	return "t_article"
}

// BeforeCreate GORM钩子: 创建时自动填充时间
func (a *Article) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	a.CreateTime = now
	return nil
}

// BeforeUpdate GORM钩子: 更新时自动刷新时间
func (a *Article) BeforeUpdate(tx *gorm.DB) error {
	now := time.Now()
	a.UpdateTime = &now
	return nil
}

// ArticleTag 文章-标签关联表 (多对多中间表)
type ArticleTag struct {
	ArticleID uint `gorm:"primaryKey" json:"articleId"`
	TagID     uint `gorm:"primaryKey" json:"tagId"`
}

func (ArticleTag) TableName() string {
	return "t_article_tag"
}
