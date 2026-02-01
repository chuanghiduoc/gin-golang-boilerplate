package http

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend-gin/internal/adapter/handler/http/response"
)

type HealthHandler struct {
	pool *pgxpool.Pool
}

func NewHealthHandler(pool *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{
		pool: pool,
	}
}

type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

// Health godoc
// @Summary Health check
// @Description Check the health status of the API and database
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=HealthResponse}
// @Failure 503 {object} response.Response
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	dbStatus := "healthy"
	if err := h.pool.Ping(c.Request.Context()); err != nil {
		dbStatus = "unhealthy"
	}

	status := "healthy"
	if dbStatus != "healthy" {
		status = "unhealthy"
	}

	resp := HealthResponse{
		Status:   status,
		Database: dbStatus,
	}

	if status == "unhealthy" {
		c.JSON(503, response.Response{
			Success: false,
			Data:    resp,
		})
		return
	}

	response.Success(c, resp)
}
