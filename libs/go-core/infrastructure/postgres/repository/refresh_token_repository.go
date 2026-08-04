package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"sonora.dev/go-core/domain/identity"
	"sonora.dev/go-core/infrastructure/postgres/models"
)

type RefreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) FindByHash(ctx context.Context, tokenHash string) (*identity.RefreshToken, error) {
	var row models.RefreshToken
	if err := r.db.WithContext(ctx).First(&row, "token_hash = ?", tokenHash).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, identity.ErrNotFound
		}
		return nil, err
	}
	token := refreshTokenFromModel(row)
	return &token, nil
}

func (r *RefreshTokenRepository) Create(ctx context.Context, token *identity.RefreshToken) error {
	row := models.RefreshToken{
		ID:        token.ID,
		UserID:    token.UserID,
		DeviceID:  token.DeviceID,
		TokenHash: token.TokenHash,
		ExpiresAt: token.ExpiresAt,
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return err
	}
	token.CreatedAt = row.CreatedAt
	return nil
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, id uuid.UUID, revokedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&models.RefreshToken{}).Where("id = ?", id).
		Update("revoked_at", revokedAt).Error
}

func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID, revokedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", revokedAt).Error
}

func (r *RefreshTokenRepository) RevokeAllForDevice(ctx context.Context, deviceID uuid.UUID, revokedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&models.RefreshToken{}).
		Where("device_id = ? AND revoked_at IS NULL", deviceID).
		Update("revoked_at", revokedAt).Error
}

func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("expires_at < ?", before).Delete(&models.RefreshToken{})
	return result.RowsAffected, result.Error
}

func refreshTokenFromModel(row models.RefreshToken) identity.RefreshToken {
	return identity.RefreshToken{
		ID:        row.ID,
		UserID:    row.UserID,
		DeviceID:  row.DeviceID,
		TokenHash: row.TokenHash,
		ExpiresAt: row.ExpiresAt,
		RevokedAt: row.RevokedAt,
		CreatedAt: row.CreatedAt,
	}
}
