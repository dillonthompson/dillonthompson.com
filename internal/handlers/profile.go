package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/dillonthompson/dillonthompson.com/internal/repository"
	"github.com/gin-gonic/gin"
)

type ProfileHandler struct {
	queries *repository.Queries
}

func NewProfileHandler(queries *repository.Queries) *ProfileHandler {
	return &ProfileHandler{queries: queries}
}

func (h *ProfileHandler) GetFull(c *gin.Context) {
	c.Header("Cache-Control", "no-store, max-age=0")

	rows, err := h.queries.GetFullProfile(c.Request.Context())
	if err != nil {
		slog.Error("failed to fetch profile", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	result := make(map[string]json.RawMessage, len(rows))
	for _, r := range rows {
		result[r.Key] = r.Content
	}

	c.JSON(http.StatusOK, result)
}

func (h *ProfileHandler) GetSection(c *gin.Context) {
	c.Header("Cache-Control", "no-store, max-age=0")

	key := c.Param("key")
	content, err := h.queries.GetProfileSection(c.Request.Context(), key)
	if err != nil {
		slog.Error("failed to fetch profile section", "key", key, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "section not found"})
		return
	}

	c.Data(http.StatusOK, "application/json", content)
}
