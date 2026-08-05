// Package lyrics fetches lyrics from LRCLIB (https://lrclib.net) — the
// only lyrics provider for now (Sprint 6); docs/api-design.md's
// `/admin/lyrics-providers` (Sprint 11) will let more be configured.
package lyrics

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

var ErrNotFound = fmt.Errorf("lyrics: not found")

// ErrRateLimited distinguishes a 429 from LRCLIB from a genuine "no
// lyrics for this track" 404 — the admin Lyrics Source page (Sprint 14)
// shows this as a distinct "rate_limited" health state, not just another
// miss.
var ErrRateLimited = fmt.Errorf("lyrics: rate limited by lrclib")

type Result struct {
	PlainLyrics  string
	SyncedLyrics string
}

type LRCLIBClient struct {
	httpClient *http.Client
	baseURL    string
}

func NewLRCLIBClient() *LRCLIBClient {
	return &LRCLIBClient{httpClient: http.DefaultClient, baseURL: "https://lrclib.net/api/get"}
}

type lrclibResponse struct {
	PlainLyrics  string `json:"plainLyrics"`
	SyncedLyrics string `json:"syncedLyrics"`
}

// Fetch looks up lyrics by track/artist/duration — LRCLIB uses duration
// (seconds) to disambiguate between different recordings of the same title.
func (c *LRCLIBClient) Fetch(ctx context.Context, trackName, artistName string, durationSeconds int) (*Result, error) {
	q := url.Values{}
	q.Set("track_name", trackName)
	q.Set("artist_name", artistName)
	q.Set("duration", fmt.Sprintf("%d", durationSeconds))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"?"+q.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("lyrics: build request: %w", err)
	}
	req.Header.Set("User-Agent", "Sonora/1.0 (personal music streaming)")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("lyrics: request lrclib: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, ErrRateLimited
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lyrics: lrclib status %d", resp.StatusCode)
	}

	var parsed lrclibResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("lyrics: decode lrclib response: %w", err)
	}
	if parsed.PlainLyrics == "" && parsed.SyncedLyrics == "" {
		return nil, ErrNotFound
	}
	return &Result{PlainLyrics: parsed.PlainLyrics, SyncedLyrics: parsed.SyncedLyrics}, nil
}
