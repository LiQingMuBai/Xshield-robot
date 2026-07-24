package catfee

import (
	"context"
	"strings"
	logger "ushield_bot/internal/logger"
	"ushield_bot/internal/request"

	"ushield_bot/internal/global"
	"ushield_bot/internal/infrastructure/repositories"
	. "ushield_bot/internal/infrastructure/tools"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func MenuNavigateCatfeeSmartTransactionPlans(lang string, db *gorm.DB, chatID int64, bot *tgbotapi.BotAPI, token string) {

	bundlesRepo := repositories.NewUserSmartTransactionBundlesRepository(db)
	trxlist, err := bundlesRepo.ListActiveByToken(context.Background(), token)
	if err != nil {
	}
	var allButtons []tgbotapi.InlineKeyboardButton
	var extraButtons []tgbotapi.InlineKeyboardButton
	var onlyButtons []tgbotapi.InlineKeyboardButton
	var keyboard [][]tgbotapi.InlineKeyboardButton
	for _, trx := range trxlist {
		allButtons = append(allButtons, tgbotapi.NewInlineKeyboardButtonData("🛒"+strings.ReplaceAll(trx.Name, "笔", global.Translations[lang]["笔"]), CombineInt64AndString("ST_bundle_", trx.Id)))
	}

	if token == "TRX" {
		onlyButtons = append(onlyButtons,
			tgbotapi.NewInlineKeyboardButtonData("🔁"+global.Translations[lang]["transaction_plans_usdt_payment"], "click_switch_usdt_ST"),
		)
	}
	if token == "USDT" {
		onlyButtons = append(onlyButtons,
			tgbotapi.NewInlineKeyboardButtonData("🔁"+global.Translations[lang]["transaction_plans_trx_payment"], "click_switch_trx_ST"),
		)
	}

	extraButtons = append(extraButtons,
		tgbotapi.NewInlineKeyboardButtonData("🔢"+global.Translations[lang]["smart_transaction_address_list"], "click_bundle_package_address_stats_ST"),
		tgbotapi.NewInlineKeyboardButtonData("📜"+global.Translations[lang]["billing"], "click_bundle_package_cost_records_ST"),
	)

	for i := 0; i < len(allButtons); i += 2 {
		end := i + 2
		if end > len(allButtons) {
			end = len(allButtons)
		}
		row := allButtons[i:end]
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(row...))
	}
	for i := 0; i < len(onlyButtons); i += 1 {
		end := i + 1
		if end > len(onlyButtons) {
			end = len(onlyButtons)
		}
		row := onlyButtons[i:end]
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(row...))
	}

	for i := 0; i < len(extraButtons); i += 2 {
		end := i + 2
		if end > len(extraButtons) {
			end = len(extraButtons)
		}
		row := extraButtons[i:end]
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(row...))
	}

	// 3. 创建键盘标记
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	userRepo := repositories.NewUserRepository(db)
	user, _ := userRepo.GetByChatID(chatID)
	if IsEmpty(user.Amount) {
		user.Amount = "0"
	}

	if IsEmpty(user.TronAmount) {
		user.TronAmount = "0"
	}

	msg := tgbotapi.NewMessage(chatID, "<b>"+global.Translations[lang]["smart_transaction_plans_head"]+"</b>"+"\n\n"+global.Translations[lang]["smart_transaction_plans_tips"])
	msg.ReplyMarkup = inlineKeyboard
	msg.ParseMode = "HTML"

	bot.Send(msg)
}
func ShowSmartTransactionBundlePackageMenu(lang string, db *gorm.DB, chatID int64, bot *tgbotapi.BotAPI, token string) {
	//bundlesRepo := repositories.NewUserOperationBundlesRepository(db)
	bundlesRepo := repositories.NewUserSmartTransactionBundlesRepository(db)

	trxlist, err := bundlesRepo.ListActiveByToken(context.Background(), token)

	if err != nil {

	}

	var allButtons []tgbotapi.InlineKeyboardButton
	var extraButtons []tgbotapi.InlineKeyboardButton
	var onlyButtons []tgbotapi.InlineKeyboardButton
	var keyboard [][]tgbotapi.InlineKeyboardButton
	for _, trx := range trxlist {

		allButtons = append(allButtons, tgbotapi.NewInlineKeyboardButtonData(strings.ReplaceAll(trx.Name, "笔", global.Translations[lang]["笔"]), CombineInt64AndString("ST_bundle_", trx.Id)))
	}

	if token == "TRX" {
		onlyButtons = append(onlyButtons,
			tgbotapi.NewInlineKeyboardButtonData("🔁"+global.Translations[lang]["transaction_plans_usdt_payment"], "click_switch_usdt_ST"),
		)
	}
	if token == "USDT" {
		onlyButtons = append(onlyButtons,
			tgbotapi.NewInlineKeyboardButtonData("🔁"+global.Translations[lang]["transaction_plans_trx_payment"], "click_switch_trx_ST"),
		)
	}

	extraButtons = append(extraButtons,
		tgbotapi.NewInlineKeyboardButtonData("🔢"+global.Translations[lang]["smart_transaction_address_list"], "click_bundle_package_address_stats_ST"),
		//tgbotapi.NewInlineKeyboardButtonData("➕"+global.Translations[lang]["add_address"], "click_bundle_package_address_management"),
		tgbotapi.NewInlineKeyboardButtonData("📜"+global.Translations[lang]["billing"], "click_bundle_package_cost_records_ST"),
	)

	for i := 0; i < len(allButtons); i += 2 {
		end := i + 2
		if end > len(allButtons) {
			end = len(allButtons)
		}
		row := allButtons[i:end]
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(row...))
	}
	for i := 0; i < len(onlyButtons); i += 1 {
		end := i + 1
		if end > len(onlyButtons) {
			end = len(onlyButtons)
		}
		row := onlyButtons[i:end]
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(row...))
	}

	for i := 0; i < len(extraButtons); i += 2 {
		end := i + 2
		if end > len(extraButtons) {
			end = len(extraButtons)
		}
		row := extraButtons[i:end]
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(row...))
	}

	// 3. 创建键盘标记
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	userRepo := repositories.NewUserRepository(db)
	user, _ := userRepo.GetByChatID(chatID)
	if IsEmpty(user.Amount) {
		user.Amount = "0"
	}

	if IsEmpty(user.TronAmount) {
		user.TronAmount = "0"
	}

	msg := tgbotapi.NewMessage(chatID, "<b>"+global.Translations[lang]["smart_transaction_plans_head"]+"</b>"+"\n\n"+global.Translations[lang]["smart_transaction_plans_tips"])
	msg.ReplyMarkup = inlineKeyboard
	msg.ParseMode = "HTML"

	bot.Send(msg)
}

