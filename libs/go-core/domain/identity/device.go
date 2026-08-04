package identity

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type DeviceType string

const (
	DeviceWeb     DeviceType = "web"
	DeviceMobile  DeviceType = "mobile"
	DeviceDesktop DeviceType = "desktop"
)

type Device struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Name       string
	Type       DeviceType
	IsActive   bool
	LastSeenAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type DeviceRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Device, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*Device, error)
	Create(ctx context.Context, device *Device) error
	Touch(ctx context.Context, id uuid.UUID, seenAt time.Time) error
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	// SetActive marks deviceID as the user's sole Active Device (Sprint 8
	// Transfer Playback), clearing the flag on every other device of theirs.
	SetActive(ctx context.Context, userID, deviceID uuid.UUID) error
}
