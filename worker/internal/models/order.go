package models

// status possíveis do pedido
const (
	StatusReceived  = "received"
	StatusPreparing = "preparing"
	StatusReady     = "ready"
	StatusDelivered = "delivered"
)

// OrderEvent é o evento consumido da fila do redis
type OrderEvent struct {
	OrderID int    `json:"order_id"`
	Event   string `json:"event"`
}
