package handlers

import (
	"errors"
	"io"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	appcatalog "sonora.dev/go-core/application/catalog"

	"sonora.dev/backend/internal/http/middleware"
	"sonora.dev/backend/internal/http/response"
)

type CatalogHandler struct {
	service *appcatalog.Service
}

func NewCatalogHandler(service *appcatalog.Service) *CatalogHandler {
	return &CatalogHandler{service: service}
}

func (h *CatalogHandler) GetSong(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid song id")
	}
	song, err := h.service.GetSong(c.Context(), id)
	if err != nil {
		return h.catalogError(c, err)
	}
	return response.OK(c, fiber.StatusOK, songJSON(song))
}

func (h *CatalogHandler) GetAlbum(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid album id")
	}
	album, err := h.service.GetAlbum(c.Context(), id)
	if err != nil {
		return h.catalogError(c, err)
	}
	songs, err := h.service.ListSongsByAlbum(c.Context(), id)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to load album songs")
	}
	out := albumJSON(album)
	out["songs"] = songsJSON(songs)
	return response.OK(c, fiber.StatusOK, out)
}

func (h *CatalogHandler) GetArtist(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid artist id")
	}
	artist, err := h.service.GetArtist(c.Context(), id)
	if err != nil {
		return h.catalogError(c, err)
	}
	return response.OK(c, fiber.StatusOK, fiber.Map{
		"id":        artist.ID,
		"name":      artist.Name,
		"image_url": artist.ImageURL,
	})
}

func (h *CatalogHandler) ListArtistAlbums(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid artist id")
	}
	albums, err := h.service.ListAlbumsByArtist(c.Context(), id)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to load artist albums")
	}
	out := make([]fiber.Map, 0, len(albums))
	for _, a := range albums {
		out = append(out, fiber.Map{"id": a.ID, "title": a.Title, "cover_url": a.CoverURL})
	}
	return response.OK(c, fiber.StatusOK, out)
}

func (h *CatalogHandler) ListArtistSongs(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid artist id")
	}
	songs, err := h.service.ListSongsByArtist(c.Context(), id)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to load artist songs")
	}
	return response.OK(c, fiber.StatusOK, songsJSON(songs))
}

// ListLibrarySongs/Albums/Artists back the Sprint 14 sisipan Browse
// Library page (ADR 0011) — ?search= and ?sort=alpha (default: recent).
func (h *CatalogHandler) ListLibrarySongs(c *fiber.Ctx) error {
	songs, err := h.service.ListLibrarySongs(c.Context(), c.Query("search"), c.Query("sort") == "alpha")
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to list songs")
	}
	out := make([]fiber.Map, 0, len(songs))
	for _, s := range songs {
		out = append(out, fiber.Map{
			"id": s.ID, "title": s.Title, "duration_ms": s.DurationMs,
			"artist_name": s.ArtistName, "album_title": s.AlbumTitle,
		})
	}
	return response.OK(c, fiber.StatusOK, out)
}

func (h *CatalogHandler) ListLibraryAlbums(c *fiber.Ctx) error {
	albums, err := h.service.ListLibraryAlbums(c.Context(), c.Query("search"), c.Query("sort") == "alpha")
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to list albums")
	}
	out := make([]fiber.Map, 0, len(albums))
	for _, a := range albums {
		out = append(out, fiber.Map{"id": a.ID, "title": a.Title, "cover_url": a.CoverURL, "artist_name": a.ArtistName})
	}
	return response.OK(c, fiber.StatusOK, out)
}

func (h *CatalogHandler) ListLibraryArtists(c *fiber.Ctx) error {
	artists, err := h.service.ListLibraryArtists(c.Context(), c.Query("search"), c.Query("sort") == "alpha")
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to list artists")
	}
	out := make([]fiber.Map, 0, len(artists))
	for _, a := range artists {
		out = append(out, fiber.Map{"id": a.ID, "name": a.Name, "image_url": a.ImageURL})
	}
	return response.OK(c, fiber.StatusOK, out)
}

