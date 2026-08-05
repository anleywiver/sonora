package middleware

import (
	"github.com/gofiber/fiber/v2"

	"sonora.dev/go-core/application/appsettings"

	"sonora.dev/backend/internal/http/response"
)

// MaintenanceGate blocks non-Owner traffic while maintenance_mode is on
// (Sprint 14 sisipan, ADR 0011/0012). Must be registered AFTER
// RequireAuth wherever it's used (reads UserRole from locals) — added
// to every authenticated non-admin route in main.go, never to
// adminGroup (Owner needs to keep working, including flipping the
// toggle back off) and never to /auth/* (must keep working so login/
// refresh/logout aren't themselves blocked by the switch).
func MaintenanceGate(settings *appsettings.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if UserRole(c) == "owner" {
			return c.Next()
		}
		values, err := settings.List(c.Context())
		if err == nil && values[appsettings.KeyMaintenanceMode] == "true" {
			return response.Fail(c, fiber.StatusServiceUnavailable, "maintenance_mode", "Sonora sedang dalam maintenance, coba lagi nanti")
		}
		return c.Next()
	}
}
