package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	applyrics "sonora.dev/go-core/application/lyrics"

	"sonora.dev/backend/internal/http/response"
)

// LyricsProviderHandler backs the admin Lyrics Source page (Sprint 14,
// docs/screens-spec.md #19).
type LyricsProviderHandler struct {
	service *applyrics.Service
}

func NewLyricsProviderHandler(service *applyrics.Service) *LyricsProviderHandler {
	return &LyricsProviderHandler{service: service}
}

func (h *LyricsProviderHandler) List(c *fiber.Ctx) error {
	providers, err := h.service.ListProviders(c.Context())
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to load lyrics providers")
	}
	out := make([]fiber.Map, 0, len(providers))
	for _, p := range providers {
		out = append(out, fiber.Map{
			"id":             p.ID,
			"name":           p.Name,
			"base_url":       p.BaseURL,
			"is_enabled":     p.IsEnabled,
			"priority":       p.Priority,
			"health_status":  p.HealthStatus,
			"total_lookups":  p.TotalLookups,
			"match_rate_pct": p.MatchRate(),
		})
	}
	return response.OK(c, fiber.StatusOK, out)
}

type updateProviderRequest struct {
	Priority  *int  `json:"priority"`
	IsEnabled *bool `json:"is_enabled"`
}

func (h *LyricsProviderHandler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid provider id")
	}
	var req updateProviderRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid request body")
	}
	if req.Priority != nil {
		if err := h.service.SetProviderPriority(c.Context(), id, *req.Priority); err != nil {
			return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to update priority")
		}
	}
	if req.IsEnabled != nil {
		if err := h.service.SetProviderEnabled(c.Context(), id, *req.IsEnabled); err != nil {
			return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to update enabled state")
		}
	}
	return response.OK(c, fiber.StatusOK, fiber.Map{})
}
