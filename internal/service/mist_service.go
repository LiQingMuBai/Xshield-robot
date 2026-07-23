package service

import (
	"context"
	"errors"
	"log"
	"strings"
	"ushield_bot/internal/cache"
	"ushield_bot/internal/domain"
	"ushield_bot/internal/global"
	"ushield_bot/internal/handler"
	"ushield_bot/internal/infrastructure/repositories"
	. "ushield_bot/internal/infrastructure/tools"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

var (
	ErrAddressDetectionInvalidAddress = errors.New("address detection invalid address")
)

const mistTemporaryUnavailableText = "目前我们的AI大数据态势感知系统正在对区块链地址风险进行高强度的实时分析，导致“地址风险监测”功能出现短暂的“思考卡顿”（临时失效）。"

type AddressDetectionService struct {
	db         *gorm.DB
	mistCookie string
}

type addressDetectionCosts struct {
	TRX  string
	USDT string
}

type AddressDetectionResult struct {
	Text                string
	ChargeFeedback      string
	InsufficientBalance bool
	User                domain.User
}

func NewAddressDetectionService(db *gorm.DB, mistCookie string) *AddressDetectionService {
	return &AddressDetectionService{
		db:         db,
		mistCookie: mistCookie,
	}
}

func ExtractSlowMistRiskQuery(lang string, cacheStore cache.Cache, message *tgbotapi.Message, db *gorm.DB, mistCookie string, bot *tgbotapi.BotAPI) {
	detectionService := NewAddressDetectionService(db, mistCookie)
	result, err := detectionService.Detect(context.Background(), lang, cacheStore, message.Chat.ID, strings.TrimSpace(message.Text))
	if err == ErrAddressDetectionInvalidAddress {
		msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[lang]["address_wrong_tips"]+"</b>"+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)
		return
	}
	if err != nil {
		log.Printf("address detection err: %v", err)
		return
	}

	if result.InsufficientBalance {
		sendAddressDetectionInsufficientBalance(bot, message.Chat.ID, lang, result.User)
		return
	}

	sendAddressDetectionResult(bot, message.Chat.ID, lang, result.Text)
	if result.ChargeFeedback != "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, result.ChargeFeedback)
		msg.ParseMode = "HTML"
		bot.Send(msg)
	}
}

func (s *AddressDetectionService) Detect(ctx context.Context, lang string, cacheStore cache.Cache, chatID int64, address string) (*AddressDetectionResult, error) {
	if !IsValidAddress(address) && !IsValidEthereumAddress(address) {
		return nil, ErrAddressDetectionInvalidAddress
	}

	userRepo := repositories.NewUserRepository(s.db)
	user, err := userRepo.GetByChatID(chatID)
	if err != nil {
		return nil, err
	}
	normalizeAddressDetectionUser(&user)

	costs, err := s.loadCosts()
	if err != nil {
		return nil, err
	}

	result := &AddressDetectionResult{User: user}
	if user.Times >= 0 && !hasEnoughAddressDetectionBalance(user, costs) {
		result.InsufficientBalance = true
		return result, nil
	}

	text, temporaryFailure, err := s.buildDetectionText(lang, cacheStore, address)
	if err != nil {
		return nil, err
	}
	result.Text = text

	if err := userRepo.UpdateDispatchTimesByChatID(1, chatID); err != nil {
		log.Printf("update address detection times err: %v", err)
	}

	if user.Times >= 0 && !temporaryFailure {
		feedback, chargeErr := s.chargeAndRecord(ctx, &user, chatID, address, costs)
		if chargeErr != nil {
			return nil, chargeErr
		}
		result.ChargeFeedback = feedback
		result.User = user
	}

	return result, nil
}

func (s *AddressDetectionService) loadCosts() (*addressDetectionCosts, error) {
	dictRepo := repositories.NewSysDictionariesRepo(s.db)
	trxCost, err := dictRepo.GetDictionaryDetail("address_detection_cost")
	if err != nil {
		return nil, err
	}
	usdtCost, err := dictRepo.GetDictionaryDetail("address_detection_cost_usdt")
	if err != nil {
		return nil, err
	}
	return &addressDetectionCosts{
		TRX:  trxCost,
		USDT: usdtCost,
	}, nil
}

func (s *AddressDetectionService) buildDetectionText(lang string, cacheStore cache.Cache, address string) (string, bool, error) {
	symbol, graphCoin := addressDetectionSymbol(address)
	addressInfo, err := handler.GetAddressInfo(symbol, address, s.mistCookie)
	if err != nil || !addressInfo.Success {
		return mistTemporaryUnavailableText, true, nil
	}

	text := handler.BuildRiskSummaryText(lang, cacheStore, addressInfo)
	addressProfile := handler.GetAddressProfile(symbol, address, s.mistCookie)
	labelAddressList := handler.ListRiskAddresses(graphCoin, address, s.mistCookie)

	text += global.Translations[lang]["balance"] + "：" + addressProfile.BalanceUsd + "\n"
	text += global.Translations[lang]["total_received"] + "：" + addressProfile.TotalReceivedUsd + "\n"
	text += global.Translations[lang]["total_spent"] + "：" + addressProfile.TotalSpentUsd + "\n"
	text += global.Translations[lang]["first_tx_time"] + "：" + addressProfile.FirstTxTime + "\n"
	text += global.Translations[lang]["last_tx_time"] + "：" + addressProfile.LastTxTime + "\n"
	text += global.Translations[lang]["tx_count"] + "：" + addressProfile.TxCount + "\n"
	text += global.Translations[lang]["counterparty_analysis"] + "：" + "\n"
	text += addressDetectionCounterpartyText(lang, labelAddressList)
	text += global.Translations[lang]["ushield_tips"] + "\n"

	return text, false, nil
}

