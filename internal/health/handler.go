package health

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	db *pgxpool.Pool
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{db: db}
}

func (h *Handler) RegisterRoutes(r gin.IRouter) {
	r.GET("/health", h.Check)
}

type HealthResponse struct {
	Status    string    `json:"status"    example:"ok"`
	Database  string    `json:"database"  example:"ok"`
	Timestamp time.Time `json:"timestamp"`
}

// Check godoc
// @Summary      Health check
// @Tags         health
// @Produce      json
// @Success      200  {object}  HealthResponse
// @Failure      503  {object}  HealthResponse
// @Router       /health [get]
func (h *Handler) Check(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	dbStatus := "ok"
	if err := h.db.Ping(ctx); err != nil {
		dbStatus = "unavailable"
	}

	status := http.StatusOK
	if dbStatus != "ok" {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, HealthResponse{
		Status:    dbStatus,
		Database:  dbStatus,
		Timestamp: time.Now().UTC(),
	})
}
