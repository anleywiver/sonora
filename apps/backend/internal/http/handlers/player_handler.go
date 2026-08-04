package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	appplayback "sonora.dev/go-core/application/playback"

	"sonora.dev/backend/internal/http/middleware"
	"sonora.dev/backend/internal/http/response"
	"sonora.dev/backend/internal/ws"
)

// PlayerHandler is Sprint 7's minimal read/write for playback_states — no
// Active Device authority check yet (any of the user's devices can write
// state). Sprint 8 adds that check plus the granular /player/play,pause,
// seek,next,previous,transfer endpoints from docs/api-design.md; this
// state sync is what those will call internally.
type PlayerHandler struct {
	service *appplayback.Service
	hub     *ws.Hub
}

func NewPlayerHandler(service *appplayback.Service, hub *ws.Hub) *PlayerHandler {
	return &PlayerHandler{service: service, hub: hub}
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

func stateJSON(s *appplayback.State) fiber.Map {
	return fiber.Map{
		"active_device_id": s.ActiveDeviceID,
		"current_song_id":  s.CurrentSongID,
		"position_ms":      s.PositionMs,
		"is_playing":       s.IsPlaying,
	}
}
