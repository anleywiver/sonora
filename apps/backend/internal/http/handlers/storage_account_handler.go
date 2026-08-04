package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	appstorage "sonora.dev/go-core/application/storageaccount"

	"sonora.dev/backend/internal/http/response"
)

// StorageAccountHandler is the Drive Manager admin backend (Sprint 9,
// building on the Sprint 3 minimal bootstrap — see ADR 0002). Registering
// a Drive account still needs a refresh token obtained out-of-band; no
// in-app OAuth consent flow (not worth it at personal/family scale).
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

func (h *StorageAccountHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid storage account id")
	}
	if err := h.service.Delete(c.Context(), id); err != nil {
		if errors.Is(err, appstorage.ErrNotFound) {
			return response.Fail(c, fiber.StatusNotFound, "not_found", "storage account not found")
		}
		if errors.Is(err, appstorage.ErrInUse) {
			return response.Fail(c, fiber.StatusConflict, "conflict", "storage account still has files stored on it")
		}
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to delete storage account")
	}
	return response.OK(c, fiber.StatusOK, fiber.Map{})
}

func (h *StorageAccountHandler) HealthCheck(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid storage account id")
	}
	account, err := h.service.HealthCheck(c.Context(), id)
	if err != nil {
		if errors.Is(err, appstorage.ErrNotFound) {
			return response.Fail(c, fiber.StatusNotFound, "not_found", "storage account not found")
		}
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to run health check")
	}
	return response.OK(c, fiber.StatusOK, accountJSON(account))
}

func accountJSON(a *appstorage.Account) fiber.Map {
	return fiber.Map{
		"id":                   a.ID,
		"provider":             a.Provider,
		"label":                a.Label,
		"account_email":        a.AccountEmail,
		"quota_bytes":          a.QuotaBytes,
		"used_bytes":           a.UsedBytes,
		"is_active":            a.IsActive,
		"health_status":        a.HealthStatus,
		"last_health_check_at": a.LastHealthCheckAt,
	}
}
