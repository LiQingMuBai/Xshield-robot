package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"ushield_bot/internal/cache"
	"ushield_bot/internal/domain"
	"ushield_bot/internal/global"
	trxfee "ushield_bot/internal/infrastructure/3rd"
	"ushield_bot/internal/infrastructure/repositories"
	. "ushield_bot/internal/infrastructure/tools"
	"ushield_bot/internal/request"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func CLICK_BUNDLE_PACKAGE_ADDRESS_MANAGER_REMOVE(_lang string, cache cache.Cache, bot *tgbotapi.BotAPI, message *tgbotapi.Message, db *gorm.DB) bool {
	if !IsValidAddress(message.Text) {
		msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[_lang]["invalid_address_tips"]+"</b>"+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)
		return true
	}

	userOperationPackageAddressesRepo := repositories.NewUserOperationPackageAddressesRepo(db)

	var record domain.UserOperationPackageAddresses
	record.Status = 0
	record.Address = message.Text
	record.ChatID = message.Chat.ID

	errsg := userOperationPackageAddressesRepo.Remove(context.Background(), message.Chat.ID, message.Text)
	if errsg != nil {
		log.Printf("errsg: %s", errsg)
		return true
	}
	msg := tgbotapi.NewMessage(message.Chat.ID, "✅"+"<b>"+global.Translations[_lang]["address_deleted_success"]+"</b>"+"\n")
	msg.ParseMode = "HTML"
	bot.Send(msg)
	//CLICK_BUNDLE_PACKAGE_ADDRESS_MANAGEMENT(cache, bot, message.Chat.ID, db)

	msg2 := CLICK_BUNDLE_PACKAGE_ADDRESS_STATS2(_lang, db, message.Chat.ID)
	bot.Send(msg2)
	return false
}
func CLICK_BUNDLE_PACKAGE_ADDRESS_MANAGER_ADD(_lang string, cache cache.Cache, bot *tgbotapi.BotAPI, message *tgbotapi.Message, db *gorm.DB) bool {
	if !IsValidAddress(message.Text) {
		msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[_lang]["invalid_address_tips"]+"</b>"+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)
		return true
	}

	userOperationPackageAddressesRepo := repositories.NewUserOperationPackageAddressesRepo(db)

	exitRecord, _ := userOperationPackageAddressesRepo.GetUserOperationPackageAddress(context.Background(), message.Text, message.Chat.ID)

	if exitRecord.Id > 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+global.Translations[_lang]["address_added_tips"]+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)
		CLICK_BUNDLE_PACKAGE_ADDRESS_MANAGEMENT(_lang, cache, bot, message.Chat.ID, db)
		return true
	}

	var record domain.UserOperationPackageAddresses
	record.Status = 0
	record.Address = message.Text
	record.ChatID = message.Chat.ID

	errsg := userOperationPackageAddressesRepo.Create(context.Background(), &record)
	if errsg != nil {
		log.Printf("errsg: %s", errsg)
		return true
	}
	msg := tgbotapi.NewMessage(message.Chat.ID, "✅"+"<b>"+global.Translations[_lang]["address_added_success"]+"</b>"+"\n")
	msg.ParseMode = "HTML"
	bot.Send(msg)
	CLICK_BUNDLE_PACKAGE_ADDRESS_MANAGEMENT(_lang, cache, bot, message.Chat.ID, db)
	return false
}

