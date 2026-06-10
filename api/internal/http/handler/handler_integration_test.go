package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"orderflow/api/internal/http/handler"
	"orderflow/api/internal/http/router"
	"orderflow/api/internal/models"
	"orderflow/api/internal/services"
)

const testJWTSecret = "integration-test-secret"

// mocks em memória para subir o router completo sem postgres/redis

type memoryUserRepository struct {
	users map[string]*models.User
}

func (m *memoryUserRepository) Create(user *models.User) (*models.User, error) {
	created := *user
	created.ID = len(m.users) + 1
	m.users[user.Email] = &created
	return &created, nil
}

func (m *memoryUserRepository) GetByEmail(email string) (*models.User, error) {
	user, ok := m.users[email]
	if !ok {
		return nil, nil
	}
	return user, nil
}

type memoryOrderRepository struct {
	orders map[int]*models.Order
}

func (m *memoryOrderRepository) Create(order *models.Order) (*models.Order, error) {
	created := *order
	created.ID = len(m.orders) + 1
	created.CreatedAt = time.Now()
	created.UpdatedAt = time.Now()
	m.orders[created.ID] = &created
	return &created, nil
}

func (m *memoryOrderRepository) List(page, limit int, status string) ([]models.Order, int, error) {
	orders := []models.Order{}
	for _, order := range m.orders {
		orders = append(orders, *order)
	}
	return orders, len(orders), nil
}

func (m *memoryOrderRepository) GetByID(id int) (*models.Order, error) {
	order, ok := m.orders[id]
	if !ok {
		return nil, nil
	}
	return order, nil
}

type memoryMenuRepository struct {
	items map[int]models.MenuItem
}

func (m *memoryMenuRepository) List() ([]models.MenuItem, error) {
	items := []models.MenuItem{}
	for _, item := range m.items {
		items = append(items, item)
	}
	return items, nil
}

func (m *memoryMenuRepository) GetByIDs(ids []int) ([]models.MenuItem, error) {
	items := []models.MenuItem{}
	for _, id := range ids {
		if item, ok := m.items[id]; ok {
			items = append(items, item)
		}
	}
	return items, nil
}

type memoryStatsRepository struct{}

func (m *memoryStatsRepository) StatsToday() (*models.StatsResponse, error) {
	return &models.StatsResponse{OrdersByStatus: map[string]int{}}, nil
}

type memoryPublisher struct {
	published []int
}

func (m *memoryPublisher) PublishOrderCreated(_ context.Context, orderID int) error {
	m.published = append(m.published, orderID)
	return nil
}

type memoryCache struct {
	values map[string]string
}

func (m *memoryCache) Get(_ context.Context, key string) (string, error) {
	return m.values[key], nil
}

func (m *memoryCache) Set(_ context.Context, key, value string, _ time.Duration) error {
	m.values[key] = value
	return nil
}

func setupTestRouter() (*gin.Engine, *memoryPublisher) {
	gin.SetMode(gin.TestMode)

	userRepo := &memoryUserRepository{users: map[string]*models.User{}}
	orderRepo := &memoryOrderRepository{orders: map[int]*models.Order{}}
	menuRepo := &memoryMenuRepository{items: map[int]models.MenuItem{
		1: {ID: 1, Name: "X-Burger", Price: 28.90, Available: true},
	}}
	publisher := &memoryPublisher{}
	cache := &memoryCache{values: map[string]string{}}

	authService := services.NewAuthService(userRepo, testJWTSecret)
	orderService := services.NewOrderService(orderRepo, menuRepo, publisher, cache)
	menuService := services.NewMenuService(menuRepo)
	statsService := services.NewStatsService(&memoryStatsRepository{})

	engine := router.SetupRouter(
		handler.NewAuthController(authService),
		handler.NewOrderController(orderService),
		handler.NewMenuController(menuService),
		handler.NewStatsController(statsService),
		handler.NewHealthController(nil, nil),
		handler.NewVersionController("test", "none"),
		testJWTSecret,
	)

	return engine, publisher
}

func doRequest(engine *gin.Engine, method, path, token string, body any) *httptest.ResponseRecorder {
	var reader *bytes.Reader
	if body != nil {
		payload, _ := json.Marshal(body)
		reader = bytes.NewReader(payload)
	} else {
		reader = bytes.NewReader(nil)
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)
	return recorder
}

