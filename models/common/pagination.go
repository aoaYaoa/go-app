package common

// PaginationParams 分页参数
type PaginationParams struct {
	Page     int `form:"page" json:"page" binding:"min=1"`
	PageSize int `form:"page_size" json:"page_size" binding:"min=1,max=100"`
}

// GetOffset 获取偏移量
func (p *PaginationParams) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}

// GetLimit 获取限制条数
func (p *PaginationParams) GetLimit() int {
	return p.PageSize
}

// GetDefaultPagination 获取默认分页参数
func GetDefaultPagination() *PaginationParams {
	return &PaginationParams{
		Page:     1,
		PageSize: 10,
	}
}
