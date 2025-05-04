package service

import (
	"errors"
	"fmt"
	"time"

	"go-app/config"
	"go-app/database/repositories"
	"go-app/middleware"
	"go-app/models/user"
)

// UserService 用户服务接口
type UserService interface {
	Register(req *user.RegisterRequest) (*user.User, error)
	Login(req *user.LoginRequest) (*user.User, string, error)
	GetUserByID(id uint) (*user.User, error)
	GetUsers(page, pageSize int, keyword string, status int) ([]user.User, int64, error)
	UpdateProfile(id uint, req *user.UpdateProfileRequest) (*user.User, error)
	ChangePassword(id uint, req *user.ChangePasswordRequest) error
	DeleteUser(id uint) error
}

// UserServiceImpl 用户服务实现
type UserServiceImpl struct {
	userRepo repositories.UserRepository
	cfg      *config.Config
}

// NewUserService 创建用户服务
func NewUserService(userRepo repositories.UserRepository, cfg *config.Config) UserService {
	return &UserServiceImpl{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

// Register 用户注册
func (s *UserServiceImpl) Register(req *user.RegisterRequest) (*user.User, error) {
	// 检查用户名是否存在
	if _, err := s.userRepo.FindByUsername(req.Username); err == nil {
		return nil, errors.New("用户名已被使用")
	}

	// 检查邮箱是否存在
	if _, err := s.userRepo.FindByEmail(req.Email); err == nil {
		return nil, errors.New("邮箱已被使用")
	}

	// 创建新用户
	hashedPassword, err := middleware.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("密码加密失败: " + err.Error())
	}

	newUser := &user.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  hashedPassword,
		Nickname:  req.Nickname,
		Status:    1, // 正常状态
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.userRepo.Create(newUser); err != nil {
		return nil, errors.New("创建用户失败: " + err.Error())
	}

	return newUser, nil
}

// Login 用户登录
func (s *UserServiceImpl) Login(req *user.LoginRequest) (*user.User, string, error) {
	// 调试信息
	fmt.Printf("尝试登录用户: %s\n", req.Username)

	// 根据用户名查找用户
	u, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		fmt.Printf("用户查找失败: %v\n", err)
		return nil, "", errors.New("用户名或密码错误")
	}

	// 输出调试信息
	fmt.Printf("找到用户: %s, ID: %d, 状态: %d\n", u.Username, u.ID, u.Status)
	fmt.Printf("数据库密码: %s\n", u.Password)

	// 检查用户状态
	if u.Status != 1 {
		fmt.Printf("用户状态异常: %d\n", u.Status)
		return nil, "", errors.New("用户已被禁用")
	}

	// 验证密码 - 先检查密码哈希，如果失败则检查明文密码
	passwordMatch := middleware.CheckPasswordHash(req.Password, u.Password)

	// 如果哈希验证失败，尝试直接比较明文密码（临时解决方案）
	if !passwordMatch && u.Password == req.Password {
		passwordMatch = true
		fmt.Println("警告：使用明文密码匹配成功，应更新为哈希密码")
	}

	fmt.Printf("密码匹配结果: %v\n", passwordMatch)

	if !passwordMatch {
		return nil, "", errors.New("用户名或密码错误")
	}

	// 生成JWT令牌
	token, err := middleware.GenerateToken(u.ID, s.cfg.JWT.Secret, s.cfg.JWT.Expire)
	if err != nil {
		return nil, "", errors.New("生成令牌失败: " + err.Error())
	}

	return u, token, nil
}

// GetUserByID 根据ID获取用户
func (s *UserServiceImpl) GetUserByID(id uint) (*user.User, error) {
	u, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	return u, nil
}

// GetUsers 获取用户列表
func (s *UserServiceImpl) GetUsers(page, pageSize int, keyword string, status int) ([]user.User, int64, error) {
	// 设置默认值
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 创建过滤条件
	filter := map[string]interface{}{}
	if status != 0 {
		filter["status"] = status
	}
	if keyword != "" {
		filter["keyword"] = keyword
	}

	// 获取用户列表
	return s.userRepo.FindAll(page, pageSize, filter)
}

// UpdateProfile 更新用户资料
func (s *UserServiceImpl) UpdateProfile(id uint, req *user.UpdateProfileRequest) (*user.User, error) {
	// 获取用户
	u, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	// 更新字段
	if req.Nickname != "" {
		u.Nickname = req.Nickname
	}
	if req.Avatar != "" {
		u.Avatar = req.Avatar
	}
	u.UpdatedAt = time.Now()

	// 更新用户
	if err := s.userRepo.Update(u); err != nil {
		return nil, errors.New("更新用户资料失败: " + err.Error())
	}

	return u, nil
}

// ChangePassword 修改密码
func (s *UserServiceImpl) ChangePassword(id uint, req *user.ChangePasswordRequest) error {
	// 获取用户
	u, err := s.userRepo.FindByID(id)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 验证旧密码
	if !middleware.CheckPasswordHash(req.OldPassword, u.Password) {
		return errors.New("原密码错误")
	}

	// 更新密码
	hashedPassword, err := middleware.HashPassword(req.NewPassword)
	if err != nil {
		return errors.New("密码加密失败: " + err.Error())
	}

	u.Password = hashedPassword
	u.UpdatedAt = time.Now()

	// 更新用户
	if err := s.userRepo.Update(u); err != nil {
		return errors.New("更新密码失败: " + err.Error())
	}

	return nil
}

// DeleteUser 删除用户
func (s *UserServiceImpl) DeleteUser(id uint) error {
	if err := s.userRepo.Delete(id); err != nil {
		return errors.New("删除用户失败: " + err.Error())
	}
	return nil
}
