// Package lyrics fetches and caches lyrics for a song. LRCLIB is checked
// (and the result cached in the `lyrics` table) only on a cache miss —
// repeat requests for the same song never hit the network.
package lyrics

import (
	"context"
	"errors"
	"fmt"

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

	result, err := s.client.Fetch(ctx, song.Title, song.ArtistName, song.DurationMs/1000)
	if err != nil {
		if errors.Is(err, infralyrics.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("lyrics: fetch: %w", err)
	}

	providerID, err := s.findOrCreateProvider(ctx)
	if err != nil {
		return nil, err
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
