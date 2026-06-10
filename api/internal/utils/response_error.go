package utils

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"orderflow/api/internal/services"
)

// RespondError traduz erros de domínio para o status http adequado
func RespondError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "credenciais inválidas"})
	case errors.Is(err, services.ErrEmailAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{"error": "email já cadastrado"})
	case errors.Is(err, services.ErrOrderNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "pedido não encontrado"})
	case errors.Is(err, services.ErrMenuItemNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"error": "item do cardápio não encontrado"})
	default:
		slog.Error("erro interno",
			slog.String("component", "http"),
			slog.String("request_id", c.GetString("requestId")),
			slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro interno"})
	}
}
