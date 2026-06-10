package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"

	"orderflow/api/internal/metrics"
	"orderflow/api/internal/models"
)

// OrdersQueueKey é a lista do redis usada como fila de eventos de pedidos
const OrdersQueueKey = "orderflow:orders:events"

type Publisher struct {
	client *redis.Client
}

func NewPublisher(client *redis.Client) *Publisher {
	return &Publisher{client: client}
}

func (p *Publisher) PublishOrderCreated(ctx context.Context, orderID int) error {
	event := models.OrderEvent{OrderID: orderID, Event: "order_created"}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("erro ao serializar evento: %w", err)
	}

	if err := p.client.LPush(ctx, OrdersQueueKey, payload).Err(); err != nil {
		return fmt.Errorf("erro ao publicar evento na fila: %w", err)
	}

	metrics.EventsPublished.Inc()

	return nil
}
