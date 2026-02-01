package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"backend-gin/internal/infrastructure/logger"
)

func Logger(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		fields := map[string]any{
			"request_id": requestID,
			"status":     status,
			"method":     c.Request.Method,
			"path":       path,
			"query":      query,
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
			"latency_ms": latency.Milliseconds(),
		}

		if len(c.Errors) > 0 {
			fields["errors"] = c.Errors.String()
		}

		switch {
		case status >= 500:
			log.WithFields(fields).Error("server error")
		case status >= 400:
			log.WithFields(fields).Warn("client error")
		default:
			log.WithFields(fields).Info("request completed")
		}
	}
}
