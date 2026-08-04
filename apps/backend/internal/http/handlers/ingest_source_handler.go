package handlers

import (
	"encoding/json"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	appingestsource "sonora.dev/go-core/application/ingestsource"
	"sonora.dev/go-core/infrastructure/bandcamp"
	"sonora.dev/go-core/infrastructure/dropbox"

	"sonora.dev/backend/internal/http/response"
)

// IngestSourceHandler is the admin backend for the "Ingest Sources" page
// (Sprint 10, ADR 0004) — Bandcamp/cloud sync connections, Owner-managed
// and global like Drive Manager's storage accounts.
type IngestSourceHandler struct {
	service *appingestsource.Service
}

func NewIngestSourceHandler(service *appingestsource.Service) *IngestSourceHandler {
	return &IngestSourceHandler{service: service}
}

type connectIngestSourceRequest struct {
	Provider     string `json:"provider"`
	Label        string `json:"label"`
	AccountEmail string `json:"account_email"`

	// bandcamp credentials
	IdentityCookie string `json:"identity_cookie"`
	FanID          string `json:"fan_id"`

	// cloud_sync (Dropbox) credentials
	RefreshToken string `json:"refresh_token"`
	AppKey       string `json:"app_key"`
	AppSecret    string `json:"app_secret"`
	FolderPath   string `json:"folder_path"`
}

func (h *IngestSourceHandler) Connect(c *fiber.Ctx) error {
	var req connectIngestSourceRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid request body")
	}
	if req.Label == "" {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "label is required")
	}

	var credentialsJSON []byte
	var err error
	switch req.Provider {
	case "bandcamp":
		if req.IdentityCookie == "" || req.FanID == "" {
			return response.Fail(c, fiber.StatusBadRequest, "validation_error", "identity_cookie and fan_id are required for bandcamp")
		}
		credentialsJSON, err = json.Marshal(bandcamp.Credentials{IdentityCookie: req.IdentityCookie, FanID: req.FanID})
	case "cloud_sync":
		if req.RefreshToken == "" || req.AppKey == "" || req.AppSecret == "" || req.FolderPath == "" {
			return response.Fail(c, fiber.StatusBadRequest, "validation_error", "refresh_token, app_key, app_secret, and folder_path are required for cloud_sync")
		}
		credentialsJSON, err = json.Marshal(dropbox.Credentials{
			RefreshToken: req.RefreshToken,
			AppKey:       req.AppKey,
			AppSecret:    req.AppSecret,
			FolderPath:   req.FolderPath,
		})
	default:
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "provider must be bandcamp or cloud_sync")
	}
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to prepare credentials")
	}

	connection, err := h.service.Connect(c.Context(), req.Provider, req.Label, req.AccountEmail, string(credentialsJSON))
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to create connection")
	}
	return response.OK(c, fiber.StatusCreated, connectionJSON(connection))
}

func (h *IngestSourceHandler) List(c *fiber.Ctx) error {
	connections, err := h.service.List(c.Context())
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to list connections")
	}
	out := make([]fiber.Map, 0, len(connections))
	for _, conn := range connections {
		out = append(out, connectionJSON(conn))
	}
	return response.OK(c, fiber.StatusOK, out)
}

func (h *IngestSourceHandler) Disconnect(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid connection id")
	}
	if err := h.service.Disconnect(c.Context(), id); err != nil {
		if errors.Is(err, appingestsource.ErrNotFound) {
			return response.Fail(c, fiber.StatusNotFound, "not_found", "connection not found")
		}
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to disconnect")
	}
	return response.OK(c, fiber.StatusOK, fiber.Map{})
}

// Sync triggers an immediate sync for one connection — same code path the
// scheduler calls periodically, exposed so the admin doesn't have to wait
// for the next scheduled run to test a freshly connected provider.
func (h *IngestSourceHandler) Sync(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid connection id")
	}
	if err := h.service.Sync(c.Context(), id); err != nil {
		if errors.Is(err, appingestsource.ErrNotFound) {
			return response.Fail(c, fiber.StatusNotFound, "not_found", "connection not found")
		}
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "sync failed: "+err.Error())
	}
	return response.OK(c, fiber.StatusOK, fiber.Map{})
}

func connectionJSON(conn *appingestsource.Connection) fiber.Map {
	return fiber.Map{
		"id":             conn.ID,
		"provider":       conn.Provider,
		"label":          conn.Label,
		"account_email":  conn.AccountEmail,
		"is_active":      conn.IsActive,
		"last_synced_at": conn.LastSyncedAt,
	}
}
