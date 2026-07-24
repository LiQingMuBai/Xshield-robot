package service

import (
	"context"
	"strconv"
	"strings"
	"ushield_bot/internal/cache"
	"ushield_bot/internal/domain"
	"ushield_bot/internal/global"
	"ushield_bot/internal/infrastructure/repositories"
	trxfee "ushield_bot/internal/infrastructure/thirdparty"
	. "ushield_bot/internal/infrastructure/tools"
	logger "ushield_bot/internal/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func PayPremiumOrder(bot *tgbotapi.BotAPI, chatID int64, lang string, db *gorm.DB, catfeeClient *trxfee.CatfeeService, orderNO string) {
	usdtDepositRepo := repositories.NewUserUSDTDepositsRepository(db)
	record, _ := usdtDepositRepo.GetByOrderNo(context.Background(), orderNO)

	userRepo := repositories.NewUserRepository(db)
	user, _ := userRepo.GetByChatID(chatID)
	normalizeCommerceOrderPayer(&user)
	if !hasEnoughCommerceUSDT(user.Amount, record.Amount) {
		sendCommerceInsufficientBalance(bot, chatID, lang, user)
		return
	}

	balance, _ := SubtractStringNumbers(user.Amount, record.Amount, 1)
	user.Amount = balance
	_ = userRepo.Save(context.Background(), &user)

	tgOrderDB := repositories.NewTelegramPremiumOrderRepository(db)
	orderRecord, _ := tgOrderDB.GetByOrderNo(context.Background(), orderNO)
	tgOrderDB.UpdateStatusByOrderNo(context.Background(), orderNO, 1)
	if _, err := catfeeClient.Premium(orderRecord.TGUsername, orderRecord.Month); err != nil {
		logger.Errorf("pay premium order failed: %v", err)
		return
	}

	tips := strings.ReplaceAll(global.Translations[lang]["successfully_purchased_telegram"], "{month_package}", global.Translations[lang][orderRecord.Month+"_month_premium"])
	msg := tgbotapi.NewMessage(chatID, global.Translations[lang]["order_id"]+"：TOPUP-"+orderNO+" , "+tips)
	msg.ParseMode = "HTML"
	bot.Send(msg)

	usdtPlaceholderRepo := repositories.NewUserUsdtPlaceholdersRepository(db)
	usdtPlaceholderRepo.UpdateStatusByPlaceholder(context.Background(), record.Placeholder, 0)
	usdtDepositRepo.UpdateStatusByOrderNo(context.Background(), orderNO, 2)
}

func CancelPremiumOrder(bot *tgbotapi.BotAPI, cacheStore cache.Cache, chatID int64, lang string, db *gorm.DB, orderNO string) {
	usdtDepositRepo := repositories.NewUserUSDTDepositsRepository(db)
	record, _ := usdtDepositRepo.GetByOrderNo(context.Background(), orderNO)

	usdtPlaceholderRepo := repositories.NewUserUsdtPlaceholdersRepository(db)
	usdtPlaceholderRepo.UpdateStatusByPlaceholder(context.Background(), record.Placeholder, 0)
	usdtDepositRepo.UpdateStatusByOrderNo(context.Background(), orderNO, 2)

	msg := tgbotapi.NewMessage(chatID, global.Translations[lang]["order_id"]+"：TOPUP-"+orderNO+" , "+global.Translations[lang]["cancel_order_tips"])
	msg.ParseMode = "HTML"
	bot.Send(msg)

	deleteCommerceOrderMessage(bot, cacheStore, chatID)
}

func PayStarsOrder(bot *tgbotapi.BotAPI, chatID int64, lang string, db *gorm.DB, orderNO string) {
	usdtDepositRepo := repositories.NewUserUSDTDepositsRepository(db)
	record, _ := usdtDepositRepo.GetByOrderNo(context.Background(), orderNO)

	userRepo := repositories.NewUserRepository(db)
	user, _ := userRepo.GetByChatID(chatID)
	normalizeCommerceOrderPayer(&user)
	if !hasEnoughCommerceUSDT(user.Amount, record.Amount) {
		sendCommerceInsufficientBalance(bot, chatID, lang, user)
		return
	}

	balance, _ := SubtractStringNumbers(user.Amount, record.Amount, 1)
	user.Amount = balance
	_ = userRepo.Save(context.Background(), &user)

	tgOrderDB := repositories.NewTelegramStarsOrderRepository(db)
	orderRecord, _ := tgOrderDB.GetByOrderNo(context.Background(), orderNO)
	tgOrderDB.UpdateStatusByOrderNo(context.Background(), orderNO, 1)

	tips := strings.ReplaceAll(global.Translations[lang]["successfully_purchased_stars"], "{count}", orderRecord.Stars)
	msg := tgbotapi.NewMessage(chatID, global.Translations[lang]["order_id"]+"：TOPUP-"+orderNO+" , "+tips)
	msg.ParseMode = "HTML"
	bot.Send(msg)

	usdtPlaceholderRepo := repositories.NewUserUsdtPlaceholdersRepository(db)
	usdtPlaceholderRepo.UpdateStatusByPlaceholder(context.Background(), record.Placeholder, 0)
	usdtDepositRepo.UpdateStatusByOrderNo(context.Background(), orderNO, 2)
}

func CancelStarsOrder(bot *tgbotapi.BotAPI, cacheStore cache.Cache, chatID int64, lang string, db *gorm.DB, orderNO string) {
	usdtDepositRepo := repositories.NewUserUSDTDepositsRepository(db)
	record, _ := usdtDepositRepo.GetByOrderNo(context.Background(), orderNO)

	usdtPlaceholderRepo := repositories.NewUserUsdtPlaceholdersRepository(db)
	usdtPlaceholderRepo.UpdateStatusByPlaceholder(context.Background(), record.Placeholder, 0)
	usdtDepositRepo.UpdateStatusByOrderNo(context.Background(), orderNO, 2)

	msg := tgbotapi.NewMessage(chatID, global.Translations[lang]["order_id"]+"：TOPUP-"+orderNO+" , "+global.Translations[lang]["cancel_order_tips"])
	msg.ParseMode = "HTML"
	bot.Send(msg)

	deleteCommerceOrderMessage(bot, cacheStore, chatID)
}

func deleteCommerceOrderMessage(bot *tgbotapi.BotAPI, cacheStore cache.Cache, chatID int64) {
	prevMessageIDStr, _ := cacheStore.Get(strconv.FormatInt(chatID, 10) + "_order")
	prevMessageID, err := strconv.Atoi(prevMessageIDStr)
	if err != nil {
		return
	}
	bot.Request(tgbotapi.DeleteMessageConfig{ChatID: chatID, MessageID: prevMessageID})
}

func normalizeCommerceOrderPayer(user *domain.User) {
	if IsEmpty(user.Amount) {
		user.Amount = "0"
	}
	if IsEmpty(user.TronAmount) {
		user.TronAmount = "0"
	}
}

func hasEnoughCommerceUSDT(balance string, amount string) bool {
	flag, _ := CompareNumberStrings(balance, amount)
	return flag >= 0
}

func sendCommerceInsufficientBalance(bot *tgbotapi.BotAPI, chatID int64, lang string, user domain.User) {
	msg := tgbotapi.NewMessage(
		chatID,
		"<b>"+"🔍"+global.Translations[lang]["insufficient_balance_tips"]+"</b>"+"\n"+
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
