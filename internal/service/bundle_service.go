package service

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"ushield_bot/internal/cache"
	"ushield_bot/internal/domain"
	"ushield_bot/internal/global"
	"ushield_bot/internal/infrastructure/repositories"
	. "ushield_bot/internal/infrastructure/tools"
	logger "ushield_bot/internal/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func CheckBundlePackage(lang string, cache cache.Cache, bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *gorm.DB) {
	//deductionAmount := callbackQuery.Data[7:len(callbackQuery.Data)]
	userOperationBundlesRepo := repositories.NewUserOperationBundlesRepository(db)
	bundleID := strings.ReplaceAll(callbackQuery.Data, "bundle_", "")
	bundlePackage, err := userOperationBundlesRepo.GetByID(context.Background(), bundleID)
	if err != nil {

	}

	deductionAmount := bundlePackage.Amount

	//logger.Printf("deductionAmount: %v\n", deductionAmount)
	userRepo := repositories.NewUserRepository(db)
	user, _ := userRepo.GetByChatID(callbackQuery.Message.Chat.ID)
	if IsEmpty(user.Amount) {
		user.Amount = "0"
	}

	if IsEmpty(user.TronAmount) {
		user.TronAmount = "0"
	}

	lessBalance := false
	if bundlePackage.Token == "USDT" {
		//扣usdt
		if flag, _ := CompareNumberStrings(user.Amount, deductionAmount); flag < 0 {
			lessBalance = true
		}
	} else if bundlePackage.Token == "TRX" {
		//扣trx
		if flag, _ := CompareNumberStrings(user.TronAmount, deductionAmount); flag < 0 {
			lessBalance = true
		}

	}

	if lessBalance {

		videoPath := "./static/Audi.png"

		// 创建视频消息（从本地文件）
		msg := tgbotapi.NewPhoto(callbackQuery.Message.Chat.ID, tgbotapi.FilePath(videoPath))

		if bundlePackage.Token == "TRX" {

			trxPlaceholderRepo := repositories.NewUserTRXPlaceholdersRepository(db)
			placeholder, queryErr := trxPlaceholderRepo.GetRandomAvailable(context.Background())
			if queryErr != nil {
			}
			err := trxPlaceholderRepo.UpdateStatusByID(context.Background(), placeholder.Id, 1)
			if err != nil {
				logger.Errorf("Error updating trx placeholder: %v", err)
			}
			realTransferAmount := AddStringsAsFloats(placeholder.Placeholder, bundlePackage.Amount)

			//生成订单
			trxDepositRepo := repositories.NewUserTRXDepositsRepository(db)

			orderNO := Generate6DigitOrderNo()
			var trxDeposit domain.UserTRXDeposits
			trxDeposit.OrderNO = orderNO
			trxDeposit.UserID = callbackQuery.Message.Chat.ID
			trxDeposit.Status = 0
			bundle, _ := strconv.ParseInt(bundleID, 10, 64)
			trxDeposit.BundleId = bundle
			trxDeposit.Source = 6

			trxDeposit.Placeholder = placeholder.Placeholder

			agent := os.Getenv("BOT_AGENT")
			sysUserRepo := repositories.NewSysUsersRepository(db)
			_, depositAddress, _ := sysUserRepo.GetAddressesByUsername(context.Background(), agent)
			trxDeposit.Address = depositAddress
			trxDeposit.Amount = bundlePackage.Amount
			trxDeposit.CreatedAt = time.Now()

			createErr := trxDepositRepo.Create(context.Background(), &trxDeposit)
			if createErr != nil {
				logger.Errorf("Error creating trxDeposit: %v", createErr)
			}

			msg.Caption = global.Translations[lang]["order_id"] + "：TOPUP-" + trxDeposit.OrderNO + "\n" +
				global.Translations[lang]["payment_amount"] + "：" + "<code>" + realTransferAmount + "</code>" + " TRX " + global.Translations[lang]["copy_text_tips"] + "\n" +
				global.Translations[lang]["receive_address"] + "<code>" + trxDeposit.Address + "</code>" + global.Translations[lang]["copy_text_tips"] + "\n" +
				global.Translations[lang]["tx_time_limit_tips"] + "\n" +
				global.Translations[lang]["deposit_time_label"] + FormatDateTimeValue(trxDeposit.CreatedAt) + "\n" +
				global.Translations[lang]["amount_suffix_tips"] + "\n"

			inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("⏳"+global.Translations[lang]["catfee_smart_transaction_pay_button"]+realTransferAmount+" TRX ", "noop"),
				),

				tgbotapi.NewInlineKeyboardRow(

					tgbotapi.NewInlineKeyboardButtonData("🔙"+global.Translations[lang]["back_home"], "back_home"),
					tgbotapi.NewInlineKeyboardButtonData("❌"+global.Translations[lang]["cancel_order"], "cancel_order"),
				))
			msg.ReplyMarkup = inlineKeyboard
			msg.ParseMode = "HTML"
			//msg.DisableWebPagePreview = true
			sent, _ := bot.Send(msg)

			expiration := 1 * time.Minute // 短时间缓存空值
			//设置用户状态
			cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10)+"_order", strconv.Itoa(sent.MessageID), expiration)

			//expiration := 1 * time.Minute // 短时间缓存空值

			//设置用户状态
			cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10)+"_order_no", "TRX_"+trxDeposit.OrderNO, expiration)
			return
		}

		if bundlePackage.Token == "USDT" {

			usdtPlaceholderRepo := repositories.NewUserUsdtPlaceholdersRepository(db)
			placeholder, queryErr := usdtPlaceholderRepo.GetRandomAvailable(context.Background())
			if queryErr != nil {
			}
			err := usdtPlaceholderRepo.UpdateStatusByID(context.Background(), placeholder.Id, 1)
			if err != nil {
				logger.Errorf("Error updating usdt placeholder: %v", err)
			}
			realTransferAmount := AddStringsAsFloats(placeholder.Placeholder, bundlePackage.Amount)

			//生成订单
			usdtDepositRepo := repositories.NewUserUSDTDepositsRepository(db)

			orderNO := Generate6DigitOrderNo()
			var usdtDeposit domain.UserUSDTDeposits
			usdtDeposit.OrderNO = orderNO
			usdtDeposit.UserID = callbackQuery.Message.Chat.ID
			usdtDeposit.Status = 0
			bundle, _ := strconv.ParseInt(bundleID, 10, 64)
			usdtDeposit.BundleId = bundle
			usdtDeposit.Source = 6

			usdtDeposit.Placeholder = placeholder.Placeholder

			agent := os.Getenv("BOT_AGENT")
			sysUserRepo := repositories.NewSysUsersRepository(db)
			_, depositAddress, _ := sysUserRepo.GetAddressesByUsername(context.Background(), agent)
			usdtDeposit.Address = depositAddress
			usdtDeposit.Amount = bundlePackage.Amount
			usdtDeposit.CreatedAt = time.Now()

			createErr := usdtDepositRepo.Create(context.Background(), &usdtDeposit)
			if createErr != nil {
				logger.Errorf("Error creating usdtDeposit: %v", createErr)
			}

			msg.Caption = global.Translations[lang]["order_id"] + "：TOPUP-" + usdtDeposit.OrderNO + "\n" +
				global.Translations[lang]["payment_amount"] + "：" + "<code>" + realTransferAmount + "</code>" + " USDT " + global.Translations[lang]["copy_text_tips"] + "\n" +
				global.Translations[lang]["receive_address"] + "<code>" + usdtDeposit.Address + "</code>" + global.Translations[lang]["copy_text_tips"] + "\n" +
				global.Translations[lang]["tx_time_limit_tips"] + "\n" +
				global.Translations[lang]["deposit_time_label"] + FormatDateTimeValue(usdtDeposit.CreatedAt) + "\n" +
				global.Translations[lang]["amount_suffix_tips"] + "\n"

			inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("⏳"+global.Translations[lang]["catfee_smart_transaction_pay_button"]+realTransferAmount+" USDT ", "noop"),
				),

				tgbotapi.NewInlineKeyboardRow(

					tgbotapi.NewInlineKeyboardButtonData("🔙"+global.Translations[lang]["back_home"], "back_home"),
					tgbotapi.NewInlineKeyboardButtonData("❌"+global.Translations[lang]["cancel_order"], "cancel_order"),
				))
			msg.ReplyMarkup = inlineKeyboard
			msg.ParseMode = "HTML"
			//msg.DisableWebPagePreview = true
			sent, _ := bot.Send(msg)

			expiration := 1 * time.Minute // 短时间缓存空值
			//设置用户状态
			cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10)+"_order", strconv.Itoa(sent.MessageID), expiration)
			cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10)+"_order_no", "USDT_"+usdtDeposit.OrderNO, expiration)
			//扫码支付

			//msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID,
			//	"🆔"+global.Translations[lang]["user_id"]+": "+user.Associates+"\n"+
			//		"👤"+global.Translations[lang]["username"]+": @"+user.Username+"\n"+
			//		"💰"+global.Translations[lang]["balance"]+": "+"\n"+
			//		"- TRX：   "+user.TronAmount+"\n"+
			//		"-  USDT："+user.Amount)
			//
			//msg.ParseMode = "HTML"
			//
			//inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			//	tgbotapi.NewInlineKeyboardRow(
			//		tgbotapi.NewInlineKeyboardButtonData("💵"+global.Translations[lang]["deposit"], "deposit_amount"),
			//	),
			//)
			//
			//msg.ReplyMarkup = inlineKeyboard
			//bot.Send(msg)

			return
		}
		return
	}

	//扣款

	//扣錢
	if bundlePackage.Token == "TRX" {
		balance, _ := SubtractStringNumbers(user.TronAmount, bundlePackage.Amount, 1)
		user.TronAmount = balance
	} else if bundlePackage.Token == "USDT" {
		balance, _ := SubtractStringNumbers(user.Amount, bundlePackage.Amount, 1)
		user.Amount = balance
	}

	err = userRepo.Save(context.Background(), &user)
	if err != nil {

	}

	//加入訂閲記錄

	bundleTimes := ExtractLeadingInt64(bundlePackage.Name)

	_bundleTimes := bundleTimes + user.BundleTimes
	err = userRepo.UpdateBundleTimesByChatID(_bundleTimes, callbackQuery.Message.Chat.ID)
	if err != nil {
		return
	}

	//加入訂閲記錄
	userPackageSubscriptionsRepo := repositories.NewUserPackageSubscriptionsRepository(db)
	var record domain.UserPackageSubscriptions
	record.ChatID = callbackQuery.Message.Chat.ID
	record.Address = ""
	bundle, _ := strconv.ParseInt(bundleID, 10, 64)

	record.BundleID = bundle
	record.Status = 2
	record.Amount = bundlePackage.Amount
	record.Times = ExtractLeadingInt64(bundlePackage.Name)
	record.BundleName = bundlePackage.Name
	err = userPackageSubscriptionsRepo.Create(context.Background(), &record)
	if err != nil {
		return
	}

	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "✅"+"🧾"+global.Translations[lang]["package_order_purchased_successfully"]+"\n\n"+
		global.Translations[lang]["package_name"]+"："+strings.ReplaceAll(bundlePackage.Name, "笔", global.Translations[lang]["笔"])+"\n"+
		global.Translations[lang]["payment_amount"]+"："+bundlePackage.Amount+" "+bundlePackage.Token+"\n"+
		//global.Translations[lang]["address"]+"："+message.Text+"\n\n"+
		global.Translations[lang]["order_id"]+"："+fmt.Sprintf("%d", record.Id)+""+"\n")
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
	cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), "null_apply_bundle_package_address", expiration)
	return

}

