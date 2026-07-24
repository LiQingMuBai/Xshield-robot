package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
	"ushield_bot/internal/cache"
	"ushield_bot/internal/domain"
	"ushield_bot/internal/global"
	"ushield_bot/internal/infrastructure/repositories"
	trxfee "ushield_bot/internal/infrastructure/thirdparty"
	. "ushield_bot/internal/infrastructure/tools"
	logger "ushield_bot/internal/logger"
	"ushield_bot/internal/request"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func RemoveBundlePackageAddress(lang string, cache cache.Cache, bot *tgbotapi.BotAPI, message *tgbotapi.Message, db *gorm.DB) bool {
	if !IsValidAddress(message.Text) {
		msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[lang]["invalid_address_tips"]+"</b>"+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)
		return true
	}

	userOperationPackageAddressesRepo := repositories.NewUserOperationPackageAddressesRepo(db)

	var record domain.UserOperationPackageAddresses
	record.Status = 0
	record.Address = message.Text
	record.ChatID = message.Chat.ID

	createErr := userOperationPackageAddressesRepo.DeleteByChatIDAndAddress(context.Background(), message.Chat.ID, message.Text)
	if createErr != nil {
		logger.Errorf("createErr: %v", createErr)
		return true
	}
	msg := tgbotapi.NewMessage(message.Chat.ID, "✅"+"<b>"+global.Translations[lang]["address_deleted_success"]+"</b>"+"\n")
	msg.ParseMode = "HTML"
	bot.Send(msg)
	//ShowBundlePackageAddressManagement(cache, bot, message.Chat.ID, db)

	msg2 := BuildBundlePackageAddressSummaryMessage(lang, db, message.Chat.ID)
	bot.Send(msg2)
	return false
}
func AddBundlePackageAddress(lang string, cache cache.Cache, bot *tgbotapi.BotAPI, message *tgbotapi.Message, db *gorm.DB) bool {
	if !IsValidAddress(message.Text) {
		msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[lang]["invalid_address_tips"]+"</b>"+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)
		return true
	}

	userOperationPackageAddressesRepo := repositories.NewUserOperationPackageAddressesRepo(db)

	exitRecord, _ := userOperationPackageAddressesRepo.GetByAddressAndChatID(context.Background(), message.Text, message.Chat.ID)

	if exitRecord.Id > 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+global.Translations[lang]["address_added_tips"]+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)
		ShowBundlePackageAddressManagement(lang, cache, bot, message.Chat.ID, db)
		return true
	}

	var record domain.UserOperationPackageAddresses
	record.Status = 0
	record.Address = message.Text
	record.ChatID = message.Chat.ID

	createErr := userOperationPackageAddressesRepo.Create(context.Background(), &record)
	if createErr != nil {
		logger.Errorf("createErr: %v", createErr)
		return true
	}
	msg := tgbotapi.NewMessage(message.Chat.ID, "✅"+"<b>"+global.Translations[lang]["address_added_success"]+"</b>"+"\n")
	msg.ParseMode = "HTML"
	bot.Send(msg)
	ShowBundlePackageAddressManagement(lang, cache, bot, message.Chat.ID, db)
	return false
}

