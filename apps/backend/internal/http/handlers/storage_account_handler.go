package handlers

import (
	"github.com/gofiber/fiber/v2"

	appstorage "sonora.dev/go-core/application/storageaccount"

	"sonora.dev/backend/internal/http/response"
)

// StorageAccountHandler is the Sprint 3 minimal bootstrap (create + list
// only) — see ADR 0002. Full Drive Manager (health check, quota routing,
// in-app OAuth consent) is Sprint 9.
type StorageAccountHandler struct {
	service *appstorage.Service
}

func NewStorageAccountHandler(service *appstorage.Service) *StorageAccountHandler {
	return &StorageAccountHandler{service: service}
}

type createStorageAccountRequest struct {
	Label        string `json:"label"`
	AccountEmail string `json:"account_email"`
	RefreshToken string `json:"refresh_token"`
	QuotaBytes   *int64 `json:"quota_bytes"`
}

func (h *StorageAccountHandler) Create(c *fiber.Ctx) error {
	var req createStorageAccountRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid request body")
	}
	if req.Label == "" || req.RefreshToken == "" {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "label and refresh_token are required")
	}

	account, err := h.service.Create(c.Context(), req.Label, req.AccountEmail, req.RefreshToken, req.QuotaBytes)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to create storage account")
	}
	return response.OK(c, fiber.StatusCreated, accountJSON(account))
}

func (h *StorageAccountHandler) List(c *fiber.Ctx) error {
	accounts, err := h.service.List(c.Context())
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to list storage accounts")
	}
	out := make([]fiber.Map, 0, len(accounts))
	for _, a := range accounts {
		out = append(out, accountJSON(a))
	}
	return response.OK(c, fiber.StatusOK, out)
}

func accountJSON(a *appstorage.Account) fiber.Map {
	return fiber.Map{
		"id":            a.ID,
		"provider":      a.Provider,
		"label":         a.Label,
		"account_email": a.AccountEmail,
		"quota_bytes":   a.QuotaBytes,
		"used_bytes":    a.UsedBytes,
		"is_active":     a.IsActive,
		"health_status": a.HealthStatus,
	}
}
