package repositories

import (
	"context"
	"gorm.io/gorm"
	"ushield_bot/internal/domain"
)

type UserUsdtPlaceholdersRepository struct {
	db *gorm.DB
}

func NewUserUsdtPlaceholdersRepository(db *gorm.DB) *UserUsdtPlaceholdersRepository {
	return &UserUsdtPlaceholdersRepository{
		db: db,
	}
}
func (r *UserUsdtPlaceholdersRepository) ListAvailable(ctx context.Context) ([]domain.UserUsdtPlaceholders, error) {
	var placeholders []domain.UserUsdtPlaceholders
	err := r.db.WithContext(ctx).
		Model(&domain.UserUsdtPlaceholders{}).
		Select("id", "placeholder").
		Where("status = ?", 0).
		Scan(&placeholders).Error
	return placeholders, err

}

func (r *UserUsdtPlaceholdersRepository) UpdateStatusByPlaceholder(ctx context.Context, placeholder string, status int64) error {
	return r.db.WithContext(ctx).Model(&domain.UserUsdtPlaceholders{}).
		Where("placeholder = ?", placeholder).
		Update("status", status).Error
}

func (r *UserUsdtPlaceholdersRepository) UpdateStatusByID(ctx context.Context, id int64, status int64) error {
	return r.db.WithContext(ctx).Model(&domain.UserUsdtPlaceholders{}).
		Where("id = ?", id).
		Update("status", status).Error
}
func (r *UserUsdtPlaceholdersRepository) GetFirstAvailable(ctx context.Context) (domain.UserUsdtPlaceholders, error) {
	var placeholder domain.UserUsdtPlaceholders
	err := r.db.WithContext(ctx).
		Model(&domain.UserUsdtPlaceholders{}).
		Select("id", "placeholder").
		Order("id ASC").
		Where("status = ?", 0).
		Take(&placeholder).Error
	return placeholder, err

}
func (r *UserUsdtPlaceholdersRepository) GetRandomAvailable(ctx context.Context) (domain.UserUsdtPlaceholders, error) {
	var placeholders domain.UserUsdtPlaceholders
	err := r.db.WithContext(ctx).Order("RAND()").
		Find(&placeholders, "status = ?", 0).Error
	return placeholders, err

	//err := r.db.WithContext(ctx).
	//	Find(&placeholders, "status = ?", 0).Error
	//return placeholders, err

}
