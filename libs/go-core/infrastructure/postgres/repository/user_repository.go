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

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*identity.User, error) {
	var row models.User
	if err := r.db.WithContext(ctx).First(&row, "email = ?", email).Error; err != nil {
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
		GoogleID:  strPtrOrNil(user.GoogleID),
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

// ClaimInvite fills in a pending (google_id IS NULL) invite's identity
// fields on first real login — see ADR 0009.
func (r *UserRepository) ClaimInvite(ctx context.Context, id uuid.UUID, googleID, name, avatarURL string) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Updates(map[string]any{
		"google_id":  googleID,
		"name":       name,
		"avatar_url": avatarURL,
	}).Error
}

func (r *UserRepository) Update(ctx context.Context, user *identity.User) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]any{
		"name":       user.Name,
		"avatar_url": user.AvatarURL,
	}).Error
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.User{}, "id = ?", id).Error
}

func (r *UserRepository) ListAll(ctx context.Context) ([]*identity.User, error) {
	var rows []models.User
	if err := r.db.WithContext(ctx).Order("created_at ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*identity.User, 0, len(rows))
	for _, row := range rows {
		user := userFromModel(row)
		out = append(out, &user)
	}
	return out, nil
}

func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).Count(&count).Error
	return count, err
}

func strPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (r *UserRepository) FindOwner(ctx context.Context) (*identity.User, error) {
	var row models.User
	err := r.db.WithContext(ctx).
		Where("role = ?", models.UserRoleOwner).
		Order("created_at ASC").
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, identity.ErrNotFound
		}
		return nil, err
	}
	user := userFromModel(row)
	return &user, nil
}

func userFromModel(row models.User) identity.User {
	googleID := ""
	if row.GoogleID != nil {
		googleID = *row.GoogleID
	}
	return identity.User{
		ID:        row.ID,
		GoogleID:  googleID,
		Email:     row.Email,
		Name:      row.Name,
		AvatarURL: row.AvatarURL,
		Role:      identity.Role(row.Role),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