func BuildSmartTransactionPackageCostRecordsMessage(lang string, db *gorm.DB, callbackQuery *tgbotapi.CallbackQuery) tgbotapi.MessageConfig {

	//userAddressDetectionRepo := repositories.NewUserPackageSubscriptionsRepository(db)
	userAddressDetectionRepo := repositories.NewUserSmartTransactionPackageSubscriptionsRepository(db)
	var info request.UserAddressDetectionSearch

	info.Page = 1
	info.PageSize = 5
	trxlist, total, err := userAddressDetectionRepo.ListByChatIDPage(context.Background(), info, callbackQuery.Message.Chat.ID)
	if err != nil {

		logger.Error("能量笔数套餐空", err)
	}
	var builder strings.Builder
	if total > 0 {
		//- [6.29] +3000 TRX（订单 #TOPUP-92308）
		for _, word := range trxlist {
			builder.WriteString("[")
			builder.WriteString(word.CreatedDate)
			builder.WriteString("]")
			builder.WriteString(" -")
			builder.WriteString(strings.ReplaceAll(word.BundleName, "笔", global.Translations[lang]["笔"]))

			//builder.WriteString(" TRX ")
			//builder.WriteString(" （能量笔数套餐）")

			builder.WriteString("\n") // 添加分隔符
		}
	} else {
		builder.WriteString("\n") // 添加分隔符
	}

	// 去除最后一个空格
	result := strings.TrimSpace(builder.String())

	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "🧾"+global.Translations[lang]["deduction_records"]+"\n\n "+
		result+"\n")
	msg.ParseMode = "HTML"
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[lang]["prev"], "prev_st_bundle_package_page"),
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[lang]["next"], "next_st_bundle_package_page"),
		),
		tgbotapi.NewInlineKeyboardRow(
			//tgbotapi.NewInlineKeyboardButtonData("解绑地址", "free_monitor_address"),
			tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[lang]["back_homepage"], "back_bundle_package_ST"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard
	return msg
}

