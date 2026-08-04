package models

import (
	"time"

	"github.com/google/uuid"
)

type Playlist struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID      uuid.UUID `gorm:"column:user_id;not null;index"`
	Name        string    `gorm:"not null"`
	Description string
	CoverURL    string    `gorm:"column:cover_url"`
	CreatedAt   time.Time `gorm:"not null;default:now()"`
	UpdatedAt   time.Time `gorm:"not null;default:now()"`
}

func (Playlist) TableName() string { return "playlists" }

// PlaylistSong.Position is fractional so reordering never requires
// rewriting every row in the playlist.
type PlaylistSong struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	PlaylistID uuid.UUID `gorm:"column:playlist_id;not null;index"`
	SongID     uuid.UUID `gorm:"column:song_id;not null"`
	Position   float64   `gorm:"not null"`
	AddedAt    time.Time `gorm:"column:added_at;not null;default:now()"`
}

func (PlaylistSong) TableName() string { return "playlist_songs" }

type FavoritableType string

const (
	FavoritableSong     FavoritableType = "song"
	FavoritableAlbum    FavoritableType = "album"
	FavoritableArtist   FavoritableType = "artist"
	FavoritablePlaylist FavoritableType = "playlist"
)

type Favorite struct {
	ID              uuid.UUID       `gorm:"type:uuid;primaryKey"`
	UserID          uuid.UUID       `gorm:"column:user_id;not null;index"`
	FavoritableType FavoritableType `gorm:"column:favoritable_type;not null"`
	FavoritableID   uuid.UUID       `gorm:"column:favoritable_id;not null"`
	CreatedAt       time.Time       `gorm:"not null;default:now()"`
}

func (Favorite) TableName() string { return "favorites" }

type PlayHistory struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID     uuid.UUID `gorm:"column:user_id;not null;index"`
	SongID     uuid.UUID `gorm:"column:song_id;not null"`
	ProgressMs int       `gorm:"column:progress_ms;not null;default:0"`
	PlayedAt   time.Time `gorm:"column:played_at;not null;default:now()"`
}

func (PlayHistory) TableName() string { return "play_history" }
