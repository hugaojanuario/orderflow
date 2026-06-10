package repository

import (
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Seed popula o cardápio e cria o usuário admin; é idempotente
func Seed(db *sql.DB) error {
	menuItems := []struct {
		name        string
		description string
		price       float64
	}{
		{"X-Burger", "hambúrguer artesanal com queijo e molho da casa", 28.90},
		{"X-Bacon", "hambúrguer artesanal com bacon crocante", 32.90},
		{"Batata Frita", "porção de batata frita com páprica", 14.90},
		{"Salada Caesar", "alface americana, frango grelhado e parmesão", 26.50},
		{"Pizza Margherita", "molho de tomate, muçarela e manjericão", 45.00},
		{"Refrigerante Lata", "350ml, sabores variados", 6.50},
		{"Suco Natural", "laranja, limão ou abacaxi, 500ml", 9.90},
		{"Pudim de Leite", "sobremesa da casa", 12.00},
	}

	menuQuery := `
		INSERT INTO menu_items (name, description, price)
		VALUES ($1, $2, $3)
		ON CONFLICT (name) DO NOTHING
	`
	for _, item := range menuItems {
		if _, err := db.Exec(menuQuery, item.name, item.description, item.price); err != nil {
			return fmt.Errorf("erro ao inserir item do cardápio %q: %w", item.name, err)
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("erro ao gerar hash da senha do admin: %w", err)
	}

	userQuery := `
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
		ON CONFLICT (email) DO NOTHING
	`
	if _, err := db.Exec(userQuery, "Admin", "admin@orderflow.local", string(hash)); err != nil {
		return fmt.Errorf("erro ao criar o usuário admin: %w", err)
	}

	return nil
}
