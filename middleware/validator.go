package middleware

import (
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// ValidateJSON 参数校验中间件
// 用法示例: router.POST("/users", ValidateJSON(&RegisterRequest{}), controller.Register)
func ValidateJSON(model interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 创建一个新的模型实例
		modelType := reflect.TypeOf(model)
		if modelType.Kind() == reflect.Ptr {
			modelType = modelType.Elem()
		}
		modelValue := reflect.New(modelType).Interface()

		// 绑定请求体到模型
		if err := c.ShouldBindJSON(modelValue); err != nil {
			// 处理验证错误，使用统一的错误处理工具
			ErrorWrapper(c, http.StatusBadRequest, 400, "参数验证失败", err)
			return
		}

		// 将验证后的模型存储到上下文中，以便后续处理
		c.Set("validatedData", modelValue)
		c.Next()
	}
}

// ValidateQuery 查询参数校验中间件
func ValidateQuery(model interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 创建一个新的模型实例
		modelType := reflect.TypeOf(model)
		if modelType.Kind() == reflect.Ptr {
			modelType = modelType.Elem()
		}
		modelValue := reflect.New(modelType).Interface()

		// 绑定查询参数到模型
		if err := c.ShouldBindQuery(modelValue); err != nil {
			// 处理验证错误，使用统一的错误处理工具
			ErrorWrapper(c, http.StatusBadRequest, 400, "查询参数验证失败", err)
			return
		}

		// 将验证后的模型存储到上下文中，以便后续处理
		c.Set("validatedQuery", modelValue)
		c.Next()
	}
}

// ValidateParams 路径参数校验中间件
func ValidateParams(model interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 创建一个新的模型实例
		modelType := reflect.TypeOf(model)
		if modelType.Kind() == reflect.Ptr {
			modelType = modelType.Elem()
		}
		modelValue := reflect.New(modelType).Interface()

		// 绑定路径参数到模型
		if err := c.ShouldBindUri(modelValue); err != nil {
			// 处理验证错误，使用统一的错误处理工具
			ErrorWrapper(c, http.StatusBadRequest, 400, "路径参数验证失败", err)
			return
		}

		// 将验证后的模型存储到上下文中，以便后续处理
		c.Set("validatedParams", modelValue)
		c.Next()
	}
}

// GetValidatedData 从上下文中获取验证后的数据
func GetValidatedData(c *gin.Context) interface{} {
	return c.MustGet("validatedData")
}

// GetValidatedQuery 从上下文中获取验证后的查询参数
func GetValidatedQuery(c *gin.Context) interface{} {
	return c.MustGet("validatedQuery")
}

// GetValidatedParams 从上下文中获取验证后的路径参数
func GetValidatedParams(c *gin.Context) interface{} {
	return c.MustGet("validatedParams")
}

// 自定义验证器初始化
func init() {
	// 获取验证器实例
	if _, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 注册自定义验证规则
		// 例如: v.RegisterValidation("is_valid_name", isValidName)
	}
}
