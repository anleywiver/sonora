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

type PlaylistHandler struct {
	service *applibrary.Service
}

func NewPlaylistHandler(service *applibrary.Service) *PlaylistHandler {
	return &PlaylistHandler{service: service}
}

type playlistRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *PlaylistHandler) Create(c *fiber.Ctx) error {
	var req playlistRequest
	if err := c.BodyParser(&req); err != nil || req.Name == "" {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "name is required")
	}
	p, err := h.service.CreatePlaylist(c.Context(), middleware.UserID(c), req.Name, req.Description)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to create playlist")
	}
	return response.OK(c, fiber.StatusCreated, playlistJSON(p))
}

func (h *PlaylistHandler) List(c *fiber.Ctx) error {
	playlists, err := h.service.ListPlaylists(c.Context(), middleware.UserID(c))
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to list playlists")
	}
	out := make([]fiber.Map, 0, len(playlists))
	for _, p := range playlists {
		out = append(out, playlistJSON(p))
	}
	return response.OK(c, fiber.StatusOK, out)
}

func (h *PlaylistHandler) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid playlist id")
	}
	userID := middleware.UserID(c)

	p, err := h.service.GetPlaylist(c.Context(), id, userID)
	if err != nil {
		return h.playlistError(c, err)
	}
	songs, err := h.service.ListPlaylistSongs(c.Context(), id, userID)
	if err != nil {
		return h.playlistError(c, err)
	}
	out := playlistJSON(p)
	out["songs"] = playlistSongsJSON(songs)
	return response.OK(c, fiber.StatusOK, out)
}

func (h *PlaylistHandler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid playlist id")
	}
	var req playlistRequest
	if err := c.BodyParser(&req); err != nil || req.Name == "" {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "name is required")
	}
	p, err := h.service.UpdatePlaylist(c.Context(), id, middleware.UserID(c), req.Name, req.Description)
	if err != nil {
		return h.playlistError(c, err)
	}
	return response.OK(c, fiber.StatusOK, playlistJSON(p))
}

func (h *PlaylistHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid playlist id")
	}
	if err := h.service.DeletePlaylist(c.Context(), id, middleware.UserID(c)); err != nil {
		return h.playlistError(c, err)
	}
	return response.OK(c, fiber.StatusOK, fiber.Map{})
}

type addSongRequest struct {
	SongID string `json:"song_id"`
}

func (h *PlaylistHandler) AddSong(c *fiber.Ctx) error {
	playlistID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid playlist id")
	}
	var req addSongRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid request body")
	}
	songID, err := uuid.Parse(req.SongID)
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid song_id")
	}

	row, err := h.service.AddSongToPlaylist(c.Context(), playlistID, middleware.UserID(c), songID)
	if err != nil {
		return h.playlistError(c, err)
	}
	return response.OK(c, fiber.StatusCreated, fiber.Map{
		"id":       row.ID,
		"song_id":  row.SongID,
		"position": row.Position,
	})
}

type updatePositionRequest struct {
	Position float64 `json:"position"`
}

func (h *PlaylistHandler) UpdateSongPosition(c *fiber.Ctx) error {
	playlistID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid playlist id")
	}
	rowID, err := uuid.Parse(c.Params("song_row_id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid song row id")
	}
	var req updatePositionRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid request body")
	}
	if err := h.service.UpdateSongPosition(c.Context(), playlistID, middleware.UserID(c), rowID, req.Position); err != nil {
		return h.playlistError(c, err)
	}
	return response.OK(c, fiber.StatusOK, fiber.Map{})
}

func (h *PlaylistHandler) RemoveSong(c *fiber.Ctx) error {
	playlistID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid playlist id")
	}
	rowID, err := uuid.Parse(c.Params("song_row_id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid song row id")
	}
	if err := h.service.RemoveSongFromPlaylist(c.Context(), playlistID, middleware.UserID(c), rowID); err != nil {
		return h.playlistError(c, err)
	}
	return response.OK(c, fiber.StatusOK, fiber.Map{})
}

func (h *PlaylistHandler) playlistError(c *fiber.Ctx, err error) error {
	if errors.Is(err, applibrary.ErrNotFound) {
		return response.Fail(c, fiber.StatusNotFound, "not_found", "playlist not found")
	}
	return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "playlist operation failed")
}

func playlistJSON(p *domainlibrary.Playlist) fiber.Map {
	return fiber.Map{
		"id":          p.ID,
		"name":        p.Name,
		"description": p.Description,
		"cover_url":   p.CoverURL,
		"created_at":  p.CreatedAt,
		"updated_at":  p.UpdatedAt,
	}
}

func playlistSongsJSON(songs []*applibrary.PlaylistSongView) []fiber.Map {
	out := make([]fiber.Map, 0, len(songs))
	for _, s := range songs {
		out = append(out, fiber.Map{
			"row_id":      s.RowID,
			"position":    s.Position,
			"id":          s.Song.ID,
			"title":       s.Song.Title,
			"duration_ms": s.Song.DurationMs,
			"artist_name": s.Song.ArtistName,
			"album_title": s.Song.AlbumTitle,
		})
	}
	return out
}
