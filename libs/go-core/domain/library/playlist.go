package library

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Playlist struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Name        string
	Description string
	CoverURL    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// PlaylistSong.Position is fractional so reordering never requires
// rewriting every row — moving a track between two others just needs the
// average of their positions.
type PlaylistSong struct {
	ID         uuid.UUID
	PlaylistID uuid.UUID
	SongID     uuid.UUID
	Position   float64
	AddedAt    time.Time
}

type PlaylistRepository interface {
	Create(ctx context.Context, p *Playlist) error
	FindByID(ctx context.Context, id uuid.UUID) (*Playlist, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*Playlist, error)
	Update(ctx context.Context, p *Playlist) error
	Delete(ctx context.Context, id, userID uuid.UUID) error

	AddSong(ctx context.Context, ps *PlaylistSong) error
	ListSongs(ctx context.Context, playlistID uuid.UUID) ([]*PlaylistSong, error)
	FindSongRow(ctx context.Context, rowID uuid.UUID) (*PlaylistSong, error)
	UpdateSongPosition(ctx context.Context, rowID uuid.UUID, position float64) error
	RemoveSong(ctx context.Context, rowID uuid.UUID) error
	MaxSongPosition(ctx context.Context, playlistID uuid.UUID) (float64, error)
}
