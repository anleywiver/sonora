package library

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type FavoritableType string

const (
	FavoritableSong     FavoritableType = "song"
	FavoritableAlbum    FavoritableType = "album"
	FavoritableArtist   FavoritableType = "artist"
	FavoritablePlaylist FavoritableType = "playlist"
)

type Favorite struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	FavoritableType FavoritableType
	FavoritableID   uuid.UUID
	CreatedAt       time.Time
}

type FavoriteRepository interface {
	Create(ctx context.Context, f *Favorite) error
	Delete(ctx context.Context, userID uuid.UUID, favType FavoritableType, favID uuid.UUID) error
	ListByUser(ctx context.Context, userID uuid.UUID, favType FavoritableType) ([]*Favorite, error)
	Exists(ctx context.Context, userID uuid.UUID, favType FavoritableType, favID uuid.UUID) (bool, error)
}
