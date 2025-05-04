package router

import (
	"go-app/controller/user"

	"github.com/gin-gonic/gin"
)

// SetupUserRoutes 设置用户相关路由
func SetupUserRoutes(controller *user.Controller, public, authorized *gin.RouterGroup) {
	// 公开路由
	users := public.Group("/users")
	{
		// 注册
		users.POST("/register", controller.Register)
		// 登录
		users.POST("/login", controller.Login)
	}

	// 需要认证的路由
	authUsers := authorized.Group("/users")
	{
		// 获取用户列表
		authUsers.GET("", controller.GetUsers)
		// 获取用户详情
		authUsers.GET("/:id", controller.GetUser)
		// 删除用户
		authUsers.DELETE("/:id", controller.DeleteUser)
		// 获取个人资料
		authUsers.GET("/profile", controller.GetProfile)
		// 更新个人资料
		authUsers.PUT("/profile", controller.UpdateProfile)
		// 修改密码
		authUsers.POST("/change-password", controller.ChangePassword)
	}
}
