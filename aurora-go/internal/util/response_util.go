package util

import (
	"fmt"
	"net/http"

	apperrors "github.com/aurora-go/aurora/internal/errors"

	"github.com/gin-gonic/gin"
)

// ResultVO 统一API响应结构（对齐Java版ResultVO）
// Java版本返回格式: { flag: true/false, data: ..., message: ... }
// Go版本需要对齐这个格式，同时添加 code 字段适配前端拦截器
type ResultVO struct {
	Code    int         `json:"code"`
	Flag    bool        `json:"flag"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, ResultVO{
		Code: 200,
		Flag: true,
		Data: data,
	})
}

// SuccessWithMessage 带自定义消息的成功响应
func SuccessWithMessage(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, ResultVO{
		Code:    200,
		Flag:    true,
		Data:    data,
		Message: msg,
	})
}

// Fail 错误响应
func Fail(c *gin.Context, code int, message string) {
	// 确保始终返回JSON格式,即使连接已中断也要设置正确的Content-Type
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.JSON(http.StatusOK, ResultVO{
		Flag:    false,
		Message: message,
	})
}

// PageResultVO 分页结果
// 对齐Java MyBatis-Plus的IPage结构，前端期望records字段
type PageResultVO struct {
	Records  interface{} `json:"records"`   // 数据列表（对齐前端records字段）
	List     interface{} `json:"list"`      // 兼容旧版list字段
	Count    int64       `json:"count"`     // 总记录数
	PageNum  int         `json:"pageNum"`
	PageSize int         `json:"pageSize"`
}

// NewPageResult 创建分页结果
func NewPageResult(list interface{}, count int64, pageNum, pageSize int) PageResultVO {
	return PageResultVO{
		Records:  list,   // 前端期望的字段
		List:     list,   // 兼容旧版
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
// 支持3种调用方式:
//   ResponseError(c, code, message)   - 标准方式
//   ResponseError(c, appError)        - 传入 *AppError
//   ResponseError(c, err)             - 传入普通 error
func ResponseError(c *gin.Context, args ...interface{}) {
	// 确保Content-Type为JSON
	c.Header("Content-Type", "application/json; charset=utf-8")
	
	switch len(args) {
	case 1:
		// 单参数: *AppError 或 error
		switch v := args[0].(type) {
		case *apperrors.AppError:
			Fail(c, v.Code, v.Message)
		case error:
			Fail(c, 500, v.Error())
		default:
			Fail(c, 500, fmt.Sprintf("%v", v))
		}
	case 2:
		// 双参数: code int, message string
		if code, ok := args[0].(int); ok {
			if msg, ok := args[1].(string); ok {
				Fail(c, code, msg)
				return
			}
		}
		Fail(c, 500, "invalid ResponseError arguments")
	default:
		Fail(c, 500, "invalid ResponseError arguments")
	}
}

// ResponseSuccessWithPage 分页成功响应别名
func ResponseSuccessWithPage(c *gin.Context, list interface{}, count int64, pageNum, pageSize int) {
	Success(c, NewPageResult(list, count, pageNum, pageSize))
}

// ResponseErrorWithAppError 使用AppError响应
func ResponseErrorWithAppError(c *gin.Context, appErr error) {
	if ae, ok := appErr.(*apperrors.AppError); ok {
		Fail(c, ae.Code, ae.Message)
	} else {
		Fail(c, 500, appErr.Error())
	}
}
