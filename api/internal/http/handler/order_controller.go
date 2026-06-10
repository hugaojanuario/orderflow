package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"orderflow/api/internal/models"
	"orderflow/api/internal/services"
	"orderflow/api/internal/utils"
)

type OrderController struct {
	service *services.OrderService
}

func NewOrderController(service *services.OrderService) *OrderController {
	return &OrderController{service: service}
}

func (o *OrderController) Create(c *gin.Context) {
	req := &models.CreateOrderRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body invalid"})
		return
	}

	order, err := o.service.Create(c.Request.Context(), req)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, order)
}

func (o *OrderController) List(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	status := c.Query("status")

	response, err := o.service.List(c.Request.Context(), page, limit, status)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (o *OrderController) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	order, err := o.service.Get(id)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, order)
}
