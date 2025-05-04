package controller

import (
	"go-app/config"
	"go-app/controller/user"
	"go-app/database/repositories"
	"go-app/service"
)

// Manager 控制器管理器
type Manager struct {
	User    *user.Controller
	
}

// NewManager 初始化所有控制器
func NewManager(cfg *config.Config, repoManager *repositories.RepositoryManager) *Manager {
	// 初始化用户服务
	userService := service.NewUserService(repoManager.User, cfg)

	return &Manager{
		User:    user.NewController(userService, cfg),
	}
}
