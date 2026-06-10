package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Logging emite um log estruturado por requisição
func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		slog.Info("requisição processada",
			slog.String("component", "http"),
			slog.String("request_id", c.GetString("requestId")),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Float64("latency_ms", float64(time.Since(start).Microseconds())/1000))
	}
}
