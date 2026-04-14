package vo

// TalkVO 说说请求对象（对标 Java TalkVO）
// 对标Java: TalkVO { id, content, images, isTop, status }
type TalkVO struct {
	ID      uint   `json:"id"`
	Content string `json:"content" binding:"required"` // 说说内容不能为空
	Images  string `json:"images"`                     // 说说图片JSON数组字符串
	IsTop   int8   `json:"isTop"`                      // 置顶状态(0否 1是)，去掉required避免0值校验失败
	Status  int8   `json:"status"`                     // 说说状态(1公开 2私密)，去掉required避免0值校验失败
}
