// Package library implements playlist CRUD/reorder and favorites — the
// Sprint 5 "Library" feature. Named after the domain package it wraps
// (domain/library), aliased here to avoid a name clash in this file.
package library

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	appcatalog "sonora.dev/go-core/application/catalog"
	domainlibrary "sonora.dev/go-core/domain/library"
)

var (
	ErrNotFound = domainlibrary.ErrNotFound
	ErrConflict = errors.New("library: already exists")
)

// positionStep spaces new songs 1000 apart, leaving room to insert
// between any two without renumbering the whole playlist.
const positionStep = 1000

type Service struct {
	playlists domainlibrary.PlaylistRepository
	favorites domainlibrary.FavoriteRepository
	catalog   *appcatalog.Service
}

func NewService(playlists domainlibrary.PlaylistRepository, favorites domainlibrary.FavoriteRepository, catalog *appcatalog.Service) *Service {
	return &Service{playlists: playlists, favorites: favorites, catalog: catalog}
}

func (s *Service) CreatePlaylist(ctx context.Context, userID uuid.UUID, name, description string) (*domainlibrary.Playlist, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("library: generate playlist id: %w", err)
	}
	p := &domainlibrary.Playlist{ID: id, UserID: userID, Name: name, Description: description}
	if err := s.playlists.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("library: create playlist: %w", err)
	}
	return p, nil
}

func (s *Service) ListPlaylists(ctx context.Context, userID uuid.UUID) ([]*domainlibrary.Playlist, error) {
	return s.playlists.ListByUser(ctx, userID)
}

func (s *Service) GetPlaylist(ctx context.Context, id, userID uuid.UUID) (*domainlibrary.Playlist, error) {
	p, err := s.playlists.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.UserID != userID {
		return nil, domainlibrary.ErrNotFound
	}
	return p, nil
}

func (s *Service) UpdatePlaylist(ctx context.Context, id, userID uuid.UUID, name, description string) (*domainlibrary.Playlist, error) {
	p, err := s.GetPlaylist(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	p.Name = name
	p.Description = description
	if err := s.playlists.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) DeletePlaylist(ctx context.Context, id, userID uuid.UUID) error {
	return s.playlists.Delete(ctx, id, userID)
}

// PlaylistSongView is a playlist row joined with the song's catalog
// details, for display without a second round-trip per song from the
// frontend.
type PlaylistSongView struct {
	RowID    uuid.UUID
	Position float64
	Song     *appcatalog.SongDetail
}

func (s *Service) ListPlaylistSongs(ctx context.Context, playlistID, userID uuid.UUID) ([]*PlaylistSongView, error) {
	if _, err := s.GetPlaylist(ctx, playlistID, userID); err != nil {
		return nil, err
	}
	rows, err := s.playlists.ListSongs(ctx, playlistID)
	if err != nil {
		return nil, fmt.Errorf("library: list playlist songs: %w", err)
	}
	out := make([]*PlaylistSongView, 0, len(rows))
	for _, row := range rows {
		song, err := s.catalog.GetSong(ctx, row.SongID)
		if err != nil {
			continue // song was deleted from the catalog since being added
		}
		out = append(out, &PlaylistSongView{RowID: row.ID, Position: row.Position, Song: song})
	}
	return out, nil
}

func (s *Service) AddSongToPlaylist(ctx context.Context, playlistID, userID, songID uuid.UUID) (*domainlibrary.PlaylistSong, error) {
	if _, err := s.GetPlaylist(ctx, playlistID, userID); err != nil {
		return nil, err
	}
	if _, err := s.catalog.GetSong(ctx, songID); err != nil {
		return nil, err
	}

	maxPos, err := s.playlists.MaxSongPosition(ctx, playlistID)
	if err != nil {
		return nil, fmt.Errorf("library: max position: %w", err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("library: generate row id: %w", err)
	}
	ps := &domainlibrary.PlaylistSong{ID: id, PlaylistID: playlistID, SongID: songID, Position: maxPos + positionStep}
	if err := s.playlists.AddSong(ctx, ps); err != nil {
		return nil, fmt.Errorf("library: add song: %w", err)
	}
	return ps, nil
}

// UpdateSongPosition moves a track within the playlist. The caller
// computes the target position client-side (typically the midpoint of
// the two rows it's being dropped between), per the fractional-position
// scheme (ADR: no full-playlist renumbering on reorder).
func (s *Service) UpdateSongPosition(ctx context.Context, playlistID, userID, rowID uuid.UUID, position float64) error {
	if _, err := s.GetPlaylist(ctx, playlistID, userID); err != nil {
		return err
	}
	row, err := s.playlists.FindSongRow(ctx, rowID)
	if err != nil {
		return err
	}
	if row.PlaylistID != playlistID {
		return domainlibrary.ErrNotFound
	}
	return s.playlists.UpdateSongPosition(ctx, rowID, position)
}

func (s *Service) RemoveSongFromPlaylist(ctx context.Context, playlistID, userID, rowID uuid.UUID) error {
	if _, err := s.GetPlaylist(ctx, playlistID, userID); err != nil {
		return err
	}
	row, err := s.playlists.FindSongRow(ctx, rowID)
	if err != nil {
		return err
	}
	if row.PlaylistID != playlistID {
		return domainlibrary.ErrNotFound
	}
	return s.playlists.RemoveSong(ctx, rowID)
}

func (s *Service) AddFavorite(ctx context.Context, userID uuid.UUID, favType domainlibrary.FavoritableType, favID uuid.UUID) (*domainlibrary.Favorite, error) {
	exists, err := s.favorites.Exists(ctx, userID, favType, favID)
	if err != nil {
		return nil, fmt.Errorf("library: check favorite: %w", err)
	}
	if exists {
		return nil, ErrConflict
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("library: generate favorite id: %w", err)
	}
	f := &domainlibrary.Favorite{ID: id, UserID: userID, FavoritableType: favType, FavoritableID: favID}
	if err := s.favorites.Create(ctx, f); err != nil {
		return nil, fmt.Errorf("library: create favorite: %w", err)
	}
	return f, nil
}

func (s *Service) RemoveFavorite(ctx context.Context, userID uuid.UUID, favType domainlibrary.FavoritableType, favID uuid.UUID) error {
	return s.favorites.Delete(ctx, userID, favType, favID)
}

func (s *Service) ListFavorites(ctx context.Context, userID uuid.UUID, favType domainlibrary.FavoritableType) ([]*domainlibrary.Favorite, error) {
	return s.favorites.ListByUser(ctx, userID, favType)
}
