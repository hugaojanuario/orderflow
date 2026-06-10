package repository

import (
	"database/sql"
	"fmt"
	"time"

	"orderflow/api/internal/models"
)

type StatsRepository struct {
	db *sql.DB
}

func NewStatsRepository(db *sql.DB) *StatsRepository {
	return &StatsRepository{db: db}
}

// StatsToday agrega os números do dia corrente
func (r *StatsRepository) StatsToday() (*models.StatsResponse, error) {
	stats := &models.StatsResponse{
		Date:           time.Now().Format("2006-01-02"),
		OrdersByStatus: map[string]int{},
	}

	statusQuery := `
		SELECT status, COUNT(*)
		FROM orders
		WHERE created_at::date = CURRENT_DATE
		GROUP BY status
	`
	rows, err := r.db.Query(statusQuery)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar pedidos por status: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int
		if err = rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("erro ao ler pedidos por status: %w", err)
		}
		stats.OrdersByStatus[status] = count
		stats.TotalOrders += count
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("erro ao percorrer pedidos por status: %w", err)
	}

	revenueQuery := `
		SELECT COALESCE(SUM(total), 0)
		FROM orders
		WHERE created_at::date = CURRENT_DATE
	`
	if err = r.db.QueryRow(revenueQuery).Scan(&stats.Revenue); err != nil {
		return nil, fmt.Errorf("erro ao calcular o faturamento: %w", err)
	}

	// tempo médio entre o pedido ser recebido e ficar pronto
	prepQuery := `
		SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (ready.created_at - received.created_at))), 0)
		FROM order_status_history received
		JOIN order_status_history ready
			ON ready.order_id = received.order_id AND ready.status = 'ready'
		WHERE received.status = 'received'
			AND received.created_at::date = CURRENT_DATE
	`
	if err = r.db.QueryRow(prepQuery).Scan(&stats.AvgPrepSeconds); err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("erro ao calcular o tempo médio de preparo: %w", err)
	}

	return stats, nil
}
