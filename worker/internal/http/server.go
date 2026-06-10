package http

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

// NewServer monta o servidor de health checks e métricas do worker
func NewServer(port string, db *sql.DB, redisClient *redis.Client) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, `{"status":"ok"}`)
	})

	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, `{"status":"unavailable","error":"postgres indisponível"}`)
			return
		}

		if err := redisClient.Ping(ctx).Err(); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, `{"status":"unavailable","error":"redis indisponível"}`)
			return
		}

		writeJSON(w, http.StatusOK, `{"status":"ok"}`)
	})

	mux.Handle("/metrics", promhttp.Handler())

	return &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
}

func writeJSON(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}
