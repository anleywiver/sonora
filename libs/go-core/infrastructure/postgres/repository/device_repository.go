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

type DeviceRepository struct {
	db *gorm.DB
}

func NewDeviceRepository(db *gorm.DB) *DeviceRepository {
	return &DeviceRepository{db: db}
}

func (r *DeviceRepository) FindByID(ctx context.Context, id uuid.UUID) (*identity.Device, error) {
	var row models.Device
	if err := r.db.WithContext(ctx).First(&row, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, identity.ErrNotFound
		}
		return nil, err
	}
	device := deviceFromModel(row)
	return &device, nil
}

func (r *DeviceRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*identity.Device, error) {
	var rows []models.Device
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	devices := make([]*identity.Device, 0, len(rows))
	for _, row := range rows {
		d := deviceFromModel(row)
		devices = append(devices, &d)
	}
	return devices, nil
}

func (r *DeviceRepository) Create(ctx context.Context, device *identity.Device) error {
	row := models.Device{
		ID:       device.ID,
		UserID:   device.UserID,
		Name:     device.Name,
		Type:     models.DeviceType(device.Type),
		IsActive: device.IsActive,
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return err
	}
	device.CreatedAt = row.CreatedAt
	device.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *DeviceRepository) Touch(ctx context.Context, id uuid.UUID, seenAt time.Time) error {
	return r.db.WithContext(ctx).Model(&models.Device{}).Where("id = ?", id).
		Update("last_seen_at", seenAt).Error
}

func (r *DeviceRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	res := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&models.Device{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return identity.ErrNotFound
	}
	return nil
}

func deviceFromModel(row models.Device) identity.Device {
	return identity.Device{
		ID:         row.ID,
		UserID:     row.UserID,
		Name:       row.Name,
		Type:       identity.DeviceType(row.Type),
		IsActive:   row.IsActive,
		LastSeenAt: row.LastSeenAt,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
}
