package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	httpctx "backend-gin/internal/adapter/handler/http/context"
	"backend-gin/internal/adapter/handler/http/response"
	"backend-gin/internal/usecase/auth"
)

// Keep for backward compatibility
type ContextKey = httpctx.ContextKey

const (
	UserIDKey    = httpctx.UserIDKey
	UserEmailKey = httpctx.UserEmailKey
	UserRoleKey  = httpctx.UserRoleKey
	LangKey      = httpctx.LangKey
)

func Auth(authService auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "authorization header required")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Unauthorized(c, "invalid authorization header format")
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := authService.ValidateToken(token)
		if err != nil {
			response.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}

		c.Set(string(UserIDKey), claims.UserID)
		c.Set(string(UserEmailKey), claims.Email)
		c.Set(string(UserRoleKey), claims.Role)

		c.Next()
	}
}
