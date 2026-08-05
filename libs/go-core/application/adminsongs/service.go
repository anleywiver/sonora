// Package adminsongs implements the admin Manage Songs page (Sprint 14
// sisipan, ADR 0010): list+search, metadata edit (find-or-create
// artist/album, same pattern as application/ingest), and soft delete.
// Kept separate from application/catalog (the stable user-facing read
// path) so admin-only mutation logic doesn't mix into it.
package adminsongs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"sonora.dev/go-core/infrastructure/meilisearch"
	"sonora.dev/go-core/infrastructure/postgres/sqlc"
)

var ErrNotFound = errors.New("adminsongs: song not found")

type Song struct {
	ID              uuid.UUID
	Title           string
	ArtistName      string
	AlbumTitle      string
	DurationMs      int
	StorageProvider string
	CreatedAt       time.Time
}

type Service struct {
	q      *sqlc.Queries
	search *meilisearch.Client
}

func NewService(q *sqlc.Queries, search *meilisearch.Client) *Service {
	return &Service{q: q, search: search}
}

func (s *Service) List(ctx context.Context, search, cursor string, limit int32) ([]Song, string, bool, error) {
	params := sqlc.ListAdminSongsParams{LimitCount: limit + 1}
	if search != "" {
		params.Search = &search
	}
	if cursor != "" {
		createdAt, id, err := decodeCursor(cursor)
		if err != nil {
			return nil, "", false, err
		}
		params.CursorCreatedAt = pgtype.Timestamptz{Time: createdAt, Valid: true}
		params.CursorID = toPgUUID(id)
	}

	rows, err := s.q.ListAdminSongs(ctx, params)
	if err != nil {
		return nil, "", false, fmt.Errorf("adminsongs: list: %w", err)
	}

	hasMore := len(rows) > int(limit)
	if hasMore {
		rows = rows[:limit]
	}

	songs := make([]Song, 0, len(rows))
	for _, row := range rows {
		var albumTitle string
		if row.AlbumTitle != nil {
			albumTitle = *row.AlbumTitle
		}
		songs = append(songs, Song{
			ID:              fromPgUUID(row.ID),
			Title:           row.Title,
			ArtistName:      row.ArtistName,
			AlbumTitle:      albumTitle,
			DurationMs:      int(row.DurationMs),
			StorageProvider: row.StorageProvider,
			CreatedAt:       row.CreatedAt.Time,
		})
	}

	nextCursor := ""
	if hasMore && len(songs) > 0 {
		last := songs[len(songs)-1]
		nextCursor = encodeCursor(last.CreatedAt, last.ID)
	}
	return songs, nextCursor, hasMore, nil
}

type UpdateInput struct {
	Title      *string
	ArtistName *string
	AlbumTitle *string
	GenreName  *string
}

// Update edits a song's metadata. artist_name/album_title are resolved
// via find-or-create (same as the ingest pipeline) since songs reference
// artists/albums by ID, not free text. genre_name replaces the song's
// entire genre set with that one genre (the admin UI edits a single
// "genre" field, not a multi-select list).
func (s *Service) Update(ctx context.Context, songID uuid.UUID, in UpdateInput) error {
	song, err := s.q.GetSongByID(ctx, toPgUUID(songID))
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("adminsongs: get song: %w", err)
	}

	if in.Title != nil {
		if err := s.q.UpdateSongTitle(ctx, sqlc.UpdateSongTitleParams{ID: toPgUUID(songID), Title: *in.Title}); err != nil {
			return fmt.Errorf("adminsongs: update title: %w", err)
		}
	}

	if in.ArtistName != nil || in.AlbumTitle != nil {
		artistID := song.ArtistID
		if in.ArtistName != nil {
			artist, err := s.findOrCreateArtist(ctx, *in.ArtistName)
			if err != nil {
				return err
			}
			artistID = artist.ID
		}
		albumID := song.AlbumID
		if in.AlbumTitle != nil {
			album, err := s.findOrCreateAlbum(ctx, artistID, *in.AlbumTitle)
			if err != nil {
				return err
			}
			albumID = album.ID
		}
		if err := s.q.UpdateSongArtistAlbum(ctx, sqlc.UpdateSongArtistAlbumParams{ID: toPgUUID(songID), ArtistID: artistID, AlbumID: albumID}); err != nil {
			return fmt.Errorf("adminsongs: update artist/album: %w", err)
		}
	}

	if in.GenreName != nil {
		genre, err := s.findOrCreateGenre(ctx, *in.GenreName)
		if err != nil {
			return err
		}
		if err := s.q.ClearSongGenres(ctx, toPgUUID(songID)); err != nil {
			return fmt.Errorf("adminsongs: clear genres: %w", err)
		}
		if err := s.q.AddSongGenre(ctx, sqlc.AddSongGenreParams{SongID: toPgUUID(songID), GenreID: genre.ID}); err != nil {
			return fmt.Errorf("adminsongs: add genre: %w", err)
		}
	}

	return nil
}

// Delete soft-deletes the song and removes it from search — see ADR
// 0010 for the deliberate scope limits (playback via an existing
// playlist/favorite link isn't blocked by this).
func (s *Service) Delete(ctx context.Context, songID uuid.UUID) error {
	affected, err := s.q.SoftDeleteSong(ctx, toPgUUID(songID))
	if err != nil {
		return fmt.Errorf("adminsongs: soft delete: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	if err := s.search.DeleteSong(ctx, songID.String()); err != nil {
		return fmt.Errorf("adminsongs: remove from search: %w", err)
	}
	return nil
}

func (s *Service) findOrCreateArtist(ctx context.Context, name string) (sqlc.Artist, error) {
	artist, err := s.q.GetArtistByName(ctx, name)
	if err == nil {
		return artist, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return sqlc.Artist{}, fmt.Errorf("adminsongs: lookup artist: %w", err)
	}
	id, err := uuid.NewV7()
	if err != nil {
		return sqlc.Artist{}, fmt.Errorf("adminsongs: generate artist id: %w", err)
	}
	return s.q.CreateArtist(ctx, sqlc.CreateArtistParams{ID: toPgUUID(id), Name: name})
}

func (s *Service) findOrCreateAlbum(ctx context.Context, artistID pgtype.UUID, title string) (sqlc.Album, error) {
	album, err := s.q.GetAlbumByArtistAndTitle(ctx, sqlc.GetAlbumByArtistAndTitleParams{ArtistID: artistID, Title: title})
	if err == nil {
		return album, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return sqlc.Album{}, fmt.Errorf("adminsongs: lookup album: %w", err)
	}
	id, err := uuid.NewV7()
	if err != nil {
		return sqlc.Album{}, fmt.Errorf("adminsongs: generate album id: %w", err)
	}
	return s.q.CreateAlbum(ctx, sqlc.CreateAlbumParams{ID: toPgUUID(id), ArtistID: artistID, Title: title})
}

func (s *Service) findOrCreateGenre(ctx context.Context, name string) (sqlc.Genre, error) {
	genre, err := s.q.GetGenreByName(ctx, name)
	if err == nil {
		return genre, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return sqlc.Genre{}, fmt.Errorf("adminsongs: lookup genre: %w", err)
	}
	id, err := uuid.NewV7()
	if err != nil {
		return sqlc.Genre{}, fmt.Errorf("adminsongs: generate genre id: %w", err)
	}
	return s.q.CreateGenre(ctx, sqlc.CreateGenreParams{ID: toPgUUID(id), Name: name})
}
