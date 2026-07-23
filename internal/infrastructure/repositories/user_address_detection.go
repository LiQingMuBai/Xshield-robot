package repositories

import (
	"context"
	_ "github.com/go-sql-driver/mysql"
	"ushield_bot/internal/request"

	"gorm.io/gorm"
	"ushield_bot/internal/domain"
)

type UserAddressDetectionRepo struct {
	db *gorm.DB
}

func NewUserAddressDetectionRepository(db *gorm.DB) *UserAddressDetectionRepo {
	return &UserAddressDetectionRepo{
		db: db,
	}
}

func (r *UserAddressDetectionRepo) Create(ctx context.Context, detection *domain.UserAddressDetection) error {
	return r.db.WithContext(ctx).Create(detection).Error
}

func (r *UserAddressDetectionRepo) ListHistoryByChatIDAndStatus(ctx context.Context, chatID int64, status int64) ([]domain.UserAddressDetection, error) {
	var detections []domain.UserAddressDetection

	err := r.db.WithContext(ctx).
		Select("id,amount,address, DATE_FORMAT(created_at, '%m-%d') as created_date").
		Where("chat_id = ?", chatID).
		Where("status = ?", status).
		Find(&detections).Error
	return detections, err

}
func (r *UserAddressDetectionRepo) ListByChatIDPage(ctx context.Context, info request.UserAddressDetectionSearch, chatID int64) (list []domain.UserAddressDetection, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	// 创建db
	db := r.db.WithContext(ctx).Model(&domain.UserAddressDetection{}).Select("id,amount,address, DATE_FORMAT(created_at, '%m-%d') as created_date").Where("chat_id = ?", chatID)
	var detections []domain.UserAddressDetection
	// 如果有条件搜索 下方会自动创建搜索语句

	err = db.Count(&total).Error
	if err != nil {
		return
	}

	if limit != 0 {
		db = db.Limit(int(limit)).Offset(int(offset)).Order("id DESC")
	}

	err = db.Find(&detections).Error
	return detections, total, err
}
