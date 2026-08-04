package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	appqueue "sonora.dev/go-core/application/queue"

	"sonora.dev/backend/internal/http/middleware"
	"sonora.dev/backend/internal/http/response"
)

type QueueHandler struct {
	service *appqueue.Service
}

func NewQueueHandler(service *appqueue.Service) *QueueHandler {
	return &QueueHandler{service: service}
}

func (h *QueueHandler) List(c *fiber.Ctx) error {
	items, err := h.service.List(c.Context(), middleware.UserID(c))
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to load queue")
	}
	out := make([]fiber.Map, 0, len(items))
	for _, i := range items {
		out = append(out, queueItemJSON(i))
	}
	return response.OK(c, fiber.StatusOK, out)
}

type addQueueItemRequest struct {
	SongID string `json:"song_id"`
}

func (h *QueueHandler) Add(c *fiber.Ctx) error {
	var req addQueueItemRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid request body")
	}
	songID, err := uuid.Parse(req.SongID)
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid song_id")
	}
	item, err := h.service.Add(c.Context(), middleware.UserID(c), songID)
	if err != nil {
		return h.queueError(c, err)
	}
	return response.OK(c, fiber.StatusCreated, queueItemJSON(item))
}

type updateQueueItemRequest struct {
	Position float64 `json:"position"`
}

func (h *QueueHandler) UpdatePosition(c *fiber.Ctx) error {
	itemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid queue item id")
	}
	var req updateQueueItemRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid request body")
	}
	if err := h.service.UpdatePosition(c.Context(), middleware.UserID(c), itemID, req.Position); err != nil {
		return h.queueError(c, err)
	}
	return response.OK(c, fiber.StatusOK, fiber.Map{})
}

func (h *QueueHandler) Remove(c *fiber.Ctx) error {
	itemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid queue item id")
	}
	if err := h.service.Remove(c.Context(), middleware.UserID(c), itemID); err != nil {
		return h.queueError(c, err)
	}
	return response.OK(c, fiber.StatusOK, fiber.Map{})
}

func (h *QueueHandler) Clear(c *fiber.Ctx) error {
	if err := h.service.Clear(c.Context(), middleware.UserID(c)); err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to clear queue")
	}
	return response.OK(c, fiber.StatusOK, fiber.Map{})
}

func (h *QueueHandler) queueError(c *fiber.Ctx, err error) error {
	if errors.Is(err, appqueue.ErrNotFound) {
		return response.Fail(c, fiber.StatusNotFound, "not_found", "queue item not found")
	}
	return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "queue operation failed")
}

func queueItemJSON(i *appqueue.Item) fiber.Map {
	return fiber.Map{
		"id":          i.ID,
		"position":    i.Position,
		"song_id":     i.Song.ID,
		"title":       i.Song.Title,
		"artist_name": i.Song.ArtistName,
		"duration_ms": i.Song.DurationMs,
	}
}
