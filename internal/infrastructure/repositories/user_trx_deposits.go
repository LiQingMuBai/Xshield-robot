package repositories

import (
	"context"
	"ushield_bot/internal/request"

	_ "github.com/go-sql-driver/mysql"

	"ushield_bot/internal/domain"

	"gorm.io/gorm"
)

type UserTRXDepositsRepo struct {
	db *gorm.DB
}

func NewUserTRXDepositsRepository(db *gorm.DB) *UserTRXDepositsRepo {
	return &UserTRXDepositsRepo{
		db: db,
	}
}

func (r *UserTRXDepositsRepo) Create(ctx context.Context, trxDeposit *domain.UserTRXDeposits) error {
	return r.db.WithContext(ctx).Create(trxDeposit).Error
}
func (r *UserTRXDepositsRepo) ListByUserIDAndStatus(ctx context.Context, userID int64, status int64) ([]domain.UserTRXDeposits, error) {
	var subscriptions []domain.UserTRXDeposits

	err := r.db.Select("id,amount,order_no, DATE_FORMAT(created_at, '%m-%d') as created_date").
		Where("user_id = ?", userID).
		Where("status = ?", status).
		Find(&subscriptions).Error
	return subscriptions, err

}
func (r *UserTRXDepositsRepo) ListTRXDepositsByPage(ctx context.Context, info request.UserTrxDepositsSearch, chatID int64) (list []domain.UserTRXDeposits, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	// 创建db
	db := r.db.Model(&domain.UserTRXDeposits{}).Select("id,amount,order_no, DATE_FORMAT(created_at, '%m-%d') as created_date").Where("user_id = ?", chatID).Where("status = ?", 1)
	var deposits []domain.UserTRXDeposits
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

func (r *UserTRXDepositsRepo) GetByOrderNo(ctx context.Context, orderNo string) (domain.UserTRXDeposits, error) {
	var depositRecord domain.UserTRXDeposits
	err := r.db.WithContext(ctx).
		Find(&depositRecord, "order_no = ?", orderNo).Error
	return depositRecord, err

}

func (r *UserTRXDepositsRepo) Update(ctx context.Context, id int64, status int64) error {
	return r.db.WithContext(ctx).Model(&domain.UserTRXDeposits{}).
		Where("id = ?", id).
		Update("status", status).Error
}
