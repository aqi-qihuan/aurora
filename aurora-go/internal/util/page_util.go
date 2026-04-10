package util

import (
	"fmt"
	"math"
	"reflect"

	"github.com/gin-gonic/gin"
)

// PageResult 通用分页结果（对标 Java 的 PageResult / IPage）
type PageResult struct {
	List      interface{} `json:"records"`     // 数据列表
	Total     int64       `json:"total"`       // 总记录数
	PageNum   int         `json:"pageNum"`     // 当前页码
	PageSize  int         `json:"pageSize"`    // 每页大小
	TotalPages int        `json:"totalPages"`  // 总页数
	HasPrev   bool        `json:"hasPrevious"` // 是否有上一页
	HasNext   bool        `json:"hasNext"`     // 是否有下一页
}

// PageQuery 分页查询参数（从请求中提取）
// 对标 Java 的 ConditionVO + MyBatis-Plus Page
func PageQuery(c *gin.Context) (pageNum int, pageSize int) {
	pageNum = DefaultIntQuery(c, "current", 1)
	pageSize = DefaultIntQuery(c, "size", 10)

	// 边界保护
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100 // 防止一次性拉取过多数据
	}
	return
}

// BuildPageResult 构建分页结果
func BuildPageResult(list interface{}, total int64, pageNum, pageSize int) *PageResult {
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	if totalPages == 0 {
		totalPages = 1
	}
	return &PageResult{
		List:       list,
		Total:      total,
		PageNum:    pageNum,
		PageSize:   pageSize,
		TotalPages: totalPages,
		HasPrev:    pageNum > 1,
		HasNext:    pageNum < totalPages,
	}
}

// Offset 计算数据库 OFFSET 值
func Offset(pageNum, pageSize int) int {
	return (pageNum - 1) * pageSize
}

// GetOffset 从 Gin context 获取 offset（便捷方法）
func GetOffset(c *gin.Context) int {
	pageNum, pageSize := PageQuery(c)
	return Offset(pageNum, pageSize)
}

// DefaultIntQuery 获取整数查询参数，带默认值
func DefaultIntQuery(c *gin.Context, key string, defaultVal int) int {
	val, exists := c.GetQuery(key)
	if !exists {
		return defaultVal
	}
	var result int
	if _, err := ParseInt(val, &result); err != nil || result < 0 {
		return defaultVal
	}
	return result
}

// ParseInt 通用整数解析器（替代 strconv.Atoi，支持更宽松的输入）
func ParseInt(s string, target interface{}) (int, error) {
	var result int
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			result = result*10 + int(ch-'0')
		} else if result == 0 && ch == '-' {
			continue
		} else {
			break
		}
	}
	v := reflect.ValueOf(target)
	if v.Kind() == reflect.Ptr && v.Elem().CanSet() {
		switch v.Elem().Kind() {
		case reflect.Int:
			v.Elem().SetInt(int64(result))
		case reflect.Int8:
			v.Elem().SetInt(int64(result))
		case reflect.Int16:
			v.Elem().SetInt(int64(result))
		case reflect.Int32:
			v.Elem().SetInt(int64(result))
		case reflect.Int64:
			v.Elem().SetInt(int64(result))
		default:
			return 0, ErrUnsupportedType
		}
	}
	return result, nil
}

var ErrUnsupportedType = fmt.Errorf("unsupported target type for ParseInt")
