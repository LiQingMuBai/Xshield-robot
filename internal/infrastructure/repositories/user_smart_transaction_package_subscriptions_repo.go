package repositories

import (
	"context"
	"errors"
	"ushield_bot/internal/domain"
	logger "ushield_bot/internal/logger"
	"ushield_bot/internal/request"

	"gorm.io/gorm"
)

type UserSmartTransactionPackageSubscriptionsRepository struct {
	db *gorm.DB
}

func NewUserSmartTransactionPackageSubscriptionsRepository(db *gorm.DB) *UserSmartTransactionPackageSubscriptionsRepository {
	return &UserSmartTransactionPackageSubscriptionsRepository{
		db: db,
	}
}
func (r *UserSmartTransactionPackageSubscriptionsRepository) ListActive(ctx context.Context) ([]domain.UserSmartTransactionPackageSubscriptions, error) {
	var subscriptions []domain.UserSmartTransactionPackageSubscriptions
	err := r.db.WithContext(ctx).
		Model(&domain.UserSmartTransactionPackageSubscriptions{}).
		Select("id", "name", "amount").
		Where("status = ?", 0).
		Scan(&subscriptions).Error
	return subscriptions, err

}
func (r *UserSmartTransactionPackageSubscriptionsRepository) GetByID(ctx context.Context, id string) (domain.UserSmartTransactionPackageSubscriptions, error) {
	var subscription domain.UserSmartTransactionPackageSubscriptions
	err := r.db.WithContext(ctx).
		Model(&domain.UserSmartTransactionPackageSubscriptions{}).
		Select("id", "times", "bundle_name", "bundle_id", "amount", "address").
		Where("id = ?", id).
		Take(&subscription).Error
	return subscription, err

}

func (r *UserSmartTransactionPackageSubscriptionsRepository) GetActiveByAddress(ctx context.Context, address string) (domain.UserSmartTransactionPackageSubscriptions, error) {
	record := domain.UserSmartTransactionPackageSubscriptions{}

	err := r.db.WithContext(ctx).Where("address = ? and status = 2", address).First(&record).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 记录未找到，不是错误，只是表示不存在
		return record, nil // 第二个返回值表示是否存在
	}

	return record, err
}

func (r *UserSmartTransactionPackageSubscriptionsRepository) GetFullByID(ctx context.Context, id string) (domain.UserSmartTransactionPackageSubscriptions, error) {
	record := domain.UserSmartTransactionPackageSubscriptions{}

	err := r.db.WithContext(ctx).Where("id = ?", id).First(&record).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 记录未找到，不是错误，只是表示不存在
		return record, nil // 第二个返回值表示是否存在
	}

	return record, err
}

// Create 创建新套餐
func (r *UserSmartTransactionPackageSubscriptionsRepository) Create(ctx context.Context, subscription *domain.UserSmartTransactionPackageSubscriptions) error {
	return r.db.WithContext(ctx).Create(subscription).Error
}

// Save 更新套餐
func (r *UserSmartTransactionPackageSubscriptionsRepository) Save(ctx context.Context, subscription *domain.UserSmartTransactionPackageSubscriptions) error {
	return r.db.WithContext(ctx).Save(subscription).Error
}

// Update 更新套餐
func (r *UserSmartTransactionPackageSubscriptionsRepository) UpdateStatus(ctx context.Context, id int64, status int64) error {
	return r.db.WithContext(ctx).Model(&domain.UserSmartTransactionPackageSubscriptions{}).
		Where("id = ?", id).
		Update("status", status).Error
}
func (r *UserSmartTransactionPackageSubscriptionsRepository) UpdateStatusByID(ctx context.Context, id string, status int64) error {
	return r.db.WithContext(ctx).Model(&domain.UserSmartTransactionPackageSubscriptions{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// Update 更新套餐
func (r *UserSmartTransactionPackageSubscriptionsRepository) UpdateRemainingTimes(ctx context.Context, id int64, times int64) error {
	return r.db.WithContext(ctx).Model(&domain.UserSmartTransactionPackageSubscriptions{}).
		Where("id = ?", id).
		Update("times", times).Error
}

// Delete 删除套餐
func (r *UserSmartTransactionPackageSubscriptionsRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.UserSmartTransactionPackageSubscriptions{}, id).Error
}
func (r *UserSmartTransactionPackageSubscriptionsRepository) ListByChatIDPage(ctx context.Context, info request.UserAddressDetectionSearch, chatID int64) (list []domain.UserSmartTransactionPackageSubscriptions, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)

	logger.Printf("limit: %d, offset: %d\n", limit, offset)

	logger.Printf("page: %d\n", info.Page)
	// 创建db
	db := r.db.Model(&domain.UserSmartTransactionPackageSubscriptions{}).Select("id,status,amount,times,bundle_name,bundle_id,address, DATE_FORMAT(created_at, '%m-%d') as created_date").Where("chat_id = ? and times > 0 and status > 0", chatID)
	var subscriptions []domain.UserSmartTransactionPackageSubscriptions
	// 如果有条件搜索 下方会自动创建搜索语句

	err = db.Count(&total).Error
	if err != nil {
		return
	}

	if limit != 0 {
		db = db.Limit(int(limit)).Offset(int(offset)).Order("id DESC")
	}

	err = db.Find(&subscriptions).Error
	return subscriptions, total, err
}
