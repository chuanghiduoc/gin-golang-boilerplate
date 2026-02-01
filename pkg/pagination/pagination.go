package pagination

const (
	DefaultPage     = 1
	DefaultPageSize = 10
	MaxPageSize     = 100
)

// Request represents pagination parameters from query string
type Request struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// Result represents pagination metadata for response
type Result struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
}

// Normalize applies default values if not set
func (r *Request) Normalize() {
	if r.Page <= 0 {
		r.Page = DefaultPage
	}
	if r.PageSize <= 0 {
		r.PageSize = DefaultPageSize
	}
	if r.PageSize > MaxPageSize {
		r.PageSize = MaxPageSize
	}
}

// Offset calculates the offset for database query
func (r *Request) Offset() int32 {
	r.Normalize()
	return int32((r.Page - 1) * r.PageSize)
}

// Limit returns the limit for database query
func (r *Request) Limit() int32 {
	r.Normalize()
	return int32(r.PageSize)
}

// NewResult creates a pagination result from total count and request
func NewResult(total int64, req *Request) *Result {
	req.Normalize()

	totalPages := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPages++
	}

	return &Result{
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}
}