func (s *AddressDetectionService) chargeAndRecord(ctx context.Context, user *domain.User, chatID int64, address string, costs *addressDetectionCosts) (string, error) {
	userRepo := repositories.NewUserRepository(s.db)
	detectionRepo := repositories.NewUserAddressDetectionRepository(s.db)

	record := domain.UserAddressDetection{
		Status:  1,
		ChatID:  chatID,
		Address: address,
		Network: addressDetectionNetwork(address),
	}

	if CompareStringsWithFloat(user.TronAmount, costs.TRX, 1) {
		tronAmount, err := SubtractStringNumbers(user.TronAmount, costs.TRX, 1)
		if err != nil {
			return "", err
		}
		user.TronAmount = tronAmount
		record.Amount = costs.TRX
		if err := userRepo.Save(ctx, user); err != nil {
			return "", err
		}
		if err := detectionRepo.Create(ctx, &record); err != nil {
			return "", err
		}
		return "✅🧾" + global.Translations[user.Lang]["address_detection_payment_tips"] + costs.TRX + " TRX \n\n", nil
	}

	amount, err := SubtractStringNumbers(user.Amount, costs.USDT, 1)
	if err != nil {
		return "", err
	}
	user.Amount = amount
	record.Amount = costs.USDT
	if err := userRepo.Save(ctx, user); err != nil {
		return "", err
	}
	if err := detectionRepo.Create(ctx, &record); err != nil {
		return "", err
	}
	return "✅🧾" + global.Translations[user.Lang]["address_detection_payment_tips"] + costs.USDT + " USDT \n\n", nil
}

func sendAddressDetectionResult(bot *tgbotapi.BotAPI, chatID int64, lang, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔍"+global.Translations[lang]["detect_again"], "back_address_detection_home"),
		),
	)
	bot.Send(msg)
}

func sendAddressDetectionInsufficientBalance(bot *tgbotapi.BotAPI, chatID int64, lang string, user domain.User) {
	msg := tgbotapi.NewMessage(
		chatID,
		"<b>"+"🔍"+global.Translations[lang]["daily_free_limit"]+"</b>"+"\n"+
			"🆔"+global.Translations[lang]["user_id"]+": <code>"+user.Associates+"</code>\n"+
			"👤"+global.Translations[lang]["username"]+": @"+user.Username+"\n"+
			"💰"+global.Translations[lang]["balance"]+"\n"+
			"- TRX：   "+user.TronAmount+"\n"+
			"-  USDT："+user.Amount,
	)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💵"+global.Translations[lang]["deposit"], "deposit_amount"),
		),
	)
	bot.Send(msg)
}

func normalizeAddressDetectionUser(user *domain.User) {
	if IsEmpty(user.Amount) {
		user.Amount = "0"
	}
	if IsEmpty(user.TronAmount) {
		user.TronAmount = "0"
	}
	if IsEmpty(user.Lang) {
		user.Lang = "zh"
	}
}

func hasEnoughAddressDetectionBalance(user domain.User, costs *addressDetectionCosts) bool {
	return CompareStringsWithFloat(user.Amount, costs.USDT, 1) || CompareStringsWithFloat(user.TronAmount, costs.TRX, 1)
}

func addressDetectionSymbol(address string) (string, string) {
	if IsValidEthereumAddress(address) {
		return "USDT-ERC20", "ETH"
	}
	return "USDT-TRC20", "USDT-TRC20"
}

func addressDetectionNetwork(address string) string {
	if IsValidEthereumAddress(address) {
		return "Ethereum"
	}
	return "Tron"
}

func addressDetectionCounterpartyText(lang string, labelAddressList handler.LabeledAddressList) string {
	var builder strings.Builder
	for _, data := range labelAddressList.GraphDic.NodeList {
		shortTitle := abbreviateAddressDetectionTitle(data.Title)
		switch {
		case strings.Contains(data.Label, "huione"):
			builder.WriteString(shortTitle + global.Translations[lang]["huione"] + "\n")
		case strings.Contains(data.Label, "Theft"):
			builder.WriteString(shortTitle + global.Translations[lang]["theft"] + "\n")
		case strings.Contains(data.Label, "Drainer"):
			builder.WriteString(shortTitle + global.Translations[lang]["scam"] + "\n")
		case strings.Contains(data.Label, "Banned"):
			builder.WriteString(shortTitle + global.Translations[lang]["sanctioned"] + "\n")
		}
	}
	return builder.String()
}

func abbreviateAddressDetectionTitle(title string) string {
	runes := []rune(title)
	if len(runes) <= 10 {
		return title
	}
	if len(runes) <= 34 {
		return string(runes[:5]) + "..." + string(runes[len(runes)-5:])
	}
	return string(runes[:5]) + "..." + string(runes[29:34])
}
