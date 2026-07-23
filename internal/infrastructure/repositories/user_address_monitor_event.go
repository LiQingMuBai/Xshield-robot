package repositories

import (
	"context"
	"ushield_bot/internal/request"

	_ "github.com/go-sql-driver/mysql"

	"ushield_bot/internal/domain"

	"gorm.io/gorm"
)

type UserAddressMonitorEventRepo struct {
	db *gorm.DB
}

func NewUserAddressMonitorEventRepo(db *gorm.DB) *UserAddressMonitorEventRepo {
	return &UserAddressMonitorEventRepo{
		db: db,
	}
}

func (r *UserAddressMonitorEventRepo) GetByID(ctx context.Context, id string) (domain.UserAddressMonitorEvent, error) {
	var event domain.UserAddressMonitorEvent
	err := r.db.WithContext(ctx).
		Find(&event, "id = ?", id).Error
	return event, err

}

func (r *UserAddressMonitorEventRepo) Create(ctx context.Context, event *domain.UserAddressMonitorEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}
func (r *UserAddressMonitorEventRepo) DeleteByChatIDAndAddress(ctx context.Context, chatID int64, address string) error {
	return r.db.WithContext(ctx).Delete(&domain.UserAddressMonitorEvent{}, "chat_id = ? AND address = ?", chatID, address).Error
}
func (r *UserAddressMonitorEventRepo) DeleteByID(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.UserAddressMonitorEvent{}, "id = ? ", id).Error
}

func (r *UserAddressMonitorEventRepo) ListByChatID(ctx context.Context, chatID int64) ([]domain.UserAddressMonitorEvent, error) {
	var subscriptions []domain.UserAddressMonitorEvent
	err := r.db.WithContext(ctx).
		Model(&domain.UserAddressMonitorEvent{}).
		Select("id", "days", "address", "network").
		Where("chat_id = ? and status = 1", chatID).
		Scan(&subscriptions).Error
	return subscriptions, err

}
func (r *UserAddressMonitorEventRepo) DeleteAllByChatID(ctx context.Context, chatID int64) error {
	return r.db.WithContext(ctx).Delete(&domain.UserAddressMonitorEvent{}, "chat_id = ?", chatID).Error
}

func (r *UserAddressMonitorEventRepo) ListByChatIDPage(ctx context.Context, info request.UserAddressDetectionSearch, chatID int64) (list []domain.UserAddressMonitorEvent, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	// 创建db
	db := r.db.Model(&domain.UserAddressMonitorEvent{}).Select("id,amount,address, DATE_FORMAT(created_at, '%m-%d') as created_date").Where("chat_id = ? ", chatID)
	var events []domain.UserAddressMonitorEvent
	// 如果有条件搜索 下方会自动创建搜索语句

	err = db.Count(&total).Error
	if err != nil {
		return
	}

	if limit != 0 {
		db = db.Limit(int(limit)).Offset(int(offset)).Order("id DESC")
	}

	err = db.Find(&events).Error
	return events, total, err
}