func APPLY_BUNDLE_PACKAGE(_lang string, cache cache.Cache, bot *tgbotapi.BotAPI, message *tgbotapi.Message, db *gorm.DB, status string) bool {
	if !IsValidAddress(message.Text) {
		msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[_lang]["invalid_address_tips"]+"</b>"+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)
		return true
	}

	bundleID := strings.ReplaceAll(status, "apply_bundle_package_", "")
	userOperationBundlesRepo := repositories.NewUserOperationBundlesRepository(db)
	bundlePackage, err := userOperationBundlesRepo.Query(context.Background(), bundleID)

	if err != nil {
		fmt.Println(err)
	}
	userRepo := repositories.NewUserRepository(db)
	user, _ := userRepo.GetByUserID(message.Chat.ID)

	lessBalance := false
	if bundlePackage.Token == "USDT" {
		//扣usdt
		if flag, _ := CompareNumberStrings(user.Amount, bundlePackage.Amount); flag < 0 {
			lessBalance = true
		}
		fmt.Printf("bundle %v is USDT\n", bundlePackage)
	} else if bundlePackage.Token == "TRX" {
		//扣trx
		if flag, _ := CompareNumberStrings(user.TronAmount, bundlePackage.Amount); flag < 0 {
			lessBalance = true
		}

		fmt.Printf("bundle %v is trx\n", bundlePackage)
	}

	if lessBalance {
		msg := tgbotapi.NewMessage(message.Chat.ID,
			//"💬"+"<b>"+"用户姓名: "+"</b>"+user.Username+"\n"+
			//	"👤"+"<b>"+"用户电报ID: "+"</b>"+user.Associates+"\n"+
			//	"💵"+"<b>"+"余额不足 "+"</b>"+"\n"+
			//	"💴"+"<b>"+"当前TRX余额:  "+"</b>"+user.TronAmount+" TRX"+"\n"+
			//	"💴"+"<b>"+"当前USDT余额:  "+"</b>"+user.Amount+" USDT")

			"🆔"+global.Translations[_lang]["user_id"]+": "+user.Associates+"\n"+
				"👤"+global.Translations[_lang]["username"]+": @"+user.Username+"\n"+
				"💰"+global.Translations[_lang]["balance"]+": "+"\n"+
				"- TRX：   "+user.TronAmount+"\n"+
				"- USDT："+user.Amount)

		msg.ParseMode = "HTML"

		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("💵"+global.Translations[_lang]["deposit"], "deposit_amount"),
			),
		)

		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)

		return false
	}

	//扣錢
	if bundlePackage.Token == "TRX" {
		balance, _ := SubtractStringNumbers(user.TronAmount, bundlePackage.Amount, 1)
		fmt.Printf("TRX balance %s", balance)
		user.TronAmount = balance
	} else if bundlePackage.Token == "USDT" {
		balance, _ := SubtractStringNumbers(user.Amount, bundlePackage.Amount, 1)
		fmt.Printf("USDT balance %s", balance)

		user.Amount = balance
	}

	err = userRepo.Update2(context.Background(), &user)
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
	msg := tgbotapi.NewMessage(message.Chat.ID, "✅"+"🧾"+global.Translations[_lang]["package_order_purchased_successfully"]+"\n\n"+
		global.Translations[_lang]["package_name"]+"："+strings.ReplaceAll(bundlePackage.Name, "笔", global.Translations[_lang]["笔"])+"\n\n"+
		global.Translations[_lang]["payment_amount"]+"："+bundlePackage.Amount+" "+bundlePackage.Token+"\n\n"+
		global.Translations[_lang]["address"]+"："+message.Text+"\n\n"+
		global.Translations[_lang]["order_id"]+"："+fmt.Sprintf("%d", record.Id)+""+"\n\n")
	msg.ParseMode = "HTML"
	// 当点击"按钮 1"时显示内联键盘
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🧾"+global.Translations[_lang]["package_address_list"], "click_bundle_package_address_stats"),
			tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[_lang]["back_homepage"], "back_bundle_package"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard

	bot.Send(msg)

	expiration := 1 * time.Minute // 短时间缓存空值

	//设置用户状态
	cache.Set(strconv.FormatInt(message.Chat.ID, 10), "null_apply_bundle_package_address", expiration)
	return false
}
func CLICK_BUNDLE_PACKAGE_ADDRESS_STATS2(_lang string, db *gorm.DB, chatID int64) tgbotapi.MessageConfig {

	//fmt.Println("ExtractBundlePackage")
	//userAddressDetectionRepo := repositories.NewUserPackageSubscriptionsRepository(db)
	//var info request.UserAddressDetectionSearch
	//
	//info.Page = 1
	//info.PageSize = 5
	userRepo := repositories.NewUserRepository(db)
	user, _ := userRepo.GetByUserID(chatID)

	userOperationPackageAddressesRepo := repositories.NewUserOperationPackageAddressesRepo(db)
	orderlist, err := userOperationPackageAddressesRepo.Query(context.Background(), chatID)
	//orderlist, total, err := userAddressDetectionRepo.GetUserPackageSubscriptionsInfoList(context.Background(), info, chatID)

	energyRepo := repositories.NewUserEnergyOrdersRepo(db)
	usedTimes, _ := energyRepo.Count(context.Background(), chatID)

	if err != nil {

		fmt.Println("能量笔数套餐空", err)
	}
	var builder strings.Builder
	if len(orderlist) > 0 {

		builder.WriteString("\n")

		builder.WriteString(global.Translations[_lang]["remaining"])
		builder.WriteString(strconv.FormatInt(user.BundleTimes, 10))
		builder.WriteString(" " + global.Translations[_lang]["笔"])

		//usedTimes := ExtractLeadingInt64(order.BundleName) - order.Times
		builder.WriteString("     " + global.Translations[_lang]["used"])
		builder.WriteString(strconv.FormatInt(usedTimes, 10))
		builder.WriteString(" " + global.Translations[_lang]["笔"])

		//builder.WriteString("\n") // 添加分隔符
		//- [6.29] +3000 TRX（订单 #TOPUP-92308）
		for _, order := range orderlist {
			//builder.WriteString(global.Translations[_lang]["address"]+"：")
			//builder.WriteString("\n")
			//builder.WriteString("<code>" + order.Address + "</code>")
			//builder.WriteString("\n")
			//builder.WriteString("状态：")
			////0默认初始化状态  1 自动派送 2 手动 3 结束
			//if order.Status == 3 {
			//	builder.WriteString("<b>" + "已结束" + "</b>")
			//} else if order.Status == 2 {
			//	builder.WriteString("<b>" + "已停止" + "</b>")
			//} else if order.Status == 1 {
			//	builder.WriteString("<b>" + "已开启" + "</b>")
			//} else if order.Status == 0 {
			//	builder.WriteString("<b>" + "初始化" + "</b>")
			//}
			//

			////builder.WriteString(" （能量笔数套餐）")

			builder.WriteString("\n") // 添加分隔符

			//if user.BundleTimes <= 0 {
			//	builder.WriteString("笔数为空")
			//	builder.WriteString("\n")
			//	//break
			//}

			//builder.WriteString(global.Translations[_lang]["address"]+"：")
			//builder.WriteString("\n")
			builder.WriteString("<code>" + order.Address + "</code>")
			builder.WriteString("\n")

			//if order.Status == 2 {
			//	builder.WriteString("开启自动发能:/startDispatch")
			//	builder.WriteString(strconv.FormatInt(order.Id, 10))
			//}
			//if order.Status == 1 {
			//	builder.WriteString("关闭自动发能:/stopDispatch")
			//	builder.WriteString(strconv.FormatInt(order.Id, 10))
			//}
			//builder.WriteString("\n") // 添加分隔符
			if user.BundleTimes > 0 {
				builder.WriteString(global.Translations[_lang]["dispatch_now"] + ":/dispatchNow")
				builder.WriteString(strconv.FormatInt(order.Id, 10))
				builder.WriteString("_1")
				builder.WriteString("\n") // 添加分隔符

				//builder.WriteString(strconv.FormatInt(order.Id, 10))
				//builder.WriteString("\n") // 添加分隔符
			}
			if user.BundleTimes > 1 {
				builder.WriteString("\n") // 添加分隔符
				builder.WriteString(global.Translations[_lang]["dispatch_now_2"] + ":/dispatchNow")
				builder.WriteString(strconv.FormatInt(order.Id, 10))
				builder.WriteString("_2")
				builder.WriteString("\n") // 添加分隔符

			}

			//builder.WriteString("\n")
			builder.WriteString("➖➖➖➖➖➖➖➖➖➖➖➖➖") // 添加分隔符
			//builder.WriteString("\n")            // 添加分隔符
		}
		//if user.BundleTimes > 0 {
		//	builder.WriteString(global.Translations[_lang]["address_list_empty_tips"]+"\n\n") // 添加分隔符
		//	builder.WriteString("<b>向其他地址发能</b>:/dispatchOthers")
		//	builder.WriteString("\n") // 添加分隔符
		//}
	} else {
		builder.WriteString("\n")

		builder.WriteString(global.Translations[_lang]["remaining"])
		builder.WriteString(strconv.FormatInt(user.BundleTimes, 10))
		builder.WriteString(" " + global.Translations[_lang]["笔"])

		//usedTimes := ExtractLeadingInt64(order.BundleName) - order.Times
		builder.WriteString("     " + global.Translations[_lang]["used"])
		builder.WriteString(strconv.FormatInt(usedTimes, 10))
		builder.WriteString(" " + global.Translations[_lang]["笔"])
		builder.WriteString("\n")
		builder.WriteString(global.Translations[_lang]["address_list_empty_tips"] + "\n\n") // 添加分隔符
	}

	// 去除最后一个空格
	result := strings.TrimSpace(builder.String())

	msg := tgbotapi.NewMessage(chatID, "🧾"+global.Translations[_lang]["package_address_list"]+"\n"+
		result+"\n")
	msg.ParseMode = "HTML"
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			//tgbotapi.NewInlineKeyboardButtonData("解绑地址", "free_monitor_address"),
			tgbotapi.NewInlineKeyboardButtonData("⚡️"+global.Translations[_lang]["dispatch_other"], "dispatch_Now_Others"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕"+global.Translations[_lang]["add_address"], "click_bundle_package_address_manager_add"),

			tgbotapi.NewInlineKeyboardButtonData("➖"+global.Translations[_lang]["remove_address"], "click_bundle_package_address_manager_remove"),
		),

		tgbotapi.NewInlineKeyboardRow(
			//tgbotapi.NewInlineKeyboardButtonData("解绑地址", "free_monitor_address"),
			tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[_lang]["back_homepage"], "back_bundle_package"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard
	return msg
}
func CLICK_BUNDLE_PACKAGE_ADDRESS_STATS(_lang string, db *gorm.DB, chatID int64) tgbotapi.MessageConfig {

	//fmt.Println("ExtractBundlePackage")
	userAddressDetectionRepo := repositories.NewUserPackageSubscriptionsRepository(db)
	var info request.UserAddressDetectionSearch

	info.Page = 1
	info.PageSize = 5
	orderlist, total, err := userAddressDetectionRepo.GetUserPackageSubscriptionsInfoList(context.Background(), info, chatID)
	if err != nil {

		fmt.Println("能量笔数套餐空", err)
	}
	var builder strings.Builder
	if total > 0 {
		//- [6.29] +3000 TRX（订单 #TOPUP-92308）
		for _, order := range orderlist {
			//builder.WriteString(global.Translations[_lang]["address"]+"：")
			builder.WriteString("\n")
			builder.WriteString("<code>" + order.Address + "</code>")
			builder.WriteString("\n")
			//builder.WriteString("状态：")
			////0默认初始化状态  1 自动派送 2 手动 3 结束
			//if order.Status == 3 {
			//	builder.WriteString("<b>" + "已结束" + "</b>")
			//} else if order.Status == 2 {
			//	builder.WriteString("<b>" + "已停止" + "</b>")
			//} else if order.Status == 1 {
			//	builder.WriteString("<b>" + "已开启" + "</b>")
			//} else if order.Status == 0 {
			//	builder.WriteString("<b>" + "初始化" + "</b>")
			//}
			//
			//builder.WriteString("\n")

			builder.WriteString(global.Translations[_lang]["remaining"])
			builder.WriteString(strconv.FormatInt(order.Times, 10))
			builder.WriteString(" " + global.Translations[_lang]["笔"])

			usedTimes := ExtractLeadingInt64(order.BundleName) - order.Times
			builder.WriteString("     " + global.Translations[_lang]["used"])
			builder.WriteString(strconv.FormatInt(usedTimes, 10))
			builder.WriteString(" " + global.Translations[_lang]["笔"])

			////builder.WriteString(" （能量笔数套餐）")

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
			//builder.WriteString("\n")
			builder.WriteString("➖➖➖➖➖➖➖➖➖➖➖➖➖") // 添加分隔符
			//builder.WriteString("\n")            // 添加分隔符
		}
	} else {
		builder.WriteString(global.Translations[_lang]["address_list_empty_tips"] + "\n\n") // 添加分隔符
	}

	// 去除最后一个空格
	result := strings.TrimSpace(builder.String())

	msg := tgbotapi.NewMessage(chatID, "🧾"+global.Translations[_lang]["package_address_list"]+"\n\n"+
		result+"\n")
	msg.ParseMode = "HTML"
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["prev"], "next_bundle_package_address_stats"),
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["next"], "prev_bundle_package_address_stats"),
		),
		tgbotapi.NewInlineKeyboardRow(
			//tgbotapi.NewInlineKeyboardButtonData("解绑地址", "free_monitor_address"),
			tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[_lang]["back_homepage"], "back_bundle_package"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard
	return msg
}
func NEXT_BUNDLE_PACKAGE_ADDRESS_STATS(_lang string, callbackQuery *tgbotapi.CallbackQuery, db *gorm.DB, bot *tgbotapi.BotAPI) bool {
	state := global.DepositStates[callbackQuery.Message.Chat.ID]
	if state == nil {
		var state2 global.DepositState
		state2.CurrentPage = 1
		state = &state2
	}
	//if state != nil && state.CurrentPage > 1 {
	state.CurrentPage = state.CurrentPage + 1
	userAddressDetectionRepo := repositories.NewUserPackageSubscriptionsRepository(db)
	var info request.UserAddressDetectionSearch
	info.PageInfo.Page = state.CurrentPage
	info.PageInfo.PageSize = 10
	orderlist, total, _ := userAddressDetectionRepo.GetUserPackageSubscriptionsInfoList(context.Background(), info, callbackQuery.Message.Chat.ID)

	fmt.Printf("currentpage : %d", state.CurrentPage)
	fmt.Printf("total: %v\n", total)
	totalPages := (total + 5 - 1) / 5

	fmt.Printf("totalPages : %d", totalPages)
	if int64(state.CurrentPage) > totalPages {
		state.CurrentPage = totalPages
		return true
	}
	var builder strings.Builder
	if total > 0 {
		//- [6.29] +3000 TRX（订单 #TOPUP-92308）
		for _, order := range orderlist {
			//builder.WriteString(global.Translations[_lang]["address"]+"：")
			builder.WriteString("\n")
			builder.WriteString("<code>" + order.Address + "</code>")
			builder.WriteString("\n")
			//builder.WriteString("状态：")
			////0默认初始化状态  1 自动派送 2 手动 3 结束
			//if order.Status == 3 {
			//	builder.WriteString("<b>" + "已结束" + "</b>")
			//} else if order.Status == 2 {
			//	builder.WriteString("<b>" + "已停止" + "</b>")
			//} else if order.Status == 1 {
			//	builder.WriteString("<b>" + "已开启" + "</b>")
			//} else if order.Status == 0 {
			//	builder.WriteString("<b>" + "初始化" + "</b>")
			//}
			//
			//builder.WriteString("\n")

			builder.WriteString(global.Translations[_lang]["remaining"])
			builder.WriteString(strconv.FormatInt(order.Times, 10))
			builder.WriteString(" " + global.Translations[_lang]["笔"])

			usedTimes := ExtractLeadingInt64(order.BundleName) - order.Times
			builder.WriteString("     " + global.Translations[_lang]["used"])
			builder.WriteString(strconv.FormatInt(usedTimes, 10))
			builder.WriteString(" " + global.Translations[_lang]["笔"])

			////builder.WriteString(" （能量笔数套餐）")

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
			//builder.WriteString("\n")
			builder.WriteString("➖➖➖➖➖➖➖➖➖➖➖➖➖") // 添加分隔符
			//builder.WriteString("\n")            // 添加分隔符
		}
	} else {
		builder.WriteString(global.Translations[_lang]["address_list_empty_tips"] + "\n\n") // 添加分隔符
	}

	// 去除最后一个空格
	result := strings.TrimSpace(builder.String())
	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "🧾"+global.Translations[_lang]["package_address_list"]+"：</b>\n\n "+
		result+"\n")
	msg.ParseMode = "HTML"
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["prev"], "next_bundle_package_address_stats"),
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["next"], "prev_bundle_package_address_stats"),
		),
		tgbotapi.NewInlineKeyboardRow(
			//tgbotapi.NewInlineKeyboardButtonData("解绑地址", "free_monitor_address"),
			tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[_lang]["back_homepage"], "back_bundle_package"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard
	bot.Send(msg)
	//}
	fmt.Printf("state: %v\n", state)

	global.DepositStates[callbackQuery.Message.Chat.ID] = state
	return false
}

