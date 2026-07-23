package service

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

func DepositPrevUSDTOrder(_lang string, cache cache.Cache, bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *gorm.DB) {
	transferAmount := callbackQuery.Data[13:len(callbackQuery.Data)]

	fmt.Printf("transferAmount: %s\n", transferAmount)

	usdtPlaceholderRepo := repositories.NewUserUsdtPlaceholdersRepository(db)
	placeholder, queryErr := usdtPlaceholderRepo.GetAvailable(context.Background())

	//err := trxPlaceholderRepo.Update(context.Background(), placeholder.Id, 1)
	if queryErr != nil {
		fmt.Print("Failed to update user: " + queryErr.Error())
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
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID,
			global.Translations[_lang]["placeholder_array_size_warning"])

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
	realTransferAmount := AddStringsAsFloats(placeholder.Placeholder, transferAmount)

	fmt.Printf("realTransferAmount: %s\n", realTransferAmount)

	//生成订单
	usdtDepositRepo := repositories.NewUserUSDTDepositsRepository(db)

	orderNO := Generate6DigitOrderNo()
	var usdtDeposit domain.UserUSDTDeposits
	usdtDeposit.OrderNO = orderNO
	usdtDeposit.UserID = callbackQuery.Message.Chat.ID
	usdtDeposit.Status = 0
	usdtDeposit.Placeholder = placeholder.Placeholder

	//dictRepo := repositories.NewSysDictionariesRepo(db)
	_agent := os.Getenv("BOT_AGENT")
	//depositAddress, _ := dictRepo.GetDepositAddress(_agent)
	//_agent := os.Getenv("Agent")
	sysUserRepo := repositories.NewSysUsersRepository(db)
	_, depositAddress, _ := sysUserRepo.GetAddressesByUsername(context.Background(), _agent)
	usdtDeposit.Address = depositAddress
	usdtDeposit.Amount = transferAmount
	usdtDeposit.CreatedAt = time.Now()

	createErr := usdtDepositRepo.Create(context.Background(), &usdtDeposit)
	if createErr != nil {
		log.Printf("Error creating usdtDeposit: %v", createErr)
	}

	//msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID,
	//	global.Translations[_lang]["order_id"]+"：TOPUP-"+usdtDeposit.OrderNO+"\n"+
	//		global.Translations[_lang]["payment_amount"]+"："+"<code>"+realTransferAmount+"</code>"+" USDT "+global.Translations[_lang]["copy_text_tips"]+"\n"+
	//		global.Translations[_lang]["receive_address"]+"<code>"+usdtDeposit.Address+"</code>"+global.Translations[_lang]["copy_text_tips"]+"\n"+
	//		global.Translations[_lang]["tx_time_limit_tips"]+"\n"+
	//		global.Translations[_lang]["deposit_time_label"]+FormatDateTimeValue(usdtDeposit.CreatedAt)+"\n"+
	//		global.Translations[_lang]["amount_suffix_tips"]+"\n")

	videoPath := "./static/Audi.png"

	// 创建视频消息（从本地文件）
	msg := tgbotapi.NewPhoto(callbackQuery.Message.Chat.ID, tgbotapi.FilePath(videoPath))

	msg.Caption = global.Translations[_lang]["order_id"] + "：TOPUP-" + usdtDeposit.OrderNO + "\n" +
		global.Translations[_lang]["payment_amount"] + "：" + "<code>" + realTransferAmount + "</code>" + " USDT " + global.Translations[_lang]["copy_text_tips"] + "\n" +
		global.Translations[_lang]["receive_address"] + "<code>" + usdtDeposit.Address + "</code>" + global.Translations[_lang]["copy_text_tips"] + "\n" +
		global.Translations[_lang]["tx_time_limit_tips"] + "\n" +
		global.Translations[_lang]["deposit_time_label"] + FormatDateTimeValue(usdtDeposit.CreatedAt) + "\n" +
		global.Translations[_lang]["amount_suffix_tips"] + "\n"
	//msg.ReplyMarkup = inlineKeyboard

	//originStr := global.Translations[_lang]["deposit_tips"]
	//
	//targetStr := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(originStr, "{order_no}", usdtDeposit.OrderNO), "{amount}", realTransferAmount), "{receiveAddress}", usdtDeposit.Address), "{createdAt}", FormatDateTimeValue(usdtDeposit.CreatedAt))

	//msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, targetStr)
	//"⚠️注意："+"\n"+
	//"▫️注意小数点 "+realTransferAmount+" usdt 转错金额不能到账"+"\n"+
	//"▫️请在10分钟完成付款，转错金额不能到账。"+"\n"+
	//"转账10分钟后没到账及时联系"+"\n")

	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⏳"+global.Translations[_lang]["catfee_smart_transaction_pay_button"]+realTransferAmount+" USDT ", "noop"),
		),

		tgbotapi.NewInlineKeyboardRow(

			tgbotapi.NewInlineKeyboardButtonData("🔙"+global.Translations[_lang]["back_home"], "back_home"),
			tgbotapi.NewInlineKeyboardButtonData("❌"+global.Translations[_lang]["cancel_order"], "cancel_order"),
		))
	msg.ReplyMarkup = inlineKeyboard
	msg.ParseMode = "HTML"
	//msg.DisableWebPagePreview = true
	sent, _ := bot.Send(msg)

	expiration := 1 * time.Minute // 短时间缓存空值

	//设置用户状态
	cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10)+"_order_no", "USDT_"+usdtDeposit.OrderNO, expiration)

	expirationOrder := 1 * time.Minute // 短时间缓存空值
	//设置用户状态
	cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10)+"_order", strconv.Itoa(sent.MessageID), expirationOrder)

}

