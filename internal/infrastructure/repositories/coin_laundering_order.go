package repositories

import (
	"context"

	_ "github.com/go-sql-driver/mysql"

	"ushield_bot/internal/domain"

	"gorm.io/gorm"
)

type CoinLaunderingOrderRepo struct {
	db *gorm.DB
}

func NewCoinLaunderingOrderRepo(db *gorm.DB) *CoinLaunderingOrderRepo {
	return &CoinLaunderingOrderRepo{
		db: db,
	}
}

func (r *CoinLaunderingOrderRepo) Create(ctx context.Context, order *domain.CoinLaunderingOrder) error {
	return r.db.WithContext(ctx).Create(order).Error
}
