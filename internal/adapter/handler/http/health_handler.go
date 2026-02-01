package http

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend-gin/internal/adapter/handler/http/response"
	"backend-gin/internal/infrastructure/cache"
)

type HealthHandler struct {
	pool        *pgxpool.Pool
	redisClient *cache.RedisClient
}

func NewHealthHandler(pool *pgxpool.Pool, redisClient *cache.RedisClient) *HealthHandler {
	return &HealthHandler{
		pool:        pool,
		redisClient: redisClient,
	}
}

type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
	Redis    string `json:"redis,omitempty"`
}

// Health godoc
// @Summary Health check
// @Description Check the health status of the API, database, and Redis
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=HealthResponse}
// @Failure 503 {object} response.Response
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	ctx := c.Request.Context()

	// Check database
	dbStatus := "healthy"
	if err := h.pool.Ping(ctx); err != nil {
		dbStatus = "unhealthy"
	}

	// Determine overall status
	status := "healthy"
	resp := HealthResponse{
		Database: dbStatus,
	}

	// Check Redis if configured
	if h.redisClient != nil {
		if err := h.redisClient.Ping(ctx); err != nil {
			resp.Redis = "unhealthy"
			// Redis failure = degraded, not unhealthy (Redis is optional)
			if status == "healthy" {
				status = "degraded"
			}
		} else {
			resp.Redis = "healthy"
		}
	}

	// Database failure = unhealthy (critical)
	if dbStatus != "healthy" {
		status = "unhealthy"
	}

	resp.Status = status

	if status == "unhealthy" {
		c.JSON(503, response.Response{
			Success: false,
			Data:    resp,
		})
		return
	}

	response.Success(c, resp)
}
