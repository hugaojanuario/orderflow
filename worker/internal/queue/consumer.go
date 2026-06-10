package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"orderflow/worker/internal/metrics"
	"orderflow/worker/internal/models"
)

// OrdersQueueKey é a lista do redis usada como fila de eventos de pedidos
const OrdersQueueKey = "orderflow:orders:events"

// Consumer lê eventos da fila do redis e os expõe num channel somente leitura
type Consumer struct {
	client *redis.Client
	events chan models.OrderEvent
}

func NewConsumer(client *redis.Client) *Consumer {
	return &Consumer{
		client: client,
		events: make(chan models.OrderEvent, 100),
	}
}

func (c *Consumer) Events() <-chan models.OrderEvent {
	return c.events
}

// Publish coloca um evento de volta na fila (usado na recuperação de pendentes)
func (c *Consumer) Publish(ctx context.Context, orderID int) error {
	event := models.OrderEvent{OrderID: orderID, Event: "order_created"}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("erro ao serializar evento: %w", err)
	}

	if err := c.client.LPush(ctx, OrdersQueueKey, payload).Err(); err != nil {
		return fmt.Errorf("erro ao publicar evento na fila: %w", err)
	}

	return nil
}

func (c *Consumer) Run(ctx context.Context) {
	slog.Info("consumidor da fila iniciado", slog.String("component", "queue"))

	for {
		select {
		case <-ctx.Done():
			slog.Info("encerrando consumidor da fila", slog.String("component", "queue"))
			close(c.events)
			return
		default:
		}

		result, err := c.client.BRPop(ctx, 5*time.Second, OrdersQueueKey).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) || ctx.Err() != nil {
				continue
			}
			slog.Error("erro ao consumir a fila", slog.String("component", "queue"), slog.String("error", err.Error()))
			select {
			case <-ctx.Done():
			case <-time.After(time.Second):
			}
			continue
		}

		event := models.OrderEvent{}
		if err := json.Unmarshal([]byte(result[1]), &event); err != nil {
			slog.Error("erro ao decodificar evento, descartando", slog.String("component", "queue"), slog.String("error", err.Error()))
			metrics.EventsFailed.Inc()
			continue
		}

		select {
		case c.events <- event:
		case <-ctx.Done():
		}
	}
}
