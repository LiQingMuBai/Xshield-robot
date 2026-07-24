package service

import (
	"context"
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

func DepositPrevUSDTOrder(lang string, cache cache.Cache, bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *gorm.DB) {
	transferAmount := callbackQuery.Data[13:len(callbackQuery.Data)]

	logger.Printf("transferAmount: %s\n", transferAmount)

	usdtPlaceholderRepo := repositories.NewUserUsdtPlaceholdersRepository(db)
	placeholder, queryErr := usdtPlaceholderRepo.GetRandomAvailable(context.Background())

	if queryErr != nil {
		logger.Error("Failed to update user: " + queryErr.Error())
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[lang]["placeholder_array_size_warning"])

		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🕣"+global.Translations[lang]["cancel_order"], "cancel_order"),
				tgbotapi.NewInlineKeyboardButtonData("🔙"+global.Translations[lang]["back_home"], "back_home"),
			))
		msg.ReplyMarkup = inlineKeyboard
		msg.ParseMode = "HTML"
		bot.Send(msg)
		return

	}
	if placeholder.Id == 0 {
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID,
			global.Translations[lang]["placeholder_array_size_warning"])

		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🕣"+global.Translations[lang]["cancel_order"], "cancel_order"),
				tgbotapi.NewInlineKeyboardButtonData("🔙"+global.Translations[lang]["back_home"], "back_home"),
			))
		msg.ReplyMarkup = inlineKeyboard
		msg.ParseMode = "HTML"
		bot.Send(msg)

		return
	}

	err := usdtPlaceholderRepo.UpdateStatusByID(context.Background(), placeholder.Id, 1)
	if err != nil {
		logger.Errorf("Error updating usdt placeholder: %v", err)
	}
	realTransferAmount := AddStringsAsFloats(placeholder.Placeholder, transferAmount)

	logger.Printf("realTransferAmount: %s\n", realTransferAmount)

	//生成订单
	usdtDepositRepo := repositories.NewUserUSDTDepositsRepository(db)

	orderNO := Generate6DigitOrderNo()
	var usdtDeposit domain.UserUSDTDeposits
	usdtDeposit.OrderNO = orderNO
	usdtDeposit.UserID = callbackQuery.Message.Chat.ID
	usdtDeposit.Status = 0
	usdtDeposit.Placeholder = placeholder.Placeholder

	agent := os.Getenv("BOT_AGENT")
	sysUserRepo := repositories.NewSysUsersRepository(db)
	_, depositAddress, _ := sysUserRepo.GetAddressesByUsername(context.Background(), agent)
	usdtDeposit.Address = depositAddress
	usdtDeposit.Amount = transferAmount
	usdtDeposit.CreatedAt = time.Now()

	createErr := usdtDepositRepo.Create(context.Background(), &usdtDeposit)
	if createErr != nil {
		logger.Errorf("Error creating usdtDeposit: %v", createErr)
	}

	videoPath := StaticFile("Audi.png")

	// 创建视频消息（从本地文件）
	msg := tgbotapi.NewPhoto(callbackQuery.Message.Chat.ID, tgbotapi.FilePath(videoPath))

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
	sent, _ := bot.Send(msg)

	expiration := 1 * time.Minute // 短时间缓存空值

	//设置用户状态
	cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10)+"_order_no", "USDT_"+usdtDeposit.OrderNO, expiration)

	expirationOrder := 1 * time.Minute // 短时间缓存空值
	//设置用户状态
	cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10)+"_order", strconv.Itoa(sent.MessageID), expirationOrder)

}

