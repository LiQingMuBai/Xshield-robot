package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"
	"ushield_bot/internal/domain"
	"ushield_bot/internal/infrastructure/repositories"
	trxfee "ushield_bot/internal/infrastructure/thirdparty"
	. "ushield_bot/internal/infrastructure/tools"

	"gorm.io/gorm"
)

var (
	ErrDispatchForbidden         = errors.New("dispatch forbidden")
	ErrDispatchInsufficientTimes = errors.New("dispatch insufficient times")
	ErrDispatchInvalidAddress    = errors.New("dispatch invalid address")
	ErrSmartDispatchForbidden    = errors.New("smart dispatch forbidden")
)

type EnergyDispatchService struct {
	db           *gorm.DB
	trxfeeURL    string
	trxfeeAPIKey string
	trxfeeSecret string
	catfeeClient *trxfee.CatfeeService
}

type DispatchResult struct {
	Address        string
	RemainingTimes int64
	UserLang       string
}

func NewEnergyDispatchService(db *gorm.DB, trxfeeURL, trxfeeAPIKey, trxfeeSecret string, catfeeClient *trxfee.CatfeeService) *EnergyDispatchService {
	return &EnergyDispatchService{
		db:           db,
		trxfeeURL:    trxfeeURL,
		trxfeeAPIKey: trxfeeAPIKey,
		trxfeeSecret: trxfeeSecret,
		catfeeClient: catfeeClient,
	}
}

func (s *EnergyDispatchService) ToggleSmartDispatch(ctx context.Context, subscriptionID string, chatID int64, enabled bool) (string, error) {
	subscriptionRepo := repositories.NewUserSmartTransactionPackageSubscriptionsRepository(s.db)
	record, err := subscriptionRepo.GetFullByID(ctx, subscriptionID)
	if err != nil {
		return "", err
	}
	if record.ChatID != chatID {
		return "", ErrSmartDispatchForbidden
	}

	nextStatus := int64(1)
	if enabled {
		nextStatus = 2
	}
	if err := subscriptionRepo.UpdateStatusByID(ctx, subscriptionID, nextStatus); err != nil {
		return "", err
	}

	trxfeeClient := trxfee.NewTrxfeeClient(s.trxfeeURL, s.trxfeeAPIKey, s.trxfeeSecret)
	if err := trxfeeClient.EnableTimesOrder(record.Address); err != nil {
		return "", err
	}

	userRepo := repositories.NewUserRepository(s.db)
	user, err := userRepo.GetByChatID(chatID)
	if err != nil {
		return "", err
	}

	return user.Lang, nil
}

func (s *EnergyDispatchService) DispatchFromPackageAddress(ctx context.Context, packageAddressID string, chatID int64, times int) (*DispatchResult, error) {
	addressRepo := repositories.NewUserOperationPackageAddressesRepo(s.db)
	record, err := addressRepo.GetByID(ctx, packageAddressID)
	if err != nil {
		return nil, err
	}
	if record.ChatID != chatID {
		return nil, ErrDispatchForbidden
	}

	var result *DispatchResult
	for i := 0; i < times; i++ {
		result, err = s.dispatchWithUserBundleTimes(ctx, chatID, record.Address)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (s *EnergyDispatchService) DispatchToManualAddress(ctx context.Context, chatID int64, address string) (*DispatchResult, error) {
	return s.dispatchWithUserBundleTimes(ctx, chatID, address)
}

func (s *EnergyDispatchService) DispatchFromSubscription(ctx context.Context, bundleID, address string, chatID int64) (*DispatchResult, error) {
	if len(address) <= 10 {
		return nil, ErrDispatchInvalidAddress
	}

	subscriptionRepo := repositories.NewUserPackageSubscriptionsRepository(s.db)
	record, err := subscriptionRepo.GetByID(ctx, bundleID)
	if err != nil {
		return nil, err
	}

	restTimes := record.Times - 1
	if restTimes < 0 {
		return nil, ErrDispatchInsufficientTimes
	}
	if err := subscriptionRepo.UpdateRemainingTimes(ctx, record.Id, restTimes); err != nil {
		return nil, err
	}

	userRepo := repositories.NewUserRepository(s.db)
	user, err := userRepo.GetByChatID(chatID)
	if err != nil {
		return nil, err
	}

	if err := s.createAndSendEnergyOrder(ctx, chatID, address); err != nil {
		return nil, err
	}

	return &DispatchResult{
		Address:        address,
		RemainingTimes: restTimes,
		UserLang:       user.Lang,
	}, nil
}

func (s *EnergyDispatchService) dispatchWithUserBundleTimes(ctx context.Context, chatID int64, address string) (*DispatchResult, error) {
	if len(address) <= 10 {
		return nil, ErrDispatchInvalidAddress
	}

	userRepo := repositories.NewUserRepository(s.db)
	user, err := userRepo.GetByChatID(chatID)
	if err != nil {
		return nil, err
	}

	remainingTimes := user.BundleTimes - 1
	if user.BundleTimes <= 0 || remainingTimes < 0 {
		return nil, ErrDispatchInsufficientTimes
	}

	if err := userRepo.UpdateBundleTimes(remainingTimes, chatID); err != nil {
		return nil, err
	}

	if err := s.createAndSendEnergyOrder(ctx, chatID, address); err != nil {
		return nil, err
	}

	return &DispatchResult{
		Address:        address,
		RemainingTimes: remainingTimes,
		UserLang:       user.Lang,
	}, nil
}

func (s *EnergyDispatchService) createAndSendEnergyOrder(ctx context.Context, chatID int64, address string) error {
	orderNo, err := GenerateOrderID(address, 4)
	if err != nil {
		return err
	}

	sysOrder := domain.UserEnergyOrders{
		OrderNo:     orderNo,
		TxId:        "",
		FromAddress: address,
		Amount:      65000,
		ChatId:      strconv.FormatInt(chatID, 10),
	}

	ueoRepo := repositories.NewUserEnergyOrdersRepo(s.db)
	if err := ueoRepo.Create(ctx, &sysOrder); err != nil {
		return err
	}

	flag, err := s.inTrxfeeTimeRange()
	if err != nil {
		return err
	}

	if flag {
		trxfeeClient := trxfee.NewTrxfeeClient(s.trxfeeURL, s.trxfeeAPIKey, s.trxfeeSecret)
		if err := trxfeeClient.Order(orderNo, address, 65_000); err != nil {
			return err
		}
		return nil
	}

	s.catfeeClient.Order(address)
	return nil
}

func (s *EnergyDispatchService) inTrxfeeTimeRange() (bool, error) {
	sysDictionariesRepo := repositories.NewSysDictionariesRepo(s.db)
	timeRange, err := sysDictionariesRepo.GetDictionaryDetail("time_range")
	if err != nil {
		return false, err
	}

	parts := strings.Split(timeRange, "-")
	if len(parts) != 2 {
		return false, errors.New("invalid time_range format")
	}

	startHour, err := strconv.Atoi(parts[0])
	if err != nil {
		return false, err
	}
	endHour, err := strconv.Atoi(parts[1])
	if err != nil {
		return false, err
	}

	hour := time.Now().Hour()
	return hour >= startHour && hour <= endHour, nil
}
