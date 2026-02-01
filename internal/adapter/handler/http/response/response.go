package response

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	httpctx "backend-gin/internal/adapter/handler/http/context"
	"backend-gin/internal/infrastructure/i18n"
	"backend-gin/pkg/apperror"
	"backend-gin/pkg/pagination"
)

// Meta contains request metadata for tracing
type Meta struct {
	RequestID string `json:"request_id"`
	Timestamp string `json:"timestamp"`
}

// Response represents the standard API response
type Response struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   *Error `json:"error,omitempty"`
	Meta    Meta   `json:"meta"`
}

// Error represents the error structure
type Error struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details []*FieldError  `json:"details,omitempty"`
}

// FieldError represents a validation error for a specific field
type FieldError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Pagination contains pagination metadata
type Pagination struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Success    bool       `json:"success"`
	Data       any        `json:"data"`
	Pagination Pagination `json:"pagination"`
	Meta       Meta       `json:"meta"`
}

// buildMeta creates metadata from gin context
func buildMeta(c *gin.Context) Meta {
	return Meta{
		RequestID: httpctx.GetRequestID(c),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// getTranslator returns the translator from context or default
func getTranslator(c *gin.Context) *i18n.Translator {
	if t, exists := c.Get("translator"); exists {
		if translator, ok := t.(*i18n.Translator); ok {
			return translator
		}
	}
	return i18n.New(i18n.DefaultLanguage)
}

// T translates a key using the context's language
func T(c *gin.Context, key string, params ...map[string]string) string {
	return getTranslator(c).T(key, params...)
}

// Success sends a successful response with status 200
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
		Meta:    buildMeta(c),
	})
}

// Created sends a successful response with status 201
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    data,
		Meta:    buildMeta(c),
	})
}

// NoContent sends a successful response with status 204
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Paginated sends a paginated response
func Paginated(c *gin.Context, data any, total int64, page, pageSize, totalPages int) {
	c.JSON(http.StatusOK, PaginatedResponse{
		Success: true,
		Data:    data,
		Pagination: Pagination{
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
			HasNext:    page < totalPages,
			HasPrev:    page > 1,
		},
		Meta: buildMeta(c),
	})
}

// PaginatedWithMeta sends a paginated response using pagination.Result
func PaginatedWithMeta(c *gin.Context, data any, meta *pagination.Result) {
	c.JSON(http.StatusOK, PaginatedResponse{
		Success: true,
		Data:    data,
		Pagination: Pagination{
			Total:      meta.Total,
			Page:       meta.Page,
			PageSize:   meta.PageSize,
			TotalPages: meta.TotalPages,
			HasNext:    meta.Page < meta.TotalPages,
			HasPrev:    meta.Page > 1,
		},
		Meta: buildMeta(c),
	})
}

// Err sends an error response based on the error type
func Err(c *gin.Context, err error) {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		var details []*FieldError
		for _, d := range appErr.Details {
			details = append(details, &FieldError{
				Field:   d.Field,
				Code:    d.Code,
				Message: d.Message,
			})
		}

		// Translate message if it has i18n key
		message := appErr.Message
		if appErr.HasI18nKey() {
			t := getTranslator(c)
			if appErr.GetI18nParams() != nil {
				message = t.T(appErr.GetI18nKey(), appErr.GetI18nParams())
			} else {
				message = t.T(appErr.GetI18nKey())
			}
		}

		c.JSON(appErr.HTTPCode, Response{
			Success: false,
			Error: &Error{
				Code:    appErr.ErrorCode,
				Message: message,
				Details: details,
			},
			Meta: buildMeta(c),
		})
		return
	}

	t := getTranslator(c)
	c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Error: &Error{
			Code:    apperror.CodeInternalError,
			Message: t.T("common.internal_error"),
		},
		Meta: buildMeta(c),
	})
}

// ValidationErr sends a validation error response with field details
func ValidationErr(c *gin.Context, err error) {
	t := getTranslator(c)
	var details []*FieldError

	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		for _, fe := range ve {
			details = append(details, &FieldError{
				Field:   toSnakeCase(fe.Field()),
				Code:    mapValidationTag(fe.Tag()),
				Message: buildValidationMessageI18n(t, fe),
			})
		}
	}

	c.JSON(http.StatusUnprocessableEntity, Response{
		Success: false,
		Error: &Error{
			Code:    apperror.CodeValidationError,
			Message: t.T("common.validation_error"),
			Details: details,
		},
		Meta: buildMeta(c),
	})
}

// BadRequest sends a bad request error response
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Error: &Error{
			Code:    apperror.CodeBadRequest,
			Message: message,
		},
		Meta: buildMeta(c),
	})
}

// Unauthorized sends an unauthorized error response
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Success: false,
		Error: &Error{
			Code:    apperror.CodeUnauthorized,
			Message: message,
		},
		Meta: buildMeta(c),
	})
}

// Forbidden sends a forbidden error response
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{
		Success: false,
		Error: &Error{
			Code:    apperror.CodeForbidden,
			Message: message,
		},
		Meta: buildMeta(c),
	})
}

// NotFound sends a not found error response
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Response{
		Success: false,
		Error: &Error{
			Code:    apperror.CodeNotFound,
			Message: message,
		},
		Meta: buildMeta(c),
	})
}