func DepositCancelOrder(_lang string, cache cache.Cache, bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *gorm.DB) {
	//设置用户状态
	orderNO, _ := cache.Get(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10) + "_order_no")
	msg_order := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID,
		global.Translations[_lang]["order_id"]+"：TOPUP-"+orderNO+" , "+global.Translations[_lang]["cancel_order_tips"])
	msg_order.ParseMode = "HTML"
	//msg.DisableWebPagePreview = true
	bot.Send(msg_order)

	prevMessageIDStr, _ := cache.Get(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10) + "_order")

	prevMessageID, err := strconv.Atoi(prevMessageIDStr)
	if err != nil {
		fmt.Println("转换失败:", err)
		//return
	}

	bot.Request(tgbotapi.DeleteMessageConfig{ChatID: callbackQuery.Message.Chat.ID, MessageID: prevMessageID})

	//prevMessageID

	if strings.Contains(orderNO, "TRX_") {

		_orderNO := strings.ReplaceAll(orderNO, "TRX_", "")
		userTRXDepositsRepo := repositories.NewUserTRXDepositsRepository(db)
		record, _ := userTRXDepositsRepo.GetByOrderNo(context.Background(), _orderNO)

		//update
		userTRXDepositsRepo.Update(context.Background(), record.Id, 2)

		fmt.Printf("record: %v\n", record)

		userTRXPlaceholdersRepo := repositories.NewUserTRXPlaceholdersRepository(db)
		userTRXPlaceholdersRepo.UpdateByPlaceholder(context.Background(), record.Placeholder, 0)
		fmt.Printf("placeholder重置 %s\n", record.Placeholder)
	}

	if strings.Contains(orderNO, "USDT_") {
		_orderNO := strings.ReplaceAll(orderNO, "USDT_", "")
		userUSDTDepositsRepo := repositories.NewUserUSDTDepositsRepository(db)
		record, _ := userUSDTDepositsRepo.GetByOrderNo(context.Background(), _orderNO)
		//update
		userUSDTDepositsRepo.UpdateStatusByID(context.Background(), record.Id, 2)

		fmt.Printf("record: %v\n", record)
		userUSDTPlaceholdersRepo := repositories.NewUserUsdtPlaceholdersRepository(db)
		userUSDTPlaceholdersRepo.UpdateByPlaceholder(context.Background(), record.Placeholder, 0)
		fmt.Printf("placeholder重置 %s\n", record.Placeholder)
	}

	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		//tgbotapi.NewInlineKeyboardRow(
		//	tgbotapi.NewInlineKeyboardButtonData("🆔我的账户", "click_my_account"),
		//
		//),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💳"+global.Translations[_lang]["deposit"], "deposit_amount"),
			//tgbotapi.NewInlineKeyboardButtonData("🔗第二通知人", "click_backup_account"),
			tgbotapi.NewInlineKeyboardButtonData("📜"+global.Translations[_lang]["billing"], "click_my_recepit"),
			tgbotapi.NewInlineKeyboardButtonData("🛎️"+global.Translations[_lang]["support"], "click_callcenter"),
			//tgbotapi.NewInlineKeyboardButtonData("🛠️我的服务", "click_my_service"),
		),
		tgbotapi.NewInlineKeyboardRow(
			//tgbotapi.NewInlineKeyboardButtonData("🔗绑定备用帐号", "click_backup_account"),
			tgbotapi.NewInlineKeyboardButtonData("👥"+global.Translations[_lang]["business"], "click_business_cooperation"),
			tgbotapi.NewInlineKeyboardButtonData("💬"+global.Translations[_lang]["channel"], "click_offical_channel"),

			tgbotapi.NewInlineKeyboardButtonData("❓"+global.Translations[_lang]["tutorials"], "click_QA"),
		),
		//tgbotapi.NewInlineKeyboardRow(),
	)

	userRepo := repositories.NewUserRepository(db)
	user, _ := userRepo.GetByChatID(callbackQuery.Message.Chat.ID)

	if IsEmpty(user.Amount) {
		user.Amount = "0"
	}

	if IsEmpty(user.TronAmount) {
		user.TronAmount = "0"
	}

	str := ""
	if len(user.BackupChatID) > 0 {
		//id, _ := strconv.ParseInt(user.BackupChatID, 10, 64)
		//backup_user, _ := userRepo.GetByChatID(id)
		str = "🔗 " + global.Translations[_lang]["secondary_contact"] + "：  " + "@" + user.BackupChatID
	} else {
		str = global.Translations[_lang]["secondary_contact_none"]
	}

	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "🆔 "+global.Translations[_lang]["user_id"]+"：<code>"+user.Associates+"</code>\n\n👤 "+global.Translations[_lang]["username"]+"：@"+user.Username+"\n\n"+
		str+"\n\n💰"+
		global.Translations[_lang]["balance"]+"：\n\n"+
		"- TRX："+user.TronAmount+"\n"+
		"- USDT："+user.Amount)
	msg.ReplyMarkup = inlineKeyboard
	msg.ParseMode = "HTML"
	bot.Send(msg)
}

