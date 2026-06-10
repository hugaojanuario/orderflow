package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"orderflow/api/internal/services"
	"orderflow/api/internal/utils"
)

type MenuController struct {
	service *services.MenuService
}

func NewMenuController(service *services.MenuService) *MenuController {
	return &MenuController{service: service}
}

func (m *MenuController) List(c *gin.Context) {
	items, err := m.service.List()
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, items)
}
