package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/dillonthompson/dillonthompson.com/internal/repository"
	"github.com/gin-gonic/gin"
)

type ExperienceHandler struct {
	queries *repository.Queries
}

func NewExperienceHandler(queries *repository.Queries) *ExperienceHandler {
	return &ExperienceHandler{queries: queries}
}

type experienceResponse struct {
	ID                 string          `json:"id"`
	Company            string          `json:"company"`
	Role               string          `json:"role"`
	StartDate          string          `json:"start_date"`
	EndDate            *string         `json:"end_date"`
	Description        json.RawMessage `json:"description"`
	TechStack          json.RawMessage `json:"tech_stack"`
	SecurityHighlights json.RawMessage `json:"security_highlights"`
	Metadata           json.RawMessage `json:"metadata"`
	CreatedAt          time.Time       `json:"created_at"`
}

func (h *ExperienceHandler) List(c *gin.Context) {
	c.Header("Cache-Control", "no-store, max-age=0")

	rows, err := h.queries.GetPublishedExperiences(c.Request.Context())
	if err != nil {
		slog.Error("failed to fetch experiences", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	experiences := make([]experienceResponse, len(rows))
	for i, r := range rows {
		var endDate *string
		if r.EndDate.Valid {
			s := r.EndDate.Time.Format("2006-01-02")
			endDate = &s
		}
		experiences[i] = experienceResponse{
			ID:                 r.ID,
			Company:            r.Company,
			Role:               r.Role,
			StartDate:          r.StartDate.Format("2006-01-02"),
			EndDate:            endDate,
			Description:        r.Description,
			TechStack:          r.TechStack,
			SecurityHighlights: r.SecurityHighlights,
			Metadata:           r.Metadata,
			CreatedAt:          r.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, experiences)
}
