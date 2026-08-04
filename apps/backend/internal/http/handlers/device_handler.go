package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	appauth "sonora.dev/go-core/application/auth"
	"sonora.dev/go-core/domain/identity"

	"sonora.dev/backend/internal/http/middleware"
	"sonora.dev/backend/internal/http/response"
)

type DeviceHandler struct {
	service *appauth.Service
}

func NewDeviceHandler(service *appauth.Service) *DeviceHandler {
	return &DeviceHandler{service: service}
}

func (h *DeviceHandler) List(c *fiber.Ctx) error {
	devices, err := h.service.ListDevices(c.Context(), middleware.UserID(c))
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to load devices")
	}

	out := make([]fiber.Map, 0, len(devices))
	for _, d := range devices {
		out = append(out, fiber.Map{
			"id":           d.ID,
			"name":         d.Name,
			"type":         d.Type,
			"is_active":    d.IsActive,
			"last_seen_at": d.LastSeenAt,
			"created_at":   d.CreatedAt,
		})
	}
	return response.OK(c, fiber.StatusOK, out)
}

func (h *DeviceHandler) Delete(c *fiber.Ctx) error {
	deviceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid device id")
	}

	if err := h.service.RemoveDevice(c.Context(), deviceID, middleware.UserID(c)); err != nil {
		if errors.Is(err, identity.ErrNotFound) {
			return response.Fail(c, fiber.StatusNotFound, "not_found", "device not found")
		}
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to remove device")
	}
	return response.OK(c, fiber.StatusOK, fiber.Map{})
}
