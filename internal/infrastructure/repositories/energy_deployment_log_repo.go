package repositories

import (
	"context"

	"gorm.io/gorm"
	"ushield_bot/internal/domain"
)

type EnergyDeploymentLogRepo struct {
	db *gorm.DB
}

func NewEnergyDeploymentLogRepo(db *gorm.DB) *EnergyDeploymentLogRepo {
	return &EnergyDeploymentLogRepo{db: db}
}

func (r *EnergyDeploymentLogRepo) Create(ctx context.Context, record *domain.EnergyDeploymentLog) error {
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *EnergyDeploymentLogRepo) ListByChatID(ctx context.Context, chatID int64, limit int) ([]domain.EnergyDeploymentLog, error) {
	var records []domain.EnergyDeploymentLog
	if limit <= 0 {
		limit = 100
	}
	err := r.db.WithContext(ctx).
		Model(&domain.EnergyDeploymentLog{}).
		Where("chat_id = ?", chatID).
		Order("id DESC").
		Limit(limit).
		Find(&records).Error
	return records, err
}

func (r *EnergyDeploymentLogRepo) CountByChatID(ctx context.Context, chatID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.EnergyDeploymentLog{}).
		Where("chat_id = ?", chatID).
		Count(&count).Error
	return count, err
}