func ShowNextBundlePackagePage(lang string, callbackQuery *tgbotapi.CallbackQuery, db *gorm.DB, bot *tgbotapi.BotAPI) bool {
	state := global.DepositStates[callbackQuery.Message.Chat.ID]
	if state == nil {
		var state2 global.DepositState
		state2.CurrentPage = 1
		state = &state2
	}
	//if state != nil && state.CurrentPage > 1 {
	state.CurrentPage = state.CurrentPage + 1
	userAddressDetectionRepo := repositories.NewUserSmartTransactionPackageSubscriptionsRepository(db)
	var info request.UserAddressDetectionSearch
	info.PageInfo.Page = state.CurrentPage
	info.PageInfo.PageSize = 5
	trxlist, total, _ := userAddressDetectionRepo.ListByChatIDPage(context.Background(), info, callbackQuery.Message.Chat.ID)

	totalPages := (total + 5 - 1) / 5
	if int64(state.CurrentPage) > totalPages {
		state.CurrentPage = totalPages
		return true
	}
	var builder strings.Builder
	builder.WriteString("\n") // 添加分隔符
	//- [6.29] +3000 TRX（订单 #TOPUP-92308）
	for _, word := range trxlist {
		builder.WriteString("[")
		builder.WriteString(word.CreatedDate)
		builder.WriteString("]")
		builder.WriteString(" -")
		builder.WriteString(strings.ReplaceAll(word.BundleName, "笔", global.Translations[lang]["笔"]))
		//builder.WriteString(" （能量笔数套餐）")

		builder.WriteString("\n") // 添加分隔符
	}

	// 去除最后一个空格
	result := strings.TrimSpace(builder.String())
	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "🧾"+global.Translations[lang]["deduction_records"]+"\n\n "+
		result+"\n")
	msg.ParseMode = "HTML"
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[lang]["prev"], "prev_st_bundle_package_page"),
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[lang]["next"], "next_st_bundle_package_page"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[lang]["back_homepage"], "back_bundle_package_ST"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard
	bot.Send(msg)
	//}
	global.DepositStates[callbackQuery.Message.Chat.ID] = state
	return false
}

func ShowPrevBundlePackagePage(lang string, callbackQuery *tgbotapi.CallbackQuery, db *gorm.DB, bot *tgbotapi.BotAPI) (*global.DepositState, bool) {
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
		//- [6.29] +3000 TRX（订单 #TOPUP-92308）
		for _, word := range trxlist {
			builder.WriteString("[")
			builder.WriteString(word.CreatedDate)
			builder.WriteString("]")
			builder.WriteString(" -")
			builder.WriteString(strings.ReplaceAll(word.BundleName, "笔", global.Translations[lang]["笔"]))
			//builder.WriteString(" （能量笔数套餐）")

			builder.WriteString("\n") // 添加分隔符
		}

		// 去除最后一个空格
		result := strings.TrimSpace(builder.String())
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "🧾"+global.Translations[lang]["deduction_records"]+"\n\n "+
			result+"\n")
		msg.ParseMode = "HTML"
		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(global.Translations[lang]["prev"], "prev_st_bundle_package_page"),
				tgbotapi.NewInlineKeyboardButtonData(global.Translations[lang]["next"], "next_st_bundle_package_page"),
			),
			tgbotapi.NewInlineKeyboardRow(
				//tgbotapi.NewInlineKeyboardButtonData("解绑地址", "free_monitor_address"),
				tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[lang]["back_homepage"], "back_bundle_package_ST"),
			),
		)
		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)
	} else {
		state.CurrentPage = state.CurrentPage - 1
		userAddressDetectionRepo := repositories.NewUserSmartTransactionPackageSubscriptionsRepository(db)
		var info request.UserAddressDetectionSearch
		info.PageInfo.Page = state.CurrentPage
		info.PageSize = 5
		trxlist, _, _ := userAddressDetectionRepo.ListByChatIDPage(context.Background(), info, callbackQuery.Message.Chat.ID)
		var builder strings.Builder
		builder.WriteString("\n") // 添加分隔符
		//- [6.29] +3000 TRX（订单 #TOPUP-92308）
		for _, word := range trxlist {
			builder.WriteString("[")
			builder.WriteString(word.CreatedDate)
			builder.WriteString("]")
			builder.WriteString(" -")
			builder.WriteString(strings.ReplaceAll(word.BundleName, "笔", global.Translations[lang]["笔"]))
			//builder.WriteString(" （能量笔数套餐）")

			builder.WriteString("\n") // 添加分隔符
		}

		// 去除最后一个空格
		result := strings.TrimSpace(builder.String())
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "🧾"+global.Translations[lang]["deduction_records"]+"\n\n "+
			result+"\n")
		msg.ParseMode = "HTML"
		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(global.Translations[lang]["prev"], "prev_st_bundle_package_page"),
				tgbotapi.NewInlineKeyboardButtonData(global.Translations[lang]["next"], "next_st_bundle_package_page"),
			),
			tgbotapi.NewInlineKeyboardRow(
				//tgbotapi.NewInlineKeyboardButtonData("解绑地址", "free_monitor_address"),
				tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[lang]["back_homepage"], "back_bundle_package_ST"),
			),
		)
		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)
	}
	return state, false
}
