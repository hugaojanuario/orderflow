package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"orderflow/api/internal/models"
)

type mockOrderRepository struct {
	orders    map[int]*models.Order
	listCalls int
}

func newMockOrderRepository() *mockOrderRepository {
	return &mockOrderRepository{orders: map[int]*models.Order{}}
}

func (m *mockOrderRepository) Create(order *models.Order) (*models.Order, error) {
	created := *order
	created.ID = len(m.orders) + 1
	created.CreatedAt = time.Now()
	created.UpdatedAt = time.Now()
	m.orders[created.ID] = &created
	return &created, nil
}

func (m *mockOrderRepository) List(page, limit int, status string) ([]models.Order, int, error) {
	m.listCalls++
	orders := []models.Order{}
	for _, order := range m.orders {
		orders = append(orders, *order)
	}
	return orders, len(orders), nil
}

func (m *mockOrderRepository) GetByID(id int) (*models.Order, error) {
	order, ok := m.orders[id]
	if !ok {
		return nil, nil
	}
	return order, nil
}

type mockMenuRepository struct {
	items map[int]models.MenuItem
}

func (m *mockMenuRepository) List() ([]models.MenuItem, error) {
	items := []models.MenuItem{}
	for _, item := range m.items {
		items = append(items, item)
	}
	return items, nil
}

func (m *mockMenuRepository) GetByIDs(ids []int) ([]models.MenuItem, error) {
	items := []models.MenuItem{}
	for _, id := range ids {
		if item, ok := m.items[id]; ok {
			items = append(items, item)
		}
	}
	return items, nil
}

type mockPublisher struct {
	published []int
	err       error
}

func (m *mockPublisher) PublishOrderCreated(_ context.Context, orderID int) error {
	if m.err != nil {
		return m.err
	}
	m.published = append(m.published, orderID)
	return nil
}

type mockCache struct {
	values map[string]string
}

func newMockCache() *mockCache {
	return &mockCache{values: map[string]string{}}
}

func (m *mockCache) Get(_ context.Context, key string) (string, error) {
	return m.values[key], nil
}

func (m *mockCache) Set(_ context.Context, key, value string, _ time.Duration) error {
	m.values[key] = value
	return nil
}

func newOrderServiceForTest() (*OrderService, *mockOrderRepository, *mockPublisher, *mockCache) {
	repo := newMockOrderRepository()
	menuRepo := &mockMenuRepository{items: map[int]models.MenuItem{
		1: {ID: 1, Name: "X-Burger", Price: 28.90, Available: true},
		2: {ID: 2, Name: "Batata Frita", Price: 14.90, Available: true},
	}}
	publisher := &mockPublisher{}
	cache := newMockCache()
	service := NewOrderService(repo, menuRepo, publisher, cache)
	return service, repo, publisher, cache
}

func TestCreateOrderComputesTotalAndPublishesEvent(t *testing.T) {
	service, _, publisher, _ := newOrderServiceForTest()

	order, err := service.Create(context.Background(), &models.CreateOrderRequest{
		CustomerName: "Maria",
		Items: []models.CreateOrderItemRequest{
			{MenuItemID: 1, Quantity: 2},
			{MenuItemID: 2, Quantity: 1},
		},
	})
	if err != nil {
		t.Fatalf("esperava sucesso, recebeu erro: %v", err)
	}

	expectedTotal := 28.90*2 + 14.90
	if order.Total != expectedTotal {
		t.Fatalf("esperava total %.2f, recebeu %.2f", expectedTotal, order.Total)
	}
	if order.Status != models.StatusReceived {
		t.Fatalf("esperava status %q, recebeu %q", models.StatusReceived, order.Status)
	}
	if len(publisher.published) != 1 || publisher.published[0] != order.ID {
		t.Fatalf("esperava evento publicado para o pedido %d, recebeu: %v", order.ID, publisher.published)
	}
}

func TestCreateOrderMenuItemNotFound(t *testing.T) {
	service, _, publisher, _ := newOrderServiceForTest()

	_, err := service.Create(context.Background(), &models.CreateOrderRequest{
		CustomerName: "Maria",
		Items:        []models.CreateOrderItemRequest{{MenuItemID: 99, Quantity: 1}},
	})
	if !errors.Is(err, ErrMenuItemNotFound) {
		t.Fatalf("esperava ErrMenuItemNotFound, recebeu: %v", err)
	}
	if len(publisher.published) != 0 {
		t.Fatal("não deveria publicar evento quando o pedido falha")
	}
}

func TestListFirstPageUsesCache(t *testing.T) {
	service, repo, _, _ := newOrderServiceForTest()

	if _, err := service.Create(context.Background(), &models.CreateOrderRequest{
		CustomerName: "Maria",
		Items:        []models.CreateOrderItemRequest{{MenuItemID: 1, Quantity: 1}},
	}); err != nil {
		t.Fatalf("erro ao criar pedido: %v", err)
	}

	if _, err := service.List(context.Background(), 1, 20, ""); err != nil {
		t.Fatalf("erro na primeira listagem: %v", err)
	}
	if _, err := service.List(context.Background(), 1, 20, ""); err != nil {
		t.Fatalf("erro na segunda listagem: %v", err)
	}

	// a segunda chamada deve vir do cache, sem bater no repository
	if repo.listCalls != 1 {
		t.Fatalf("esperava 1 chamada ao repository, recebeu %d", repo.listCalls)
	}
}

func TestGetOrderNotFound(t *testing.T) {
	service, _, _, _ := newOrderServiceForTest()

	_, err := service.Get(42)
	if !errors.Is(err, ErrOrderNotFound) {
		t.Fatalf("esperava ErrOrderNotFound, recebeu: %v", err)
	}
}
