package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"ushield_bot/internal/domain"
	"ushield_bot/internal/infrastructure/repositories"
	. "ushield_bot/internal/infrastructure/tools"

	"gorm.io/gorm"
)

var (
	ErrFreezeAlertInvalidAddress      = errors.New("freeze alert invalid address")
	ErrFreezeAlertInsufficientBalance = errors.New("freeze alert insufficient balance")
	ErrFreezeAlertMonitorNotFound     = errors.New("freeze alert monitor not found")
	ErrFreezeAlertForbidden           = errors.New("freeze alert forbidden")
)

type FreezeAlertService struct {
	db *gorm.DB
}

type FreezeAlertPreview struct {
	Address   string
	TRXPrice  string
	USDTPrice string
}

type FreezeAlertConfirmResult struct {
	Address string
	Network string
	Amount  string
}

type FreezeAlertMonitorItem struct {
	ID      int64
	Address string
}

type FreezeAlertClosePreview struct {
	ID            int64
	Address       string
	RemainingDays int64
}

func NewFreezeAlertService(db *gorm.DB) *FreezeAlertService {
	return &FreezeAlertService{db: db}
}

func (s *FreezeAlertService) Start(chatID int64) error {
	userRepo := repositories.NewUserRepository(s.db)
	user, err := userRepo.GetByUserID(chatID)
	if err != nil {
		return err
	}

	trxPrice, usdtPrice, err := s.loadPrices()
	if err != nil {
		return err
	}

	if !hasEnoughFreezeAlertBalance(user, trxPrice, usdtPrice) {
		return ErrFreezeAlertInsufficientBalance
	}

	return nil
}

func (s *FreezeAlertService) Preview(address string) (*FreezeAlertPreview, error) {
	normalizedAddress := strings.TrimSpace(address)
	if !IsValidAddress(normalizedAddress) && !IsValidEthereumAddress(normalizedAddress) {
		return nil, ErrFreezeAlertInvalidAddress
	}

	trxPrice, usdtPrice, err := s.loadPrices()
	if err != nil {
		return nil, err
	}

	return &FreezeAlertPreview{
		Address:   normalizedAddress,
		TRXPrice:  trxPrice,
		USDTPrice: usdtPrice,
	}, nil
}

func (s *FreezeAlertService) Confirm(ctx context.Context, chatID int64, address string) (*FreezeAlertConfirmResult, error) {
	preview, err := s.Preview(address)
	if err != nil {
		return nil, err
	}

	userRepo := repositories.NewUserRepository(s.db)
	user, err := userRepo.GetByUserID(chatID)
	if err != nil {
		return nil, err
	}
	normalizeUserBalances(&user)

	if !hasEnoughFreezeAlertBalance(user, preview.TRXPrice, preview.USDTPrice) {
		return nil, ErrFreezeAlertInsufficientBalance
	}

	result := &FreezeAlertConfirmResult{
		Address: preview.Address,
		Network: freezeAlertNetwork(preview.Address),
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txUserRepo := repositories.NewUserRepository(tx)
		txEventRepo := repositories.NewUserAddressMonitorEventRepo(tx)

		if CompareStringsWithFloat(user.TronAmount, preview.TRXPrice, 1) {
			rest, calcErr := SubtractStringNumbers(user.TronAmount, preview.TRXPrice, 1)
			if calcErr != nil {
				return calcErr
			}
			user.TronAmount = rest
			result.Amount = preview.TRXPrice + " TRX"
		} else {
			rest, calcErr := SubtractStringNumbers(user.Amount, preview.USDTPrice, 1)
			if calcErr != nil {
				return calcErr
			}
			user.Amount = rest
			result.Amount = preview.USDTPrice + " USDT"
		}

		if err := txUserRepo.Update2(ctx, &user); err != nil {
			return err
		}

		event := domain.UserAddressMonitorEvent{
			ChatID:  chatID,
			Status:  1,
			Address: preview.Address,
			Network: result.Network,
			Days:    1,
			Amount:  result.Amount,
		}

		return txEventRepo.Create(ctx, &event)
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *FreezeAlertService) ListActive(chatID int64) ([]FreezeAlertMonitorItem, error) {
	eventRepo := repositories.NewUserAddressMonitorEventRepo(s.db)
	events, err := eventRepo.Query(context.Background(), chatID)
	if err != nil {
		return nil, err
	}

	items := make([]FreezeAlertMonitorItem, 0, len(events))
	for _, event := range events {
		items = append(items, FreezeAlertMonitorItem{
			ID:      event.Id,
			Address: event.Address,
		})
	}

	return items, nil
}

func (s *FreezeAlertService) PreviewClose(chatID int64, eventID string) (*FreezeAlertClosePreview, error) {
	eventRepo := repositories.NewUserAddressMonitorEventRepo(s.db)
	event, err := eventRepo.Find(context.Background(), eventID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFreezeAlertMonitorNotFound
		}
		return nil, err
	}
	if event.ChatID != chatID {
		return nil, ErrFreezeAlertForbidden
	}

	remainingDays := int64(30) - event.Days
	if remainingDays < 0 {
		remainingDays = 0
	}

	return &FreezeAlertClosePreview{
		ID:            event.Id,
		Address:       event.Address,
		RemainingDays: remainingDays,
	}, nil
}

func (s *FreezeAlertService) Close(ctx context.Context, chatID int64, eventID string) error {
	preview, err := s.PreviewClose(chatID, eventID)
	if err != nil {
		return err
	}

	eventRepo := repositories.NewUserAddressMonitorEventRepo(s.db)
	return eventRepo.Close(ctx, formatFreezeAlertID(preview.ID))
}

func (s *FreezeAlertService) loadPrices() (string, string, error) {
	dictRepo := repositories.NewSysDictionariesRepo(s.db)
	trxPrice, err := dictRepo.GetDictionaryDetail("server_trx_price")
	if err != nil {
		return "", "", err
	}
	usdtPrice, err := dictRepo.GetDictionaryDetail("server_usdt_price")
	if err != nil {
		return "", "", err
	}
	return trxPrice, usdtPrice, nil
}

func normalizeUserBalances(user *domain.User) {
	if IsEmpty(user.Amount) {
		user.Amount = "0"
	}
	if IsEmpty(user.TronAmount) {
		user.TronAmount = "0"
	}
}

func hasEnoughFreezeAlertBalance(user domain.User, trxPrice, usdtPrice string) bool {
	normalizeUserBalances(&user)
	return CompareStringsWithFloat(user.TronAmount, trxPrice, 1) || CompareStringsWithFloat(user.Amount, usdtPrice, 1)
}

func freezeAlertNetwork(address string) string {
	switch {
	case IsValidEthereumAddress(address):
		return "Ethereum"
	case IsValidAddress(address):
		return "Tron"
	default:
		return ""
	}
}

func formatFreezeAlertID(id int64) string {
	return strconv.FormatInt(id, 10)
}
