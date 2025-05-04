package middleware

import (
	"go-app/config"

	"github.com/gin-gonic/gin"
)

// SetupMiddlewares 统一设置所有中间件
func SetupMiddlewares(r *gin.Engine, cfg *config.Config) {
	// 日志中间件（放在最前面，记录所有请求）
	r.Use(Logger())

	// 全局错误处理中间件
	r.Use(ErrorHandler())

	// 跨域中间件
	r.Use(Cors(cfg))

	// 签名验证中间件
	r.Use(Signature(&SignatureConfig{
		AppKey:    cfg.Signature.AppKey,
		AppSecret: cfg.Signature.AppSecret,
		Expire:    cfg.Signature.Expire,
	}))
}

// SetupAuthMiddleware 设置认证中间件
func SetupAuthMiddleware(r *gin.RouterGroup, cfg *config.Config) {
	// JWT认证
	r.Use(JWTAuth(cfg))
}
