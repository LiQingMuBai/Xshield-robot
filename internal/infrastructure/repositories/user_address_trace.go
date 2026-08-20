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

func (r *UserAddressTraceRepo) DeleteByChatIDAndAddress(ctx context.Context, chatID int64, address string) (int64, error) {
	tx := r.db.WithContext(ctx).
		Exec("DELETE FROM user_address_trace WHERE chat_id = ? AND LOWER(address) = LOWER(?)", chatID, address)
	return tx.RowsAffected, tx.Error
}

func (r *UserAddressTraceRepo) DeleteByChatIDAddressAndNetwork(ctx context.Context, chatID int64, address, network string) (int64, error) {
	tx := r.db.WithContext(ctx).
		Exec("DELETE FROM user_address_trace WHERE chat_id = ? AND LOWER(address) = LOWER(?) AND network = ?",
			chatID, address, network)
	return tx.RowsAffected, tx.Error
}

func (r *UserAddressTraceRepo) GetByChatIDAndAddress(ctx context.Context, chatID int64, address string) (domain.UserAddressTrace, error) {
	var item domain.UserAddressTrace
	err := r.db.WithContext(ctx).
		Where("chat_id = ? AND LOWER(address) = LOWER(?) AND status = 1", chatID, address).
		First(&item).Error
	return item, err
}

func (r *UserAddressTraceRepo) GetByChatIDAddressAndNetwork(ctx context.Context, chatID int64, address, network string) (domain.UserAddressTrace, error) {
	var item domain.UserAddressTrace
	err := r.db.WithContext(ctx).
		Where("chat_id = ? AND LOWER(address) = LOWER(?) AND network = ? AND status = 1", chatID, address, network).
		First(&item).Error
	return item, err
}

func (r *UserAddressTraceRepo) CountByChatID(ctx context.Context, chatID int64) (count int64, err error) {
	err = r.db.WithContext(ctx).Model(&domain.UserAddressTrace{}).Where("chat_id = ? AND status = 1", chatID).Count(&count).Error
	if err != nil {
		return
	}
	return count, nil
}

func (r *UserAddressTraceRepo) CountByChatIDAndNetwork(ctx context.Context, chatID int64, network string) (count int64, err error) {
	err = r.db.WithContext(ctx).Model(&domain.UserAddressTrace{}).
		Where("chat_id = ? AND network = ? AND status = 1", chatID, network).Count(&count).Error
	return count, err
}

func (r *UserAddressTraceRepo) ListByChatID(ctx context.Context, chatID int64) ([]domain.UserAddressTrace, error) {
	var subscriptions []domain.UserAddressTrace
	err := r.db.WithContext(ctx).
		Model(&domain.UserAddressTrace{}).
		Select("id", "address", "network").
		Where("chat_id = ? AND status = 1", chatID).
		Scan(&subscriptions).Error
	return subscriptions, err
}

func (r *UserAddressTraceRepo) ListByChatIDAndNetwork(ctx context.Context, chatID int64, network string) ([]domain.UserAddressTrace, error) {
	var subscriptions []domain.UserAddressTrace
	err := r.db.WithContext(ctx).
		Model(&domain.UserAddressTrace{}).
		Select("id", "address", "network").
		Where("chat_id = ? AND network = ? AND status = 1", chatID, network).
		Scan(&subscriptions).Error
	return subscriptions, err
}
