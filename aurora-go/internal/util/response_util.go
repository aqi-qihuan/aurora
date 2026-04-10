package util

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ResultVO 统一API响应结构（对齐Java版ResultVO）
type ResultVO struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, ResultVO{
		Code:    200,
		Message: "操作成功",
		Data:    data,
	})
}

// SuccessWithMessage 带自定义消息的成功响应
func SuccessWithMessage(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, ResultVO{
		Code:    200,
		Message: msg,
		Data:    data,
	})
}

// Fail 错误响应
func Fail(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, ResultVO{
		Code:    code,
		Message: message,
	})
}

// PageResultVO 分页结果
type PageResultVO struct {
	List     interface{} `json:"list"`
	Count    int64       `json:"count"`    // 总记录数 (非当前页)
	PageNum  int         `json:"pageNum"`
	PageSize int         `json:"pageSize"`
}

// NewPageResult 创建分页结果
func NewPageResult(list interface{}, count int64, pageNum, pageSize int) PageResultVO {
	return PageResultVO{
		List:     list,
		Count:    count,
		PageNum:  pageNum,
		PageSize: pageSize,
	}
}

// ===== Handler层常用别名 (兼容 handler 中的调用) =====

// ResponseSuccess 成功响应别名
func ResponseSuccess(c *gin.Context, data interface{}) {
	Success(c, data)
}

// ResponseError 错误响应别名
func ResponseError(c *gin.Context, code int, message string) {
	Fail(c, code, message)
}

// ResponseSuccessWithPage 分页成功响应别名
func ResponseSuccessWithPage(c *gin.Context, list interface{}, count int64, pageNum, pageSize int) {
	Success(c, NewPageResult(list, count, pageNum, pageSize))
}

// ResponseErrorWithAppError 使用AppError响应
func ResponseErrorWithAppError(c *gin.Context, appErr error) {
	if ae, ok := appErr.(*interface {
		Code    int
		Message string
	}); ok {
		Fail(c, ae.Code, ae.Message)
	} else {
		Fail(c, 500, appErr.Error())
	}
}
