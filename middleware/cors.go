package middleware

import (
	"time"

	"go-app/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Cors 跨域中间件
func Cors(cfg *config.Config) gin.HandlerFunc {
	// 配置跨域源
	allowOrigins := []string{"http://localhost:3000", "http://localhost:8080"}
	if len(cfg.CORS.AllowOrigins) > 0 {
		allowOrigins = cfg.CORS.AllowOrigins
	}

	// 配置有效期
	maxAge := 12 * time.Hour
	if cfg.CORS.MaxAge > 0 {
		maxAge = cfg.CORS.MaxAge
	}

	return cors.New(cors.Config{
		// 允许的源
		AllowOrigins: allowOrigins,
		// 允许的请求方法
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		// 允许的请求头
		AllowHeaders: []string{
			"Origin", "Content-Length", "Content-Type", "Authorization",
			"Accept", "X-Requested-With", "X-CSRF-Token", "signature",
			"app_key", "timestamp", "nonce", "sign",
		},
		// 是否允许携带认证信息（如cookies）
		AllowCredentials: cfg.CORS.AllowCredentials,
		// 预检请求的有效期
		MaxAge: maxAge,
	})
}