// Conflict sends a conflict error response
func Conflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, Response{
		Success: false,
		Error: &Error{
			Code:    apperror.CodeConflict,
			Message: message,
		},
		Meta: buildMeta(c),
	})
}

// TooManyRequests sends a rate limit error response
func TooManyRequests(c *gin.Context, message string) {
	c.JSON(http.StatusTooManyRequests, Response{
		Success: false,
		Error: &Error{
			Code:    apperror.CodeTooManyRequests,
			Message: message,
		},
		Meta: buildMeta(c),
	})
}

// InternalError sends an internal server error response
func InternalError(c *gin.Context) {
	t := getTranslator(c)
	c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Error: &Error{
			Code:    apperror.CodeInternalError,
			Message: t.T("common.internal_error"),
		},
		Meta: buildMeta(c),
	})
}

// I18n response functions - use i18n key instead of raw message

// UnauthorizedI18n sends an unauthorized error with i18n key
func UnauthorizedI18n(c *gin.Context, key string, params ...map[string]string) {
	t := getTranslator(c)
	c.JSON(http.StatusUnauthorized, Response{
		Success: false,
		Error: &Error{
			Code:    apperror.CodeUnauthorized,
			Message: t.T(key, params...),
		},
		Meta: buildMeta(c),
	})
}

// ForbiddenI18n sends a forbidden error with i18n key
func ForbiddenI18n(c *gin.Context, key string, params ...map[string]string) {
	t := getTranslator(c)
	c.JSON(http.StatusForbidden, Response{
		Success: false,
		Error: &Error{
			Code:    apperror.CodeForbidden,
			Message: t.T(key, params...),
		},
		Meta: buildMeta(c),
	})
}

// NotFoundI18n sends a not found error with i18n key
func NotFoundI18n(c *gin.Context, key string, params ...map[string]string) {
	t := getTranslator(c)
	c.JSON(http.StatusNotFound, Response{
		Success: false,
		Error: &Error{
			Code:    apperror.CodeNotFound,
			Message: t.T(key, params...),
		},
		Meta: buildMeta(c),
	})
}

// BadRequestI18n sends a bad request error with i18n key
func BadRequestI18n(c *gin.Context, key string, params ...map[string]string) {
	t := getTranslator(c)
	c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Error: &Error{
			Code:    apperror.CodeBadRequest,
			Message: t.T(key, params...),
		},
		Meta: buildMeta(c),
	})
}

// ConflictI18n sends a conflict error with i18n key
func ConflictI18n(c *gin.Context, key string, params ...map[string]string) {
	t := getTranslator(c)
	c.JSON(http.StatusConflict, Response{
		Success: false,
		Error: &Error{
			Code:    apperror.CodeConflict,
			Message: t.T(key, params...),
		},
		Meta: buildMeta(c),
	})
}

// Helper functions

// toSnakeCase converts PascalCase/camelCase to snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// mapValidationTag maps validator tags to error codes
func mapValidationTag(tag string) string {
	tagMap := map[string]string{
		"required":  "REQUIRED",
		"email":     "INVALID_EMAIL",
		"min":       "TOO_SHORT",
		"max":       "TOO_LONG",
		"len":       "INVALID_LENGTH",
		"uuid":      "INVALID_UUID",
		"url":       "INVALID_URL",
		"oneof":     "INVALID_VALUE",
		"gte":       "TOO_SMALL",
		"lte":       "TOO_LARGE",
		"gt":        "TOO_SMALL",
		"lt":        "TOO_LARGE",
		"alphanum":  "INVALID_FORMAT",
		"alpha":     "INVALID_FORMAT",
		"numeric":   "INVALID_FORMAT",
		"lowercase": "MUST_BE_LOWERCASE",
		"uppercase": "MUST_BE_UPPERCASE",
	}

	if code, ok := tagMap[tag]; ok {
		return code
	}
	return "INVALID"
}

// buildValidationMessage creates user-friendly validation messages (deprecated: use buildValidationMessageI18n)
func buildValidationMessage(fe validator.FieldError) string {
	field := toSnakeCase(fe.Field())

	switch fe.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email address"
	case "min":
		return field + " must be at least " + fe.Param() + " characters"
	case "max":
		return field + " must be at most " + fe.Param() + " characters"
	case "len":
		return field + " must be exactly " + fe.Param() + " characters"
	case "uuid":
		return field + " must be a valid UUID"
	case "url":
		return field + " must be a valid URL"
	case "oneof":
		return field + " must be one of: " + fe.Param()
	case "gte":
		return field + " must be greater than or equal to " + fe.Param()
	case "lte":
		return field + " must be less than or equal to " + fe.Param()
	case "gt":
		return field + " must be greater than " + fe.Param()
	case "lt":
		return field + " must be less than " + fe.Param()
	default:
		return field + " is invalid"
	}
}

// buildValidationMessageI18n creates i18n validation messages
func buildValidationMessageI18n(t *i18n.Translator, fe validator.FieldError) string {
	field := toSnakeCase(fe.Field())
	params := map[string]string{"field": field}

	switch fe.Tag() {
	case "required":
		return t.T("validation.required", params)
	case "email":
		return t.T("validation.email", params)
	case "min":
		params["min"] = fe.Param()
		return t.T("validation.min_length", params)
	case "max":
		params["max"] = fe.Param()
		return t.T("validation.max_length", params)
	case "uuid":
		return t.T("validation.uuid", params)
	default:
		return field + " is invalid"
	}
}
