// Package lyrics fetches and caches lyrics for a song. LRCLIB is checked
// (and the result cached in the `lyrics` table) only on a cache miss —
// repeat requests for the same song never hit the network.
package lyrics

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	appcatalog "sonora.dev/go-core/application/catalog"
	infralyrics "sonora.dev/go-core/infrastructure/lyrics"
	"sonora.dev/go-core/infrastructure/postgres/sqlc"
)

var ErrNotFound = errors.New("lyrics: not available for this song")

type Lyrics struct {
	SyncedContent string
	PlainContent  string
}

type Service struct {
	q       *sqlc.Queries
	catalog *appcatalog.Service
	client  *infralyrics.LRCLIBClient
}

func NewService(q *sqlc.Queries, catalog *appcatalog.Service, client *infralyrics.LRCLIBClient) *Service {
	return &Service{q: q, catalog: catalog, client: client}
}

func (s *Service) GetLyrics(ctx context.Context, songID uuid.UUID) (*Lyrics, error) {
	cached, err := s.q.GetLyricsBySongID(ctx, toPgUUID(songID))
	if err == nil {
		return &Lyrics{
			SyncedContent: strOrEmpty(cached.SyncedContent),
			PlainContent:  strOrEmpty(cached.PlainContent),
		}, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("lyrics: check cache: %w", err)
	}

	song, err := s.catalog.GetSong(ctx, songID)
	if err != nil {
		return nil, err
	}

	providerID, err := s.findOrCreateProvider(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.client.Fetch(ctx, song.Title, song.ArtistName, song.DurationMs/1000)
	if err != nil {
		// Recorded either way (Sprint 14, docs/screens-spec.md #19's "Match
		// Rate" needs misses counted too, not just hits) — a miss previously
		// left no trace in the DB at all.
		if recErr := s.q.RecordLyricsLookup(ctx, sqlc.RecordLyricsLookupParams{ID: providerID, Matched: false}); recErr != nil {
			log.Printf("lyrics: record failed lookup: %v", recErr)
		}
		if errors.Is(err, infralyrics.ErrRateLimited) {
			if healthErr := s.q.SetLyricsProviderHealth(ctx, sqlc.SetLyricsProviderHealthParams{ID: providerID, HealthStatus: "rate_limited"}); healthErr != nil {
				log.Printf("lyrics: record rate limit health: %v", healthErr)
			}
		}
		if errors.Is(err, infralyrics.ErrNotFound) || errors.Is(err, infralyrics.ErrRateLimited) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("lyrics: fetch: %w", err)
	}
	if err := s.q.RecordLyricsLookup(ctx, sqlc.RecordLyricsLookupParams{ID: providerID, Matched: true}); err != nil {
		log.Printf("lyrics: record successful lookup: %v", err)
	}
	// A successful call proves the provider recovered — clear a stale
	// rate_limited health state rather than leaving it stuck forever.
	if err := s.q.SetLyricsProviderHealth(ctx, sqlc.SetLyricsProviderHealthParams{ID: providerID, HealthStatus: "online"}); err != nil {
		log.Printf("lyrics: reset health to online: %v", err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("lyrics: generate id: %w", err)
	}
	var synced, plain *string
	if result.SyncedLyrics != "" {
		synced = &result.SyncedLyrics
	}
	if result.PlainLyrics != "" {
		plain = &result.PlainLyrics
	}
	if _, err := s.q.CreateLyrics(ctx, sqlc.CreateLyricsParams{
		ID:            toPgUUID(id),
		SongID:        toPgUUID(songID),
		ProviderID:    providerID,
		SyncedContent: synced,
		PlainContent:  plain,
	}); err != nil {
		return nil, fmt.Errorf("lyrics: cache: %w", err)
	}

	return &Lyrics{SyncedContent: result.SyncedLyrics, PlainContent: result.PlainLyrics}, nil
}

type Provider struct {
	ID                uuid.UUID
	Name              string
	BaseURL           string
	IsEnabled         bool
	Priority          int
	HealthStatus      string
	TotalLookups      int64
	SuccessfulMatches int64
}

// MatchRate returns the successful-match ratio as a percentage, or -1 if
// there have been no lookups yet (avoids a 0/0 division reading as "0%
// match rate" when the real answer is "no data yet").
func (p Provider) MatchRate() float64 {
	if p.TotalLookups == 0 {
		return -1
	}
	return float64(p.SuccessfulMatches) / float64(p.TotalLookups) * 100
}

// ListProviders backs the admin Lyrics Source page (Sprint 14,
// docs/screens-spec.md #19).
func (s *Service) ListProviders(ctx context.Context) ([]Provider, error) {
	rows, err := s.q.ListLyricsProviders(ctx)
	if err != nil {
		return nil, fmt.Errorf("lyrics: list providers: %w", err)
	}
	out := make([]Provider, 0, len(rows))
	for _, row := range rows {
		out = append(out, Provider{
			ID:                uuid.UUID(row.ID.Bytes),
			Name:              row.Name,
			BaseURL:           row.BaseUrl,
			IsEnabled:         row.IsEnabled,
			Priority:          int(row.Priority),
			HealthStatus:      row.HealthStatus,
			TotalLookups:      row.TotalLookups,
			SuccessfulMatches: row.SuccessfulMatches,
		})
	}
	return out, nil
}

func (s *Service) SetProviderPriority(ctx context.Context, id uuid.UUID, priority int) error {
	return s.q.UpdateLyricsProviderPriority(ctx, sqlc.UpdateLyricsProviderPriorityParams{
		ID:       toPgUUID(id),
		Priority: int32(priority),
	})
}

func (s *Service) SetProviderEnabled(ctx context.Context, id uuid.UUID, enabled bool) error {
	return s.q.SetLyricsProviderEnabled(ctx, sqlc.SetLyricsProviderEnabledParams{
		ID:        toPgUUID(id),
		IsEnabled: enabled,
	})
}

func (s *Service) findOrCreateProvider(ctx context.Context) (pgtype.UUID, error) {
	provider, err := s.q.GetLyricsProviderByName(ctx, "lrclib")
	if err == nil {
		return provider.ID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return pgtype.UUID{}, fmt.Errorf("lyrics: lookup provider: %w", err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("lyrics: generate provider id: %w", err)
	}
	created, err := s.q.CreateLyricsProvider(ctx, sqlc.CreateLyricsProviderParams{
		ID:      toPgUUID(id),
		Name:    "lrclib",
		BaseUrl: "https://lrclib.net",
	})
	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("lyrics: create provider: %w", err)
	}
	return created.ID, nil
}
