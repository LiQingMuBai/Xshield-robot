package repositories

import (
	"context"
	"ushield_bot/internal/request"

	_ "github.com/go-sql-driver/mysql"

	"ushield_bot/internal/domain"

	"gorm.io/gorm"
)

type UserUSDTDepositsRepo struct {
	db *gorm.DB
}

func NewUserUSDTDepositsRepository(db *gorm.DB) *UserUSDTDepositsRepo {
	return &UserUSDTDepositsRepo{
		db: db,
	}
}

func (r *UserUSDTDepositsRepo) Create(ctx context.Context, deposit *domain.UserUSDTDeposits) error {
	return r.db.WithContext(ctx).Create(deposit).Error
}

func (r *UserUSDTDepositsRepo) ListHistoryByUserIDAndStatus(ctx context.Context, userID int64, status int64) ([]domain.UserUSDTDeposits, error) {
	var deposits []domain.UserUSDTDeposits
	err := r.db.WithContext(ctx).
		Select("id,amount,order_no, DATE_FORMAT(created_at, '%m-%d') as created_date").
		Where("user_id = ?", userID).
		Where("status = ?", status).
		Find(&deposits).Error
	return deposits, err

}
func (r *UserUSDTDepositsRepo) ListByPage(ctx context.Context, info request.UserUsdtDepositsSearch, chatID int64) (list []domain.UserUSDTDeposits, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	// 创建db
	db := r.db.WithContext(ctx).Model(&domain.UserUSDTDeposits{}).Select("id,amount,order_no, DATE_FORMAT(created_at, '%m-%d') as created_date").Where("user_id = ?", chatID).Where("status = ?", 1)
	var deposits []domain.UserUSDTDeposits
	// 如果有条件搜索 下方会自动创建搜索语句

	err = db.Count(&total).Error
	if err != nil {
		return
	}

	if limit != 0 {
		db = db.Limit(int(limit)).Offset(int(offset)).Order("id DESC")
	}

	err = db.Find(&deposits).Error
	return deposits, total, err
}

func (r *UserUSDTDepositsRepo) GetByOrderNo(ctx context.Context, orderNo string) (domain.UserUSDTDeposits, error) {
	var depositRecord domain.UserUSDTDeposits
	err := r.db.WithContext(ctx).
		Find(&depositRecord, "order_no = ?", orderNo).Error
	return depositRecord, err

}

func (r *UserUSDTDepositsRepo) UpdateStatusByID(ctx context.Context, id int64, status int64) error {
	return r.db.WithContext(ctx).Model(&domain.UserUSDTDeposits{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *UserUSDTDepositsRepo) UpdateStatusByOrderNo(ctx context.Context, orderNo string, status int64) error {
	return r.db.WithContext(ctx).Model(&domain.UserUSDTDeposits{}).
		Where("order_no = ?", orderNo).
		Update("status", status).Error
}
