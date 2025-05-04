package middleware

import (
	"net/http"

	"go-app/config"

	"github.com/gin-gonic/gin"
)

// WhitelistConfig 白名单配置
type WhitelistConfig struct {
	// IP白名单列表
	IPWhitelist []string
	// 路径白名单列表（不需要验证的路径）
	PathWhitelist []string
	// 是否启用IP白名单
	EnableIPWhitelist bool
	// 是否启用路径白名单
	EnablePathWhitelist bool
}

// DefaultWhitelistConfig 默认白名单配置
var DefaultWhitelistConfig = WhitelistConfig{
	IPWhitelist:         []string{},
	PathWhitelist:       []string{},
	EnableIPWhitelist:   false,
	EnablePathWhitelist: false,
}

// NewWhitelistConfig 从应用配置创建白名单配置
func NewWhitelistConfig(cfg *config.Config) WhitelistConfig {
	return WhitelistConfig{
		IPWhitelist:         cfg.Whitelist.IPWhitelist,
		PathWhitelist:       cfg.Whitelist.PathWhitelist,
		EnableIPWhitelist:   cfg.Whitelist.EnableIPWhitelist,
		EnablePathWhitelist: cfg.Whitelist.EnablePathWhitelist,
	}
}

// Whitelist 白名单中间件
func Whitelist(config WhitelistConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查路径白名单
		if config.EnablePathWhitelist {
			path := c.Request.URL.Path
			for _, whitelistPath := range config.PathWhitelist {
				if path == whitelistPath {
					c.Next()
					return
				}
			}
		}

		// 检查IP白名单
		if config.EnableIPWhitelist {
			clientIP := c.ClientIP()
			for _, whitelistIP := range config.IPWhitelist {
				if clientIP == whitelistIP {
					c.Next()
					return
				}
			}
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "IP地址不在白名单中",
			})
			return
		}

		c.Next()
	}
}

// IsIPInWhitelist 检查IP是否在白名单中
func IsIPInWhitelist(ip string, whitelist []string) bool {
	for _, whitelistIP := range whitelist {
		if ip == whitelistIP {
			return true
		}
	}
	return false
}

// IsPathInWhitelist 检查路径是否在白名单中
func IsPathInWhitelist(path string, whitelist []string) bool {
	for _, whitelistPath := range whitelist {
		if path == whitelistPath {
			return true
		}
	}
	return false
}

// AddToIPWhitelist 添加IP到白名单
func AddToIPWhitelist(ip string) {
	DefaultWhitelistConfig.IPWhitelist = append(DefaultWhitelistConfig.IPWhitelist, ip)
}

// AddToPathWhitelist 添加路径到白名单
func AddToPathWhitelist(path string) {
	DefaultWhitelistConfig.PathWhitelist = append(DefaultWhitelistConfig.PathWhitelist, path)
}

// RemoveFromIPWhitelist 从白名单中移除IP
func RemoveFromIPWhitelist(ip string) {
	for i, whitelistIP := range DefaultWhitelistConfig.IPWhitelist {
		if whitelistIP == ip {
			DefaultWhitelistConfig.IPWhitelist = append(DefaultWhitelistConfig.IPWhitelist[:i], DefaultWhitelistConfig.IPWhitelist[i+1:]...)
			break
		}
	}
}

// RemoveFromPathWhitelist 从白名单中移除路径
func RemoveFromPathWhitelist(path string) {
	for i, whitelistPath := range DefaultWhitelistConfig.PathWhitelist {
		if whitelistPath == path {
			DefaultWhitelistConfig.PathWhitelist = append(DefaultWhitelistConfig.PathWhitelist[:i], DefaultWhitelistConfig.PathWhitelist[i+1:]...)
			break
		}
	}
}
