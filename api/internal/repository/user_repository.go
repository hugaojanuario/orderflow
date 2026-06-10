package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"orderflow/api/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) (*models.User, error) {
	query := `
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, password_hash, created_at
	`
	created := &models.User{}
	err := r.db.QueryRow(query, user.Name, user.Email, user.PasswordHash).
		Scan(&created.ID, &created.Name, &created.Email, &created.PasswordHash, &created.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar o usuário: %w", err)
	}

	return created, nil
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, name, email, password_hash, created_at
		FROM users
		WHERE email = $1
	`
	user := &models.User{}
	err := r.db.QueryRow(query, email).
		Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar o usuário: %w", err)
	}

	return user, nil
}
