package repositories

import (
	"context"
	"gorm.io/gorm"
	"ushield_bot/internal/domain"
)

type UserOperationBundlesRepository struct {
	db *gorm.DB
}

func NewUserOperationBundlesRepository(db *gorm.DB) *UserOperationBundlesRepository {
	return &UserOperationBundlesRepository{
		db: db,
	}
}
func (r *UserOperationBundlesRepository) ListByToken(ctx context.Context, token string) ([]domain.UserOperationBundles, error) {
	var bundles []domain.UserOperationBundles
	err := r.db.WithContext(ctx).
		Model(&domain.UserOperationBundles{}).
		Select("id", "name", "amount").
		Where("status = ?", 0).Where("token = ?", token).
		Scan(&bundles).Error
	return bundles, err

}
func (r *UserOperationBundlesRepository) ListActive(ctx context.Context) ([]domain.UserOperationBundles, error) {
	var bundles []domain.UserOperationBundles
	err := r.db.WithContext(ctx).
		Model(&domain.UserOperationBundles{}).
		Select("id", "name", "amount").
		Where("status = ?", 0).
		Scan(&bundles).Error
	return bundles, err

}
func (r *UserOperationBundlesRepository) GetByAmount(ctx context.Context, amount string) (domain.UserOperationBundles, error) {
	var bundleRecord domain.UserOperationBundles
	err := r.db.WithContext(ctx).
		Model(&domain.UserOperationBundles{}).
		Select("id", "name").
		Where("amount = ?", amount).
		Take(&bundleRecord).Error
	return bundleRecord, err

}

func (r *UserOperationBundlesRepository) GetByID(ctx context.Context, id string) (domain.UserOperationBundles, error) {
	var bundleRecord domain.UserOperationBundles
	err := r.db.WithContext(ctx).
		Model(&domain.UserOperationBundles{}).
		//Select("id", "days", "address", "network").
		Where("id = ?", id).
		Take(&bundleRecord).Error
	return bundleRecord, err

}
