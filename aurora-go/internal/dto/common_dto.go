package dto

// PageResultDTO 通用分页结果
type PageResultDTO struct {
	List     interface{} `json:"list"`
	Count    int64       `json:"count"`    // 总记录数
	PageNum  int         `json:"pageNum"`
	PageSize int         `json:"pageSize"`
}

// ConditionVO 通用查询条件
type ConditionVO struct {
	Keywords string `form:"keywords"`   // 搜索关键词
	Status   *int8  `form:"status"`     // 状态筛选
	Type     *int8  `form:"type"`       // 类型筛选
	CategoryID *uint `form:"categoryId"`
	IsDelete *int8  `form:"isDelete"`
	DateStart string `form:"dateStart"` // 时间范围起始
	DateEnd   string `form:"dateEnd"`   // 时间范围结束
	Sort      string `form:"sort"`       // 排序字段
	Order     string `form:"order"`      // asc/desc
}

// PageVO 分页参数
type PageVO struct {
	PageNum  int `form:"pageNum,default=1" json:"pageNum"`
	PageSize int `form:"pageSize,default=10" json:"pageSize"`
}

// GetOffset 计算SQL OFFSET
func (p *PageVO) GetOffset() int {
	if p.PageNum < 1 {
		p.PageNum = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 10
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
	return (p.PageNum - 1) * p.PageSize
}

// DeleteVO 删除请求
type DeleteVO struct {
	ID uint `uri:"id" binding:"required" json:"id" binding:"required"`
}

// BatchDeleteVO 批量删除请求
type BatchDeleteVO struct {
	IDList []uint `json:"idList" binding:"required,min=1"`
}
