package catfee

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"ushield_bot/internal/cache"
	"ushield_bot/internal/domain"
	"ushield_bot/internal/global"
	"ushield_bot/internal/infrastructure/repositories"
	. "ushield_bot/internal/infrastructure/tools"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func BUNDLE_CHECK2(_lang string, cache cache.Cache, bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *gorm.DB) {
	//deductionAmount := callbackQuery.Data[7:len(callbackQuery.Data)]
	userOperationBundlesRepo := repositories.NewUserOperationBundlesRepository(db)
	bundleID := strings.ReplaceAll(callbackQuery.Data, "bundle_", "")
	bundlePackage, err := userOperationBundlesRepo.Query(context.Background(), bundleID)

	if err != nil {

	}

	deductionAmount := bundlePackage.Amount

	//fmt.Printf("deductionAmount: %v\n", deductionAmount)
	userRepo := repositories.NewUserRepository(db)
	user, _ := userRepo.GetByUserID(callbackQuery.Message.Chat.ID)
	if IsEmpty(user.Amount) {
		user.Amount = "0"
	}

	if IsEmpty(user.TronAmount) {
		user.TronAmount = "0"
	}

	fmt.Printf("user usdt balance : %s\n", user.Amount)
	fmt.Printf("user  trx balance : %s\n", user.TronAmount)
	fmt.Printf("deductionAmount : %s\n", deductionAmount)
	fmt.Printf("Token : %s\n", bundlePackage.Token)

	lessBalance := false
	if bundlePackage.Token == "USDT" {
		//扣usdt
		if flag, _ := CompareNumberStrings(user.Amount, deductionAmount); flag < 0 {
			lessBalance = true
		}
		fmt.Printf("bundle %v is USDT\n", bundlePackage)
	} else if bundlePackage.Token == "TRX" {
		//扣trx
		if flag, _ := CompareNumberStrings(user.TronAmount, deductionAmount); flag < 0 {
			lessBalance = true
		}

		fmt.Printf("bundle %v is trx\n", bundlePackage)
	}

	if lessBalance {
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID,
			//"💬"+"<b>"+"用户姓名: "+"</b>"+user.Username+"\n"+
			//	"👤"+"<b>"+"用户电报ID: "+"</b>"+user.Associates+"\n"+
			//	"💵"+"<b>"+"余额不足 "+"</b>"+"\n"+
			//	"💴"+"<b>"+"当前TRX余额:  "+"</b>"+user.TronAmount+" TRX"+"\n"+
			//	"💴"+"<b>"+"当前USDT余额:  "+"</b>"+user.Amount+" USDT")

			"🆔"+global.Translations[_lang]["user_id"]+": "+user.Associates+"\n"+
				"👤"+global.Translations[_lang]["username"]+": @"+user.Username+"\n"+
				"💰"+global.Translations[_lang]["balance"]+": "+"\n"+
				"- TRX：   "+user.TronAmount+"\n"+
				"-  USDT："+user.Amount)

		msg.ParseMode = "HTML"

		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("💵"+global.Translations[_lang]["deposit"], "deposit_amount"),
			),
		)

		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)

		return
	}

	//扣款

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

	bundleTimes := ExtractLeadingInt64(bundlePackage.Name)

	_bundleTimes := bundleTimes + user.BundleTimes
	err = userRepo.UpdateBundleTimes(_bundleTimes, callbackQuery.Message.Chat.ID)
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

	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "✅"+"🧾"+global.Translations[_lang]["package_order_purchased_successfully"]+"\n\n"+
		global.Translations[_lang]["package_name"]+"："+strings.ReplaceAll(bundlePackage.Name, "笔", global.Translations[_lang]["笔"])+"\n"+
		global.Translations[_lang]["payment_amount"]+"："+bundlePackage.Amount+" "+bundlePackage.Token+"\n"+
		//global.Translations[_lang]["address"]+"："+message.Text+"\n\n"+
		global.Translations[_lang]["order_id"]+"："+fmt.Sprintf("%d", record.Id)+""+"\n")
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
	cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), "null_apply_bundle_package_address", expiration)
	return

}

