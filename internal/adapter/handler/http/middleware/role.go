package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"backend-gin/internal/adapter/handler/http/response"
	"backend-gin/internal/domain/entity"
	"backend-gin/internal/domain/repository"
)

func RequireRole(userRepo repository.UserRepository, roles ...entity.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get(string(UserIDKey))
		if !exists {
			response.Unauthorized(c, "user not authenticated")
			c.Abort()
			return
		}

		id, ok := userID.(uuid.UUID)
		if !ok {
			response.InternalError(c)
			c.Abort()
			return
		}

		user, err := userRepo.GetByID(c.Request.Context(), id)
		if err != nil {
			response.Unauthorized(c, "user not found")
			c.Abort()
			return
		}

		hasRole := false
		for _, role := range roles {
			if user.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			response.Forbidden(c, "insufficient permissions")
			c.Abort()
			return
		}

		c.Set(string(UserRoleKey), user.Role)
		c.Next()
	}
}

func RequireAdmin(userRepo repository.UserRepository) gin.HandlerFunc {
	return RequireRole(userRepo, entity.RoleAdmin)
}

func GetUserRole(c *gin.Context) entity.Role {
	if role, exists := c.Get(string(UserRoleKey)); exists {
		if r, ok := role.(entity.Role); ok {
			return r
		}
	}
	return entity.RoleUser
}
