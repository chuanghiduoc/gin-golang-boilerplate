package middleware

import (
	"github.com/gin-gonic/gin"

	httpctx "backend-gin/internal/adapter/handler/http/context"
)

// RequestID middleware generates or extracts request ID for tracing
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(httpctx.RequestIDHeader)
		if requestID == "" {
			requestID = httpctx.GenerateRequestID()
		}

		httpctx.SetRequestID(c, requestID)

		c.Next()
	}
}
