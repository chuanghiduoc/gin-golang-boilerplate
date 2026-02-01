package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"backend-gin/internal/adapter/handler/http/response"
)

type RateLimitConfig struct {
	Max        int
	WindowSec  int
	KeyPrefix  string
	Message    string
	StatusCode int
}

func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Max:        100,
		WindowSec:  60,
		KeyPrefix:  "ratelimit:",
		Message:    "too many requests",
		StatusCode: http.StatusTooManyRequests,
	}
}

func RateLimit(redisClient *redis.Client) gin.HandlerFunc {
	return RateLimitWithConfig(redisClient, DefaultRateLimitConfig())
}

func RateLimitWithConfig(redisClient *redis.Client, cfg RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if redisClient == nil {
			c.Next()
			return
		}

		key := cfg.KeyPrefix + getClientIP(c)
		ctx := context.Background()

		count, err := redisClient.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}

		if count == 1 {
			redisClient.Expire(ctx, key, time.Duration(cfg.WindowSec)*time.Second)
		}

		ttl, _ := redisClient.TTL(ctx, key).Result()

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.Max))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, cfg.Max-int(count))))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", int(ttl.Seconds())))

		if int(count) > cfg.Max {
			c.Header("Retry-After", fmt.Sprintf("%d", int(ttl.Seconds())))
			response.TooManyRequests(c, cfg.Message)
			c.Abort()
			return
		}

		c.Next()
	}
}

func RateLimitByUserID(redisClient *redis.Client, cfg RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if redisClient == nil {
			c.Next()
			return
		}

		userID, exists := c.Get(string(UserIDKey))
		if !exists {
			c.Next()
			return
		}

		key := cfg.KeyPrefix + "user:" + fmt.Sprintf("%v", userID)
		ctx := context.Background()

		count, err := redisClient.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}

		if count == 1 {
			redisClient.Expire(ctx, key, time.Duration(cfg.WindowSec)*time.Second)
		}

		ttl, _ := redisClient.TTL(ctx, key).Result()

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.Max))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, cfg.Max-int(count))))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", int(ttl.Seconds())))

		if int(count) > cfg.Max {
			c.Header("Retry-After", fmt.Sprintf("%d", int(ttl.Seconds())))
			response.TooManyRequests(c, cfg.Message)
			c.Abort()
			return
		}

		c.Next()
	}
}

func getClientIP(c *gin.Context) string {
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}
	return c.ClientIP()
}
