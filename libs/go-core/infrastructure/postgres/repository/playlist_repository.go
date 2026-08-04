package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"sonora.dev/go-core/domain/library"
	"sonora.dev/go-core/infrastructure/postgres/models"
)

type PlaylistRepository struct {
	db *gorm.DB
}

func NewPlaylistRepository(db *gorm.DB) *PlaylistRepository {
	return &PlaylistRepository{db: db}
}

func (r *PlaylistRepository) Create(ctx context.Context, p *library.Playlist) error {
	row := models.Playlist{
		ID:          p.ID,
		UserID:      p.UserID,
		Name:        p.Name,
		Description: p.Description,
		CoverURL:    p.CoverURL,
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return err
	}
	p.CreatedAt = row.CreatedAt
	p.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *PlaylistRepository) FindByID(ctx context.Context, id uuid.UUID) (*library.Playlist, error) {
	var row models.Playlist
	if err := r.db.WithContext(ctx).First(&row, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, library.ErrNotFound
		}
		return nil, err
	}
	p := playlistFromModel(row)
	return &p, nil
}

func (r *PlaylistRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*library.Playlist, error) {
	var rows []models.Playlist
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*library.Playlist, 0, len(rows))
	for _, row := range rows {
		p := playlistFromModel(row)
		out = append(out, &p)
	}
	return out, nil
}

func (r *PlaylistRepository) Update(ctx context.Context, p *library.Playlist) error {
	res := r.db.WithContext(ctx).Model(&models.Playlist{}).Where("id = ? AND user_id = ?", p.ID, p.UserID).
		Updates(map[string]any{
			"name":        p.Name,
			"description": p.Description,
			"cover_url":   p.CoverURL,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return library.ErrNotFound
	}
	return nil
}

func (r *PlaylistRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	res := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&models.Playlist{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return library.ErrNotFound
	}
	return nil
}

func (r *PlaylistRepository) AddSong(ctx context.Context, ps *library.PlaylistSong) error {
	row := models.PlaylistSong{
		ID:         ps.ID,
		PlaylistID: ps.PlaylistID,
		SongID:     ps.SongID,
		Position:   ps.Position,
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return err
	}
	ps.AddedAt = row.AddedAt
	return nil
}

func (r *PlaylistRepository) ListSongs(ctx context.Context, playlistID uuid.UUID) ([]*library.PlaylistSong, error) {
	var rows []models.PlaylistSong
	if err := r.db.WithContext(ctx).Where("playlist_id = ?", playlistID).Order("position ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*library.PlaylistSong, 0, len(rows))
	for _, row := range rows {
		out = append(out, playlistSongFromModel(row))
	}
	return out, nil
}

func (r *PlaylistRepository) FindSongRow(ctx context.Context, rowID uuid.UUID) (*library.PlaylistSong, error) {
	var row models.PlaylistSong
	if err := r.db.WithContext(ctx).First(&row, "id = ?", rowID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, library.ErrNotFound
		}
		return nil, err
	}
	return playlistSongFromModel(row), nil
}

func (r *PlaylistRepository) UpdateSongPosition(ctx context.Context, rowID uuid.UUID, position float64) error {
	res := r.db.WithContext(ctx).Model(&models.PlaylistSong{}).Where("id = ?", rowID).Update("position", position)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return library.ErrNotFound
	}
	return nil
}

func (r *PlaylistRepository) RemoveSong(ctx context.Context, rowID uuid.UUID) error {
	res := r.db.WithContext(ctx).Where("id = ?", rowID).Delete(&models.PlaylistSong{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return library.ErrNotFound
	}
	return nil
}

// MaxSongPosition returns the highest position in the playlist, or 0 if
// empty — the caller adds a fixed increment on top for the new song.
func (r *PlaylistRepository) MaxSongPosition(ctx context.Context, playlistID uuid.UUID) (float64, error) {
	var max *float64
	if err := r.db.WithContext(ctx).Model(&models.PlaylistSong{}).
		Where("playlist_id = ?", playlistID).
		Select("MAX(position)").Scan(&max).Error; err != nil {
		return 0, err
	}
	if max == nil {
		return 0, nil
	}
	return *max, nil
}

func playlistFromModel(row models.Playlist) library.Playlist {
	return library.Playlist{
		ID:          row.ID,
		UserID:      row.UserID,
		Name:        row.Name,
		Description: row.Description,
		CoverURL:    row.CoverURL,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func playlistSongFromModel(row models.PlaylistSong) *library.PlaylistSong {
	return &library.PlaylistSong{
		ID:         row.ID,
		PlaylistID: row.PlaylistID,
		SongID:     row.SongID,
		Position:   row.Position,
		AddedAt:    row.AddedAt,
	}
}
