package member

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"
	"ushield_bot/crawler"
	"ushield_bot/internal/cache"
	"ushield_bot/internal/domain"
	"ushield_bot/internal/global"
	"ushield_bot/internal/infrastructure/repositories"
	"ushield_bot/internal/infrastructure/tools"
	logger "ushield_bot/internal/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func Rent(lang string, cache cache.Cache, db *gorm.DB, bot *tgbotapi.BotAPI, username string, chatID int64, month string) {
	username = strings.ReplaceAll(username, "@", "")
	name := global.Translations[lang][month+"_month_premium"]

	tips := global.Translations[lang]["premium_activation_tips"]

	tips = strings.ReplaceAll(tips, "{premium_package}", name)
	tips = strings.ReplaceAll(tips, "{username}", username)

	firstName, lastName, _ := crawler.GetTelegramUserInfo(username)
	tips = strings.ReplaceAll(tips, "{nickname}", firstName+lastName)

	premiumUserDB := repositories.NewTelegramPremiumConfigRepository(db)

	monthRecord, _ := premiumUserDB.GetByEnName(context.Background(), month+"_month_premium_fee")

	//生成订单
	usdtDepositRepo := repositories.NewUserUSDTDepositsRepository(db)
	usdtPlaceholderRepo := repositories.NewUserUsdtPlaceholdersRepository(db)
	placeholder, _ := usdtPlaceholderRepo.GetRandomAvailable(context.Background())
	orderNO := tools.Generate6DigitOrderNo()
	var usdtDeposit domain.UserUSDTDeposits
	usdtDeposit.OrderNO = orderNO
	usdtDeposit.UserID = chatID
	usdtDeposit.Status = 0
	usdtDeposit.Placeholder = placeholder.Placeholder

	//来自于波场伴侣  //source  0代表充值、1代表智能托管、2代表检测、3代表预警、4代表VIP会员
	usdtDeposit.Source = 4
	usdtDeposit.BundleId = monthRecord.Id

	//dictRepo := repositories.NewSysDictionariesRepo(db)
	agent := os.Getenv("BOT_AGENT")
	//depositAddress, _ := dictRepo.GetDepositAddress(agent)
	//agent := os.Getenv("Agent")
	sysUserRepo := repositories.NewSysUsersRepository(db)
	_, depositAddress, _ := sysUserRepo.GetAddressesByUsername(context.Background(), agent)
	usdtDeposit.Address = depositAddress
	usdtDeposit.Amount = monthRecord.Amount
	usdtDeposit.CreatedAt = time.Now()

	createErr := usdtDepositRepo.Create(context.Background(), &usdtDeposit)
	if createErr != nil {
		logger.Errorf("Error creating usdtDeposit: %v", createErr)
	}

	err := usdtPlaceholderRepo.UpdateStatusByID(context.Background(), placeholder.Id, 1)
	if err != nil {
		logger.Errorf("Error updating usdt placeholder: %v", err)
	}

	//新增会员订单
	tgOrderDB := repositories.NewTelegramPremiumOrderRepository(db)
	var tgOrder domain.TelegramPremiumOrder
	tgOrder.OrderNO = orderNO
	tgOrder.Amount = monthRecord.Amount
	tgOrder.Month = month
	tgOrder.ChatID = chatID
	tgOrder.Status = 0
	tgOrder.CreatedAt = time.Now()
	tgOrder.TGUsername = username

	tgOrderDB.Create(context.Background(), &tgOrder)

	logger.Printf("无小数点：%s\n", monthRecord.Amount)
	logger.Printf("有小数点：%s\n", tools.AddStringsAsFloats(monthRecord.Amount, usdtDeposit.Placeholder))
	logger.Printf("小数点：%s\n", usdtDeposit.Placeholder)
	tips = strings.ReplaceAll(tips, "{amount}", tools.AddStringsAsFloats(monthRecord.Amount, usdtDeposit.Placeholder))

	//agent := os.Getenv("Agent")
	//sysUserRepo := repositories.NewSysUsersRepository(db)
	//receiveAddress, _, _ := sysUserRepo.GetAddressesByUsername(context.Background(), agent)

	tips = strings.ReplaceAll(tips, "{address}", depositAddress)

	videoPath := "./static/Audi.png"

	// 创建视频消息（从本地文件）
	msg := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(videoPath))
	msg.Caption = tips

	//msg := tgbotapi.NewMessage(chatID, tips)
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(

		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[lang]["balance_pay_order"], "pay_premium_order_"+usdtDeposit.OrderNO),
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[lang]["cancel_order"], "cancel_premium_order_"+usdtDeposit.OrderNO),
		),
	)
	msg.ReplyMarkup = inlineKeyboard

	msg.ParseMode = "HTML"
	sent, _ := bot.Send(msg)

	expiration := 1 * time.Minute // 短时间缓存空值
	//设置用户状态
	cache.Set(strconv.FormatInt(chatID, 10)+"_order", strconv.Itoa(sent.MessageID), expiration)

}
