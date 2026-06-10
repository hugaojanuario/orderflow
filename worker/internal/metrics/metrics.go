package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	EventsConsumed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "orderflow_worker_events_consumed_total",
		Help: "total de eventos consumidos da fila",
	})

	EventsFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "orderflow_worker_events_failed_total",
		Help: "total de eventos com erro de processamento",
	})

	ProcessingDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "orderflow_worker_processing_duration_seconds",
		Help:    "duração do processamento completo de um pedido",
		Buckets: []float64{1, 2.5, 5, 10, 15, 30, 60, 120},
	})

	QueueSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "orderflow_worker_queue_size",
		Help: "tamanho atual da fila de eventos no redis",
	})
)
