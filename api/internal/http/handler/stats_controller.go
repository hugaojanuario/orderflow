package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"orderflow/api/internal/services"
	"orderflow/api/internal/utils"
)

type StatsController struct {
	service *services.StatsService
}

func NewStatsController(service *services.StatsService) *StatsController {
	return &StatsController{service: service}
}

func (s *StatsController) Today(c *gin.Context) {
	stats, err := s.service.Today()
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, stats)
}
