package launder

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

func MenuLaunderNavigate(lang string, db *gorm.DB, chatID int64, bot *tgbotapi.BotAPI) {

	coinLaunderingConfigRepo := repositories.NewCoinLaunderingConfigRepository(db)

	configs, _ := coinLaunderingConfigRepo.ListActive(context.Background())

	var allButtons []tgbotapi.InlineKeyboardButton
	var keyboard [][]tgbotapi.InlineKeyboardButton
	for _, config := range configs {
		allButtons = append(allButtons, tgbotapi.NewInlineKeyboardButtonData(config.Name, "click_launder_"+config.Amount))
	}

	var extraButtons []tgbotapi.InlineKeyboardButton

	btn := tgbotapi.NewInlineKeyboardButtonURL("FixedFloat rules", "https://ff.io/terms-of-service")

	extraButtons = append(extraButtons, tgbotapi.NewInlineKeyboardButtonData(global.Translations[lang]["other_blockchain_usdt_tips"], "click_callcenter"), tgbotapi.NewInlineKeyboardButtonData(global.Translations[lang]["customize_usdt_amount_tips"], "click_callcenter"), btn)

	for i := 0; i < len(allButtons); i += 4 {
		end := i + 4
		if end > len(allButtons) {
			end = len(allButtons)
		}
		row := allButtons[i:end]
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
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	videoPath := tools.StaticFile("fixedfloat.jpg")

	// 创建视频消息（从本地文件）
	videoMsg := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(videoPath))
	videoMsg.Caption = global.Translations[lang]["fixedfloat_rules"] + "\n\n" + global.Translations[lang]["coin_laundering_tips"]
	videoMsg.ParseMode = "HTML"
	videoMsg.ReplyMarkup = inlineKeyboard

	// 发送视频
	if _, err := bot.Send(videoMsg); err != nil {
		logger.Errorf("发送视频失败: %v", err)
	}

}

func MenuNavigateForLaunder(cache cache.Cache, lang string, db *gorm.DB, chatID int64, username string, bot *tgbotapi.BotAPI, amount string) {
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[lang]["bsc_usdt_name"], "click_laundering_USDTBSC_"+amount),
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[lang]["ethereum_usdt_name"], "click_laundering_USDT_"+amount),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Polygon-USDT", "click_laundering_USDTMATIC_"+amount),
			tgbotapi.NewInlineKeyboardButtonData("Arbitrum-USDT", "click_laundering_USDTARBITRUM_"+amount),
		),

		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("BTC", "click_laundering_BTC_"+amount),
			tgbotapi.NewInlineKeyboardButtonData("ETH", "click_laundering_ETH_"+amount),
			tgbotapi.NewInlineKeyboardButtonData("BNB", "click_laundering_BSC_"+amount),
		),
	)

	videoMsg := tgbotapi.NewMessage(chatID, global.Translations[lang]["coin_laundering_choose_coin_tips"])

	videoMsg.ReplyMarkup = inlineKeyboard

	// 发送视频
	if _, err := bot.Send(videoMsg); err != nil {
		logger.Errorf("发送视频失败: %v", err)
	}

	expiration := 1 * time.Minute // 短时间缓存空值

	//设置用户状态
	cache.Set(strconv.FormatInt(chatID, 10), "purchase_telegram_stars"+amount, expiration)

}

func Purchase(lang string, cache cache.Cache, db *gorm.DB, bot *tgbotapi.BotAPI, username string, chatID int64, count string) {
	username = strings.ReplaceAll(username, "@", "")

	tips := global.Translations[lang]["purchase_telegram_stars"]

	tips = strings.ReplaceAll(tips, "{stars}", count)
	tips = strings.ReplaceAll(tips, "{username}", username)

	firstName, lastName, _ := crawler.GetTelegramUserInfo(username)
	tips = strings.ReplaceAll(tips, "{nickname}", firstName+lastName)

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

	//来自于波场伴侣  //source  0代表充值、1代表智能托管、2代表检测、3代表预警、4代表VIP会员、5代表星星支付
	usdtDeposit.Source = 5
	_count, _ := strconv.ParseInt(count, 10, 64)
	usdtDeposit.BundleId = _count
	agent := os.Getenv("BOT_AGENT")
	sysUserRepo := repositories.NewSysUsersRepository(db)
	_, depositAddress, _ := sysUserRepo.GetAddressesByUsername(context.Background(), agent)
	usdtDeposit.Address = depositAddress

	dictDetailRepo := repositories.NewSysDictionariesRepo(db)
	star_unit, _ := dictDetailRepo.GetDictionaryDetail("star_unit")
	logger.Printf("star_unit: %s\n", star_unit)
	unitPrice, _ := strconv.ParseFloat(star_unit, 64)
	price, _ := tools.StringMultiply(count, unitPrice)
	usdtDeposit.Amount = price
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
	tgOrderDB := repositories.NewTelegramStarsOrderRepository(db)
	var tgOrder domain.TelegramStarsOrder
	tgOrder.OrderNO = orderNO
	tgOrder.Amount = price
	tgOrder.Stars = count
	tgOrder.ChatID = chatID
	tgOrder.Status = 0
	tgOrder.CreatedAt = time.Now()
	tgOrder.TGUsername = username

	tgOrderDB.Create(context.Background(), &tgOrder)

	logger.Printf("无小数点：%s\n", price)
	logger.Printf("有小数点：%s\n", tools.AddStringsAsFloats(price, usdtDeposit.Placeholder))
	logger.Printf("小数点：%s\n", usdtDeposit.Placeholder)
	tips = strings.ReplaceAll(tips, "{amount}", tools.AddStringsAsFloats(price, usdtDeposit.Placeholder))

	tips = strings.ReplaceAll(tips, "{address}", depositAddress)

	msg := tgbotapi.NewMessage(chatID, tips)
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(

		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[lang]["balance_pay_order"], "launder_"+orderNO),
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[lang]["cancel_order"], "cancel_stars_"+orderNO),
		),
	)
	msg.ReplyMarkup = inlineKeyboard

	msg.ParseMode = "HTML"
	bot.Send(msg)

}
