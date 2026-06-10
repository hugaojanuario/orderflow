package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "orderflow_api_http_requests_total",
		Help: "total de requisições http por método, rota e status",
	}, []string{"method", "route", "status"})

	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "orderflow_api_http_request_duration_seconds",
		Help:    "latência das requisições http",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "route"})

	OrdersByStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "orderflow_orders_by_status",
		Help: "quantidade de pedidos por status",
	}, []string{"status"})

	EventsPublished = promauto.NewCounter(prometheus.CounterOpts{
		Name: "orderflow_api_queue_events_published_total",
		Help: "total de eventos publicados na fila",
	})
)

// Middleware registra contador e histograma de latência por rota
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unknown"
		}
		status := strconv.Itoa(c.Writer.Status())

		HTTPRequestsTotal.WithLabelValues(c.Request.Method, route, status).Inc()
		HTTPRequestDuration.WithLabelValues(c.Request.Method, route).Observe(time.Since(start).Seconds())
	}
}
