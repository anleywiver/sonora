package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"sonora.dev/go-core/domain/library"
	"sonora.dev/go-core/infrastructure/postgres/models"
)

type FavoriteRepository struct {
	db *gorm.DB
}

func NewFavoriteRepository(db *gorm.DB) *FavoriteRepository {
	return &FavoriteRepository{db: db}
}

func (r *FavoriteRepository) Create(ctx context.Context, f *library.Favorite) error {
	row := models.Favorite{
		ID:              f.ID,
		UserID:          f.UserID,
		FavoritableType: models.FavoritableType(f.FavoritableType),
		FavoritableID:   f.FavoritableID,
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return err
	}
	f.CreatedAt = row.CreatedAt
	return nil
}

func (r *FavoriteRepository) Delete(ctx context.Context, userID uuid.UUID, favType library.FavoritableType, favID uuid.UUID) error {
	res := r.db.WithContext(ctx).
		Where("user_id = ? AND favoritable_type = ? AND favoritable_id = ?", userID, string(favType), favID).
		Delete(&models.Favorite{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return library.ErrNotFound
	}
	return nil
}

func (r *FavoriteRepository) ListByUser(ctx context.Context, userID uuid.UUID, favType library.FavoritableType) ([]*library.Favorite, error) {
	q := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if favType != "" {
		q = q.Where("favoritable_type = ?", string(favType))
	}
	var rows []models.Favorite
	if err := q.Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*library.Favorite, 0, len(rows))
	for _, row := range rows {
		out = append(out, &library.Favorite{
			ID:              row.ID,
			UserID:          row.UserID,
			FavoritableType: library.FavoritableType(row.FavoritableType),
			FavoritableID:   row.FavoritableID,
			CreatedAt:       row.CreatedAt,
		})
	}
	return out, nil
}

func (r *FavoriteRepository) Exists(ctx context.Context, userID uuid.UUID, favType library.FavoritableType, favID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Favorite{}).
		Where("user_id = ? AND favoritable_type = ? AND favoritable_id = ?", userID, string(favType), favID).
		Count(&count).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}
