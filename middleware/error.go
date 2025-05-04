package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
)

// ErrorResponse 统一错误响应结构
type ErrorResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Error   string      `json:"error,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

// ErrorHandler 错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 使用defer来确保在请求处理完毕后执行错误捕获和处理
		defer func() {
			// 捕获panic
			if err := recover(); err != nil {
				// 获取堆栈信息
				stack := string(debug.Stack())
				stackLines := strings.Split(stack, "\n")

				// 打印简化的堆栈信息到日志
				fmt.Printf("Panic recovered: %v\nStack trace:\n%s\n", err, stack)

				// 对客户端隐藏完整堆栈信息，只显示必要的错误信息
				errMsg := fmt.Sprintf("%v", err)
				response := ErrorResponse{
					Code:    http.StatusInternalServerError,
					Message: "服务器内部错误",
					Error:   errMsg,
				}

				// 在开发模式下，可以返回更多调试信息
				if gin.Mode() == gin.DebugMode {
					if len(stackLines) > 10 {
						stackLines = stackLines[:10] // 限制堆栈行数
					}
					response.Details = stackLines
				}

				// 中止请求并返回JSON错误响应
				c.AbortWithStatusJSON(http.StatusInternalServerError, response)
			}
		}()

		// 处理下一个中间件或控制器
		c.Next()

		// 检查是否有错误设置
		if len(c.Errors) > 0 {
			// 获取最后一个错误
			err := c.Errors.Last()

			// 构建错误响应
			var response ErrorResponse
			var statusCode int

			// 根据错误类型设置状态码和错误信息
			switch err.Type {
			case gin.ErrorTypeBind:
				// 参数绑定错误
				statusCode = http.StatusBadRequest
				response = ErrorResponse{
					Code:    400,
					Message: "请求参数错误",
					Error:   err.Error(),
				}
			case gin.ErrorTypePrivate:
				// 业务逻辑错误
				statusCode = http.StatusBadRequest
				response = ErrorResponse{
					Code:    400,
					Message: err.Error(),
				}
			case gin.ErrorTypePublic:
				// 公开错误消息
				statusCode = http.StatusInternalServerError
				response = ErrorResponse{
					Code:    500,
					Message: err.Error(),
				}
			default:
				// 其他错误
				statusCode = http.StatusInternalServerError
				response = ErrorResponse{
					Code:    500,
					Message: "服务器内部错误",
					Error:   err.Error(),
				}
			}

			// 如果响应已经被写入，则不再重写
			if !c.Writer.Written() {
				c.JSON(statusCode, response)
			}
		}
	}
}

// ErrorWrapper 错误处理包装函数，用于在控制器中快速抛出错误
func ErrorWrapper(c *gin.Context, statusCode int, code int, message string, err error) {
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}

	response := ErrorResponse{
		Code:    code,
		Message: message,
		Error:   errStr,
	}

	// 在开发模式下，打印错误信息
	if gin.Mode() == gin.DebugMode && err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	c.AbortWithStatusJSON(statusCode, response)
}
