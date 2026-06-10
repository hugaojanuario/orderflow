package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"orderflow/worker/internal/models"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// GetStatus retorna o status atual do pedido; string vazia quando não existe
func (r *OrderRepository) GetStatus(orderID int) (string, error) {
	query := `
		SELECT status
		FROM orders
		WHERE id = $1
	`
	var status string
	err := r.db.QueryRow(query, orderID).Scan(&status)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("erro ao buscar status do pedido: %w", err)
	}

	return status, nil
}

// AdvanceStatus avança o status de forma idempotente: só aplica a transição
// se o pedido ainda estiver no status esperado, registrando o histórico na
// mesma transação. Retorna false quando outro processo já avançou o pedido.
func (r *OrderRepository) AdvanceStatus(orderID int, from, to string) (bool, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return false, fmt.Errorf("erro ao abrir transação: %w", err)
	}
	defer tx.Rollback()

	updateQuery := `
		UPDATE orders
		SET status = $1, updated_at = NOW()
		WHERE id = $2 AND status = $3
	`
	result, err := tx.Exec(updateQuery, to, orderID, from)
	if err != nil {
		return false, fmt.Errorf("erro ao atualizar status do pedido: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("erro ao verificar atualização do pedido: %w", err)
	}
	if rows == 0 {
		return false, nil
	}

	historyQuery := `
		INSERT INTO order_status_history (order_id, status)
		VALUES ($1, $2)
	`
	if _, err = tx.Exec(historyQuery, orderID, to); err != nil {
		return false, fmt.Errorf("erro ao registrar histórico do pedido: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return false, fmt.Errorf("erro ao confirmar transação: %w", err)
	}

	return true, nil
}

// ListUnfinished retorna os ids de pedidos que ainda não chegaram ao status final
func (r *OrderRepository) ListUnfinished() ([]int, error) {
	query := `
		SELECT id
		FROM orders
		WHERE status != $1
		ORDER BY created_at
	`
	rows, err := r.db.Query(query, models.StatusDelivered)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar pedidos pendentes: %w", err)
	}
	defer rows.Close()

	ids := []int{}
	for rows.Next() {
		var id int
		if err = rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("erro ao ler pedido pendente: %w", err)
		}
		ids = append(ids, id)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("erro ao percorrer pedidos pendentes: %w", err)
	}

	return ids, nil
}
