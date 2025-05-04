package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"go-app/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuth JWT认证中间件
func JWTAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 临时禁用JWT验证
		c.Set("userID", uint(1)) // 设置一个默认用户ID
		c.Next()
		return

		// 从请求头中获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "请先登录",
			})
			c.Abort()
			return
		}

		// 检查token格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的认证格式",
			})
			c.Abort()
			return
		}

		// 解析token
		token := parts[1]
		claims, err := ParseToken(token, cfg.JWT.Secret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "认证失败: " + err.Error(),
			})
			c.Abort()
			return
		}

		// 将用户信息保存到上下文
		c.Set("userID", claims.UserID)
		c.Next()
	}
}

// Claims JWT claims
type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT令牌
func GenerateToken(userID uint, secret string, expire time.Duration) (string, error) {
	// 创建claims
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "user_token",
		},
	}

	// 创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名token
	return token.SignedString([]byte(secret))
}

// ParseToken 解析JWT令牌
func ParseToken(tokenString string, secret string) (*Claims, error) {
	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	// 验证token
	if !token.Valid {
		return nil, errors.New("令牌无效")
	}

	// 获取claims
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("无法获取令牌声明")
	}

	return claims, nil
}
