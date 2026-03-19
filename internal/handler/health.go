package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/lvtao/go-gin-api-admin/internal/config"
	"github.com/lvtao/go-gin-api-admin/pkg/response"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Index(c *gin.Context) {
	response.Success(c, gin.H{
		"message": "Go Gin API Admin is running",
	})
}

func (h *HealthHandler) Health(c *gin.Context) {
	cfg := config.GetConfig()
	response.Success(c, gin.H{
		"status":  "ok",
		"name":    cfg.App.Name,
		"version": cfg.App.Version,
	})
}
