package vo

// TalkVO 说说请求对象（对标 Java TalkVO）
// 对标Java: TalkVO { id, content, images, isTop, status }
type TalkVO struct {
	ID      uint   `json:"id"`
	Content string `json:"content" binding:"required"` // 说说内容不能为空
	Images  string `json:"images"`                     // 说说图片JSON数组字符串
	IsTop   int8   `json:"isTop" binding:"required"`   // 置顶状态不能为空
	Status  int8   `json:"status" binding:"required"`  // 说说状态不能为空
}
