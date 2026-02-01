package middleware

import (
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"

	httpctx "backend-gin/internal/adapter/handler/http/context"
	"backend-gin/internal/adapter/handler/http/response"
	"backend-gin/internal/infrastructure/logger"
	"backend-gin/pkg/apperror"
)

func Recovery(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID := httpctx.GetRequestID(c)

				log.WithFields(map[string]any{
					"request_id": requestID,
					"error":      err,
					"stack":      string(debug.Stack()),
				}).Error("panic recovered")

				c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
					Success: false,
					Error: &response.Error{
						Code:    apperror.CodeInternalError,
						Message: "internal server error",
					},
					Meta: response.Meta{
						RequestID: requestID,
						Timestamp: time.Now().UTC().Format(time.RFC3339),
					},
				})
			}
		}()
		c.Next()
	}
}
