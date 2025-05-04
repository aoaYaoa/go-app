package middleware

import (
	"crypto/md5"
	"encoding/hex"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// SignatureConfig 签名配置
type SignatureConfig struct {
	AppKey    string        // 应用key
	AppSecret string        // 应用密钥
	Expire    time.Duration // 签名有效期
}

// SignatureParams 签名参数
type SignatureParams struct {
	AppKey    string `form:"app_key"`
	Timestamp int64  `form:"timestamp"`
	Nonce     string `form:"nonce"`
	Sign      string `form:"sign"`
}

// Signature 签名验证中间件
func Signature(config *SignatureConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 临时禁用签名验证
		c.Next()
		return

		// 调试信息
		log.Printf("收到请求: %s %s", c.Request.Method, c.Request.URL.Path)

		// OPTIONS请求直接放行
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// 从请求头中获取签名参数
		sign := c.Request.Header.Get("signature")
		if sign == "" {
			// 尝试从查询参数获取
			var params SignatureParams
			if err := c.ShouldBindQuery(&params); err != nil {
				log.Printf("签名验证失败: %v", err)
				ErrorWrapper(c, http.StatusBadRequest, 400, "签名参数错误", err)
				return
			}

			// 验证AppKey
			if params.AppKey != config.AppKey {
				ErrorWrapper(c, http.StatusBadRequest, 400, "无效的AppKey", nil)
				return
			}

			// 验证时间戳
			now := time.Now().Unix()
			if now-params.Timestamp > int64(config.Expire.Seconds()) {
				ErrorWrapper(c, http.StatusBadRequest, 400, "签名已过期", nil)
				return
			}

			// 获取所有请求参数
			queryParams := c.Request.URL.Query()
			formParams := c.Request.PostForm

			// 合并所有参数
			allParams := make(map[string]string)
			for key, values := range queryParams {
				if key != "sign" { // 排除签名参数
					allParams[key] = values[0]
				}
			}
			for key, values := range formParams {
				if key != "sign" { // 排除签名参数
					allParams[key] = values[0]
				}
			}

			// 按参数名排序
			var keys []string
			for k := range allParams {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			// 构建签名字符串
			var signStr strings.Builder
			for _, k := range keys {
				signStr.WriteString(k)
				signStr.WriteString("=")
				signStr.WriteString(allParams[k])
				signStr.WriteString("&")
			}
			signStr.WriteString("app_secret=")
			signStr.WriteString(config.AppSecret)

			// 计算MD5签名
			hash := md5.New()
			hash.Write([]byte(signStr.String()))
			calculatedSign := hex.EncodeToString(hash.Sum(nil))

			// 验证签名
			if calculatedSign != params.Sign {
				ErrorWrapper(c, http.StatusBadRequest, 400, "签名验证失败", nil)
				return
			}

			// 将参数存储到上下文中，以便后续使用
			c.Set("signatureParams", params)
		}

		c.Next()
	}
}

// GetSignatureParams 从上下文中获取签名参数
func GetSignatureParams(c *gin.Context) *SignatureParams {
	if params, exists := c.Get("signatureParams"); exists {
		return params.(*SignatureParams)
	}
	return nil
}
