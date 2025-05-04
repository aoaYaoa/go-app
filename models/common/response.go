package common

// Response 通用响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// NewResponse 创建新的响应
func NewResponse(code int, message string, data interface{}) *Response {
	return &Response{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// SuccessResponse 成功响应
func SuccessResponse(data interface{}) *Response {
	return &Response{
		Code:    200,
		Message: "success",
		Data:    data,
	}
}

// ErrorResponse 错误响应
func ErrorResponse(code int, message string) *Response {
	return &Response{
		Code:    code,
		Message: message,
	}
}

// PaginatedResponse 分页响应结构
type PaginatedResponse struct {
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	Data     interface{} `json:"data"`
}

// NewPaginatedResponse 创建新的分页响应
func NewPaginatedResponse(total int64, page, pageSize int, data interface{}) *PaginatedResponse {
	return &PaginatedResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Data:     data,
	}
}
