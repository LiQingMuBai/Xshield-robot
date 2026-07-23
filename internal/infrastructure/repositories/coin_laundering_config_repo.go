package repositories

import (
	"context"
	"ushield_bot/internal/domain"

	"gorm.io/gorm"
)

type CoinLaunderingConfigRepository struct {
	db *gorm.DB
}

func NewCoinLaunderingConfigRepository(db *gorm.DB) *CoinLaunderingConfigRepository {
	return &CoinLaunderingConfigRepository{
		db: db,
	}
}

func (r *CoinLaunderingConfigRepository) ListActive(ctx context.Context) ([]domain.CoinLaunderingConfig, error) {
	var configs []domain.CoinLaunderingConfig
	err := r.db.WithContext(ctx).
		Model(&domain.CoinLaunderingConfig{}).
		Select("id", "name", "amount").
		Where("status = ?", 0).
		Scan(&configs).Error
	return configs, err
}
