package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	appauth "sonora.dev/go-core/application/auth"
	appplayback "sonora.dev/go-core/application/playback"
	"sonora.dev/go-core/domain/identity"

	"sonora.dev/backend/internal/http/middleware"
	"sonora.dev/backend/internal/http/response"
	"sonora.dev/backend/internal/ws"
)

// PlayerHandler covers playback_states read/write plus Transfer Playback
// (Sprint 8). Any device can still call UpdateState directly (no "only the
// Active Device may push state" enforcement) — the real authority
// boundary in this app is client-side: a non-active device's UI sends
// player:command over WS instead of calling this endpoint (see
// ws_handler.go's relay), so in practice only the active device calls
// this. Enforcing it server-side too is future hardening, not required
// for the personal/family scale this targets.
type PlayerHandler struct {
	service *appplayback.Service
	auth    *appauth.Service
	hub     *ws.Hub
}

func NewPlayerHandler(service *appplayback.Service, auth *appauth.Service, hub *ws.Hub) *PlayerHandler {
	return &PlayerHandler{service: service, auth: auth, hub: hub}
}

func (h *PlayerHandler) GetState(c *fiber.Ctx) error {
	state, err := h.service.GetState(c.Context(), middleware.UserID(c))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return response.OK(c, fiber.StatusOK, fiber.Map{})
		}
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to load playback state")
	}
	return response.OK(c, fiber.StatusOK, stateJSON(state))
}

type updateStateRequest struct {
	ActiveDeviceID *string `json:"active_device_id"`
	CurrentSongID  *string `json:"current_song_id"`
	PositionMs     int     `json:"position_ms"`
	IsPlaying      bool    `json:"is_playing"`
}

func (h *PlayerHandler) UpdateState(c *fiber.Ctx) error {
	var req updateStateRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid request body")
	}

	var activeDeviceID, currentSongID *uuid.UUID
	if req.ActiveDeviceID != nil {
		id, err := uuid.Parse(*req.ActiveDeviceID)
		if err != nil {
			return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid active_device_id")
		}
		activeDeviceID = &id
	}
	if req.CurrentSongID != nil {
		id, err := uuid.Parse(*req.CurrentSongID)
		if err != nil {
			return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid current_song_id")
		}
		currentSongID = &id
	}

	userID := middleware.UserID(c)
	state, err := h.service.UpsertState(c.Context(), userID, activeDeviceID, currentSongID, req.PositionMs, req.IsPlaying)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to update playback state")
	}

	h.hub.Broadcast(userID, fiber.Map{"type": "player:state", "data": stateJSON(state)})
	return response.OK(c, fiber.StatusOK, stateJSON(state))
}

type transferRequest struct {
	DeviceID string `json:"device_id"`
}

// Transfer hands off Active Device status to another of the user's own
// devices (Spotify Connect-style "Transfer Playback"). Position/song
// carry over as-is; the newly active device is expected to start
// playing locally once it sees itself become active via the broadcast.
func (h *PlayerHandler) Transfer(c *fiber.Ctx) error {
	var req transferRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid request body")
	}
	deviceID, err := uuid.Parse(req.DeviceID)
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid device_id")
	}
	userID := middleware.UserID(c)

	if err := h.auth.SetActiveDevice(c.Context(), userID, deviceID); err != nil {
		if errors.Is(err, identity.ErrNotFound) {
			return response.Fail(c, fiber.StatusNotFound, "not_found", "device not found")
		}
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to transfer playback")
	}

	current, err := h.service.GetState(c.Context(), userID)
	positionMs, currentSongID, isPlaying := 0, (*uuid.UUID)(nil), false
	if err == nil {
		positionMs, currentSongID, isPlaying = current.PositionMs, current.CurrentSongID, current.IsPlaying
	}

	state, err := h.service.UpsertState(c.Context(), userID, &deviceID, currentSongID, positionMs, isPlaying)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to update playback state")
	}

	h.hub.Broadcast(userID, fiber.Map{"type": "player:state", "data": stateJSON(state)})
	return response.OK(c, fiber.StatusOK, stateJSON(state))
}

func stateJSON(s *appplayback.State) fiber.Map {
	return fiber.Map{
		"active_device_id": s.ActiveDeviceID,
		"current_song_id":  s.CurrentSongID,
		"position_ms":      s.PositionMs,
		"is_playing":       s.IsPlaying,
	}
}
