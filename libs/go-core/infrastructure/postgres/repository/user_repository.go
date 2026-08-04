package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"sonora.dev/go-core/domain/identity"
	"sonora.dev/go-core/infrastructure/postgres/models"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*identity.User, error) {
	var row models.User
	if err := r.db.WithContext(ctx).First(&row, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, identity.ErrNotFound
		}
		return nil, err
	}
	user := userFromModel(row)
	return &user, nil
}

func (r *UserRepository) FindByGoogleID(ctx context.Context, googleID string) (*identity.User, error) {
	var row models.User
	if err := r.db.WithContext(ctx).First(&row, "google_id = ?", googleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, identity.ErrNotFound
		}
		return nil, err
	}
	user := userFromModel(row)
	return &user, nil
}

func (r *UserRepository) Create(ctx context.Context, user *identity.User) error {
	row := models.User{
		ID:        user.ID,
		GoogleID:  user.GoogleID,
		Email:     user.Email,
		Name:      user.Name,
		AvatarURL: user.AvatarURL,
		Role:      models.UserRole(user.Role),
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return err
	}
	user.CreatedAt = row.CreatedAt
	user.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).Count(&count).Error
	return count, err
}

func userFromModel(row models.User) identity.User {
	return identity.User{
		ID:        row.ID,
		GoogleID:  row.GoogleID,
		Email:     row.Email,
		Name:      row.Name,
		AvatarURL: row.AvatarURL,
		Role:      identity.Role(row.Role),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
