// Package search implements the /search endpoints backed by Meilisearch.
package search

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"sonora.dev/go-core/infrastructure/meilisearch"
)

type Service struct {
	client *meilisearch.Client
}

func NewService(client *meilisearch.Client) *Service {
	return &Service{client: client}
}

type SongResult struct {
	ID         uuid.UUID
	Title      string
	ArtistID   uuid.UUID
	ArtistName string
	AlbumID    *uuid.UUID
	AlbumTitle string
	DurationMs int
}

func (s *Service) SearchSongs(ctx context.Context, query string, limit int64) ([]*SongResult, error) {
	res, err := s.client.SearchSongs(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	out := make([]*SongResult, 0, len(res.Hits))
	for _, hit := range res.Hits {
		id, err := uuid.Parse(hit.ID)
		if err != nil {
			continue
		}
		artistID, err := uuid.Parse(hit.ArtistID)
		if err != nil {
			continue
		}
		var albumID *uuid.UUID
		if hit.AlbumID != "" {
			if parsed, err := uuid.Parse(hit.AlbumID); err == nil {
				albumID = &parsed
			}
		}
		out = append(out, &SongResult{
			ID:         id,
			Title:      hit.Title,
			ArtistID:   artistID,
			ArtistName: hit.ArtistName,
			AlbumID:    albumID,
			AlbumTitle: hit.AlbumTitle,
			DurationMs: hit.DurationMs,
		})
	}
	return out, nil
}
