package repositories

import (
	"context"
	"ushield_bot/internal/domain"

	"gorm.io/gorm"
)

type UserSmartTransactionBundlesRepository struct {
	db *gorm.DB
}

func NewUserSmartTransactionBundlesRepository(db *gorm.DB) *UserSmartTransactionBundlesRepository {
	return &UserSmartTransactionBundlesRepository{
		db: db,
	}
}
func (r *UserSmartTransactionBundlesRepository) ListByToken(ctx context.Context, token string) ([]domain.UserSmartTransactionBundles, error) {
	var bundles []domain.UserSmartTransactionBundles
	err := r.db.WithContext(ctx).
		Model(&domain.UserSmartTransactionBundles{}).
		Select("id", "name", "amount").
		Where("status = ?", 0).Where("token = ?", token).
		Scan(&bundles).Error
	return bundles, err

}
func (r *UserSmartTransactionBundlesRepository) ListActive(ctx context.Context) ([]domain.UserSmartTransactionBundles, error) {
	var bundles []domain.UserSmartTransactionBundles
	err := r.db.WithContext(ctx).
		Model(&domain.UserSmartTransactionBundles{}).
		Select("id", "name", "amount").
		Where("status = ?", 0).
		Scan(&bundles).Error
	return bundles, err

}
func (r *UserSmartTransactionBundlesRepository) GetByAmount(ctx context.Context, amount string) (domain.UserSmartTransactionBundles, error) {
	var bundleRecord domain.UserSmartTransactionBundles
	err := r.db.WithContext(ctx).
		Model(&domain.UserSmartTransactionBundles{}).
		Select("id", "name").
		Where("amount = ?", amount).
		Take(&bundleRecord).Error
	return bundleRecord, err

}

func (r *UserSmartTransactionBundlesRepository) GetByID(ctx context.Context, id string) (domain.UserSmartTransactionBundles, error) {
	var bundleRecord domain.UserSmartTransactionBundles
	err := r.db.WithContext(ctx).
		Model(&domain.UserSmartTransactionBundles{}).
		//Select("id", "days", "address", "network").
		Where("id = ?", id).
		Take(&bundleRecord).Error
	return bundleRecord, err

}