func HandleBundleSubscriptionInput(lang string, message *tgbotapi.Message, bot *tgbotapi.BotAPI, db *gorm.DB, status string) bool {
	if !IsValidAddress(message.Text) {
		msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[lang]["invalid_address_tips"]+"</b>"+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)
		return true
	}

	userRepo := repositories.NewUserRepository(db)
	user, _ := userRepo.GetByChatID(message.Chat.ID)

	fee := status[7:len(status)]
	if CompareStringsWithFloat(fee, user.Amount, 1) {
		//余额不足，需充值
		msg := tgbotapi.NewMessage(message.Chat.ID,
			//"💬"+"<b>"+"余额不足: "+"</b>"+"\n"+
			//	"💬"+"<b>"+"用户姓名: "+"</b>"+user.Username+"\n"+
			//	"👤"+"<b>"+"用户电报ID: "+"</b>"+user.Associates+"\n"+
			//	"💵"+"<b>"+"当前TRX余额:  "+"</b>"+user.TronAmount+" TRX"+"\n"+
			//	"💴"+"<b>"+"当前USDT余额:  "+"</b>"+user.Amount+" USDT")
			"🆔"+global.Translations[lang]["user_id"]+": <code>"+user.Associates+"</code>\n"+
				"👤"+global.Translations[lang]["username"]+": @"+user.Username+"\n"+
				"💰"+global.Translations[lang]["balance"]+": "+"\n"+
				"- TRX：   "+user.TronAmount+"\n"+
				"-  USDT："+user.Amount)
		msg.ParseMode = "HTML"
		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("💵"+global.Translations[lang]["deposit"], "deposit_amount"),
			),
		)

		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)
	} else {
		//扣款
		//调用trxfee接口

		//trxfeeHandler := handler.NewTrxfeeHandler()

		//trxfeeHandler.RequestTimesOrder(context.Background(),"","",message.Text,)
		rest, _ := SubtractStringNumbers(user.Amount, fee, 1)
		user.Amount = rest
		userRepo.Save(context.Background(), &user)
		msg := tgbotapi.NewMessage(message.Chat.ID,
			"<b>"+"✅笔数套餐订阅成功"+"</b>"+"\n"+
				//"💬"+"<b>"+"用户姓名: "+"</b>"+user.Username+"\n"+
				//"👤"+"<b>"+"用户电报ID: "+"</b>"+user.Associates+"\n"+
				//"💵"+"<b>"+"当前TRX余额:  "+"</b>"+user.TronAmount+" TRX"+"\n"+
				//"💴"+"<b>"+"当前USDT余额:  "+"</b>"+user.Amount+" USDT")
				"🆔"+global.Translations[lang]["user_id"]+": <code>"+user.Associates+"</code>\n"+
				"👤"+global.Translations[lang]["username"]+": @"+user.Username+"\n"+
				"💰"+global.Translations[lang]["balance"]+": "+"\n"+
				"- TRX：   "+user.TronAmount+"\n"+
				"-  USDT："+user.Amount)
		msg.ParseMode = "HTML"
		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("💵"+global.Translations[lang]["deposit"], "deposit_amount"),
			),
		)

		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)
	}
	return false
}
