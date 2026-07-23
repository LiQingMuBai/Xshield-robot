package repositories

import (
	"context"
	"ushield_bot/internal/domain"

	"gorm.io/gorm"
)

type TelegramStarsOrderRepository struct {
	db *gorm.DB
}

func NewTelegramStarsOrderRepository(db *gorm.DB) *TelegramStarsOrderRepository {
	return &TelegramStarsOrderRepository{
		db: db,
	}
}

func (r *TelegramStarsOrderRepository) Query(ctx context.Context, orderNO string) (domain.TelegramStarsOrder, error) {
	var config domain.TelegramStarsOrder
	err := r.db.WithContext(ctx).
		Find(&config, "order_no = ?", orderNO).Error
	return config, err
}

func (r *TelegramStarsOrderRepository) Create(ctx context.Context, order *domain.TelegramStarsOrder) error {
	return r.db.WithContext(ctx).Create(order).Error
}
func (r *TelegramStarsOrderRepository) Update(ctx context.Context, orderNo string, _status int64) error {
	return r.db.WithContext(ctx).Model(&domain.TelegramStarsOrder{}).
		Where("order_no = ?", orderNo).
		Update("status", _status).Error
}
