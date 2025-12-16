package repositories

import (
	"context"
	"ushield_bot/internal/domain"

	"gorm.io/gorm"
)

type TelegramPremiumConfigRepository struct {
	db *gorm.DB
}

func NewTelegramPremiumConfigRepository(db *gorm.DB) *TelegramPremiumConfigRepository {
	return &TelegramPremiumConfigRepository{
		db: db,
	}
}

func (r *TelegramPremiumConfigRepository) Query(ctx context.Context, enName string) (domain.TelegramPremiumConfig, error) {
	var config domain.TelegramPremiumConfig
	err := r.db.WithContext(ctx).
		Find(&config, "en_name = ?", enName).Error
	return config, err
}