func PREV_BUNDLE_PACKAGE_ADDRESS_STATS(_lang string, callbackQuery *tgbotapi.CallbackQuery, db *gorm.DB, bot *tgbotapi.BotAPI) (*global.DepositState, bool) {
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
		orderlist, total, _ := userAddressDetectionRepo.GetUserPackageSubscriptionsInfoList(context.Background(), info, callbackQuery.Message.Chat.ID)
		var builder strings.Builder
		if total > 0 {
			//- [6.29] +3000 TRX（订单 #TOPUP-92308）
			for _, order := range orderlist {
				builder.WriteString("\n")
				//builder.WriteString(global.Translations[_lang]["address"]+"：")
				builder.WriteString("<code>" + order.Address + "</code>")
				builder.WriteString("\n")
				//builder.WriteString("状态：")
				////0默认初始化状态  1 自动派送 2 手动 3 结束
				//if order.Status == 3 {
				//	builder.WriteString("<b>" + "已结束" + "</b>")
				//} else if order.Status == 2 {
				//	builder.WriteString("<b>" + "已停止" + "</b>")
				//} else if order.Status == 1 {
				//	builder.WriteString("<b>" + "已开启" + "</b>")
				//} else if order.Status == 0 {
				//	builder.WriteString("<b>" + "初始化" + "</b>")
				//}
				//
				//builder.WriteString("\n")

				builder.WriteString(global.Translations[_lang]["remaining"])
				builder.WriteString(strconv.FormatInt(order.Times, 10))
				builder.WriteString(" " + global.Translations[_lang]["笔"])

				usedTimes := ExtractLeadingInt64(order.BundleName) - order.Times
				builder.WriteString("     " + global.Translations[_lang]["used"])
				builder.WriteString(strconv.FormatInt(usedTimes, 10))
				builder.WriteString(" " + global.Translations[_lang]["笔"])

				////builder.WriteString(" （能量笔数套餐）")

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
				//builder.WriteString("\n")
				builder.WriteString("➖➖➖➖➖➖➖➖➖➖➖➖➖") // 添加分隔符
				//builder.WriteString("\n")            // 添加分隔符
			}
		} else {
			builder.WriteString(global.Translations[_lang]["address_list_empty_tips"] + "\n\n") // 添加分隔符
		}

		// 去除最后一个空格
		result := strings.TrimSpace(builder.String())
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "🧾"+global.Translations[_lang]["package_address_list"]+"：</b>\n\n "+
			result+"\n")
		msg.ParseMode = "HTML"
		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["prev"], "next_bundle_package_address_stats"),
				tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["next"], "prev_bundle_package_address_stats"),
			),
			tgbotapi.NewInlineKeyboardRow(
				//tgbotapi.NewInlineKeyboardButtonData("解绑地址", "free_monitor_address"),
				tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[_lang]["back_homepage"], "back_bundle_package"),
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
		orderlist, total, _ := userAddressDetectionRepo.GetUserPackageSubscriptionsInfoList(context.Background(), info, callbackQuery.Message.Chat.ID)
		var builder strings.Builder
		if total > 0 {
			//- [6.29] +3000 TRX（订单 #TOPUP-92308）
			for _, order := range orderlist {
				//builder.WriteString(global.Translations[_lang]["address"]+"：")
				builder.WriteString("\n")
				builder.WriteString("<code>" + order.Address + "</code>")
				builder.WriteString("\n")
				//builder.WriteString("状态：")
				////0默认初始化状态  1 自动派送 2 手动 3 结束
				//if order.Status == 3 {
				//	builder.WriteString("<b>" + "已结束" + "</b>")
				//} else if order.Status == 2 {
				//	builder.WriteString("<b>" + "已停止" + "</b>")
				//} else if order.Status == 1 {
				//	builder.WriteString("<b>" + "已开启" + "</b>")
				//} else if order.Status == 0 {
				//	builder.WriteString("<b>" + "初始化" + "</b>")
				//}
				//
				//builder.WriteString("\n")

				builder.WriteString(global.Translations[_lang]["remaining"])
				builder.WriteString(strconv.FormatInt(order.Times, 10))
				builder.WriteString(" " + global.Translations[_lang]["笔"])

				usedTimes := ExtractLeadingInt64(order.BundleName) - order.Times
				builder.WriteString("     " + global.Translations[_lang]["used"])
				builder.WriteString(strconv.FormatInt(usedTimes, 10))
				builder.WriteString(" " + global.Translations[_lang]["笔"])

				////builder.WriteString(" （能量笔数套餐）")

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
				//builder.WriteString("\n")
				builder.WriteString("➖➖➖➖➖➖➖➖➖➖➖➖➖") // 添加分隔符
				//builder.WriteString("\n")            // 添加分隔符
			}
		} else {
			builder.WriteString(global.Translations[_lang]["address_list_empty_tips"] + "\n\n") // 添加分隔符
		}

		// 去除最后一个空格
		result := strings.TrimSpace(builder.String())
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "🧾"+global.Translations[_lang]["package_address_list"]+"：</b>\n\n "+
			result+"\n")
		msg.ParseMode = "HTML"
		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["prev"], "next_bundle_package_address_stats"),
				tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["next"], "prev_bundle_package_address_stats"),
			),
			tgbotapi.NewInlineKeyboardRow(
				//tgbotapi.NewInlineKeyboardButtonData("解绑地址", "free_monitor_address"),
				tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[_lang]["back_homepage"], "back_bundle_package"),
			),
		)
		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)
	}
	return state, false
}
func APPLY_ST_BUNDLE_PACKAGE(trxfeeClient *trxfee.TrxfeeClient, _lang string, cache cache.Cache, bot *tgbotapi.BotAPI, message *tgbotapi.Message, db *gorm.DB, status string) bool {
	if !IsValidAddress(message.Text) {
		msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[_lang]["invalid_address_tips"]+"</b>"+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)
		return true
	}

	bundleID := strings.ReplaceAll(status, "apply_ST_bundle_package_", "")
	userOperationBundlesRepo := repositories.NewUserSmartTransactionBundlesRepository(db)
	bundlePackage, err := userOperationBundlesRepo.Query(context.Background(), bundleID)

	if err != nil {
		fmt.Println(err)
	}
	userRepo := repositories.NewUserRepository(db)
	user, _ := userRepo.GetByUserID(message.Chat.ID)

	lessBalance := false
	if bundlePackage.Token == "USDT" {
		//扣usdt
		if flag, _ := CompareNumberStrings(user.Amount, bundlePackage.Amount); flag < 0 {
			lessBalance = true
		}
		fmt.Printf("bundle %v is USDT\n", bundlePackage)
	} else if bundlePackage.Token == "TRX" {
		//扣trx
		if flag, _ := CompareNumberStrings(user.TronAmount, bundlePackage.Amount); flag < 0 {
			lessBalance = true
		}

		fmt.Printf("bundle %v is trx\n", bundlePackage)
	}

	if lessBalance {
		msg := tgbotapi.NewMessage(message.Chat.ID,
			"🆔"+global.Translations[_lang]["user_id"]+": "+user.Associates+"\n"+
				"👤"+global.Translations[_lang]["username"]+": @"+user.Username+"\n"+
				"💰"+global.Translations[_lang]["balance"]+": "+"\n"+
				"- TRX：   "+user.TronAmount+"\n"+
				"- USDT："+user.Amount)

		msg.ParseMode = "HTML"

		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("💵"+global.Translations[_lang]["deposit"], "deposit_amount"),
			),
		)

		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)

		return false
	}

	//加入訂閲記錄
	userPackageSubscriptionsRepo := repositories.NewUserSmartTransactionPackageSubscriptionsRepository(db)

	//判断是否已经购买的地址，在进行中的
	item, err := userPackageSubscriptionsRepo.Get(message.Text)

	if err != nil {
		return false
	}

	if item.Id > 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, global.Translations[_lang]["smart_transaction_plans_repeat_order"]+
			"🆔"+global.Translations[_lang]["user_id"]+": "+user.Associates+"\n"+
			"👤"+global.Translations[_lang]["username"]+": @"+user.Username+"\n"+
			"💰"+global.Translations[_lang]["balance"]+": "+"\n"+
			"- TRX：   "+user.TronAmount+"\n"+
			"- USDT："+user.Amount)

		msg.ParseMode = "HTML"

		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("💵"+global.Translations[_lang]["deposit"], "deposit_amount"),
			),
		)

		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)

		return false
	}

	//扣錢
	if bundlePackage.Token == "TRX" {
		balance, _ := SubtractStringNumbers(user.TronAmount, bundlePackage.Amount, 1)
		fmt.Printf("TRX balance %s\n", balance)
		user.TronAmount = balance
	} else if bundlePackage.Token == "USDT" {
		balance, _ := SubtractStringNumbers(user.Amount, bundlePackage.Amount, 1)
		fmt.Printf("USDT balance %s\n", balance)

		user.Amount = balance
	}

	err = userRepo.Update2(context.Background(), &user)
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

	fmt.Printf("address %s\n", record.Address)
	fmt.Printf("times %d\n", record.Times)

	trxfeeClient.TimesOrder(record.Address, int(record.Times))

	msg := tgbotapi.NewMessage(message.Chat.ID, "✅"+global.Translations[_lang]["smart_transaction_package_order_purchased_successfully"]+"\n"+
		global.Translations[_lang]["package_name"]+"："+strings.ReplaceAll(bundlePackage.Name, "笔", global.Translations[_lang]["笔"])+"\n"+
		global.Translations[_lang]["payment_amount"]+"："+bundlePackage.Amount+" "+bundlePackage.Token+"\n"+
		global.Translations[_lang]["address"]+"："+message.Text+"\n"+
		global.Translations[_lang]["order_id"]+"："+fmt.Sprintf("%d", record.Id)+""+"\n")
	msg.ParseMode = "HTML"
	// 当点击"按钮 1"时显示内联键盘
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🧾"+global.Translations[_lang]["smart_transaction_package_address_list"], "click_bundle_package_address_stats_ST"),
			tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[_lang]["back_homepage"], "back_bundle_package_ST"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard

	bot.Send(msg)

	expiration := 1 * time.Minute // 短时间缓存空值

	//设置用户状态
	cache.Set(strconv.FormatInt(message.Chat.ID, 10), "null_apply_bundle_package_address_ST", expiration)
	return false
}
func CLICK_BUNDLE_PACKAGE_ADDRESS_STATS_ST(_lang string, db *gorm.DB, chatID int64) tgbotapi.MessageConfig {

	//fmt.Println("ExtractBundlePackage")
	userAddressDetectionRepo := repositories.NewUserSmartTransactionPackageSubscriptionsRepository(db)
	var info request.UserAddressDetectionSearch

	info.Page = 1
	info.PageSize = 100000
	orderlist, total, err := userAddressDetectionRepo.GetUserSmartTransactionPackageSubscriptionsInfoList(context.Background(), info, chatID)
	if err != nil {

		fmt.Println("能量笔数套餐空", err)
	}
	var builder strings.Builder
	if total > 0 {
		//- [6.29] +3000 TRX（订单 #TOPUP-92308）
		for _, order := range orderlist {
			//builder.WriteString(global.Translations[_lang]["address"]+"：")
			builder.WriteString("\n")
			builder.WriteString("<code>" + order.Address + "</code>")
			builder.WriteString("\n")
			//builder.WriteString("状态：")
			////0默认初始化状态  1 自动派送 2 手动 3 结束
			//if order.Status == 3 {
			//	builder.WriteString("<b>" + "已结束" + "</b>")
			//} else if order.Status == 2 {
			//	builder.WriteString("<b>" + "已停止" + "</b>")
			//} else if order.Status == 1 {
			//	builder.WriteString("<b>" + "已开启" + "</b>")
			//} else if order.Status == 0 {
			//	builder.WriteString("<b>" + "初始化" + "</b>")
			//}
			//
			//builder.WriteString("\n")
			if order.Status == 2 {
				builder.WriteString(global.Translations[_lang]["smart_transaction_disable"])
			}
			if order.Status == 1 {
				builder.WriteString(global.Translations[_lang]["smart_transaction_enable"])
			}
			builder.WriteString(global.Translations[_lang]["remaining"])
			builder.WriteString(strconv.FormatInt(order.Times, 10))
			builder.WriteString(" " + global.Translations[_lang]["笔"])

			usedTimes := ExtractLeadingInt64(order.BundleName) - order.Times
			builder.WriteString("     " + global.Translations[_lang]["used"])
			builder.WriteString(strconv.FormatInt(usedTimes, 10))
			builder.WriteString(" " + global.Translations[_lang]["笔"])

			////builder.WriteString(" （能量笔数套餐）")

			builder.WriteString("\n") // 添加分隔符
			if order.Times > 0 {
				if order.Status == 2 {
					builder.WriteString(global.Translations[_lang]["close_auto_dispatch_energy"] + ":/close_ST")
					builder.WriteString(strconv.FormatInt(order.Id, 10))
				}
				if order.Status == 1 {
					builder.WriteString(global.Translations[_lang]["open_auto_dispatch_energy"] + ":/open_ST")
					builder.WriteString(strconv.FormatInt(order.Id, 10))
				}
				builder.WriteString("\n") // 添加分隔符
			}
			//builder.WriteString("\n")
			builder.WriteString("➖➖➖➖➖➖➖➖➖➖➖➖➖") // 添加分隔符
			//builder.WriteString("\n")            // 添加分隔符
		}
	} else {
		builder.WriteString(global.Translations[_lang]["smart_transaction__list_empty_tips"] + "\n\n") // 添加分隔符
	}

	// 去除最后一个空格
	result := strings.TrimSpace(builder.String())

	msg := tgbotapi.NewMessage(chatID, "🧾"+global.Translations[_lang]["smart_transaction_package_address_list"]+"\n\n"+
		result+"\n")
	msg.ParseMode = "HTML"
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		//tgbotapi.NewInlineKeyboardRow(
		//	tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["prev"], "next_bundle_package_address_stats"),
		//	tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["next"], "prev_bundle_package_address_stats"),
		//),
		tgbotapi.NewInlineKeyboardRow(
			//tgbotapi.NewInlineKeyboardButtonData("解绑地址", "free_monitor_address"),
			tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[_lang]["back_homepage"], "back_bundle_package_ST"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard
	return msg
}
