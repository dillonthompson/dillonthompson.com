package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	db *sql.DB
}

func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) Check(c *gin.Context) {
	dbStatus := "up"
	if err := h.db.PingContext(c.Request.Context()); err != nil {
		dbStatus = "down"
	}

	status := http.StatusOK
	if dbStatus != "up" {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, gin.H{
		"status":    dbStatus,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