func DepositCancelOrder(lang string, cache cache.Cache, bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *gorm.DB) {
	//设置用户状态
	orderNO, _ := cache.Get(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10) + "_order_no")
	msg_order := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID,
		global.Translations[lang]["order_id"]+"：TOPUP-"+orderNO+" , "+global.Translations[lang]["cancel_order_tips"])
	msg_order.ParseMode = "HTML"
	bot.Send(msg_order)

	prevMessageIDStr, _ := cache.Get(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10) + "_order")

	prevMessageID, err := strconv.Atoi(prevMessageIDStr)
	if err != nil {
		logger.Error("转换失败:", err)
	}

	bot.Request(tgbotapi.DeleteMessageConfig{ChatID: callbackQuery.Message.Chat.ID, MessageID: prevMessageID})

	//prevMessageID

	if strings.Contains(orderNO, "TRX_") {

		_orderNO := strings.ReplaceAll(orderNO, "TRX_", "")
		userTRXDepositsRepo := repositories.NewUserTRXDepositsRepository(db)
		record, _ := userTRXDepositsRepo.GetByOrderNo(context.Background(), _orderNO)

		// update deposit and release placeholder
		userTRXDepositsRepo.UpdateStatusByID(context.Background(), record.Id, 2)

		logger.Printf("record: %v\n", record)

		userTRXPlaceholdersRepo := repositories.NewUserTRXPlaceholdersRepository(db)
		userTRXPlaceholdersRepo.UpdateStatusByPlaceholder(context.Background(), record.Placeholder, 0)
		logger.Printf("placeholder重置 %s\n", record.Placeholder)
	}

	if strings.Contains(orderNO, "USDT_") {
		_orderNO := strings.ReplaceAll(orderNO, "USDT_", "")
		userUSDTDepositsRepo := repositories.NewUserUSDTDepositsRepository(db)
		record, _ := userUSDTDepositsRepo.GetByOrderNo(context.Background(), _orderNO)
		// update deposit and release placeholder
		userUSDTDepositsRepo.UpdateStatusByID(context.Background(), record.Id, 2)

		logger.Printf("record: %v\n", record)
		userUSDTPlaceholdersRepo := repositories.NewUserUsdtPlaceholdersRepository(db)
		userUSDTPlaceholdersRepo.UpdateStatusByPlaceholder(context.Background(), record.Placeholder, 0)
		logger.Printf("placeholder重置 %s\n", record.Placeholder)
	}

	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💳"+global.Translations[lang]["deposit"], "deposit_amount"),
			tgbotapi.NewInlineKeyboardButtonData("📜"+global.Translations[lang]["billing"], "click_my_recepit"),
			tgbotapi.NewInlineKeyboardButtonData("🛎️"+global.Translations[lang]["support"], "click_callcenter"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👥"+global.Translations[lang]["business"], "click_business_cooperation"),
			tgbotapi.NewInlineKeyboardButtonData("💬"+global.Translations[lang]["channel"], "click_offical_channel"),

			tgbotapi.NewInlineKeyboardButtonData("❓"+global.Translations[lang]["tutorials"], "click_QA"),
		),
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
		str = "🔗 " + global.Translations[lang]["secondary_contact"] + "：  " + "@" + user.BackupChatID
	} else {
		str = global.Translations[lang]["secondary_contact_none"]
	}

	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "🆔 "+global.Translations[lang]["user_id"]+"：<code>"+user.Associates+"</code>\n\n👤 "+global.Translations[lang]["username"]+"：@"+user.Username+"\n\n"+
		str+"\n\n💰"+
		global.Translations[lang]["balance"]+"：\n\n"+
		"- TRX："+user.TronAmount+"\n"+
		"- USDT："+user.Amount)
	msg.ReplyMarkup = inlineKeyboard
	msg.ParseMode = "HTML"
	bot.Send(msg)
}

func DepositPrevOrder(lang string, cache cache.Cache, bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *gorm.DB) {
	transferAmount := callbackQuery.Data[12:len(callbackQuery.Data)]

	logger.Printf("transferAmount: %s\n", transferAmount)

	trxPlaceholderRepo := repositories.NewUserTRXPlaceholdersRepository(db)
	placeholder, queryErr := trxPlaceholderRepo.GetRandomAvailable(context.Background())

	if queryErr != nil {
		logger.Error("Failed to update user: " + queryErr.Error())
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[lang]["tron_network_tips"])

		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🕣"+global.Translations[lang]["cancel_order"], "cancel_order"),
				tgbotapi.NewInlineKeyboardButtonData("🔙"+global.Translations[lang]["back_home"], "back_home"),
			))
		msg.ReplyMarkup = inlineKeyboard
		msg.ParseMode = "HTML"
		bot.Send(msg)

		return

	}
	if placeholder.Id == 0 {
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[lang]["tron_network_tips"])

		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🕣"+global.Translations[lang]["cancel_order"], "cancel_order"),
				tgbotapi.NewInlineKeyboardButtonData("🔙"+global.Translations[lang]["back_home"], "back_home"),
			))
		msg.ReplyMarkup = inlineKeyboard
		msg.ParseMode = "HTML"
		bot.Send(msg)

		return
	}

	err := trxPlaceholderRepo.UpdateStatusByID(context.Background(), placeholder.Id, 1)
	if err != nil {
		logger.Errorf("Error updating trx placeholder: %v", err)
	}
	realTransferAmount := AddStringsAsFloats(placeholder.Placeholder, transferAmount)

	logger.Printf("realTransferAmount: %s\n", realTransferAmount)

	//生成订单
	trxDepositRepo := repositories.NewUserTRXDepositsRepository(db)

	orderNO := Generate6DigitOrderNo()
	var trxDeposit domain.UserTRXDeposits
	trxDeposit.OrderNO = orderNO
	trxDeposit.UserID = callbackQuery.Message.Chat.ID
	trxDeposit.Status = 0
	trxDeposit.Placeholder = placeholder.Placeholder

	agent := os.Getenv("BOT_AGENT")
	//depositAddress, _ := dictRepo.GetDepositAddress(agent)
	sysUserRepo := repositories.NewSysUsersRepository(db)
	_, depositAddress, _ := sysUserRepo.GetAddressesByUsername(context.Background(), agent)
	trxDeposit.Address = depositAddress
	trxDeposit.Amount = transferAmount
	trxDeposit.CreatedAt = time.Now()

	createErr := trxDepositRepo.Create(context.Background(), &trxDeposit)
	if createErr != nil {
		logger.Errorf("Error creating trxDeposit: %v", createErr)
	}

	videoPath := StaticFile("Audi.png")

	// 创建视频消息（从本地文件）
	msg := tgbotapi.NewPhoto(callbackQuery.Message.Chat.ID, tgbotapi.FilePath(videoPath))

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
	sent, _ := bot.Send(msg)
	expiration := 1 * time.Minute // 短时间缓存空值

	//设置用户状态
	cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10)+"_order_no", "TRX_"+trxDeposit.OrderNO, expiration)
	expirationOrder := 1 * time.Minute // 短时间缓存空值
	//设置用户状态
	cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10)+"_order", strconv.Itoa(sent.MessageID), expirationOrder)

}
