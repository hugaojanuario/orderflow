package metrics

import (
	"context"
	"log/slog"
	"time"

	"orderflow/api/internal/models"
)

// OrderCounter é o que o poller precisa do repository de pedidos
type OrderCounter interface {
	CountByStatus() (map[string]int, error)
}

// StatusPoller atualiza periodicamente o gauge de pedidos por status
type StatusPoller struct {
	repo     OrderCounter
	interval time.Duration
}

func NewStatusPoller(repo OrderCounter, interval time.Duration) *StatusPoller {
	return &StatusPoller{repo: repo, interval: interval}
}

func (p *StatusPoller) Run(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	p.collect()

	for {
		select {
		case <-ctx.Done():
			slog.Info("encerrando poller de métricas", slog.String("component", "metrics"))
			return
		case <-ticker.C:
			p.collect()
		}
	}
}

func (p *StatusPoller) collect() {
	counts, err := p.repo.CountByStatus()
	if err != nil {
		slog.Error("erro ao coletar pedidos por status", slog.String("component", "metrics"), slog.String("error", err.Error()))
		return
	}

	// zera os status conhecidos para não deixar valor antigo no gauge
	statuses := []string{models.StatusReceived, models.StatusPreparing, models.StatusReady, models.StatusDelivered}
	for _, status := range statuses {
		OrdersByStatus.WithLabelValues(status).Set(float64(counts[status]))
	}
}