func ST_BUNDLE_CHECK(_lang string, cache cache.Cache, bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *gorm.DB) {
	//deductionAmount := callbackQuery.Data[7:len(callbackQuery.Data)]
	userOperationBundlesRepo := repositories.NewUserSmartTransactionBundlesRepository(db)
	bundleID := strings.ReplaceAll(callbackQuery.Data, "ST_bundle_", "")
	bundlePackage, err := userOperationBundlesRepo.Query(context.Background(), bundleID)

	if err != nil {

	}

	deductionAmount := bundlePackage.Amount

	//fmt.Printf("deductionAmount: %v\n", deductionAmount)
	userRepo := repositories.NewUserRepository(db)
	user, _ := userRepo.GetByUserID(callbackQuery.Message.Chat.ID)
	if IsEmpty(user.Amount) {
		user.Amount = "0"
	}

	if IsEmpty(user.TronAmount) {
		user.TronAmount = "0"
	}

	fmt.Printf("user usdt balance : %s\n", user.Amount)
	fmt.Printf("user  trx balance : %s\n", user.TronAmount)
	fmt.Printf("deductionAmount : %s\n", deductionAmount)
	fmt.Printf("Token : %s\n", bundlePackage.Token)

	lessBalance := false
	if bundlePackage.Token == "USDT" {
		//扣usdt
		if flag, _ := CompareNumberStrings(user.Amount, deductionAmount); flag < 0 {
			lessBalance = true
		}
		fmt.Printf("bundle %v is USDT\n", bundlePackage)
	} else if bundlePackage.Token == "TRX" {
		//扣trx
		if flag, _ := CompareNumberStrings(user.TronAmount, deductionAmount); flag < 0 {
			lessBalance = true
		}

		fmt.Printf("bundle %v is trx\n", bundlePackage)
	}

	if lessBalance {
		if bundlePackage.Token == "TRX" {
			trxPlaceholderRepo := repositories.NewUserTRXPlaceholdersRepository(db)
			placeholder, esg := trxPlaceholderRepo.Query(context.Background())
			//err := trxPlaceholderRepo.Update(context.Background(), placeholder.Id, 1)
			if esg != nil {
				fmt.Printf("Failed to update user: " + esg.Error())
				msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[_lang]["placeholder_array_size_warning"])

				inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("🕣"+global.Translations[_lang]["cancel_order"], "cancel_order"),
						tgbotapi.NewInlineKeyboardButtonData("🔙"+global.Translations[_lang]["back_home"], "back_home"),
					))
				msg.ReplyMarkup = inlineKeyboard
				msg.ParseMode = "HTML"
				//msg.DisableWebPagePreview = true
				bot.Send(msg)
				return

			}
			if placeholder.Id == 0 {
				msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[_lang]["placeholder_array_size_warning"])
				inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("🕣"+global.Translations[_lang]["cancel_order"], "cancel_order"),
						tgbotapi.NewInlineKeyboardButtonData("🔙"+global.Translations[_lang]["back_home"], "back_home"),
					))
				msg.ReplyMarkup = inlineKeyboard
				msg.ParseMode = "HTML"
				//msg.DisableWebPagePreview = true
				bot.Send(msg)

				return
			}

			err := trxPlaceholderRepo.Update(context.Background(), placeholder.Id, 1)
			if err != nil {
				log.Printf("Error updating usdt placeholder: %v", err)
			}
			realTransferAmount := AddStringsAsFloats(placeholder.Placeholder, bundlePackage.Amount)

			fmt.Printf("realTransferAmount: %s\n", realTransferAmount)

			//生成订单
			trxDepositRepo := repositories.NewUserTRXDepositsRepository(db)

			orderNO := Generate6DigitOrderNo()
			var trxDeposit domain.UserTRXDeposits
			trxDeposit.OrderNO = orderNO
			trxDeposit.UserID = callbackQuery.Message.Chat.ID
			trxDeposit.Status = 0
			trxDeposit.Placeholder = placeholder.Placeholder

			//来自于波场伴侣  //source  0代表充值、1代表智能托管、2代表检测、3代表预警
			trxDeposit.Source = 1
			value, err := strconv.ParseInt(bundleID, 10, 64)
			if err != nil {
				fmt.Printf("套餐ID转换失败: %v\n", err)
				return
			}

			trxDeposit.BundleId = value

			//dictRepo := repositories.NewSysDictionariesRepo(db)
			_agent := os.Getenv("Agent")
			//depositAddress, _ := dictRepo.GetDepositAddress(_agent)
			//_agent := os.Getenv("Agent")
			sysUserRepo := repositories.NewSysUsersRepository(db)
			_, depositAddress, _ := sysUserRepo.Find(context.Background(), _agent)
			trxDeposit.Address = depositAddress
			trxDeposit.Amount = bundlePackage.Amount
			trxDeposit.CreatedAt = time.Now()

			errsg := trxDepositRepo.Create(context.Background(), &trxDeposit)
			if errsg != nil {
				log.Printf("Error creating trxDeposit: %v", errsg)
			}

			//msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID,
			//	global.Translations[_lang]["order_id"]+"：TOPUP-"+usdtDeposit.OrderNO+"\n"+
			//		global.Translations[_lang]["payment_amount"]+"："+"<code>"+realTransferAmount+"</code>"+" USDT "+global.Translations[_lang]["copy_text_tips"]+"\n"+
			//		global.Translations[_lang]["receive_address"]+"<code>"+usdtDeposit.Address+"</code>"+global.Translations[_lang]["copy_text_tips"]+"\n"+
			//		global.Translations[_lang]["tx_time_limit_tips"]+"\n"+
			//		global.Translations[_lang]["deposit_time_label"]+Format4Chinesese(usdtDeposit.CreatedAt)+"\n"+
			//		global.Translations[_lang]["amount_suffix_tips"]+"\n")

			videoPath := "./static/Audi.png"

			// 创建视频消息（从本地文件）
			msg := tgbotapi.NewPhoto(callbackQuery.Message.Chat.ID, tgbotapi.FilePath(videoPath))

			msg.Caption =
				//"🧾 智能托管套餐订单创建成功\n\n套餐："+bundlePackage.Name+"\n\n支付金额："+realTransferAmount+bundlePackage.Token+"\n收款地址："+"<code>"+trxDeposit.Address+"</code>"+"\n\n"+global.Translations[_lang]["order_id"]+"：PKG-"+trxDeposit.OrderNO+"\n\n订单有效期：10 分钟\n\n⚠️ 请务必准确输入尾数金额，否则将无法入账！")

				global.Translations[_lang]["catfee_smart_transaction_head_1"] +
					bundlePackage.Name + global.Translations[_lang]["catfee_smart_transaction_head_2"] +
					"<code>" + realTransferAmount + "</code>" +
					bundlePackage.Token +
					global.Translations[_lang]["catfee_smart_transaction_head_3"] +
					"<code>" + trxDeposit.Address + "</code>" + "\n\n" +
					global.Translations[_lang]["order_id"] + "：PKG-" +
					trxDeposit.OrderNO +
					global.Translations[_lang]["catfee_smart_transaction_head_4"]

			msg.ParseMode = "HTML"

			inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("⏳"+global.Translations[_lang]["catfee_smart_transaction_pay_button"]+realTransferAmount+bundlePackage.Token, "noop"),
					tgbotapi.NewInlineKeyboardButtonData("❌"+global.Translations[_lang]["cancel_order"], "cancel_catfee_order"),
				),
			)

			msg.ReplyMarkup = inlineKeyboard
			sent, _ := bot.Send(msg)
			expiration := 1 * time.Minute // 短时间缓存空值

			//设置用户状态
			cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10)+"_order_no", "TRX_"+trxDeposit.OrderNO, expiration)

			expirationOrder := 1 * time.Minute // 短时间缓存空值
			//设置用户状态
			cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10)+"_order", strconv.Itoa(sent.MessageID), expirationOrder)

			return
		}

		if bundlePackage.Token == "USDT" {
			usdtPlaceholderRepo := repositories.NewUserUsdtPlaceholdersRepository(db)
			placeholder, esg := usdtPlaceholderRepo.Query(context.Background())
			//err := trxPlaceholderRepo.Update(context.Background(), placeholder.Id, 1)
			if esg != nil {
				fmt.Printf("Failed to update user: " + esg.Error())
				msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[_lang]["placeholder_array_size_warning"])

				inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("🕣"+global.Translations[_lang]["cancel_order"], "cancel_order"),
						tgbotapi.NewInlineKeyboardButtonData("🔙"+global.Translations[_lang]["back_home"], "back_home"),
					))
				msg.ReplyMarkup = inlineKeyboard
				msg.ParseMode = "HTML"
				//msg.DisableWebPagePreview = true
				bot.Send(msg)
				return

			}
			if placeholder.Id == 0 {
				msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[_lang]["placeholder_array_size_warning"])
				inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("🕣"+global.Translations[_lang]["cancel_order"], "cancel_order"),
						tgbotapi.NewInlineKeyboardButtonData("🔙"+global.Translations[_lang]["back_home"], "back_home"),
					))
				msg.ReplyMarkup = inlineKeyboard
				msg.ParseMode = "HTML"
				//msg.DisableWebPagePreview = true
				bot.Send(msg)

				return
			}

			err := usdtPlaceholderRepo.Update(context.Background(), placeholder.Id, 1)
			if err != nil {
				log.Printf("Error updating usdt placeholder: %v", err)
			}
			realTransferAmount := AddStringsAsFloats(placeholder.Placeholder, bundlePackage.Amount)

			fmt.Printf("realTransferAmount: %s\n", realTransferAmount)

			//生成订单
			usdtDepositRepo := repositories.NewUserUSDTDepositsRepository(db)

			orderNO := Generate6DigitOrderNo()
			var usdtDeposit domain.UserUSDTDeposits
			usdtDeposit.OrderNO = orderNO
			usdtDeposit.UserID = callbackQuery.Message.Chat.ID
			usdtDeposit.Status = 0
			usdtDeposit.Placeholder = placeholder.Placeholder

			//来自于波场伴侣  //source  0代表充值、1代表智能托管、2代表检测、3代表预警
			usdtDeposit.Source = 1
			value, err := strconv.ParseInt(bundleID, 10, 64)
			if err != nil {
				fmt.Printf("套餐ID转换失败: %v\n", err)
				return
			}

			usdtDeposit.BundleId = value

			//dictRepo := repositories.NewSysDictionariesRepo(db)
			_agent := os.Getenv("Agent")
			//depositAddress, _ := dictRepo.GetDepositAddress(_agent)
			//_agent := os.Getenv("Agent")
			sysUserRepo := repositories.NewSysUsersRepository(db)
			_, depositAddress, _ := sysUserRepo.Find(context.Background(), _agent)
			usdtDeposit.Address = depositAddress
			usdtDeposit.Amount = bundlePackage.Amount
			usdtDeposit.CreatedAt = time.Now()

			errsg := usdtDepositRepo.Create(context.Background(), &usdtDeposit)
			if errsg != nil {
				log.Printf("Error creating usdtDeposit: %v", errsg)
			}

			//msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID,
			//	global.Translations[_lang]["order_id"]+"：TOPUP-"+usdtDeposit.OrderNO+"\n"+
			//		global.Translations[_lang]["payment_amount"]+"："+"<code>"+realTransferAmount+"</code>"+" USDT "+global.Translations[_lang]["copy_text_tips"]+"\n"+
			//		global.Translations[_lang]["receive_address"]+"<code>"+usdtDeposit.Address+"</code>"+global.Translations[_lang]["copy_text_tips"]+"\n"+
			//		global.Translations[_lang]["tx_time_limit_tips"]+"\n"+
			//		global.Translations[_lang]["deposit_time_label"]+Format4Chinesese(usdtDeposit.CreatedAt)+"\n"+
			//		global.Translations[_lang]["amount_suffix_tips"]+"\n")

			videoPath := "./static/Audi.png"

			// 创建视频消息（从本地文件）
			msg := tgbotapi.NewPhoto(callbackQuery.Message.Chat.ID, tgbotapi.FilePath(videoPath))

			msg.Caption = global.Translations[_lang]["catfee_smart_transaction_head_1"] +
				bundlePackage.Name + global.Translations[_lang]["catfee_smart_transaction_head_2"] +
				"<code>" + realTransferAmount + "</code>" +
				bundlePackage.Token +
				global.Translations[_lang]["catfee_smart_transaction_head_3"] +
				"<code>" + usdtDeposit.Address + "</code>" + "\n\n" +
				global.Translations[_lang]["order_id"] + "：PKG-" +
				usdtDeposit.OrderNO +
				global.Translations[_lang]["catfee_smart_transaction_head_4"]

			msg.ParseMode = "HTML"

			inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("⏳"+global.Translations[_lang]["catfee_smart_transaction_pay_button"]+realTransferAmount+bundlePackage.Token, "noop"),
					tgbotapi.NewInlineKeyboardButtonData("❌"+global.Translations[_lang]["cancel_order"], "cancel_catfee_order"),
				),
			)

			msg.ReplyMarkup = inlineKeyboard
			sent, _ := bot.Send(msg)

			expiration := 1 * time.Minute // 短时间缓存空值

			//设置用户状态
			cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10)+"_order_no", "USDT_"+usdtDeposit.OrderNO, expiration)

			expirationOrder := 1 * time.Minute // 短时间缓存空值
			//设置用户状态
			cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10)+"_order", strconv.Itoa(sent.MessageID), expirationOrder)
			return
		}
	}

	//修改用户智能笔数&扣款
	nums, _ := ExtractNumberBeforeBi(bundlePackage.Name)
	stTimes := user.StTimes + int64(nums)

	fmt.Printf("\n更新用户%d，金额：%s\n", callbackQuery.Message.Chat.ID, bundlePackage.Name)
	if bundlePackage.Token == "USDT" {
		n := 1
		value, _ := SubtractStringNumbers(user.Amount, bundlePackage.Amount, float64(n))
		userRepo.UpdateUSDTAmount(value, callbackQuery.Message.Chat.ID)
	}
	if bundlePackage.Token == "TRX" {
		n := 1
		value, _ := SubtractStringNumbers(user.TronAmount, bundlePackage.Amount, float64(n))
		userRepo.UpdateTrxAmount(value, callbackQuery.Message.Chat.ID)
	}
	err = userRepo.UpdateSTTimes(stTimes, callbackQuery.Message.Chat.ID)
	if err != nil {
		log.Printf("Error updating stTimes: %v", err)
	}
	fmt.Printf("\n更新用户%d，最新笔数：%d\n", callbackQuery.Message.Chat.ID, stTimes)
	//新增用户订阅表

	//加入訂閲記錄
	userPackageSubscriptionsRepo := repositories.NewUserSmartTransactionPackageSubscriptionsRepository(db)

	var record domain.UserSmartTransactionPackageSubscriptions
	record.ChatID = callbackQuery.Message.Chat.ID
	//record.Address = message.Text
	bundle, _ := strconv.ParseInt(bundleID, 10, 64)

	record.BundleID = bundle
	record.Status = 1
	record.Amount = bundlePackage.Amount

	record.Times = ExtractLeadingInt64(bundlePackage.Name)
	record.BundleName = bundlePackage.Name
	record.CreatedAt = time.Now()

	err = userPackageSubscriptionsRepo.Create(context.Background(), &record)

	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, strings.ReplaceAll(global.Translations[_lang]["catfee_smart_transaction_tips"], "{bundle_package_name}", bundlePackage.Name)+"\n")
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		//tgbotapi.NewInlineKeyboardRow(
		//	tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["prev"], "next_bundle_package_address_stats"),
		//	tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["next"], "prev_bundle_package_address_stats"),
		//),
		tgbotapi.NewInlineKeyboardRow(
			//tgbotapi.NewInlineKeyboardButtonData("解绑地址", "free_monitor_address"),
			tgbotapi.NewInlineKeyboardButtonData("🔢"+global.Translations[_lang]["smart_transaction_address_list"], "click_bundle_package_address_stats_ST"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard
	msg.ParseMode = "HTML"

	bot.Send(msg)

	expiration := 1 * time.Minute // 短时间缓存空值
	//设置用户状态
	cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), "apply_ST_bundle_package_"+bundleID, expiration)
	//扣款
}

