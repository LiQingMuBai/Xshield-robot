package catfee

import (
	"context"
	"strconv"
	"strings"
	"time"
	"ushield_bot/internal/cache"
	"ushield_bot/internal/domain"
	"ushield_bot/internal/global"
	"ushield_bot/internal/infrastructure/repositories"
	thirdparty "ushield_bot/internal/infrastructure/thirdparty"
	"ushield_bot/internal/infrastructure/tools"
	logger "ushield_bot/internal/logger"
	"ushield_bot/internal/request"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func PromptCustodyAddressAdd(lang string, cache cache.Cache, db *gorm.DB, bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery) {
	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[lang]["catfee_custody_address_tips"]+"\n")
	msg.ParseMode = "HTML"
	bot.Send(msg)
	expiration := 1 * time.Minute // 短时间缓存空值
	//设置用户状态
	cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), callbackQuery.Data, expiration)
}

func AddCustodyAddress(lang string, cache cache.Cache, db *gorm.DB, bot *tgbotapi.BotAPI, message *tgbotapi.Message, trxfeeClient *thirdparty.TrxfeeClient) {
	address := message.Text
	chatID := message.Chat.ID
	if !tools.IsValidAddress(address) {
		msg := tgbotapi.NewMessage(chatID, "❌"+"<b>"+global.Translations[lang]["address_wrong_tips"]+"</b>"+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)
		return
	}

	userSmartTransactionAddressesRepo := repositories.NewUserSmartTransactionAddressesRepository(db)

	//要查下是否已经有绑定的地址

	total, _ := userSmartTransactionAddressesRepo.CountByChatID(context.Background(), chatID)

	if total >= 8 {
		msg := tgbotapi.NewMessage(chatID, "<b>"+global.Translations[lang]["catfee_energy_address_limit_tips"]+"</b>"+"\n")
		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔢"+global.Translations[lang]["smart_transaction_address_list"], "click_bundle_package_address_stats_ST"),
			),
		)
		msg.ReplyMarkup = inlineKeyboard
		msg.ParseMode = "HTML"
		bot.Send(msg)

		return
	}
	record, _ := userSmartTransactionAddressesRepo.GetActiveByAddress(context.Background(), address)

	if record.ID > 0 {
		msg := tgbotapi.NewMessage(chatID, "❌"+"<b>"+global.Translations[lang]["catfee_add_address_already_exit_tips"]+"</b>"+"\n")
		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔢"+global.Translations[lang]["smart_transaction_address_list"], "click_bundle_package_address_stats_ST"),
			),
		)
		msg.ReplyMarkup = inlineKeyboard
		msg.ParseMode = "HTML"
		bot.Send(msg)

		return
	}
	var userAddress domain.UserSmartTransactionAddresses

	userAddress.Status = "0"
	userAddress.CreatedAt = time.Now()
	userAddress.ChatID = strconv.FormatInt(chatID, 10)
	userAddress.QuotaMode = "UNLIMITED"
	userAddress.Address = address
	userSmartTransactionAddressesRepo.Create(context.Background(), &userAddress)

	//添加成功
	msg := tgbotapi.NewMessage(chatID, "✅"+"<b>"+global.Translations[lang]["address_added_success"]+"</b>"+"\n")
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(

		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔢"+global.Translations[lang]["smart_transaction_address_list"], "click_bundle_package_address_stats_ST"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard
	msg.ParseMode = "HTML"
	bot.Send(msg)

	//添加激活地址
	if err := trxfeeClient.Activation(address); err != nil {
		logger.Errorf("activate custody address failed: %v", err)
	}
}

func PromptCustodyAddressRemove(lang string, cache cache.Cache, db *gorm.DB, bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery) {
	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[lang]["energy_address_remove_tips"]+"\n")
	msg.ParseMode = "HTML"
	bot.Send(msg)
	expiration := 1 * time.Minute // 短时间缓存空值
	//设置用户状态
	cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), callbackQuery.Data, expiration)
}