func DepositPrevOrder(_lang string, cache cache.Cache, bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *gorm.DB) {
	transferAmount := callbackQuery.Data[12:len(callbackQuery.Data)]

	fmt.Printf("transferAmount: %s\n", transferAmount)

	trxPlaceholderRepo := repositories.NewUserTRXPlaceholdersRepository(db)
	placeholder, queryErr := trxPlaceholderRepo.GetAvailable(context.Background())

	//err := trxPlaceholderRepo.Update(context.Background(), placeholder.Id, 1)
	if queryErr != nil {
		fmt.Print("Failed to update user: " + queryErr.Error())
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[_lang]["tron_network_tips"])

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
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[_lang]["tron_network_tips"])

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
		log.Printf("Error updating trx placeholder: %v", err)
	}
	realTransferAmount := AddStringsAsFloats(placeholder.Placeholder, transferAmount)

	fmt.Printf("realTransferAmount: %s\n", realTransferAmount)

	//生成订单
	trxDepositRepo := repositories.NewUserTRXDepositsRepository(db)

	orderNO := Generate6DigitOrderNo()
	var trxDeposit domain.UserTRXDeposits
	trxDeposit.OrderNO = orderNO
	trxDeposit.UserID = callbackQuery.Message.Chat.ID
	trxDeposit.Status = 0
	trxDeposit.Placeholder = placeholder.Placeholder

	//dictRepo := repositories.NewSysDictionariesRepo(db)
	_agent := os.Getenv("BOT_AGENT")
	//depositAddress, _ := dictRepo.GetDepositAddress(_agent)
	sysUserRepo := repositories.NewSysUsersRepository(db)
	_, depositAddress, _ := sysUserRepo.GetAddressesByUsername(context.Background(), _agent)
	trxDeposit.Address = depositAddress
	trxDeposit.Amount = transferAmount
	trxDeposit.CreatedAt = time.Now()

	createErr := trxDepositRepo.Create(context.Background(), &trxDeposit)
	if createErr != nil {
		log.Printf("Error creating trxDeposit: %v", createErr)
	}

	//msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID,
	//	global.Translations[_lang]["order_id"]+"：TOPUP-"+trxDeposit.OrderNO+"\n"+
	//		"转账金额："+"<code>"+realTransferAmount+"</code>"+" TRX （点击即可复制）"+"\n"+
	//		"转账地址："+"<code>"+trxDeposit.Address+"</code>"+"（点击即可复制）"+"\n"+
	//		global.Translations[_lang]["deposit_time_label"]+FormatDateTimeValue(trxDeposit.CreatedAt)+"\n"+
	//		"⚠️注意："+"\n"+
	//		"▫️注意小数点 "+realTransferAmount+" TRX 转错金额不能到账"+"\n"+
	//		"▫️请在10分钟完成付款，转错金额不能到账。"+"\n"+
	//		"转账10分钟后没到账及时联系"+"\n")

	videoPath := "./static/Audi.png"

	// 创建视频消息（从本地文件）
	msg := tgbotapi.NewPhoto(callbackQuery.Message.Chat.ID, tgbotapi.FilePath(videoPath))

	msg.Caption = global.Translations[_lang]["order_id"] + "：TOPUP-" + trxDeposit.OrderNO + "\n" +
		global.Translations[_lang]["payment_amount"] + "：" + "<code>" + realTransferAmount + "</code>" + " TRX " + global.Translations[_lang]["copy_text_tips"] + "\n" +
		global.Translations[_lang]["receive_address"] + "<code>" + trxDeposit.Address + "</code>" + global.Translations[_lang]["copy_text_tips"] + "\n" +
		global.Translations[_lang]["tx_time_limit_tips"] + "\n" +
		global.Translations[_lang]["deposit_time_label"] + FormatDateTimeValue(trxDeposit.CreatedAt) + "\n" +
		global.Translations[_lang]["amount_suffix_tips"] + "\n"

	//"⚠️注意："+"\n"+
	//"▫️注意小数点 "+realTransferAmount+" usdt 转错金额不能到账"+"\n"+
	//"▫️请在10分钟完成付款，转错金额不能到账。"+"\n"+
	//"转账10分钟后没到账及时联系"+"\n")
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⏳"+global.Translations[_lang]["catfee_smart_transaction_pay_button"]+realTransferAmount+" TRX ", "noop"),
		),

		tgbotapi.NewInlineKeyboardRow(

			tgbotapi.NewInlineKeyboardButtonData("🔙"+global.Translations[_lang]["back_home"], "back_home"),
			tgbotapi.NewInlineKeyboardButtonData("❌"+global.Translations[_lang]["cancel_order"], "cancel_order"),
		))
	msg.ReplyMarkup = inlineKeyboard
	msg.ParseMode = "HTML"
	//msg.DisableWebPagePreview = true
	sent, _ := bot.Send(msg)
	expiration := 1 * time.Minute // 短时间缓存空值

	//设置用户状态
	cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10)+"_order_no", "TRX_"+trxDeposit.OrderNO, expiration)
	expirationOrder := 1 * time.Minute // 短时间缓存空值
	//设置用户状态
	cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10)+"_order", strconv.Itoa(sent.MessageID), expirationOrder)

}
