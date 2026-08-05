// Package models holds GORM-mapped structs for simple CRUD domains
// (identity, library). Performance-critical domains (catalog, playback)
// go through sqlc instead — see ../sqlc.
package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	UserRoleOwner  UserRole = "owner"
	UserRoleMember UserRole = "member"
)

type User struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`
	// GoogleID is nullable (Sprint 14 sisipan, ADR 0009) — NULL means
	// "invited but never logged in yet", claimed (filled in) on first
	// real Google login. NULL rather than "" so the unique index allows
	// any number of pending invites without colliding.
	GoogleID  *string   `gorm:"column:google_id;uniqueIndex"`
	Email     string    `gorm:"uniqueIndex;not null"`
	Name      string    `gorm:"not null"`
	AvatarURL string    `gorm:"column:avatar_url"`
	Role      UserRole  `gorm:"not null;default:member"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
}

func (User) TableName() string { return "users" }

type DeviceType string

const (
	DeviceTypeWeb     DeviceType = "web"
	DeviceTypeMobile  DeviceType = "mobile"
	DeviceTypeDesktop DeviceType = "desktop"
)

type Device struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID     uuid.UUID  `gorm:"column:user_id;not null;index"`
	Name       string     `gorm:"not null"`
	Type       DeviceType `gorm:"not null;default:web"`
	IsActive   bool       `gorm:"column:is_active;not null;default:false"`
	LastSeenAt *time.Time `gorm:"column:last_seen_at"`
	CreatedAt  time.Time  `gorm:"not null;default:now()"`
	UpdatedAt  time.Time  `gorm:"not null;default:now()"`
}

func (Device) TableName() string { return "devices" }

// RefreshToken stores only the hash — the raw token is never persisted.
// Rotated on every use and bound to a single device.
type RefreshToken struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID  `gorm:"column:user_id;not null;index"`
	DeviceID  uuid.UUID  `gorm:"column:device_id;not null;index"`
	TokenHash string     `gorm:"column:token_hash;uniqueIndex;not null"`
	ExpiresAt time.Time  `gorm:"column:expires_at;not null"`
	RevokedAt *time.Time `gorm:"column:revoked_at"`
	CreatedAt time.Time  `gorm:"not null;default:now()"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }
