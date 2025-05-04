package router

import (
	"go-app/config"
	"go-app/controller"
	"go-app/database/repositories"
	"go-app/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Setup 初始化所有路由
func Setup(r *gin.Engine, cfg *config.Config, repoManager *repositories.RepositoryManager) {
	// 初始化控制器管理器
	controllerManager := controller.NewManager(cfg, repoManager)

	// 设置健康检查
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// API路由组
	api := r.Group("/api/v1")
	{
		// 添加重定向，将/api/v1/login重定向到/api/v1/users/login
		api.Any("/login", func(c *gin.Context) {
			c.Redirect(http.StatusMovedPermanently, "/api/v1/users/login")
		})

		// 公开路由组
		public := api.Group("")
		// 需要认证的路由组
		authorized := api.Group("")
		// 添加JWT认证
		middleware.SetupAuthMiddleware(authorized, cfg)

		// 设置用户路由
		SetupUserRoutes(controllerManager.User, public, authorized)
	}
}

// SetupRouter 设置并返回配置好的路由器
func SetupRouter(cfg *config.Config, repoManager *repositories.RepositoryManager) *gin.Engine {
	r := gin.Default()

	// 使用白名单中间件
	r.Use(middleware.Whitelist(middleware.NewWhitelistConfig(cfg)))

	// 初始化路由
	Setup(r, cfg, repoManager)

	return r
}
