// Package musicbrainz enriches ingested songs with canonical MusicBrainz
// IDs and, when available, an album cover from the Cover Art Archive.
// Both are public APIs needing no credential — the only real-world
// constraint is MusicBrainz's usage policy: 1 request/second and a
// descriptive User-Agent, both enforced here. See
// docs/decisions/0005-sprint11-waveform-musicbrainz-analytics.md.
package musicbrainz

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/time/rate"
)

const (
	searchURL     = "https://musicbrainz.org/ws/2/recording/"
	coverArtURL   = "https://coverartarchive.org/release/%s/front-250"
	userAgent     = "Sonora/1.0 (personal music streaming platform, self-hosted)"
	durationSlack = 2 * time.Second
)

type Match struct {
	RecordingMBID string
	ArtistMBID    string
	ReleaseMBID   string
	// CoverURL is set only when Cover Art Archive confirmed an image
	// exists for ReleaseMBID (a HEAD request, not just a guessed URL).
	CoverURL string
}

type Client struct {
	httpClient *http.Client
	limiter    *rate.Limiter
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		// MusicBrainz usage policy: max 1 request/second per IP.
		limiter: rate.NewLimiter(rate.Every(time.Second), 1),
	}
}

type searchResponse struct {
	Recordings []struct {
		ID           string `json:"id"`
		Title        string `json:"title"`
		Length       int    `json:"length"`
		ArtistCredit []struct {
			Artist struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"artist"`
		} `json:"artist-credit"`
		Releases []struct {
			ID string `json:"id"`
		} `json:"releases"`
	} `json:"recordings"`
}

// FindRecording searches for a recording by title/artist and returns the
// best match whose duration is within durationSlack of durationMs. Returns
// (nil, nil) — not an error — when nothing close enough is found, since a
// miss is an expected, non-fatal outcome (ADR 0005).
func (c *Client) FindRecording(ctx context.Context, title, artist string, durationMs int) (*Match, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("musicbrainz: rate limiter: %w", err)
	}

	query := fmt.Sprintf(`recording:"%s" AND artist:"%s"`, escapeQuery(title), escapeQuery(artist))
	reqURL := searchURL + "?" + url.Values{
		"query": {query},
		"fmt":   {"json"},
		"limit": {"5"},
	}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("musicbrainz: build search request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("musicbrainz: search request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("musicbrainz: search returned %d", resp.StatusCode)
	}

	var parsed searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("musicbrainz: decode search response: %w", err)
	}

	targetMs := float64(durationMs)
	slackMs := float64(durationSlack.Milliseconds())
	for _, rec := range parsed.Recordings {
		if rec.Length == 0 || math.Abs(float64(rec.Length)-targetMs) > slackMs {
			continue
		}
		match := &Match{RecordingMBID: rec.ID}
		if len(rec.ArtistCredit) > 0 {
			match.ArtistMBID = rec.ArtistCredit[0].Artist.ID
		}
		if len(rec.Releases) > 0 {
			match.ReleaseMBID = rec.Releases[0].ID
			if coverURL, ok := c.checkCoverArt(ctx, rec.Releases[0].ID); ok {
				match.CoverURL = coverURL
			}
		}
		return match, nil
	}
	return nil, nil
}

// checkCoverArt HEAD-checks whether the Cover Art Archive actually has an
// image for releaseMBID before handing back a URL — a guessed URL that
// 404s would otherwise get stored as a broken image in albums.cover_url.
func (c *Client) checkCoverArt(ctx context.Context, releaseMBID string) (string, bool) {
	coverURL := fmt.Sprintf(coverArtURL, releaseMBID)
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, coverURL, nil)
	if err != nil {
		return "", false
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()
	return coverURL, resp.StatusCode == http.StatusOK
}

func escapeQuery(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if r == '"' || r == '\\' {
			out = append(out, '\\')
		}
		out = append(out, r)
	}
	return string(out)
}