func ApplyBundlePackage(lang string, cache cache.Cache, bot *tgbotapi.BotAPI, message *tgbotapi.Message, db *gorm.DB, status string) bool {
	if !IsValidAddress(message.Text) {
		msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[lang]["invalid_address_tips"]+"</b>"+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)
		return true
	}

	bundleID := strings.ReplaceAll(status, "apply_bundle_package_", "")
	userOperationBundlesRepo := repositories.NewUserOperationBundlesRepository(db)
	bundlePackage, err := userOperationBundlesRepo.GetByID(context.Background(), bundleID)

	if err != nil {
		logger.Println(err)
	}
	userRepo := repositories.NewUserRepository(db)
	user, _ := userRepo.GetByChatID(message.Chat.ID)

	lessBalance := false
	if bundlePackage.Token == "USDT" {
		//扣usdt
		if flag, _ := CompareNumberStrings(user.Amount, bundlePackage.Amount); flag < 0 {
			lessBalance = true
		}
		logger.Printf("bundle %v is USDT\n", bundlePackage)
	} else if bundlePackage.Token == "TRX" {
		//扣trx
		if flag, _ := CompareNumberStrings(user.TronAmount, bundlePackage.Amount); flag < 0 {
			lessBalance = true
		}

		logger.Printf("bundle %v is trx\n", bundlePackage)
	}

	if lessBalance {
		msg := tgbotapi.NewMessage(message.Chat.ID,

			"🆔"+global.Translations[lang]["user_id"]+": <code>"+user.Associates+"</code>\n"+
				"👤"+global.Translations[lang]["username"]+": @"+user.Username+"\n"+
				"💰"+global.Translations[lang]["balance"]+": "+"\n"+
				"- TRX：   "+user.TronAmount+"\n"+
				"- USDT："+user.Amount)

		msg.ParseMode = "HTML"

		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("💵"+global.Translations[lang]["deposit"], "deposit_amount"),
			),
		)

		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)

		return false
	}

	//扣錢
	if bundlePackage.Token == "TRX" {
		balance, _ := SubtractStringNumbers(user.TronAmount, bundlePackage.Amount, 1)
		logger.Printf("TRX balance %s", balance)
		user.TronAmount = balance
	} else if bundlePackage.Token == "USDT" {
		balance, _ := SubtractStringNumbers(user.Amount, bundlePackage.Amount, 1)
		logger.Printf("USDT balance %s", balance)

		user.Amount = balance
	}

	err = userRepo.Save(context.Background(), &user)
	if err != nil {

	}

	//加入訂閲記錄
	userPackageSubscriptionsRepo := repositories.NewUserPackageSubscriptionsRepository(db)
	var record domain.UserPackageSubscriptions
	record.ChatID = message.Chat.ID
	record.Address = message.Text
	bundle, _ := strconv.ParseInt(bundleID, 10, 64)

	record.BundleID = bundle
	record.Status = 2
	record.Amount = bundlePackage.Amount
	record.Times = ExtractLeadingInt64(bundlePackage.Name)
	record.BundleName = bundlePackage.Name

	err = userPackageSubscriptionsRepo.Create(context.Background(), &record)
	if err != nil {
		return true
	}
	msg := tgbotapi.NewMessage(message.Chat.ID, "✅"+"🧾"+global.Translations[lang]["package_order_purchased_successfully"]+"\n\n"+
		global.Translations[lang]["package_name"]+"："+strings.ReplaceAll(bundlePackage.Name, "笔", global.Translations[lang]["笔"])+"\n\n"+
		global.Translations[lang]["payment_amount"]+"："+bundlePackage.Amount+" "+bundlePackage.Token+"\n\n"+
		global.Translations[lang]["address"]+"："+message.Text+"\n\n"+
		global.Translations[lang]["order_id"]+"："+fmt.Sprintf("%d", record.Id)+""+"\n\n")
	msg.ParseMode = "HTML"
	// 当点击"按钮 1"时显示内联键盘
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🧾"+global.Translations[lang]["package_address_list"], "click_bundle_package_address_stats"),
			tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[lang]["back_homepage"], "back_bundle_package"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard

	bot.Send(msg)

	expiration := 1 * time.Minute // 短时间缓存空值

	//设置用户状态
	cache.Set(strconv.FormatInt(message.Chat.ID, 10), "null_apply_bundle_package_address", expiration)
	return false
}
func BuildBundlePackageAddressSummaryMessage(lang string, db *gorm.DB, chatID int64) tgbotapi.MessageConfig {

	userRepo := repositories.NewUserRepository(db)
	user, _ := userRepo.GetByChatID(chatID)

	userOperationPackageAddressesRepo := repositories.NewUserOperationPackageAddressesRepo(db)
	orderlist, err := userOperationPackageAddressesRepo.ListByChatID(context.Background(), chatID)
	//orderlist, total, err := userAddressDetectionRepo.ListByChatIDPage(context.Background(), info, chatID)

	energyRepo := repositories.NewUserEnergyOrdersRepo(db)
	usedTimes, _ := energyRepo.CountByChatID(context.Background(), chatID)

	if err != nil {

		logger.Error("能量笔数套餐空", err)
	}
	var builder strings.Builder
	if len(orderlist) > 0 {

		builder.WriteString("\n")

		builder.WriteString(global.Translations[lang]["remaining"])
		builder.WriteString(strconv.FormatInt(user.BundleTimes, 10))
		builder.WriteString(" " + global.Translations[lang]["笔"])

		builder.WriteString("     " + global.Translations[lang]["used"])
		builder.WriteString(strconv.FormatInt(usedTimes, 10))
		builder.WriteString(" " + global.Translations[lang]["笔"])

		for _, order := range orderlist {

			builder.WriteString("\n") // 添加分隔符

			builder.WriteString("<code>" + order.Address + "</code>")
			builder.WriteString("\n")

			if user.BundleTimes > 0 {
				builder.WriteString(global.Translations[lang]["dispatch_now"] + ":/dispatchNow")
				builder.WriteString(strconv.FormatInt(order.Id, 10))
				builder.WriteString("_1")
				builder.WriteString("\n") // 添加分隔符

			}
			if user.BundleTimes > 1 {
				builder.WriteString(global.Translations[lang]["dispatch_now_2"] + ":/dispatchNow")
				builder.WriteString(strconv.FormatInt(order.Id, 10))
				builder.WriteString("_2")
				builder.WriteString("\n") // 添加分隔符

			}

			builder.WriteString("➖➖➖➖➖➖➖➖➖➖➖➖➖") // 添加分隔符
		}
	} else {
		builder.WriteString("\n")

		builder.WriteString(global.Translations[lang]["remaining"])
		builder.WriteString(strconv.FormatInt(user.BundleTimes, 10))
		builder.WriteString(" " + global.Translations[lang]["笔"])

		builder.WriteString("     " + global.Translations[lang]["used"])
		builder.WriteString(strconv.FormatInt(usedTimes, 10))
		builder.WriteString(" " + global.Translations[lang]["笔"])
		builder.WriteString("\n")
		builder.WriteString(global.Translations[lang]["address_list_empty_tips"] + "\n\n") // 添加分隔符
	}

	// 去除最后一个空格
	result := strings.TrimSpace(builder.String())

	msg := tgbotapi.NewMessage(chatID, "🧾"+global.Translations[lang]["package_address_list"]+"\n"+
		result+"\n")
	msg.ParseMode = "HTML"
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚡️"+global.Translations[lang]["dispatch_other"], "dispatch_Now_Others"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕"+global.Translations[lang]["add_address"], "click_bundle_package_address_manager_add"),

			tgbotapi.NewInlineKeyboardButtonData("➖"+global.Translations[lang]["remove_address"], "click_bundle_package_address_manager_remove"),
		),

		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[lang]["back_homepage"], "back_bundle_package"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard
	return msg
}
func BuildBundlePackageSubscriptionStatsMessage(lang string, db *gorm.DB, chatID int64) tgbotapi.MessageConfig {

	userAddressDetectionRepo := repositories.NewUserPackageSubscriptionsRepository(db)
	var info request.UserAddressDetectionSearch

	info.Page = 1
	info.PageSize = 5
	orderlist, total, err := userAddressDetectionRepo.ListByChatIDPage(context.Background(), info, chatID)
	if err != nil {

		logger.Error("能量笔数套餐空", err)
	}
	var builder strings.Builder
	if total > 0 {
		for _, order := range orderlist {
			builder.WriteString("\n")
			builder.WriteString("<code>" + order.Address + "</code>")
			builder.WriteString("\n")

			builder.WriteString(global.Translations[lang]["remaining"])
			builder.WriteString(strconv.FormatInt(order.Times, 10))
			builder.WriteString(" " + global.Translations[lang]["笔"])

			usedTimes := ExtractLeadingInt64(order.BundleName) - order.Times
			builder.WriteString("     " + global.Translations[lang]["used"])
			builder.WriteString(strconv.FormatInt(usedTimes, 10))
			builder.WriteString(" " + global.Translations[lang]["笔"])

			builder.WriteString("\n") // 添加分隔符
			if order.Times > 0 {
				if order.Status == 2 {
					builder.WriteString("开启自动发能:/startDispatch")
					builder.WriteString(strconv.FormatInt(order.Id, 10))
				}
				if order.Status == 1 {
					builder.WriteString("关闭自动发能:/stopDispatch")
					builder.WriteString(strconv.FormatInt(order.Id, 10))
				}
				builder.WriteString("\n") // 添加分隔符
				builder.WriteString("手动发能:/dispatchNow")
				builder.WriteString(strconv.FormatInt(order.Id, 10))
				builder.WriteString("\n") // 添加分隔符
				builder.WriteString("其他地址发能:/dispatchOthers")
				builder.WriteString(strconv.FormatInt(order.Id, 10))
				builder.WriteString("\n") // 添加分隔符
			}
			builder.WriteString("➖➖➖➖➖➖➖➖➖➖➖➖➖") // 添加分隔符
		}
	} else {
		builder.WriteString(global.Translations[lang]["address_list_empty_tips"] + "\n\n") // 添加分隔符
	}

	// 去除最后一个空格
	result := strings.TrimSpace(builder.String())

	msg := tgbotapi.NewMessage(chatID, "🧾"+global.Translations[lang]["package_address_list"]+"\n\n"+
		result+"\n")
	msg.ParseMode = "HTML"
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[lang]["prev"], "next_bundle_package_address_stats"),
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[lang]["next"], "prev_bundle_package_address_stats"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[lang]["back_homepage"], "back_bundle_package"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard
	return msg
}
func ShowNextBundlePackageSubscriptionStatsPage(lang string, callbackQuery *tgbotapi.CallbackQuery, db *gorm.DB, bot *tgbotapi.BotAPI) bool {
	state := global.DepositStates[callbackQuery.Message.Chat.ID]
	if state == nil {
		var state2 global.DepositState
		state2.CurrentPage = 1
		state = &state2
	}
	state.CurrentPage = state.CurrentPage + 1
	userAddressDetectionRepo := repositories.NewUserPackageSubscriptionsRepository(db)
	var info request.UserAddressDetectionSearch
	info.PageInfo.Page = state.CurrentPage
	info.PageInfo.PageSize = 10
	orderlist, total, _ := userAddressDetectionRepo.ListByChatIDPage(context.Background(), info, callbackQuery.Message.Chat.ID)

	logger.Printf("currentpage : %d", state.CurrentPage)
	logger.Printf("total: %v\n", total)
	totalPages := (total + 5 - 1) / 5

	logger.Printf("totalPages : %d", totalPages)
	if int64(state.CurrentPage) > totalPages {
		state.CurrentPage = totalPages
		return true
	}
	var builder strings.Builder
	if total > 0 {
		for _, order := range orderlist {
			builder.WriteString("\n")
			builder.WriteString("<code>" + order.Address + "</code>")
			builder.WriteString("\n")

			builder.WriteString(global.Translations[lang]["remaining"])
			builder.WriteString(strconv.FormatInt(order.Times, 10))
			builder.WriteString(" " + global.Translations[lang]["笔"])

			usedTimes := ExtractLeadingInt64(order.BundleName) - order.Times
			builder.WriteString("     " + global.Translations[lang]["used"])
			builder.WriteString(strconv.FormatInt(usedTimes, 10))
			builder.WriteString(" " + global.Translations[lang]["笔"])

			builder.WriteString("\n") // 添加分隔符
			if order.Times > 0 {
				if order.Status == 2 {
					builder.WriteString("开启自动发能:/startDispatch")
					builder.WriteString(strconv.FormatInt(order.Id, 10))
				}
				if order.Status == 1 {
					builder.WriteString("关闭自动发能:/stopDispatch")
					builder.WriteString(strconv.FormatInt(order.Id, 10))
				}
				builder.WriteString("\n") // 添加分隔符
				builder.WriteString("手动发能:/dispatchNow")
				builder.WriteString(strconv.FormatInt(order.Id, 10))

				builder.WriteString("\n") // 添加分隔符
				builder.WriteString("其他地址发能:/dispatchOthers")
				builder.WriteString(strconv.FormatInt(order.Id, 10))
				builder.WriteString("\n") // 添加分隔符
			}
			builder.WriteString("➖➖➖➖➖➖➖➖➖➖➖➖➖") // 添加分隔符
		}
	} else {
		builder.WriteString(global.Translations[lang]["address_list_empty_tips"] + "\n\n") // 添加分隔符
	}

	// 去除最后一个空格
	result := strings.TrimSpace(builder.String())
	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "🧾"+global.Translations[lang]["package_address_list"]+"：</b>\n\n "+
		result+"\n")
	msg.ParseMode = "HTML"
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[lang]["prev"], "next_bundle_package_address_stats"),
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[lang]["next"], "prev_bundle_package_address_stats"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[lang]["back_homepage"], "back_bundle_package"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard
	bot.Send(msg)
	logger.Printf("state: %v\n", state)

	global.DepositStates[callbackQuery.Message.Chat.ID] = state
	return false
}

