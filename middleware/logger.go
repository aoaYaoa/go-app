package middleware

import (
	"fmt"
	"time"

	"go-app/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 结束时间
		end := time.Now()
		latency := end.Sub(start)

		// 获取状态
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		userAgent := c.Request.UserAgent()

		// 构建日志字段
		fields := []zap.Field{
			zap.Int("status", status),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", clientIP),
			zap.String("user-agent", userAgent),
			zap.Duration("latency", latency),
		}

		// 收集错误信息
		var errorMsg string
		if len(c.Errors) > 0 {
			for _, e := range c.Errors.Errors() {
				fields = append(fields, zap.String("error", e))
				if errorMsg != "" {
					errorMsg += "; "
				}
				errorMsg += e
			}
		}

		// 根据状态码记录不同级别的日志
		msg := fmt.Sprintf("[GIN] %d %s %s", status, method, path)
		if status >= 500 {
			utils.Error(msg, fields...)
		} else if status >= 400 {
			utils.Warn(msg, fields...)
		} else {
			utils.Info(msg, fields...)
		}

		// 记录详细的请求日志到专门的日志文件
		reqLog := utils.RequestLog{
			Time:      time.Now(),
			Method:    method,
			Path:      path,
			Query:     query,
			Status:    status,
			IP:        clientIP,
			UserAgent: userAgent,
			LatencyMs: float64(latency.Microseconds()) / 1000.0, // 转换为毫秒
			Error:     errorMsg,
			// 收集更多信息
			Params:  extractParams(c),
			Headers: extractHeaders(c),
		}

		// 异步记录请求日志，不阻塞请求
		go utils.LogRequest(reqLog)
	}
}

// 从Gin上下文中提取路径参数
func extractParams(c *gin.Context) map[string]string {
	params := make(map[string]string)
	for _, param := range c.Params {
		params[param.Key] = param.Value
	}
	return params
}

// 从Gin上下文中提取请求头信息
func extractHeaders(c *gin.Context) map[string]string {
	headers := make(map[string]string)
	// 只收集重要的请求头，避免日志过大
	importantHeaders := []string{
		"Content-Type", "Accept", "Origin", "Referer",
		"X-Forwarded-For", "X-Real-IP", "User-Agent",
	}

	for _, name := range importantHeaders {
		if value := c.GetHeader(name); value != "" {
			headers[name] = value
		}
	}
	return headers
}
