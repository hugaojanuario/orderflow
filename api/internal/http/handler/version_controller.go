package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"orderflow/api/internal/models"
)

type VersionController struct {
	version string
	commit  string
}

func NewVersionController(version, commit string) *VersionController {
	return &VersionController{version: version, commit: commit}
}

func (v *VersionController) Version(c *gin.Context) {
	c.JSON(http.StatusOK, models.VersionResponse{Version: v.version, Commit: v.commit})
}