func ShowPrevBundlePackageSubscriptionStatsPage(lang string, callbackQuery *tgbotapi.CallbackQuery, db *gorm.DB, bot *tgbotapi.BotAPI) (*global.DepositState, bool) {
	state := global.DepositStates[callbackQuery.Message.Chat.ID]

	if state != nil && state.CurrentPage == 1 {
		return nil, true
	}
	if state == nil {
		var state global.DepositState
		state.CurrentPage = 1
		global.DepositStates[callbackQuery.Message.Chat.ID] = &state
		userAddressDetectionRepo := repositories.NewUserPackageSubscriptionsRepository(db)
		var info request.UserAddressDetectionSearch
		info.PageInfo.Page = 1
		info.PageInfo.PageSize = 10
		orderlist, total, _ := userAddressDetectionRepo.ListByChatIDPage(context.Background(), info, callbackQuery.Message.Chat.ID)
		var builder strings.Builder
		if total > 0 {
			for _, order := range orderlist {
				builder.WriteString("\n")
				builder.WriteString("<code>" + order.Address + "</code>")
				builder.WriteString("\n")

				builder.WriteString(global.Translations[lang]["remaining"])
				builder.WriteString(strconv.FormatInt(order.Times, 10))
				builder.WriteString(" " + global.Translations[lang]["笔"])

				usedTimes := ExtractLeadingInt64(order.BundleName) - order.Times
				builder.WriteString("     " + global.Translations[lang]["used"])
				builder.WriteString(strconv.FormatInt(usedTimes, 10))
				builder.WriteString(" " + global.Translations[lang]["笔"])

				builder.WriteString("\n") // 添加分隔符
				if order.Times > 0 {
					if order.Status == 2 {
						builder.WriteString("开启自动发能:/startDispatch")
						builder.WriteString(strconv.FormatInt(order.Id, 10))
					}
					if order.Status == 1 {
						builder.WriteString("关闭自动发能:/stopDispatch")
						builder.WriteString(strconv.FormatInt(order.Id, 10))
					}
					builder.WriteString("\n") // 添加分隔符
					builder.WriteString("手动发能:/dispatchNow")
					builder.WriteString(strconv.FormatInt(order.Id, 10))
					builder.WriteString("\n") // 添加分隔符
					builder.WriteString("其他地址发能:/dispatchOthers")
					builder.WriteString(strconv.FormatInt(order.Id, 10))
					builder.WriteString("\n") // 添加分隔符
				}
				builder.WriteString("➖➖➖➖➖➖➖➖➖➖➖➖➖") // 添加分隔符
			}
		} else {
			builder.WriteString(global.Translations[lang]["address_list_empty_tips"] + "\n\n") // 添加分隔符
		}

		// 去除最后一个空格
		result := strings.TrimSpace(builder.String())
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "🧾"+global.Translations[lang]["package_address_list"]+"：</b>\n\n "+
			result+"\n")
		msg.ParseMode = "HTML"
		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(global.Translations[lang]["prev"], "next_bundle_package_address_stats"),
				tgbotapi.NewInlineKeyboardButtonData(global.Translations[lang]["next"], "prev_bundle_package_address_stats"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[lang]["back_homepage"], "back_bundle_package"),
			),
		)
		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)
	} else {
		state.CurrentPage = state.CurrentPage - 1
		userAddressDetectionRepo := repositories.NewUserPackageSubscriptionsRepository(db)
		var info request.UserAddressDetectionSearch
		info.PageInfo.Page = state.CurrentPage
		info.PageInfo.PageSize = 10
		orderlist, total, _ := userAddressDetectionRepo.ListByChatIDPage(context.Background(), info, callbackQuery.Message.Chat.ID)
		var builder strings.Builder
		if total > 0 {
			for _, order := range orderlist {
				builder.WriteString("\n")
				builder.WriteString("<code>" + order.Address + "</code>")
				builder.WriteString("\n")

				builder.WriteString(global.Translations[lang]["remaining"])
				builder.WriteString(strconv.FormatInt(order.Times, 10))
				builder.WriteString(" " + global.Translations[lang]["笔"])

				usedTimes := ExtractLeadingInt64(order.BundleName) - order.Times
				builder.WriteString("     " + global.Translations[lang]["used"])
				builder.WriteString(strconv.FormatInt(usedTimes, 10))
				builder.WriteString(" " + global.Translations[lang]["笔"])

				builder.WriteString("\n") // 添加分隔符
				if order.Times > 0 {
					if order.Status == 2 {
						builder.WriteString("开启自动发能:/startDispatch")
						builder.WriteString(strconv.FormatInt(order.Id, 10))
					}
					if order.Status == 1 {
						builder.WriteString("关闭自动发能:/stopDispatch")
						builder.WriteString(strconv.FormatInt(order.Id, 10))
					}
					builder.WriteString("\n") // 添加分隔符
					builder.WriteString("手动发能:/dispatchNow")
					builder.WriteString(strconv.FormatInt(order.Id, 10))
					builder.WriteString("\n") // 添加分隔符
					builder.WriteString("其他地址发能:/dispatchOthers")
					builder.WriteString(strconv.FormatInt(order.Id, 10))
					builder.WriteString("\n") // 添加分隔符
				}
				builder.WriteString("➖➖➖➖➖➖➖➖➖➖➖➖➖") // 添加分隔符
			}
		} else {
			builder.WriteString(global.Translations[lang]["address_list_empty_tips"] + "\n\n") // 添加分隔符
		}

		// 去除最后一个空格
		result := strings.TrimSpace(builder.String())
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "🧾"+global.Translations[lang]["package_address_list"]+"：</b>\n\n "+
			result+"\n")
		msg.ParseMode = "HTML"
		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(global.Translations[lang]["prev"], "next_bundle_package_address_stats"),
				tgbotapi.NewInlineKeyboardButtonData(global.Translations[lang]["next"], "prev_bundle_package_address_stats"),
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
func ApplySmartTransactionBundlePackage(trxfeeClient *trxfee.TrxfeeClient, lang string, cache cache.Cache, bot *tgbotapi.BotAPI, message *tgbotapi.Message, db *gorm.DB, status string) bool {
	if !IsValidAddress(message.Text) {
		msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[lang]["invalid_address_tips"]+"</b>"+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)
		return true
	}

	bundleID := strings.ReplaceAll(status, "apply_ST_bundle_package_", "")
	userOperationBundlesRepo := repositories.NewUserSmartTransactionBundlesRepository(db)
	bundlePackage, err := userOperationBundlesRepo.GetByID(context.Background(), bundleID)

	if err != nil {
		logger.Println(err)
	}
	userRepo := repositories.NewUserRepository(db)
	user, _ := userRepo.GetByChatID(message.Chat.ID)

	lessBalance := false
	if bundlePackage.Token == "USDT" {
		//扣usdt
		if flag, _ := CompareNumberStrings(user.Amount, bundlePackage.Amount); flag < 0 {
			lessBalance = true
		}
		logger.Printf("bundle %v is USDT\n", bundlePackage)
	} else if bundlePackage.Token == "TRX" {
		//扣trx
		if flag, _ := CompareNumberStrings(user.TronAmount, bundlePackage.Amount); flag < 0 {
			lessBalance = true
		}

		logger.Printf("bundle %v is trx\n", bundlePackage)
	}

	if lessBalance {
		msg := tgbotapi.NewMessage(message.Chat.ID,
			"🆔"+global.Translations[lang]["user_id"]+": <code>"+user.Associates+"</code>\n"+
				"👤"+global.Translations[lang]["username"]+": @"+user.Username+"\n"+
				"💰"+global.Translations[lang]["balance"]+": "+"\n"+
				"- TRX：   "+user.TronAmount+"\n"+
				"- USDT："+user.Amount)

		msg.ParseMode = "HTML"

		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("💵"+global.Translations[lang]["deposit"], "deposit_amount"),
			),
		)

		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)

		return false
	}

	//加入訂閲記錄
	userPackageSubscriptionsRepo := repositories.NewUserSmartTransactionPackageSubscriptionsRepository(db)

	//判断是否已经购买的地址，在进行中的
	item, err := userPackageSubscriptionsRepo.GetActiveByAddress(context.Background(), message.Text)

	if err != nil {
		return false
	}

	if item.Id > 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, global.Translations[lang]["smart_transaction_plans_repeat_order"]+
			"🆔"+global.Translations[lang]["user_id"]+": <code>"+user.Associates+"</code>\n"+
			"👤"+global.Translations[lang]["username"]+": @"+user.Username+"\n"+
			"💰"+global.Translations[lang]["balance"]+": "+"\n"+
			"- TRX：   "+user.TronAmount+"\n"+
			"- USDT："+user.Amount)

		msg.ParseMode = "HTML"

		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("💵"+global.Translations[lang]["deposit"], "deposit_amount"),
			),
		)

		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)

		return false
	}

	//扣錢
	if bundlePackage.Token == "TRX" {
		balance, _ := SubtractStringNumbers(user.TronAmount, bundlePackage.Amount, 1)
		logger.Printf("TRX balance %s\n", balance)
		user.TronAmount = balance
	} else if bundlePackage.Token == "USDT" {
		balance, _ := SubtractStringNumbers(user.Amount, bundlePackage.Amount, 1)
		logger.Printf("USDT balance %s\n", balance)

		user.Amount = balance
	}

	err = userRepo.Save(context.Background(), &user)
	if err != nil {

		return false
	}

	var record domain.UserSmartTransactionPackageSubscriptions
	record.ChatID = message.Chat.ID
	record.Address = message.Text
	bundle, _ := strconv.ParseInt(bundleID, 10, 64)

	record.BundleID = bundle
	record.Status = 2
	record.Amount = bundlePackage.Amount
	record.Times = ExtractLeadingInt64(bundlePackage.Name)
	record.BundleName = bundlePackage.Name

	err = userPackageSubscriptionsRepo.Create(context.Background(), &record)
	if err != nil {
		return true
	}

	//

	logger.Printf("address %s\n", record.Address)
	logger.Printf("times %d\n", record.Times)

	if err := trxfeeClient.TimesOrder(record.Address, int(record.Times)); err != nil {
		logger.Errorf("create trxfee times order failed: %v", err)
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "✅"+global.Translations[lang]["smart_transaction_package_order_purchased_successfully"]+"\n"+
		global.Translations[lang]["package_name"]+"："+strings.ReplaceAll(bundlePackage.Name, "笔", global.Translations[lang]["笔"])+"\n"+
		global.Translations[lang]["payment_amount"]+"："+bundlePackage.Amount+" "+bundlePackage.Token+"\n"+
		global.Translations[lang]["address"]+"："+message.Text+"\n"+
		global.Translations[lang]["order_id"]+"："+fmt.Sprintf("%d", record.Id)+""+"\n")
	msg.ParseMode = "HTML"
	// 当点击"按钮 1"时显示内联键盘
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🧾"+global.Translations[lang]["smart_transaction_package_address_list"], "click_bundle_package_address_stats_ST"),
			tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[lang]["back_homepage"], "back_bundle_package_ST"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard

	bot.Send(msg)

	expiration := 1 * time.Minute // 短时间缓存空值

	//设置用户状态
	cache.Set(strconv.FormatInt(message.Chat.ID, 10), "null_apply_bundle_package_address_ST", expiration)
	return false
}
func BuildSmartTransactionAddressStatsMessage(lang string, db *gorm.DB, chatID int64) tgbotapi.MessageConfig {

	userAddressDetectionRepo := repositories.NewUserSmartTransactionPackageSubscriptionsRepository(db)
	var info request.UserAddressDetectionSearch

	info.Page = 1
	info.PageSize = 100000
	orderlist, total, err := userAddressDetectionRepo.ListByChatIDPage(context.Background(), info, chatID)
	if err != nil {

		logger.Error("能量笔数套餐空", err)
	}
	var builder strings.Builder
	if total > 0 {
		for _, order := range orderlist {
			builder.WriteString("\n")
			builder.WriteString("<code>" + order.Address + "</code>")
			builder.WriteString("\n")
			if order.Status == 2 {
				builder.WriteString(global.Translations[lang]["smart_transaction_disable"])
			}
			if order.Status == 1 {
				builder.WriteString(global.Translations[lang]["smart_transaction_enable"])
			}
			builder.WriteString(global.Translations[lang]["remaining"])
			builder.WriteString(strconv.FormatInt(order.Times, 10))
			builder.WriteString(" " + global.Translations[lang]["笔"])

			usedTimes := ExtractLeadingInt64(order.BundleName) - order.Times
			builder.WriteString("     " + global.Translations[lang]["used"])
			builder.WriteString(strconv.FormatInt(usedTimes, 10))
			builder.WriteString(" " + global.Translations[lang]["笔"])

			builder.WriteString("\n") // 添加分隔符
			if order.Times > 0 {
				if order.Status == 2 {
					builder.WriteString(global.Translations[lang]["close_auto_dispatch_energy"] + ":/close_ST")
					builder.WriteString(strconv.FormatInt(order.Id, 10))
				}
				if order.Status == 1 {
					builder.WriteString(global.Translations[lang]["open_auto_dispatch_energy"] + ":/open_ST")
					builder.WriteString(strconv.FormatInt(order.Id, 10))
				}
				builder.WriteString("\n") // 添加分隔符
			}
			builder.WriteString("➖➖➖➖➖➖➖➖➖➖➖➖➖") // 添加分隔符
		}
	} else {
		builder.WriteString(global.Translations[lang]["smart_transaction__list_empty_tips"] + "\n\n") // 添加分隔符
	}

	// 去除最后一个空格
	result := strings.TrimSpace(builder.String())

	msg := tgbotapi.NewMessage(chatID, "🧾"+global.Translations[lang]["smart_transaction_package_address_list"]+"\n\n"+
		result+"\n")
	msg.ParseMode = "HTML"
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[lang]["back_homepage"], "back_bundle_package_ST"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard
	return msg
}
