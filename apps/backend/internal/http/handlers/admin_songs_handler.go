package handlers

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	appadminsongs "sonora.dev/go-core/application/adminsongs"

	"sonora.dev/backend/internal/http/response"
)

// AdminSongsHandler backs the admin Manage Songs page (Sprint 14
// sisipan, ADR 0010).
type AdminSongsHandler struct {
	service *appadminsongs.Service
}

func NewAdminSongsHandler(service *appadminsongs.Service) *AdminSongsHandler {
	return &AdminSongsHandler{service: service}
}

func (h *AdminSongsHandler) List(c *fiber.Ctx) error {
	search := c.Query("search")
	cursor := c.Query("cursor")
	limit := 20
	if raw := c.Query("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}

	songs, nextCursor, hasMore, err := h.service.List(c.Context(), search, cursor, int32(limit))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid cursor")
	}
	out := make([]fiber.Map, 0, len(songs))
	for _, s := range songs {
		out = append(out, fiber.Map{
			"id": s.ID, "title": s.Title, "artist_name": s.ArtistName, "album_title": s.AlbumTitle,
			"duration_ms": s.DurationMs, "storage_provider": s.StorageProvider,
			"created_at": s.CreatedAt.Format("2006-01-02"),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": out, "next_cursor": nextCursor, "has_more": hasMore})
}

type updateSongRequest struct {
	Title      *string `json:"title"`
	ArtistName *string `json:"artist_name"`
	AlbumTitle *string `json:"album_title"`
	GenreName  *string `json:"genre_name"`
}

func (h *AdminSongsHandler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid song id")
	}
	var req updateSongRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid request body")
	}
	err = h.service.Update(c.Context(), id, appadminsongs.UpdateInput{
		Title: req.Title, ArtistName: req.ArtistName, AlbumTitle: req.AlbumTitle, GenreName: req.GenreName,
	})
	if err != nil {
		if errors.Is(err, appadminsongs.ErrNotFound) {
			return response.Fail(c, fiber.StatusNotFound, "not_found", "song not found")
		}
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to update song")
	}
	return response.OK(c, fiber.StatusOK, fiber.Map{})
}

func (h *AdminSongsHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid song id")
	}
	if err := h.service.Delete(c.Context(), id); err != nil {
		if errors.Is(err, appadminsongs.ErrNotFound) {
			return response.Fail(c, fiber.StatusNotFound, "not_found", "song not found")
		}
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to delete song")
	}
	return response.OK(c, fiber.StatusOK, fiber.Map{})
}
