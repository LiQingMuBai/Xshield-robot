package repositories

import (
	"context"

	_ "github.com/go-sql-driver/mysql"

	"ushield_bot/internal/domain"

	"gorm.io/gorm"
)

type UserAddressTraceRepo struct {
	db *gorm.DB
}

func NewUserAddressTraceRepo(db *gorm.DB) *UserAddressTraceRepo {
	return &UserAddressTraceRepo{
		db: db,
	}
}

func (r *UserAddressTraceRepo) Create(ctx context.Context, address *domain.UserAddressTrace) error {
	return r.db.WithContext(ctx).Create(address).Error
}

func (r *UserAddressTraceRepo) DeleteByChatIDAndAddress(ctx context.Context, chatID int64, address string) error {
	return r.db.WithContext(ctx).Delete(&domain.UserAddressTrace{}, "chat_id = ? AND address = ?", chatID, address).Error
}

func (r *UserAddressTraceRepo) GetByChatIDAndAddress(ctx context.Context, chatID int64, address string) (domain.UserAddressTrace, error) {
	var item domain.UserAddressTrace
	err := r.db.WithContext(ctx).
		Find(&item, "chat_id = ? AND address = ?", chatID, address).Error
	return item, err

}

func (r *UserAddressTraceRepo) CountByChatID(ctx context.Context, chatID int64) (count int64, err error) {
	err = r.db.WithContext(ctx).Model(&domain.UserAddressTrace{}).Where("chat_id = ?", chatID).Count(&count).Error
	if err != nil {
		return
	}
	return count, nil
}

func (r *UserAddressTraceRepo) ListByChatID(ctx context.Context, chatID int64) ([]domain.UserAddressTrace, error) {
	var subscriptions []domain.UserAddressTrace
	err := r.db.WithContext(ctx).
		Model(&domain.UserAddressTrace{}).
		Select("id", "address", "network").
		Where("chat_id = ?", chatID).
		Scan(&subscriptions).Error
	return subscriptions, err

}
