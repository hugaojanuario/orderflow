package repository

import (
	"database/sql"
	"fmt"

	"github.com/lib/pq"

	"orderflow/api/internal/models"
)

type MenuRepository struct {
	db *sql.DB
}

func NewMenuRepository(db *sql.DB) *MenuRepository {
	return &MenuRepository{db: db}
}

func (r *MenuRepository) List() ([]models.MenuItem, error) {
	query := `
		SELECT id, name, description, price, available, created_at
		FROM menu_items
		WHERE available = TRUE
		ORDER BY name
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar o cardápio: %w", err)
	}
	defer rows.Close()

	items := []models.MenuItem{}
	for rows.Next() {
		item := models.MenuItem{}
		err = rows.Scan(&item.ID, &item.Name, &item.Description, &item.Price, &item.Available, &item.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("erro ao ler item do cardápio: %w", err)
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("erro ao percorrer o cardápio: %w", err)
	}

	return items, nil
}

func (r *MenuRepository) GetByIDs(ids []int) ([]models.MenuItem, error) {
	query := `
		SELECT id, name, description, price, available, created_at
		FROM menu_items
		WHERE id = ANY($1)
	`
	rows, err := r.db.Query(query, pq.Array(ids))
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar itens do cardápio: %w", err)
	}
	defer rows.Close()

	items := []models.MenuItem{}
	for rows.Next() {
		item := models.MenuItem{}
		err = rows.Scan(&item.ID, &item.Name, &item.Description, &item.Price, &item.Available, &item.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("erro ao ler item do cardápio: %w", err)
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("erro ao percorrer itens do cardápio: %w", err)
	}

	return items, nil
}
