package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	applibrary "sonora.dev/go-core/application/library"
	domainlibrary "sonora.dev/go-core/domain/library"

	"sonora.dev/backend/internal/http/middleware"
	"sonora.dev/backend/internal/http/response"
)

type FavoriteHandler struct {
	service *applibrary.Service
}

func NewFavoriteHandler(service *applibrary.Service) *FavoriteHandler {
	return &FavoriteHandler{service: service}
}

var validFavoritableTypes = map[string]domainlibrary.FavoritableType{
	"song":     domainlibrary.FavoritableSong,
	"album":    domainlibrary.FavoritableAlbum,
	"artist":   domainlibrary.FavoritableArtist,
	"playlist": domainlibrary.FavoritablePlaylist,
}

type favoriteRequest struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

func parseFavoriteRequest(c *fiber.Ctx) (domainlibrary.FavoritableType, uuid.UUID, error) {
	var req favoriteRequest
	if err := c.BodyParser(&req); err != nil {
		return "", uuid.UUID{}, errors.New("invalid request body")
	}
	favType, ok := validFavoritableTypes[req.Type]
	if !ok {
		return "", uuid.UUID{}, errors.New("type must be one of song, album, artist, playlist")
	}
	id, err := uuid.Parse(req.ID)
	if err != nil {
		return "", uuid.UUID{}, errors.New("invalid id")
	}
	return favType, id, nil
}

func (h *FavoriteHandler) Create(c *fiber.Ctx) error {
	favType, id, err := parseFavoriteRequest(c)
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", err.Error())
	}
	f, err := h.service.AddFavorite(c.Context(), middleware.UserID(c), favType, id)
	if err != nil {
		if errors.Is(err, applibrary.ErrConflict) {
			return response.Fail(c, fiber.StatusConflict, "conflict", "already favorited")
		}
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to add favorite")
	}
	return response.OK(c, fiber.StatusCreated, favoriteJSON(f))
}

func (h *FavoriteHandler) Delete(c *fiber.Ctx) error {
	favType, id, err := parseFavoriteRequest(c)
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", err.Error())
	}
	if err := h.service.RemoveFavorite(c.Context(), middleware.UserID(c), favType, id); err != nil {
		if errors.Is(err, applibrary.ErrNotFound) {
			return response.Fail(c, fiber.StatusNotFound, "not_found", "favorite not found")
		}
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to remove favorite")
	}
	return response.OK(c, fiber.StatusOK, fiber.Map{})
}

func (h *FavoriteHandler) List(c *fiber.Ctx) error {
	favType := domainlibrary.FavoritableType(c.Query("type"))
	favorites, err := h.service.ListFavorites(c.Context(), middleware.UserID(c), favType)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to list favorites")
	}
	out := make([]fiber.Map, 0, len(favorites))
	for _, f := range favorites {
		out = append(out, favoriteJSON(f))
	}
	return response.OK(c, fiber.StatusOK, out)
}

func favoriteJSON(f *domainlibrary.Favorite) fiber.Map {
	return fiber.Map{
		"id":         f.ID,
		"type":       f.FavoritableType,
		"target_id":  f.FavoritableID,
		"created_at": f.CreatedAt,
	}
}
