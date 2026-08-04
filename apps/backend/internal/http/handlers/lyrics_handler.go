package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	applyrics "sonora.dev/go-core/application/lyrics"

	"sonora.dev/backend/internal/http/response"
)

type LyricsHandler struct {
	service *applyrics.Service
}

func NewLyricsHandler(service *applyrics.Service) *LyricsHandler {
	return &LyricsHandler{service: service}
}

func (h *LyricsHandler) Get(c *fiber.Ctx) error {
	songID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid song id")
	}
	lyrics, err := h.service.GetLyrics(c.Context(), songID)
	if err != nil {
		if errors.Is(err, applyrics.ErrNotFound) {
			return response.Fail(c, fiber.StatusNotFound, "not_found", "lyrics not available for this song")
		}
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to load lyrics")
	}
	return response.OK(c, fiber.StatusOK, fiber.Map{
		"synced_content": lyrics.SyncedContent,
		"plain_content":  lyrics.PlainContent,
	})
}
