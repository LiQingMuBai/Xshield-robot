package repositories

import (
	"context"
	"gorm.io/gorm"
	"ushield_bot/internal/domain"
	"ushield_bot/internal/request"
)

type UserPackageSubscriptionsRepository struct {
	db *gorm.DB
}

func NewUserPackageSubscriptionsRepository(db *gorm.DB) *UserPackageSubscriptionsRepository {
	return &UserPackageSubscriptionsRepository{
		db: db,
	}
}
func (r *UserPackageSubscriptionsRepository) ListActive(ctx context.Context) ([]domain.UserPackageSubscriptions, error) {
	var subscriptions []domain.UserPackageSubscriptions
	err := r.db.WithContext(ctx).
		Model(&domain.UserPackageSubscriptions{}).
		Select("id", "name", "amount").
		Where("status = ?", 0).
		Scan(&subscriptions).Error
	return subscriptions, err

}
func (r *UserPackageSubscriptionsRepository) GetByID(ctx context.Context, id string) (domain.UserPackageSubscriptions, error) {
	var subscription domain.UserPackageSubscriptions
	err := r.db.WithContext(ctx).
		Model(&domain.UserPackageSubscriptions{}).
		Select("id", "times", "bundle_name", "bundle_id", "amount", "address").
		Where("id = ?", id).
		Take(&subscription).Error
	return subscription, err

}

// Create 创建新套餐
func (r *UserPackageSubscriptionsRepository) Create(ctx context.Context, subscription *domain.UserPackageSubscriptions) error {
	return r.db.WithContext(ctx).Create(subscription).Error
}

// Save 更新套餐
func (r *UserPackageSubscriptionsRepository) Save(ctx context.Context, subscription *domain.UserPackageSubscriptions) error {
	return r.db.WithContext(ctx).Save(subscription).Error
}

// Update 更新套餐
func (r *UserPackageSubscriptionsRepository) UpdateStatus(ctx context.Context, id int64, status int64) error {
	return r.db.WithContext(ctx).Model(&domain.UserPackageSubscriptions{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// Update 更新套餐
func (r *UserPackageSubscriptionsRepository) UpdateRemainingTimes(ctx context.Context, id int64, times int64) error {
	return r.db.WithContext(ctx).Model(&domain.UserPackageSubscriptions{}).
		Where("id = ?", id).
		Update("times", times).Error
}

// Delete 删除套餐
func (r *UserPackageSubscriptionsRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.UserPackageSubscriptions{}, id).Error
}
func (r *UserPackageSubscriptionsRepository) ListByChatIDPage(ctx context.Context, info request.UserAddressDetectionSearch, chatID int64) (list []domain.UserPackageSubscriptions, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	// 创建db
	db := r.db.Model(&domain.UserPackageSubscriptions{}).Select("id,status,amount,times,bundle_name,bundle_id,address, DATE_FORMAT(created_at, '%m-%d') as created_date").Where("chat_id = ? and times > 0", chatID)
	var subscriptions []domain.UserPackageSubscriptions
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
