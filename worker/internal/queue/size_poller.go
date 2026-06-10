package queue

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"orderflow/worker/internal/metrics"
)

// SizePoller atualiza periodicamente o gauge de tamanho da fila
type SizePoller struct {
	client   *redis.Client
	interval time.Duration
}

func NewSizePoller(client *redis.Client, interval time.Duration) *SizePoller {
	return &SizePoller{client: client, interval: interval}
}

func (p *SizePoller) Run(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("encerrando poller da fila", slog.String("component", "queue"))
			return
		case <-ticker.C:
			size, err := p.client.LLen(ctx, OrdersQueueKey).Result()
			if err != nil {
				if ctx.Err() == nil {
					slog.Error("erro ao medir o tamanho da fila", slog.String("component", "queue"), slog.String("error", err.Error()))
				}
				continue
			}
			metrics.QueueSize.Set(float64(size))
		}
	}
}
