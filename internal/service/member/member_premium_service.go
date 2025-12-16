package member

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"ushield_bot/crawler"
	"ushield_bot/internal/cache"
	"ushield_bot/internal/domain"
	"ushield_bot/internal/global"
	"ushield_bot/internal/infrastructure/repositories"
	"ushield_bot/internal/infrastructure/tools"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func Rent(_lang string, cache cache.Cache, db *gorm.DB, bot *tgbotapi.BotAPI, username string, chatID int64, _month string) {
	username = strings.ReplaceAll(username, "@", "")
	name := global.Translations[_lang][_month+"_month_premium"]

	tips := global.Translations[_lang]["premium_activation_tips"]

	tips = strings.ReplaceAll(tips, "{premium_package}", name)
	tips = strings.ReplaceAll(tips, "{username}", username)

	firstName, lastName, _ := crawler.GetTelegramUserInfo(username)
	tips = strings.ReplaceAll(tips, "{nickname}", firstName+lastName)

	premiumUserDB := repositories.NewTelegramPremiumConfigRepository(db)

	monthRecord, _ := premiumUserDB.Query(context.Background(), _month+"_month_premium_fee")

	//生成订单
	usdtDepositRepo := repositories.NewUserUSDTDepositsRepository(db)
	usdtPlaceholderRepo := repositories.NewUserUsdtPlaceholdersRepository(db)
	placeholder, _ := usdtPlaceholderRepo.Query(context.Background())
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
	_agent := os.Getenv("Agent")
	//depositAddress, _ := dictRepo.GetDepositAddress(_agent)
	//_agent := os.Getenv("Agent")
	sysUserRepo := repositories.NewSysUsersRepository(db)
	_, depositAddress, _ := sysUserRepo.Find(context.Background(), _agent)
	usdtDeposit.Address = depositAddress
	usdtDeposit.Amount = monthRecord.Amount
	usdtDeposit.CreatedAt = time.Now()

	errsg := usdtDepositRepo.Create(context.Background(), &usdtDeposit)
	if errsg != nil {
		log.Printf("Error creating usdtDeposit: %v", errsg)
	}

	err := usdtPlaceholderRepo.Update(context.Background(), placeholder.Id, 1)
	if err != nil {
		log.Printf("Error updating usdt placeholder: %v", err)
	}

	//新增会员订单
	tgOrderDB := repositories.NewTelegramPremiumOrderRepository(db)
	var tgOrder domain.TelegramPremiumOrder
	tgOrder.OrderNO = orderNO
	tgOrder.Amount = monthRecord.Amount
	tgOrder.Month = _month
	tgOrder.ChatID = chatID
	tgOrder.Status = 0
	tgOrder.CreatedAt = time.Now()
	tgOrder.TGUsername = username

	tgOrderDB.Create(context.Background(), &tgOrder)

	fmt.Printf("无小数点：%s\n", monthRecord.Amount)
	fmt.Printf("有小数点：%s\n", tools.AddStringsAsFloats(monthRecord.Amount, usdtDeposit.Placeholder))
	fmt.Printf("小数点：%s\n", usdtDeposit.Placeholder)
	tips = strings.ReplaceAll(tips, "{amount}", tools.AddStringsAsFloats(monthRecord.Amount, usdtDeposit.Placeholder))

	//_agent := os.Getenv("Agent")
	//sysUserRepo := repositories.NewSysUsersRepository(db)
	//receiveAddress, _, _ := sysUserRepo.Find(context.Background(), _agent)

	tips = strings.ReplaceAll(tips, "{address}", depositAddress)

	msg := tgbotapi.NewMessage(chatID, tips)
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(

		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["balance_pay_order"], "pay_premium_order_"+usdtDeposit.OrderNO),
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["cancel_order"], "cancel_premium_order_"+usdtDeposit.OrderNO),
		),
	)
	msg.ReplyMarkup = inlineKeyboard

	msg.ParseMode = "HTML"
	bot.Send(msg)

}
