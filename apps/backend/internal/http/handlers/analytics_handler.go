package handlers

import (
	"github.com/gofiber/fiber/v2"

	appanalytics "sonora.dev/go-core/application/analytics"

	"sonora.dev/backend/internal/http/response"
)

const topPlayedLimit = 10

// AnalyticsHandler backs the admin Analytics page (Sprint 11,
// docs/screens-spec.md #21).
type AnalyticsHandler struct {
	service *appanalytics.Service
}

func NewAnalyticsHandler(service *appanalytics.Service) *AnalyticsHandler {
	return &AnalyticsHandler{service: service}
}

func (h *AnalyticsHandler) TopPlayed(c *fiber.Ctx) error {
	songs, err := h.service.TopPlayed(c.Context(), topPlayedLimit)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to load top played songs")
	}
	out := make([]fiber.Map, 0, len(songs))
	for _, s := range songs {
		out = append(out, fiber.Map{
			"song_id":     s.SongID,
			"title":       s.Title,
			"artist_name": s.ArtistName,
			"play_count":  s.PlayCount,
		})
	}
	return response.OK(c, fiber.StatusOK, out)
}

func (h *AnalyticsHandler) StorageGrowth(c *fiber.Ctx) error {
	points, err := h.service.StorageGrowth(c.Context())
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to load storage growth")
	}
	out := make([]fiber.Map, 0, len(points))
	for _, p := range points {
		out = append(out, fiber.Map{
			"month":       p.Month.Format("2006-01"),
			"total_bytes": p.TotalBytes,
		})
	}
	return response.OK(c, fiber.StatusOK, out)
}
