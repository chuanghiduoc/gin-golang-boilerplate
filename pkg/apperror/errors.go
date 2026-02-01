package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

// Error codes - Machine readable
const (
	CodeValidationError    = "VALIDATION_ERROR"
	CodeInvalidCredentials = "INVALID_CREDENTIALS"
	CodeUnauthorized       = "UNAUTHORIZED"
	CodeForbidden          = "FORBIDDEN"
	CodeNotFound           = "NOT_FOUND"
	CodeConflict           = "CONFLICT"
	CodeInternalError      = "INTERNAL_ERROR"
	CodeBadRequest         = "BAD_REQUEST"
	CodeTooManyRequests    = "TOO_MANY_REQUESTS"
)

// FieldError represents a validation error for a specific field
type FieldError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// AppError represents an application error with enterprise features
type AppError struct {
	HTTPCode   int               `json:"-"`
	ErrorCode  string            `json:"code"`
	Message    string            `json:"message"`
	Details    []*FieldError     `json:"details,omitempty"`
	Err        error             `json:"-"`
	I18nKey    string            `json:"-"` // i18n translation key
	I18nParams map[string]string `json:"-"` // i18n interpolation params
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// WithDetails adds field validation errors
func (e *AppError) WithDetails(details ...*FieldError) *AppError {
	e.Details = append(e.Details, details...)
	return e
}

// NewFieldError creates a new field validation error
func NewFieldError(field, code, message string) *FieldError {
	return &FieldError{
		Field:   field,
		Code:    code,
		Message: message,
	}
}

// New creates a new AppError
func New(httpCode int, errorCode, message string) *AppError {
	return &AppError{
		HTTPCode:  httpCode,
		ErrorCode: errorCode,
		Message:   message,
	}
}

// Wrap wraps an existing error with AppError
func Wrap(err error, httpCode int, errorCode, message string) *AppError {
	return &AppError{
		HTTPCode:  httpCode,
		ErrorCode: errorCode,
		Message:   message,
		Err:       err,
	}
}

// Predefined errors
var (
	ErrNotFound           = New(http.StatusNotFound, CodeNotFound, "resource not found")
	ErrBadRequest         = New(http.StatusBadRequest, CodeBadRequest, "bad request")
	ErrUnauthorized       = New(http.StatusUnauthorized, CodeUnauthorized, "unauthorized")
	ErrForbidden          = New(http.StatusForbidden, CodeForbidden, "forbidden")
	ErrInternalServer     = New(http.StatusInternalServerError, CodeInternalError, "internal server error")
	ErrConflict           = New(http.StatusConflict, CodeConflict, "resource already exists")
	ErrValidation         = New(http.StatusUnprocessableEntity, CodeValidationError, "validation error")
	ErrInvalidCredentials = New(http.StatusUnauthorized, CodeInvalidCredentials, "invalid credentials")
	ErrTooManyRequests    = New(http.StatusTooManyRequests, CodeTooManyRequests, "too many requests")
)

// Constructor functions
func NotFound(message string) *AppError {
	return New(http.StatusNotFound, CodeNotFound, message)
}

func BadRequest(message string) *AppError {
	return New(http.StatusBadRequest, CodeBadRequest, message)
}

func Unauthorized(message string) *AppError {
	return New(http.StatusUnauthorized, CodeUnauthorized, message)
}

func Forbidden(message string) *AppError {
	return New(http.StatusForbidden, CodeForbidden, message)
}

func InternalServer(message string) *AppError {
	return New(http.StatusInternalServerError, CodeInternalError, message)
}

func Conflict(message string) *AppError {
	return New(http.StatusConflict, CodeConflict, message)
}

func Validation(message string) *AppError {
	return New(http.StatusUnprocessableEntity, CodeValidationError, message)
}

func ValidationWithDetails(message string, details ...*FieldError) *AppError {
	return New(http.StatusUnprocessableEntity, CodeValidationError, message).WithDetails(details...)
}

func TooManyRequests(message string) *AppError {
	return New(http.StatusTooManyRequests, CodeTooManyRequests, message)
}

// I18n constructor functions - use i18n key instead of hardcoded message
// The actual translation happens in the response layer

func BadRequestI18n(i18nKey string) *AppError {
	return &AppError{
		HTTPCode:  http.StatusBadRequest,
		ErrorCode: CodeBadRequest,
		Message:   i18nKey, // Will be translated in response layer
		I18nKey:   i18nKey,
	}
}

func BadRequestI18nWithParams(i18nKey string, params map[string]string) *AppError {
	return &AppError{
		HTTPCode:   http.StatusBadRequest,
		ErrorCode:  CodeBadRequest,
		Message:    i18nKey,
		I18nKey:    i18nKey,
		I18nParams: params,
	}
}

func NotFoundI18n(i18nKey string) *AppError {
	return &AppError{
		HTTPCode:  http.StatusNotFound,
		ErrorCode: CodeNotFound,
		Message:   i18nKey,
		I18nKey:   i18nKey,
	}
}

func ConflictI18n(i18nKey string) *AppError {
	return &AppError{
		HTTPCode:  http.StatusConflict,
		ErrorCode: CodeConflict,
		Message:   i18nKey,
		I18nKey:   i18nKey,
	}
}

func ForbiddenI18n(i18nKey string) *AppError {
	return &AppError{
		HTTPCode:  http.StatusForbidden,
		ErrorCode: CodeForbidden,
		Message:   i18nKey,
		I18nKey:   i18nKey,
	}
}

func UnauthorizedI18n(i18nKey string) *AppError {
	return &AppError{
		HTTPCode:  http.StatusUnauthorized,
		ErrorCode: CodeUnauthorized,
		Message:   i18nKey,
		I18nKey:   i18nKey,
	}
}

// HasI18nKey checks if error has i18n key for translation
func (e *AppError) HasI18nKey() bool {
	return e.I18nKey != ""
}

// GetI18nKey returns i18n key
func (e *AppError) GetI18nKey() string {
	return e.I18nKey
}

// GetI18nParams returns i18n params
func (e *AppError) GetI18nParams() map[string]string {
	return e.I18nParams
}

// Type check functions
func IsNotFound(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.HTTPCode == http.StatusNotFound
	}
	return false
}

func IsConflict(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.HTTPCode == http.StatusConflict
	}
	return false
}

func IsValidation(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.HTTPCode == http.StatusUnprocessableEntity
	}
	return false
}

func IsUnauthorized(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.HTTPCode == http.StatusUnauthorized
	}
	return false
}

// GetHTTPCode returns HTTP status code, defaults to 500 for non-AppError
func GetHTTPCode(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.HTTPCode
	}
	return http.StatusInternalServerError
}

// GetErrorCode returns error code string
func GetErrorCode(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.ErrorCode
	}
	return CodeInternalError
}
