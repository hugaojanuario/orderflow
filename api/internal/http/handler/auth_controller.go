package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"orderflow/api/internal/models"
	"orderflow/api/internal/services"
	"orderflow/api/internal/utils"
)

type AuthController struct {
	service *services.AuthService
}

func NewAuthController(service *services.AuthService) *AuthController {
	return &AuthController{service: service}
}

func (a *AuthController) Register(c *gin.Context) {
	req := &models.RegisterRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body invalid"})
		return
	}

	user, err := a.service.Register(req)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (a *AuthController) Login(c *gin.Context) {
	req := &models.LoginRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body invalid"})
		return
	}

	response, err := a.service.Login(req)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}
