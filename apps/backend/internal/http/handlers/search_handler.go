package handlers

import (
	"github.com/gofiber/fiber/v2"

	appcatalog "sonora.dev/go-core/application/catalog"
	appsearch "sonora.dev/go-core/application/search"

	"sonora.dev/backend/internal/http/response"
)

type SearchHandler struct {
	service *appsearch.Service
	catalog *appcatalog.Service
}

func NewSearchHandler(service *appsearch.Service, catalog *appcatalog.Service) *SearchHandler {
	return &SearchHandler{service: service, catalog: catalog}
}

// Search only covers type=songs for now — artists/lyrics search (the
// other filter-tab options in docs/screens-spec.md's Search Result page)
// need history/lyrics data that lands in later sprints.
func (h *SearchHandler) Search(c *fiber.Ctx) error {
	query := c.Query("q")
	results, err := h.service.SearchSongs(c.Context(), query, 20)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "search failed")
	}
	return response.OK(c, fiber.StatusOK, songResultsJSON(results))
}

func (h *SearchHandler) Autocomplete(c *fiber.Ctx) error {
	query := c.Query("q")
	results, err := h.service.SearchSongs(c.Context(), query, 5)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "search failed")
	}
	return response.OK(c, fiber.StatusOK, songResultsJSON(results))
}

func (h *SearchHandler) Trending(c *fiber.Ctx) error {
	songs, err := h.catalog.ListRecent(c.Context(), 10)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to load trending")
	}
	return response.OK(c, fiber.StatusOK, songsJSON(songs))
}

func songResultsJSON(results []*appsearch.SongResult) []fiber.Map {
	out := make([]fiber.Map, 0, len(results))
	for _, r := range results {
		out = append(out, fiber.Map{
			"id":          r.ID,
			"title":       r.Title,
			"artist_id":   r.ArtistID,
			"artist_name": r.ArtistName,
			"album_id":    r.AlbumID,
			"album_title": r.AlbumTitle,
			"duration_ms": r.DurationMs,
		})
	}
	return out
}
