package enum

// ArticleStatus 文章状态枚举
type ArticleStatus int8

const (
	ArticleStatusDraft   ArticleStatus = 0 // 草稿
	ArticleStatusPublic  ArticleStatus = 1 // 公开
	ArticleStatusSecret  ArticleStatus = 2 // 密码保护
	ArticleStatusEncrypt ArticleStatus = 3 // 加密文章
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
	case ArticleStatusDraft:
		return "草稿"
	case ArticleStatusPublic:
		return "公开"
	case ArticleStatusSecret:
		return "密码保护"
	case ArticleStatusEncrypt:
		return "加密"
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
	return s == ArticleStatusPublic || s == ArticleStatusSecret || s == ArticleStatusEncrypt
}
