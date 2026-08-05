package handlers

import (
	"github.com/gofiber/fiber/v2"

	"sonora.dev/go-core/application/appsettings"

	"sonora.dev/backend/internal/http/response"
)

// AdminSettingsHandler backs GET/PATCH /admin/settings (Sprint 14
// sisipan, ADR 0012) — Owner only (enforced by the route's RequireRole
// middleware). Same generic key-value store as the Google OAuth toggle.
type AdminSettingsHandler struct {
	service *appsettings.Service
}

func NewAdminSettingsHandler(service *appsettings.Service) *AdminSettingsHandler {
	return &AdminSettingsHandler{service: service}
}

func (h *AdminSettingsHandler) List(c *fiber.Ctx) error {
	values, err := h.service.List(c.Context())
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to load settings")
	}
	return response.OK(c, fiber.StatusOK, values)
}

type updateSettingRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// allowedSettingKeys guards PATCH against writing arbitrary keys into
// the store — only the four known settings are editable via this
// endpoint.
var allowedSettingKeys = map[string]bool{
	appsettings.KeyGoogleOAuthEnabled: true,
	appsettings.KeyMaintenanceMode:    true,
	appsettings.KeyAppName:            true,
	appsettings.KeyDefaultLanguage:    true,
}

func (h *AdminSettingsHandler) Update(c *fiber.Ctx) error {
	var req updateSettingRequest
	if err := c.BodyParser(&req); err != nil || req.Key == "" {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "key is required")
	}
	if !allowedSettingKeys[req.Key] {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "unknown setting key")
	}
	if err := h.service.Set(c.Context(), req.Key, req.Value); err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to update setting")
	}
	return response.OK(c, fiber.StatusOK, fiber.Map{"key": req.Key, "value": req.Value})
}
