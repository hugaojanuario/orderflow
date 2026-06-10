package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"orderflow/api/internal/models"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// Create grava o pedido, os itens e a primeira entrada do histórico numa transação
func (r *OrderRepository) Create(order *models.Order) (*models.Order, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir transação: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO orders (customer_name, total, status)
		VALUES ($1, $2, $3)
		RETURNING id, customer_name, total, status, created_at, updated_at
	`
	created := &models.Order{}
	err = tx.QueryRow(query, order.CustomerName, order.Total, models.StatusReceived).
		Scan(&created.ID, &created.CustomerName, &created.Total, &created.Status, &created.CreatedAt, &created.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar o pedido: %w", err)
	}

	itemQuery := `
		INSERT INTO order_items (order_id, menu_item_id, quantity, price)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	for _, item := range order.Items {
		saved := item
		saved.OrderID = created.ID
		err = tx.QueryRow(itemQuery, created.ID, item.MenuItemID, item.Quantity, item.Price).Scan(&saved.ID)
		if err != nil {
			return nil, fmt.Errorf("erro ao criar item do pedido: %w", err)
		}
		created.Items = append(created.Items, saved)
	}

	historyQuery := `
		INSERT INTO order_status_history (order_id, status)
		VALUES ($1, $2)
	`
	if _, err = tx.Exec(historyQuery, created.ID, models.StatusReceived); err != nil {
		return nil, fmt.Errorf("erro ao registrar histórico do pedido: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("erro ao confirmar transação: %w", err)
	}

	return created, nil
}

// List retorna a página de pedidos e o total, com filtro opcional por status
func (r *OrderRepository) List(page, limit int, status string) ([]models.Order, int, error) {
	offset := (page - 1) * limit

	var total int
	var rows *sql.Rows
	var err error

	if status != "" {
		countQuery := `SELECT COUNT(*) FROM orders WHERE status = $1`
		if err = r.db.QueryRow(countQuery, status).Scan(&total); err != nil {
			return nil, 0, fmt.Errorf("erro ao contar os pedidos: %w", err)
		}
		query := `
			SELECT id, customer_name, total, status, created_at, updated_at
			FROM orders
			WHERE status = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
		rows, err = r.db.Query(query, status, limit, offset)
	} else {
		countQuery := `SELECT COUNT(*) FROM orders`
		if err = r.db.QueryRow(countQuery).Scan(&total); err != nil {
			return nil, 0, fmt.Errorf("erro ao contar os pedidos: %w", err)
		}
		query := `
			SELECT id, customer_name, total, status, created_at, updated_at
			FROM orders
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2
		`
		rows, err = r.db.Query(query, limit, offset)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("erro ao listar os pedidos: %w", err)
	}
	defer rows.Close()

	orders := []models.Order{}
	for rows.Next() {
		order := models.Order{}
		err = rows.Scan(&order.ID, &order.CustomerName, &order.Total, &order.Status, &order.CreatedAt, &order.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("erro ao ler pedido: %w", err)
		}
		orders = append(orders, order)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("erro ao percorrer os pedidos: %w", err)
	}

	return orders, total, nil
}

// GetByID retorna o pedido com itens e histórico; nil quando não existe
func (r *OrderRepository) GetByID(id int) (*models.Order, error) {
	query := `
		SELECT id, customer_name, total, status, created_at, updated_at
		FROM orders
		WHERE id = $1
	`
	order := &models.Order{}
	err := r.db.QueryRow(query, id).
		Scan(&order.ID, &order.CustomerName, &order.Total, &order.Status, &order.CreatedAt, &order.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar o pedido: %w", err)
	}

	itemsQuery := `
		SELECT oi.id, oi.order_id, oi.menu_item_id, mi.name, oi.quantity, oi.price
		FROM order_items oi
		JOIN menu_items mi ON mi.id = oi.menu_item_id
		WHERE oi.order_id = $1
		ORDER BY oi.id
	`
	rows, err := r.db.Query(itemsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar itens do pedido: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		item := models.OrderItem{}
		err = rows.Scan(&item.ID, &item.OrderID, &item.MenuItemID, &item.Name, &item.Quantity, &item.Price)
		if err != nil {
			return nil, fmt.Errorf("erro ao ler item do pedido: %w", err)
		}
		order.Items = append(order.Items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("erro ao percorrer itens do pedido: %w", err)
	}

	historyQuery := `
		SELECT id, order_id, status, created_at
		FROM order_status_history
		WHERE order_id = $1
		ORDER BY created_at, id
	`
	historyRows, err := r.db.Query(historyQuery, id)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar histórico do pedido: %w", err)
	}
	defer historyRows.Close()

	for historyRows.Next() {
		entry := models.OrderStatusHistory{}
		err = historyRows.Scan(&entry.ID, &entry.OrderID, &entry.Status, &entry.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("erro ao ler histórico do pedido: %w", err)
		}
		order.History = append(order.History, entry)
	}
	if err = historyRows.Err(); err != nil {
		return nil, fmt.Errorf("erro ao percorrer histórico do pedido: %w", err)
	}

	return order, nil
}

// CountByStatus retorna a contagem de pedidos agrupada por status (usado nas métricas)
func (r *OrderRepository) CountByStatus() (map[string]int, error) {
	query := `
		SELECT status, COUNT(*)
		FROM orders
		GROUP BY status
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("erro ao contar pedidos por status: %w", err)
	}
	defer rows.Close()

	counts := map[string]int{}
	for rows.Next() {
		var status string
		var count int
		if err = rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("erro ao ler contagem por status: %w", err)
		}
		counts[status] = count
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("erro ao percorrer contagem por status: %w", err)
	}

	return counts, nil
}