func registerAndLogin(t *testing.T, engine *gin.Engine) string {
	t.Helper()

	res := doRequest(engine, http.MethodPost, "/api/v1/auth/register", "", models.RegisterRequest{
		Name:     "Hugo",
		Email:    "hugo@test.local",
		Password: "senha123",
	})
	if res.Code != http.StatusCreated {
		t.Fatalf("esperava 201 no register, recebeu %d: %s", res.Code, res.Body.String())
	}

	res = doRequest(engine, http.MethodPost, "/api/v1/auth/login", "", models.LoginRequest{
		Email:    "hugo@test.local",
		Password: "senha123",
	})
	if res.Code != http.StatusOK {
		t.Fatalf("esperava 200 no login, recebeu %d: %s", res.Code, res.Body.String())
	}

	login := models.LoginResponse{}
	if err := json.Unmarshal(res.Body.Bytes(), &login); err != nil {
		t.Fatalf("erro ao decodificar resposta do login: %v", err)
	}
	if login.Token == "" {
		t.Fatal("esperava token no login")
	}

	return login.Token
}

func TestCreateOrderEndToEnd(t *testing.T) {
	engine, publisher := setupTestRouter()
	token := registerAndLogin(t, engine)

	res := doRequest(engine, http.MethodPost, "/api/v1/orders", token, models.CreateOrderRequest{
		CustomerName: "Maria",
		Items:        []models.CreateOrderItemRequest{{MenuItemID: 1, Quantity: 2}},
	})
	if res.Code != http.StatusCreated {
		t.Fatalf("esperava 201, recebeu %d: %s", res.Code, res.Body.String())
	}

	order := models.Order{}
	if err := json.Unmarshal(res.Body.Bytes(), &order); err != nil {
		t.Fatalf("erro ao decodificar pedido: %v", err)
	}
	if order.Total != 57.80 {
		t.Fatalf("esperava total 57.80, recebeu %.2f", order.Total)
	}
	if order.Status != models.StatusReceived {
		t.Fatalf("esperava status received, recebeu %q", order.Status)
	}
	if len(publisher.published) != 1 {
		t.Fatalf("esperava 1 evento publicado, recebeu %d", len(publisher.published))
	}
}

func TestCreateOrderWithoutTokenReturns401(t *testing.T) {
	engine, _ := setupTestRouter()

	res := doRequest(engine, http.MethodPost, "/api/v1/orders", "", models.CreateOrderRequest{
		CustomerName: "Maria",
		Items:        []models.CreateOrderItemRequest{{MenuItemID: 1, Quantity: 1}},
	})
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("esperava 401, recebeu %d: %s", res.Code, res.Body.String())
	}
}

func TestListOrdersEndpoint(t *testing.T) {
	engine, _ := setupTestRouter()
	token := registerAndLogin(t, engine)

	res := doRequest(engine, http.MethodPost, "/api/v1/orders", token, models.CreateOrderRequest{
		CustomerName: "Maria",
		Items:        []models.CreateOrderItemRequest{{MenuItemID: 1, Quantity: 1}},
	})
	if res.Code != http.StatusCreated {
		t.Fatalf("esperava 201 ao criar pedido, recebeu %d: %s", res.Code, res.Body.String())
	}

	res = doRequest(engine, http.MethodGet, "/api/v1/orders?page=1&limit=20", "", nil)
	if res.Code != http.StatusOK {
		t.Fatalf("esperava 200, recebeu %d: %s", res.Code, res.Body.String())
	}

	response := models.ListOrdersResponse{}
	if err := json.Unmarshal(res.Body.Bytes(), &response); err != nil {
		t.Fatalf("erro ao decodificar listagem: %v", err)
	}
	if response.Total != 1 || len(response.Orders) != 1 {
		t.Fatalf("esperava 1 pedido na listagem, recebeu total=%d itens=%d", response.Total, len(response.Orders))
	}
}

func TestLoginWithWrongPasswordReturns401(t *testing.T) {
	engine, _ := setupTestRouter()
	registerAndLogin(t, engine)

	res := doRequest(engine, http.MethodPost, "/api/v1/auth/login", "", models.LoginRequest{
		Email:    "hugo@test.local",
		Password: "senha-errada",
	})
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("esperava 401, recebeu %d: %s", res.Code, res.Body.String())
	}
}
