package ingest

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"sonora.dev/go-core/infrastructure/postgres/sqlc"
)

func (s *Service) findOrCreateArtist(ctx context.Context, name string) (sqlc.Artist, error) {
	artist, err := s.q.GetArtistByName(ctx, name)
	if err == nil {
		return artist, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return sqlc.Artist{}, fmt.Errorf("ingest: lookup artist: %w", err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return sqlc.Artist{}, fmt.Errorf("ingest: generate artist id: %w", err)
	}
	artist, err = s.q.CreateArtist(ctx, sqlc.CreateArtistParams{ID: toPgUUID(id), Name: name})
	if err != nil {
		return sqlc.Artist{}, fmt.Errorf("ingest: create artist: %w", err)
	}
	return artist, nil
}

func (s *Service) findOrCreateAlbum(ctx context.Context, artistID pgtype.UUID, title string) (sqlc.Album, error) {
	album, err := s.q.GetAlbumByArtistAndTitle(ctx, sqlc.GetAlbumByArtistAndTitleParams{ArtistID: artistID, Title: title})
	if err == nil {
		return album, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return sqlc.Album{}, fmt.Errorf("ingest: lookup album: %w", err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return sqlc.Album{}, fmt.Errorf("ingest: generate album id: %w", err)
	}
	album, err = s.q.CreateAlbum(ctx, sqlc.CreateAlbumParams{ID: toPgUUID(id), ArtistID: artistID, Title: title})
	if err != nil {
		return sqlc.Album{}, fmt.Errorf("ingest: create album: %w", err)
	}
	return album, nil
}
