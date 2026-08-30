package repositories

import (
	"context"
	"strings"
	"time"
	"ushield_bot/internal/domain"

	"gorm.io/gorm"
)

type AddressCounterpartyRepo struct {
	db *gorm.DB
}

func NewAddressCounterpartyRepo(db *gorm.DB) *AddressCounterpartyRepo {
	return &AddressCounterpartyRepo{db: db}
}

func (r *AddressCounterpartyRepo) GetByAddress(ctx context.Context, network, address string) (domain.AddressCounterparty, error) {
	var record domain.AddressCounterparty
	err := r.db.WithContext(ctx).
		Where(
			"network = ? AND LOWER(counterparty_addr) = LOWER(?)",
			network, strings.TrimSpace(address),
		).
		First(&record).Error
	return record, err
}

func (r *AddressCounterpartyRepo) Upsert(ctx context.Context, cp *domain.AddressCounterparty) error {
	cp.CounterpartyAddr = strings.TrimSpace(cp.CounterpartyAddr)
	existing, err := r.GetByAddress(ctx, cp.Network, cp.CounterpartyAddr)
	if err == nil && existing.Id > 0 {
		updates := map[string]interface{}{
			"updated_at": time.Now(),
		}
		if cp.Label != "" {
			updates["label"] = cp.Label
		}
		if cp.Title != "" {
			updates["title"] = cp.Title
		}
		if cp.RiskScore > 0 {
			updates["risk_score"] = cp.RiskScore
		}
		if cp.Layer > 0 {
			updates["layer"] = cp.Layer
		}
		if cp.Malicious > 0 {
			updates["malicious"] = cp.Malicious
		}
		if cp.Track != "" {
			updates["track"] = cp.Track
		}
		if cp.Color != "" {
			updates["color"] = cp.Color
		}
		if cp.Dex > 0 {
			updates["dex"] = cp.Dex
		}
		return r.db.WithContext(ctx).
			Model(&domain.AddressCounterparty{}).
			Where("id = ?", existing.Id).
			Updates(updates).Error
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	return r.db.WithContext(ctx).Create(cp).Error
}

func (r *AddressCounterpartyRepo) ListByNetwork(ctx context.Context, network string) ([]domain.AddressCounterparty, error) {
	var list []domain.AddressCounterparty
	err := r.db.WithContext(ctx).
		Where("network = ?", network).
		Order("risk_score DESC, id DESC").
		Find(&list).Error
	return list, err
}
