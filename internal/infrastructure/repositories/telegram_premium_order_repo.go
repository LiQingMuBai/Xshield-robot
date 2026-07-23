package repositories

import (
	"context"
	"ushield_bot/internal/domain"

	"gorm.io/gorm"
)

type TelegramPremiumOrderRepository struct {
	db *gorm.DB
}

func NewTelegramPremiumOrderRepository(db *gorm.DB) *TelegramPremiumOrderRepository {
	return &TelegramPremiumOrderRepository{
		db: db,
	}
}

func (r *TelegramPremiumOrderRepository) GetByOrderNo(ctx context.Context, orderNo string) (domain.TelegramPremiumOrder, error) {
	var config domain.TelegramPremiumOrder
	err := r.db.WithContext(ctx).
		Find(&config, "order_no = ?", orderNo).Error
	return config, err
}

func (r *TelegramPremiumOrderRepository) Create(ctx context.Context, order *domain.TelegramPremiumOrder) error {
	return r.db.WithContext(ctx).Create(order).Error
}
func (r *TelegramPremiumOrderRepository) UpdateStatusByOrderNo(ctx context.Context, orderNo string, status int64) error {
	return r.db.WithContext(ctx).Model(&domain.TelegramPremiumOrder{}).
		Where("order_no = ?", orderNo).
		Update("status", status).Error
}
