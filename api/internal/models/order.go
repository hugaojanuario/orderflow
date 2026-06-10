package models

import "time"

// status possíveis do pedido
const (
	StatusReceived  = "received"
	StatusPreparing = "preparing"
	StatusReady     = "ready"
	StatusDelivered = "delivered"
)

type Order struct {
	ID           int                  `json:"id"`
	CustomerName string               `json:"customer_name"`
	Total        float64              `json:"total"`
	Status       string               `json:"status"`
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`
	Items        []OrderItem          `json:"items,omitempty"`
	History      []OrderStatusHistory `json:"history,omitempty"`
}

type OrderItem struct {
	ID         int     `json:"id"`
	OrderID    int     `json:"order_id"`
	MenuItemID int     `json:"menu_item_id"`
	Name       string  `json:"name"`
	Quantity   int     `json:"quantity"`
	Price      float64 `json:"price"`
}

type OrderStatusHistory struct {
	ID        int       `json:"id"`
	OrderID   int       `json:"order_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// OrderEvent é o evento publicado na fila do redis quando um pedido é criado
type OrderEvent struct {
	OrderID int    `json:"order_id"`
	Event   string `json:"event"`
}

// DTOs

type CreateOrderItemRequest struct {
	MenuItemID int `json:"menu_item_id" binding:"required"`
	Quantity   int `json:"quantity" binding:"required,min=1"`
}

type CreateOrderRequest struct {
	CustomerName string                   `json:"customer_name" binding:"required"`
	Items        []CreateOrderItemRequest `json:"items" binding:"required,min=1,dive"`
}

type ListOrdersResponse struct {
	Orders []Order `json:"orders"`
	Page   int     `json:"page"`
	Limit  int     `json:"limit"`
	Total  int     `json:"total"`
}

type StatsResponse struct {
	Date           string         `json:"date"`
	TotalOrders    int            `json:"total_orders"`
	OrdersByStatus map[string]int `json:"orders_by_status"`
	Revenue        float64        `json:"revenue"`
	AvgPrepSeconds float64        `json:"avg_prep_seconds"`
}

type VersionResponse struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
}
