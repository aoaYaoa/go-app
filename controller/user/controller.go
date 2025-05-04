package user

import (
	"net/http"
	"strconv"

	"go-app/config"
	"go-app/models/common"
	"go-app/models/user"
	"go-app/service"

	"github.com/gin-gonic/gin"
)

// Controller 用户控制器
type Controller struct {
	userService service.UserService
	cfg         *config.Config
}

// NewController 创建用户控制器
func NewController(userService service.UserService, cfg *config.Config) *Controller {
	return &Controller{
		userService: userService,
		cfg:         cfg,
	}
}

// Register 用户注册
func (c *Controller) Register(ctx *gin.Context) {
	// 从上下文获取验证后的数据
	var req user.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(400, "请求参数错误: "+err.Error()))
		return
	}

	// 调用服务层注册用户
	u, err := c.userService.Register(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(400, err.Error()))
		return
	}

	// 返回成功响应
	ctx.JSON(http.StatusCreated, common.SuccessResponse(u.ToProfileResponse()))
}

// Login 用户登录
func (c *Controller) Login(ctx *gin.Context) {
	// 从上下文获取验证后的数据
	var req user.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(400, "请求参数错误: "+err.Error()))
		return
	}

	// 调用服务层登录
	u, token, err := c.userService.Login(&req)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, common.ErrorResponse(401, err.Error()))
		return
	}

	// 返回成功响应
	response := user.TokenResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int(c.cfg.JWT.Expire.Seconds()),
	}

	ctx.JSON(http.StatusOK, common.SuccessResponse(gin.H{
		"user":  u.ToProfileResponse(),
		"token": response,
	}))
}

// GetProfile 获取当前用户资料
func (c *Controller) GetProfile(ctx *gin.Context) {
	// 获取当前用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.ErrorResponse(401, "未授权"))
		return
	}

	// 调用服务层获取用户信息
	u, err := c.userService.GetUserByID(userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusNotFound, common.ErrorResponse(404, err.Error()))
		return
	}

	// 返回成功响应
	ctx.JSON(http.StatusOK, common.SuccessResponse(u.ToResponse()))
}

// GetUsers 获取用户列表
func (c *Controller) GetUsers(ctx *gin.Context) {
	// 获取分页参数
	var params common.PaginationParams
	if err := ctx.ShouldBindQuery(&params); err != nil {
		params = *common.GetDefaultPagination()
	}

	// 获取搜索参数
	keyword := ctx.Query("keyword")
	status, _ := strconv.Atoi(ctx.Query("status"))

	// 调用服务层获取用户列表
	users, total, err := c.userService.GetUsers(params.Page, params.PageSize, keyword, status)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.ErrorResponse(500, err.Error()))
		return
	}

	// 转换为响应对象
	var userResponses []*user.Response
	for _, u := range users {
		userResponses = append(userResponses, u.ToResponse())
	}

	// 返回分页响应
	paginatedResponse := common.NewPaginatedResponse(
		total,
		params.Page,
		params.PageSize,
		userResponses,
	)

	ctx.JSON(http.StatusOK, common.SuccessResponse(paginatedResponse))
}

// GetUser 获取用户详情
func (c *Controller) GetUser(ctx *gin.Context) {
	// 获取用户ID
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(400, "无效的用户ID"))
		return
	}

	// 调用服务层获取用户
	u, err := c.userService.GetUserByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, common.ErrorResponse(404, err.Error()))
		return
	}

	// 返回成功响应
	ctx.JSON(http.StatusOK, common.SuccessResponse(u.ToResponse()))
}

// UpdateProfile 更新用户资料
func (c *Controller) UpdateProfile(ctx *gin.Context) {
	// 获取当前用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.ErrorResponse(401, "未授权"))
		return
	}

	// 获取请求数据
	var req user.UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(400, "请求参数错误: "+err.Error()))
		return
	}

	// 调用服务层更新资料
	u, err := c.userService.UpdateProfile(userID.(uint), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.ErrorResponse(500, err.Error()))
		return
	}

	// 返回成功响应
	ctx.JSON(http.StatusOK, common.SuccessResponse(u.ToProfileResponse()))
}

// ChangePassword 修改密码
func (c *Controller) ChangePassword(ctx *gin.Context) {
	// 获取当前用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.ErrorResponse(401, "未授权"))
		return
	}

	// 获取请求数据
	var req user.ChangePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(400, "请求参数错误: "+err.Error()))
		return
	}

	// 调用服务层修改密码
	err := c.userService.ChangePassword(userID.(uint), &req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(400, err.Error()))
		return
	}

	// 返回成功响应
	ctx.JSON(http.StatusOK, common.SuccessResponse(nil))
}

// DeleteUser 删除用户
func (c *Controller) DeleteUser(ctx *gin.Context) {
	// 获取用户ID
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.ErrorResponse(400, "无效的用户ID"))
		return
	}

	// 调用服务层删除用户
	if err := c.userService.DeleteUser(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, common.ErrorResponse(500, err.Error()))
		return
	}

	// 返回成功响应
	ctx.JSON(http.StatusOK, common.SuccessResponse(nil))
}
