package handlers

import (
	"github.com/gofiber/fiber/v2"

	appdashboard "sonora.dev/go-core/application/dashboard"

	"sonora.dev/backend/internal/http/response"
)

// DashboardHandler backs the admin Dashboard page (Sprint 14,
// docs/screens-spec.md #16).
type DashboardHandler struct {
	service *appdashboard.Service
}

func NewDashboardHandler(service *appdashboard.Service) *DashboardHandler {
	return &DashboardHandler{service: service}
}

func (h *DashboardHandler) Get(c *fiber.Ctx) error {
	stats, err := h.service.GetStats(c.Context())
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to load dashboard stats")
	}
	distribution, err := h.service.StorageDistribution(c.Context())
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to load storage distribution")
	}
	jobsSummary, err := h.service.BackgroundJobsSummary(c.Context())
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to load jobs summary")
	}

	distOut := make([]fiber.Map, 0, len(distribution))
	for _, d := range distribution {
		usedPct := 0.0
		if d.QuotaBytes != nil && *d.QuotaBytes > 0 {
			usedPct = float64(d.UsedBytes) / float64(*d.QuotaBytes) * 100
		}
		distOut = append(distOut, fiber.Map{
			"id":          d.ID,
			"label":       d.Label,
			"used_bytes":  d.UsedBytes,
			"quota_bytes": d.QuotaBytes,
			"used_pct":    usedPct,
		})
	}

	return response.OK(c, fiber.StatusOK, fiber.Map{
		"total_songs":          stats.TotalSongs,
		"total_users":          stats.TotalUsers,
		"total_drives":         stats.TotalDrives,
		"total_storage_bytes":  stats.TotalStorageBytes,
		"storage_distribution": distOut,
		"background_jobs":      jobsSummary,
	})
}