func RemoveCustodyAddress(lang string, cache cache.Cache, db *gorm.DB, bot *tgbotapi.BotAPI, message *tgbotapi.Message, catfee *thirdparty.CatfeeService) {

	address := message.Text
	chatID := message.Chat.ID
	logger.Printf("删除用户id %d，地址 %s\v", chatID, address)
	if !tools.IsValidAddress(address) {
		msg := tgbotapi.NewMessage(chatID, "💬"+"<b>"+global.Translations[lang]["address_wrong_tips"]+"</b>"+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)
		return
	}

	userSmartTransactionAddressesRepo := repositories.NewUserSmartTransactionAddressesRepository(db)

	err := userSmartTransactionAddressesRepo.MarkDeletedByChatIDAndAddress(context.Background(), strconv.FormatInt(chatID, 10), address)

	if err != nil {
		logger.Errorf("删除地址失败%v", err)
	}

	code, err := catfee.MateOpenBasicDelete(address)

	if err != nil {
		logger.Errorf("catfee.MateOpenBasicDelete: %v", err)
	}
	logger.Printf("catfee删除状态 %d\n", code)

	msg := tgbotapi.NewMessage(chatID, "✅ "+"<b>"+global.Translations[lang]["address_deleted_success"]+"</b>"+"\n")
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(

		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔢"+global.Translations[lang]["smart_transaction_address_list"], "click_bundle_package_address_stats_ST"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard

	msg.ParseMode = "HTML"
	bot.Send(msg)

}

func DisableCustodyAddress(lang string, cache cache.Cache, db *gorm.DB, bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, catfee *thirdparty.CatfeeService) {

	address := callbackQuery.Message.Text
	chatID := callbackQuery.Message.Chat.ID
	logger.Printf("暂停用户id %d，地址 %s\v", chatID, address)
	userSmartTransactionAddressesRepo := repositories.NewUserSmartTransactionAddressesRepository(db)

	err := userSmartTransactionAddressesRepo.DisableByChatIDAndAddress(context.Background(), strconv.FormatInt(chatID, 10), address)

	if err != nil {
		logger.Errorf("暂停地址失败%v", err)
	}
	code, err := catfee.MateOpenBasicDisable(address)

	if err != nil {

	}
	logger.Errorf("catfee暂停地址失败 %d", code)

	msg := tgbotapi.NewMessage(chatID, "✅ "+"<b>"+global.Translations[lang]["address_deleted_success"]+"</b>"+"\n")
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(

		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔢"+global.Translations[lang]["smart_transaction_address_list"], "click_bundle_package_address_stats_ST"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard

	msg.ParseMode = "HTML"
	bot.Send(msg)

}

func EnableCustodyAddress(lang string, cache cache.Cache, db *gorm.DB, bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, catfee *thirdparty.CatfeeService) {

	address := callbackQuery.Message.Text
	chatID := callbackQuery.Message.Chat.ID
	logger.Printf("启用用户id %d，地址 %s\v", chatID, address)
	userSmartTransactionAddressesRepo := repositories.NewUserSmartTransactionAddressesRepository(db)

	err := userSmartTransactionAddressesRepo.EnableByChatIDAndAddress(context.Background(), strconv.FormatInt(chatID, 10), address)

	if err != nil {
		logger.Errorf("启用地址失败%v", err)
	}
	code, err := catfee.MateOpenBasicDisable(address)

	if err != nil {

	}
	logger.Errorf("catfee启用地址失败 %d", code)

	msg := tgbotapi.NewMessage(chatID, "✅ "+"<b>"+global.Translations[lang]["address_deleted_success"]+"</b>"+"\n")
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(

		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔢"+global.Translations[lang]["smart_transaction_address_list"], "click_bundle_package_address_stats_ST"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard

	msg.ParseMode = "HTML"
	bot.Send(msg)

}

func CatfeeAddressPrevePage(lang string, callbackQuery *tgbotapi.CallbackQuery, db *gorm.DB, bot *tgbotapi.BotAPI) (*global.DepositState, bool) {
	state := global.DepositStates[callbackQuery.Message.Chat.ID]

	if state != nil && state.CurrentPage == 1 {
		return nil, true
	}
	if state == nil {
		var state global.DepositState
		state.CurrentPage = 1
		global.DepositStates[callbackQuery.Message.Chat.ID] = &state
		userAddressDetectionRepo := repositories.NewUserSmartTransactionPackageSubscriptionsRepository(db)
		var info request.UserAddressDetectionSearch

		info.Page = 1
		info.PageSize = 5
		trxlist, _, _ := userAddressDetectionRepo.ListByChatIDPage(context.Background(), info, callbackQuery.Message.Chat.ID)

		var builder strings.Builder
		builder.WriteString("\n") // 添加分隔符
		for _, word := range trxlist {
			builder.WriteString("[")
			builder.WriteString(word.CreatedDate)
			builder.WriteString("]")
			builder.WriteString(" -")
			builder.WriteString(strings.ReplaceAll(word.BundleName, "笔", global.Translations[lang]["笔"]))

			builder.WriteString("\n") // 添加分隔符
		}

		// 去除最后一个空格
		result := strings.TrimSpace(builder.String())
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "🧾"+global.Translations[lang]["deduction_records"]+"\n\n "+
			result+"\n")
		msg.ParseMode = "HTML"
		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(global.Translations[lang]["prev"], "prev_bundle_package_page"),
				tgbotapi.NewInlineKeyboardButtonData(global.Translations[lang]["next"], "next_bundle_package_page"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔢"+global.Translations[lang]["smart_transaction_address_list"], "click_bundle_package_address_stats_ST"),
			),
		)
		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)
	} else {
		state.CurrentPage = state.CurrentPage - 1
		userAddressDetectionRepo := repositories.NewUserPackageSubscriptionsRepository(db)
		var info request.UserAddressDetectionSearch
		info.PageInfo.Page = state.CurrentPage
		info.PageSize = 5
		trxlist, _, _ := userAddressDetectionRepo.ListByChatIDPage(context.Background(), info, callbackQuery.Message.Chat.ID)
		var builder strings.Builder
		builder.WriteString("\n") // 添加分隔符
		for _, word := range trxlist {
			builder.WriteString("[")
			builder.WriteString(word.CreatedDate)
			builder.WriteString("]")
			builder.WriteString(" -")
			builder.WriteString(strings.ReplaceAll(word.BundleName, "笔", global.Translations[lang]["笔"]))

			builder.WriteString("\n") // 添加分隔符
		}

		// 去除最后一个空格
		result := strings.TrimSpace(builder.String())
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "🧾"+global.Translations[lang]["deduction_records"]+"\n\n "+
			result+"\n")
		msg.ParseMode = "HTML"
		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(global.Translations[lang]["prev"], "prev_bundle_package_page"),
				tgbotapi.NewInlineKeyboardButtonData(global.Translations[lang]["next"], "next_bundle_package_page"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[lang]["back_homepage"], "back_bundle_package"),
			),
		)
		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)
	}
	return state, false
}
