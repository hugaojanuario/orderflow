package models

import "time"

type MenuItem struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Available   bool      `json:"available"`
	CreatedAt   time.Time `json:"created_at"`
}
