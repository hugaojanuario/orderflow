package handler

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type HealthController struct {
	db    *sql.DB
	redis *redis.Client
}

func NewHealthController(db *sql.DB, redisClient *redis.Client) *HealthController {
	return &HealthController{db: db, redis: redisClient}
}

// Healthz é o liveness: responde 200 se o processo está vivo
func (h *HealthController) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Readyz é o readiness: responde 200 só se postgres e redis estão acessíveis
func (h *HealthController) Readyz(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": "postgres indisponível"})
		return
	}

	if err := h.redis.Ping(ctx).Err(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": "redis indisponível"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
