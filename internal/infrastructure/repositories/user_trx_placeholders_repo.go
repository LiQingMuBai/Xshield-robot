package repositories

import (
	"context"
	"gorm.io/gorm"
	"ushield_bot/internal/domain"
)

type UserTRXPlaceholdersRepository struct {
	db *gorm.DB
}

func NewUserTRXPlaceholdersRepository(db *gorm.DB) *UserTRXPlaceholdersRepository {
	return &UserTRXPlaceholdersRepository{
		db: db,
	}
}
func (r *UserTRXPlaceholdersRepository) ListAvailable(ctx context.Context) ([]domain.UserTRXPlaceholders, error) {
	var placeholders []domain.UserTRXPlaceholders
	err := r.db.WithContext(ctx).
		Model(&domain.UserTRXPlaceholders{}).
		Select("id", "placeholder").
		Where("status = ?", 0).
		Scan(&placeholders).Error
	return placeholders, err

}

func (r *UserTRXPlaceholdersRepository) UpdateStatusByID(ctx context.Context, id int64, status int64) error {
	return r.db.WithContext(ctx).Model(&domain.UserTRXPlaceholders{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *UserTRXPlaceholdersRepository) UpdateStatusByPlaceholder(ctx context.Context, placeholder string, status int64) error {
	return r.db.WithContext(ctx).Model(&domain.UserTRXPlaceholders{}).
		Where("placeholder = ?", placeholder).
		Update("status", status).Error
}

func (r *UserTRXPlaceholdersRepository) GetFirstAvailable(ctx context.Context) (domain.UserTRXPlaceholders, error) {
	var placeholder domain.UserTRXPlaceholders
	err := r.db.WithContext(ctx).
		Model(&domain.UserTRXPlaceholders{}).
		Select("id", "placeholder").
		Order("id ASC").
		Where("status = ?", 0).
		Take(&placeholder).Error
	return placeholder, err

}
func (r *UserTRXPlaceholdersRepository) GetRandomAvailable(ctx context.Context) (domain.UserTRXPlaceholders, error) {
	var placeholders domain.UserTRXPlaceholders
	err := r.db.WithContext(ctx).Order("RAND()").
		Find(&placeholders, "status = ?", 0).Error
	return placeholders, err

}
