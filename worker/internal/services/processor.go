package services

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"orderflow/worker/internal/metrics"
	"orderflow/worker/internal/models"
)

// OrderRepository é o contrato que o processor precisa do repositório de pedidos
type OrderRepository interface {
	GetStatus(orderID int) (string, error)
	AdvanceStatus(orderID int, from, to string) (bool, error)
}

// nextStatus define a máquina de estados da cozinha
var nextStatus = map[string]string{
	models.StatusReceived:  models.StatusPreparing,
	models.StatusPreparing: models.StatusReady,
	models.StatusReady:     models.StatusDelivered,
}

// Processor consome eventos e avança os pedidos pelo fluxo da cozinha
type Processor struct {
	repo        OrderRepository
	events      <-chan models.OrderEvent
	concurrency int
	delays      map[string]time.Duration

	// deduplicação de pedidos em processamento neste processo
	mu       sync.Mutex
	inFlight map[int]bool
}

func NewProcessor(repo OrderRepository, events <-chan models.OrderEvent, concurrency int, acceptDelay, prepDelay, deliveryDelay time.Duration) *Processor {
	return &Processor{
		repo:        repo,
		events:      events,
		concurrency: concurrency,
		delays: map[string]time.Duration{
			models.StatusPreparing: acceptDelay,
			models.StatusReady:     prepDelay,
			models.StatusDelivered: deliveryDelay,
		},
		inFlight: map[int]bool{},
	}
}

func (p *Processor) Run(ctx context.Context) {
	slog.Info("worker iniciado",
		slog.String("component", "worker"),
		slog.Int("concurrency", p.concurrency))

	sem := make(chan struct{}, p.concurrency)
	wg := &sync.WaitGroup{}

	for {
		select {
		case <-ctx.Done():
			slog.Info("encerrando worker, aguardando jobs em andamento", slog.String("component", "worker"))
			wg.Wait()
			return
		case event, ok := <-p.events:
			if !ok {
				slog.Info("canal de eventos fechado, aguardando jobs em andamento", slog.String("component", "worker"))
				wg.Wait()
				return
			}

			// adquire o semáforo respeitando o cancelamento
			select {
			case <-ctx.Done():
				wg.Wait()
				return
			case sem <- struct{}{}:
			}

			// deduplicação: ignora evento de pedido já em processamento
			p.mu.Lock()
			if p.inFlight[event.OrderID] {
				p.mu.Unlock()
				<-sem
				slog.Info("pedido já em processamento, evento ignorado",
					slog.String("component", "worker"),
					slog.Int("order_id", event.OrderID))
				continue
			}
			p.inFlight[event.OrderID] = true
			p.mu.Unlock()

			wg.Add(1)
			go func(event models.OrderEvent) {
				defer wg.Done()
				defer func() { <-sem }()
				defer func() {
					p.mu.Lock()
					delete(p.inFlight, event.OrderID)
					p.mu.Unlock()
				}()
				p.processOrder(ctx, event.OrderID)
			}(event)
		}
	}
}

// processOrder avança o pedido até o status final de forma idempotente:
// reprocessar o mesmo evento não duplica transições, pois cada passo só
// aplica a mudança se o pedido ainda estiver no status esperado
func (p *Processor) processOrder(ctx context.Context, orderID int) {
	metrics.EventsConsumed.Inc()
	start := time.Now()

	for {
		status, err := p.repo.GetStatus(orderID)
		if err != nil {
			slog.Error("erro ao buscar status do pedido",
				slog.String("component", "worker"),
				slog.Int("order_id", orderID),
				slog.String("error", err.Error()))
			metrics.EventsFailed.Inc()
			return
		}
		if status == "" {
			slog.Warn("pedido não encontrado, evento ignorado",
				slog.String("component", "worker"),
				slog.Int("order_id", orderID))
			return
		}

		next, ok := nextStatus[status]
		if !ok {
			// pedido já está no status final
			break
		}

		select {
		case <-ctx.Done():
			slog.Info("processamento interrompido, pedido será retomado na próxima inicialização",
				slog.String("component", "worker"),
				slog.Int("order_id", orderID),
				slog.String("status", status))
			return
		case <-time.After(p.delays[next]):
		}

		advanced, err := p.repo.AdvanceStatus(orderID, status, next)
		if err != nil {
			slog.Error("erro ao avançar status do pedido",
				slog.String("component", "worker"),
				slog.Int("order_id", orderID),
				slog.String("error", err.Error()))
			metrics.EventsFailed.Inc()
			return
		}
		if !advanced {
			// outro processo já avançou o pedido; recarrega o status e segue
			continue
		}

		slog.Info("status do pedido atualizado",
			slog.String("component", "worker"),
			slog.Int("order_id", orderID),
			slog.String("status", next))
	}

	metrics.ProcessingDuration.Observe(time.Since(start).Seconds())
}
