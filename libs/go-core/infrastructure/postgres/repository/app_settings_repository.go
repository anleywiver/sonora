package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"sonora.dev/go-core/infrastructure/postgres/models"
)

type AppSettingsRepository struct {
	db *gorm.DB
}

func NewAppSettingsRepository(db *gorm.DB) *AppSettingsRepository {
	return &AppSettingsRepository{db: db}
}

func (r *AppSettingsRepository) Get(ctx context.Context, key string) (string, error) {
	var row models.AppSetting
	if err := r.db.WithContext(ctx).First(&row, "key = ?", key).Error; err != nil {
		return "", err
	}
	return row.Value, nil
}

func (r *AppSettingsRepository) Set(ctx context.Context, key, value string) error {
	row := models.AppSetting{Key: key, Value: value, UpdatedAt: time.Now()}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(&row).Error
}

func (r *AppSettingsRepository) List(ctx context.Context) (map[string]string, error) {
	var rows []models.AppSetting
	if err := r.db.WithContext(ctx).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[string]string, len(rows))
	for _, row := range rows {
		out[row.Key] = row.Value
	}
	return out, nil
}
