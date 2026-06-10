package services

import (
	"fmt"

	"orderflow/api/internal/models"
)

type MenuService struct {
	repo MenuRepository
}

func NewMenuService(repo MenuRepository) *MenuService {
	return &MenuService{repo: repo}
}

func (s *MenuService) List() ([]models.MenuItem, error) {
	items, err := s.repo.List()
	if err != nil {
		return nil, fmt.Errorf("erro ao listar o cardápio: %w", err)
	}

	return items, nil
}