func (h *CatalogHandler) ListGenres(c *fiber.Ctx) error {
	genres, err := h.service.ListGenres(c.Context())
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to load genres")
	}
	out := make([]fiber.Map, 0, len(genres))
	for _, g := range genres {
		out = append(out, fiber.Map{"id": g.ID, "name": g.Name})
	}
	return response.OK(c, fiber.StatusOK, out)
}

// StreamToken issues a 5-minute token scoped to this song, since the
// browser's <audio> element can't send a custom Authorization header.
func (h *CatalogHandler) StreamToken(c *fiber.Ctx) error {
	songID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid song id")
	}
	token, expiresAt, err := h.service.IssueStreamToken(c.Context(), songID, middleware.UserID(c))
	if err != nil {
		return h.catalogError(c, err)
	}
	return response.OK(c, fiber.StatusOK, fiber.Map{
		"token":      token,
		"expires_in": int(time.Until(expiresAt).Seconds()),
	})
}

// Stream proxies the song's bytes from the storage provider, forwarding
// the client's Range header so <audio> can seek. Auth is the query-param
// stream token, not the normal Bearer JWT (this route has no RequireAuth).
func (h *CatalogHandler) Stream(c *fiber.Ctx) error {
	songID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid song id")
	}

	if _, err := h.service.ParseStreamToken(c.Query("token"), songID); err != nil {
		return response.Fail(c, fiber.StatusUnauthorized, "unauthenticated", "invalid or expired stream token")
	}

	result, err := h.service.Stream(c.Context(), songID, c.Get(fiber.HeaderRange))
	if err != nil {
		return h.catalogError(c, err)
	}
	defer result.Body.Close()

	c.Set(fiber.HeaderContentType, result.MimeType)
	c.Set(fiber.HeaderAcceptRanges, "bytes")
	if result.Partial {
		c.Set(fiber.HeaderContentRange, result.ContentRange)
		c.Status(fiber.StatusPartialContent)
	} else {
		c.Status(fiber.StatusOK)
	}
	if result.ContentLength > 0 {
		c.Set(fiber.HeaderContentLength, strconv.FormatInt(result.ContentLength, 10))
	}

	_, err = io.Copy(c.Response().BodyWriter(), result.Body)
	return err
}

func (h *CatalogHandler) catalogError(c *fiber.Ctx, err error) error {
	if errors.Is(err, appcatalog.ErrNotFound) {
		return response.Fail(c, fiber.StatusNotFound, "not_found", "not found")
	}
	return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "catalog operation failed")
}

func songJSON(s *appcatalog.SongDetail) fiber.Map {
	return fiber.Map{
		"id":              s.ID,
		"title":           s.Title,
		"duration_ms":     s.DurationMs,
		"track_number":    s.TrackNumber,
		"artist_id":       s.ArtistID,
		"artist_name":     s.ArtistName,
		"album_id":        s.AlbumID,
		"album_title":     s.AlbumTitle,
		"album_cover_url": s.AlbumCoverURL,
	}
}

func songsJSON(songs []*appcatalog.Song) []fiber.Map {
	out := make([]fiber.Map, 0, len(songs))
	for _, s := range songs {
		out = append(out, fiber.Map{
			"id":           s.ID,
			"title":        s.Title,
			"duration_ms":  s.DurationMs,
			"track_number": s.TrackNumber,
		})
	}
	return out
}

func albumJSON(a *appcatalog.AlbumDetail) fiber.Map {
	var releasedAt *string
	if a.ReleasedAt != nil {
		formatted := a.ReleasedAt.Format("2006-01-02")
		releasedAt = &formatted
	}
	return fiber.Map{
		"id":          a.ID,
		"title":       a.Title,
		"cover_url":   a.CoverURL,
		"artist_id":   a.ArtistID,
		"artist_name": a.ArtistName,
		"released_at": releasedAt,
	}
}
