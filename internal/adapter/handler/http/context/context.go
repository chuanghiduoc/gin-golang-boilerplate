package context

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ContextKey string

const (
	RequestIDKey ContextKey = "request_id"
	UserIDKey    ContextKey = "user_id"
	UserEmailKey ContextKey = "user_email"
	UserRoleKey  ContextKey = "user_role"
	LangKey      ContextKey = "lang"
)

const RequestIDHeader = "X-Request-ID"

// GetRequestID extracts request ID from gin context
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get(string(RequestIDKey)); exists {
		return requestID.(string)
	}
	return ""
}

// SetRequestID sets request ID in gin context
func SetRequestID(c *gin.Context, requestID string) {
	c.Set(string(RequestIDKey), requestID)
	c.Header(RequestIDHeader, requestID)
}

// GenerateRequestID generates a new request ID
func GenerateRequestID() string {
	return "req_" + uuid.New().String()[:8]
}

// GetUserID extracts user ID from gin context
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	if userID, exists := c.Get(string(UserIDKey)); exists {
		return userID.(uuid.UUID), true
	}
	return uuid.Nil, false
}

// GetLang extracts language from gin context
func GetLang(c *gin.Context) string {
	if lang, exists := c.Get(string(LangKey)); exists {
		return lang.(string)
	}
	return "en"
}