func ExtractBundleService(_lang string, message *tgbotapi.Message, bot *tgbotapi.BotAPI, db *gorm.DB, status string) bool {
	if !IsValidAddress(message.Text) {
		msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[_lang]["invalid_address_tips"]+"</b>"+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)
		return true
	}

	userRepo := repositories.NewUserRepository(db)
	user, _ := userRepo.GetByUserID(message.Chat.ID)

	fee := status[7:len(status)]
	fmt.Println("status : ", status)
	fmt.Println("fee : ", fee)
	fmt.Println("amount :", user.Amount)

	if CompareStringsWithFloat(fee, user.Amount, 1) {
		//余额不足，需充值
		msg := tgbotapi.NewMessage(message.Chat.ID,
			//"💬"+"<b>"+"余额不足: "+"</b>"+"\n"+
			//	"💬"+"<b>"+"用户姓名: "+"</b>"+user.Username+"\n"+
			//	"👤"+"<b>"+"用户电报ID: "+"</b>"+user.Associates+"\n"+
			//	"💵"+"<b>"+"当前TRX余额:  "+"</b>"+user.TronAmount+" TRX"+"\n"+
			//	"💴"+"<b>"+"当前USDT余额:  "+"</b>"+user.Amount+" USDT")
			"🆔"+global.Translations[_lang]["user_id"]+": "+user.Associates+"\n"+
				"👤"+global.Translations[_lang]["username"]+": @"+user.Username+"\n"+
				"💰"+global.Translations[_lang]["balance"]+": "+"\n"+
				"- TRX：   "+user.TronAmount+"\n"+
				"-  USDT："+user.Amount)
		msg.ParseMode = "HTML"
		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("💵"+global.Translations[_lang]["deposit"], "deposit_amount"),
			),
		)

		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)
	} else {
		bundlesRepo := repositories.NewUserOperationBundlesRepository(db)

		bundleRecord, _ := bundlesRepo.Find(context.Background(), fee)
		//10笔（12U）
		bundleNum := bundleRecord.Name
		count, _ := ExtractNumberBeforeBi(bundleNum)

		fmt.Printf("笔数count : %d", count)
		//扣款
		//调用trxfee接口

		//trxfeeHandler := handler.NewTrxfeeHandler()

		//trxfeeHandler.RequestTimesOrder(context.Background(),"","",message.Text,)
		rest, _ := SubtractStringNumbers(user.Amount, fee, 1)
		user.Amount = rest
		userRepo.Update2(context.Background(), &user)
		fmt.Println("rest :", rest)

		msg := tgbotapi.NewMessage(message.Chat.ID,
			"<b>"+"✅笔数套餐订阅成功"+"</b>"+"\n"+
				//"💬"+"<b>"+"用户姓名: "+"</b>"+user.Username+"\n"+
				//"👤"+"<b>"+"用户电报ID: "+"</b>"+user.Associates+"\n"+
				//"💵"+"<b>"+"当前TRX余额:  "+"</b>"+user.TronAmount+" TRX"+"\n"+
				//"💴"+"<b>"+"当前USDT余额:  "+"</b>"+user.Amount+" USDT")
				"🆔"+global.Translations[_lang]["user_id"]+": "+user.Associates+"\n"+
				"👤"+global.Translations[_lang]["username"]+": @"+user.Username+"\n"+
				"💰"+global.Translations[_lang]["balance"]+": "+"\n"+
				"- TRX：   "+user.TronAmount+"\n"+
				"-  USDT："+user.Amount)
		msg.ParseMode = "HTML"
		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("💵"+global.Translations[_lang]["deposit"], "deposit_amount"),
			),
		)

		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)
	}
	return false
}
