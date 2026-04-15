package enum

// ArticleStatus 文章状态枚举 (对齐Java ArticleStatusEnum)
// Java版: PUBLIC(1), SECRET(2), DRAFT(3)
type ArticleStatus int8

const (
	ArticleStatusPublic ArticleStatus = 1 // 公开
	ArticleStatusSecret ArticleStatus = 2 // 密码保护
	ArticleStatusDraft  ArticleStatus = 3 // 草稿
)

// ArticleType 文章类型枚举
type ArticleType int8

const (
	ArticleTypeOriginal    ArticleType = 1 // 原创
	ArticleTypeReprint     ArticleType = 2 // 转载
	ArticleTypeTranslation ArticleType = 3 // 翻译
)

// String 返回状态的字符串描述
func (s ArticleStatus) String() string {
	switch s {
	case ArticleStatusPublic:
		return "公开"
	case ArticleStatusSecret:
		return "密码保护"
	case ArticleStatusDraft:
		return "草稿"
	default:
		return "未知"
	}
}

func (t ArticleType) String() string {
	switch t {
	case ArticleTypeOriginal:
		return "原创"
	case ArticleTypeReprint:
		return "转载"
	case ArticleTypeTranslation:
		return "翻译"
	default:
		return "未知"
	}
}

// IsPublished 判断文章是否已发布（前台可见）
func (s ArticleStatus) IsPublished() bool {
	return s == ArticleStatusPublic || s == ArticleStatusSecret
}
