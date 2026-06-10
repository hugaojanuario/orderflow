package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"orderflow/api/internal/models"
)

var ErrOrderNotFound = errors.New("order not found")
var ErrMenuItemNotFound = errors.New("menu item not found")

// cache da primeira página de listagem
const listCacheTTL = 10 * time.Second

// OrderRepository é o contrato que o service precisa do repositório de pedidos
type OrderRepository interface {
	Create(order *models.Order) (*models.Order, error)
	List(page, limit int, status string) ([]models.Order, int, error)
	GetByID(id int) (*models.Order, error)
}

// MenuRepository é o contrato que o service precisa do repositório do cardápio
type MenuRepository interface {
	List() ([]models.MenuItem, error)
	GetByIDs(ids []int) ([]models.MenuItem, error)
}

// EventPublisher publica eventos de pedidos na fila
type EventPublisher interface {
	PublishOrderCreated(ctx context.Context, orderID int) error
}

// Cache é o contrato do cache de leitura
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
}

type OrderService struct {
	repo      OrderRepository
	menuRepo  MenuRepository
	publisher EventPublisher
	cache     Cache
}

func NewOrderService(repo OrderRepository, menuRepo MenuRepository, publisher EventPublisher, cache Cache) *OrderService {
	return &OrderService{repo: repo, menuRepo: menuRepo, publisher: publisher, cache: cache}
}

// Create monta o pedido com preços do cardápio, grava e publica o evento na fila
func (s *OrderService) Create(ctx context.Context, req *models.CreateOrderRequest) (*models.Order, error) {
	ids := make([]int, 0, len(req.Items))
	for _, item := range req.Items {
		ids = append(ids, item.MenuItemID)
	}

	menuItems, err := s.menuRepo.GetByIDs(ids)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar itens do cardápio: %w", err)
	}

	byID := map[int]models.MenuItem{}
	for _, item := range menuItems {
		byID[item.ID] = item
	}

	order := &models.Order{
		CustomerName: req.CustomerName,
		Status:       models.StatusReceived,
	}
	for _, item := range req.Items {
		menuItem, ok := byID[item.MenuItemID]
		if !ok {
			return nil, ErrMenuItemNotFound
		}
		order.Items = append(order.Items, models.OrderItem{
			MenuItemID: menuItem.ID,
			Name:       menuItem.Name,
			Quantity:   item.Quantity,
			Price:      menuItem.Price,
		})
		order.Total += menuItem.Price * float64(item.Quantity)
	}

	created, err := s.repo.Create(order)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar o pedido: %w", err)
	}

	// o pedido já está salvo: se a publicação falhar, logamos e seguimos —
	// o worker recupera pedidos pendentes na inicialização
	if err := s.publisher.PublishOrderCreated(ctx, created.ID); err != nil {
		slog.Error("erro ao publicar evento de pedido criado",
			slog.String("component", "orders"),
			slog.Int("order_id", created.ID),
			slog.String("error", err.Error()))
	}

	return created, nil
}

// List pagina os pedidos, usando cache no redis para a primeira página
func (s *OrderService) List(ctx context.Context, page, limit int, status string) (*models.ListOrdersResponse, error) {
	cacheKey := fmt.Sprintf("orderflow:orders:list:status=%s:limit=%d", status, limit)

	if page == 1 {
		cached, err := s.cache.Get(ctx, cacheKey)
		if err != nil {
			slog.Warn("erro ao ler cache da listagem", slog.String("component", "orders"), slog.String("error", err.Error()))
		}
		if cached != "" {
			response := &models.ListOrdersResponse{}
			if err := json.Unmarshal([]byte(cached), response); err == nil {
				return response, nil
			}
		}
	}

	orders, total, err := s.repo.List(page, limit, status)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar os pedidos: %w", err)
	}

	response := &models.ListOrdersResponse{Orders: orders, Page: page, Limit: limit, Total: total}

	if page == 1 {
		if payload, err := json.Marshal(response); err == nil {
			if err := s.cache.Set(ctx, cacheKey, string(payload), listCacheTTL); err != nil {
				slog.Warn("erro ao gravar cache da listagem", slog.String("component", "orders"), slog.String("error", err.Error()))
			}
		}
	}

	return response, nil
}

func (s *OrderService) Get(id int) (*models.Order, error) {
	order, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar o pedido: %w", err)
	}
	if order == nil {
		return nil, ErrOrderNotFound
	}

	return order, nil
}
