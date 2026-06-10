package services

import (
	"fmt"

	"orderflow/api/internal/models"
)

// StatsRepository é o contrato que o service precisa do repositório de stats
type StatsRepository interface {
	StatsToday() (*models.StatsResponse, error)
}

type StatsService struct {
	repo StatsRepository
}

func NewStatsService(repo StatsRepository) *StatsService {
	return &StatsService{repo: repo}
}

func (s *StatsService) Today() (*models.StatsResponse, error) {
	stats, err := s.repo.StatsToday()
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar as estatísticas do dia: %w", err)
	}

	return stats, nil
}
