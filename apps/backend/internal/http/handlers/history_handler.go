package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	appcatalog "sonora.dev/go-core/application/catalog"
	apphistory "sonora.dev/go-core/application/history"

	"sonora.dev/backend/internal/http/middleware"
	"sonora.dev/backend/internal/http/response"
)

type HistoryHandler struct {
	service *apphistory.Service
	catalog *appcatalog.Service
}

func NewHistoryHandler(service *apphistory.Service, catalog *appcatalog.Service) *HistoryHandler {
	return &HistoryHandler{service: service, catalog: catalog}
}

type recordHistoryRequest struct {
	SongID     string `json:"song_id"`
	ProgressMs int    `json:"progress_ms"`
}

func (h *HistoryHandler) Create(c *fiber.Ctx) error {
	var req recordHistoryRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid request body")
	}
	songID, err := uuid.Parse(req.SongID)
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid song_id")
	}
	entry, err := h.service.RecordPlay(c.Context(), middleware.UserID(c), songID, req.ProgressMs)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to record history")
	}
	return response.OK(c, fiber.StatusCreated, fiber.Map{
		"id":          entry.ID,
		"song_id":     entry.SongID,
		"progress_ms": entry.ProgressMs,
		"played_at":   entry.PlayedAt,
	})
}

func (h *HistoryHandler) List(c *fiber.Ctx) error {
	cursor := c.Query("cursor")
	limit := 20
	if raw := c.Query("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	entries, nextCursor, hasMore, err := h.service.ListHistory(c.Context(), middleware.UserID(c), cursor, int32(limit))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid cursor")
	}
	out := make([]fiber.Map, 0, len(entries))
	for _, e := range entries {
		out = append(out, fiber.Map{
			"id":          e.ID,
			"song_id":     e.SongID,
			"progress_ms": e.ProgressMs,
			"played_at":   e.PlayedAt,
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":        out,
		"next_cursor": nextCursor,
		"has_more":    hasMore,
	})
}

func (h *HistoryHandler) ContinueListening(c *fiber.Ctx) error {
	entries, err := h.service.ListContinueListening(c.Context(), middleware.UserID(c), 10)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to load continue listening")
	}
	out := make([]fiber.Map, 0, len(entries))
	for _, e := range entries {
		song, err := h.catalog.GetSong(c.Context(), e.SongID)
		if err != nil {
			continue // song was deleted from the catalog since being played
		}
		out = append(out, fiber.Map{
			"song_id":     e.SongID,
			"title":       song.Title,
			"artist_name": song.ArtistName,
			"progress_ms": e.ProgressMs,
			"duration_ms": e.DurationMs,
			"played_at":   e.PlayedAt,
		})
	}
	return response.OK(c, fiber.StatusOK, out)
}
