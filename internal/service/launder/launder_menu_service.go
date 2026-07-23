package launder

import (
	"context"
	"fmt"
	"log"
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

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func MenuLaunderNavigate(_lang string, db *gorm.DB, _chatID int64, bot *tgbotapi.BotAPI) {

	coinLaunderingConfigRepo := repositories.NewCoinLaunderingConfigRepository(db)

	configs, _ := coinLaunderingConfigRepo.QueryAll(context.Background())
	//star_unit, _ := dictDetailRepo.GetDictionaryDetail("star_unit")
	//fmt.Printf("star_unit: %s\n", star_unit)
	//unitPrice, _ := strconv.ParseFloat(star_unit, 64)

	var allButtons []tgbotapi.InlineKeyboardButton
	var keyboard [][]tgbotapi.InlineKeyboardButton
	for _, config := range configs {
		allButtons = append(allButtons, tgbotapi.NewInlineKeyboardButtonData(config.Name, "click_launder_"+config.Amount))
	}

	var extraButtons []tgbotapi.InlineKeyboardButton

	btn := tgbotapi.NewInlineKeyboardButtonURL("FixedFloat rules", "https://ff.io/terms-of-service")
	//row := tgbotapi.NewInlineKeyboardRow(btn)

	extraButtons = append(extraButtons, tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["other_blockchain_usdt_tips"], "click_callcenter"), tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["customize_usdt_amount_tips"], "click_callcenter"), btn)
	//
	//inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
	//	//tgbotapi.NewInlineKeyboardRow(
	//	//
	//	//	tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["member_telegram_menu"], "purchase_telegram_premium"),
	//	//	tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["telegram_id_menu"], "purchase_anonymous_mobile"),
	//	//),
	//	tgbotapi.NewInlineKeyboardRow(
	//		tgbotapi.NewInlineKeyboardButtonData("100"+"U", "click_launder_"+"100"),
	//		tgbotapi.NewInlineKeyboardButtonData("200"+"U", "click_launder_"+"200"),
	//		tgbotapi.NewInlineKeyboardButtonData("400"+"U", "click_launder_"+"400"),
	//		tgbotapi.NewInlineKeyboardButtonData("800"+"U", "click_launder_"+"800"),
	//	),
	//	tgbotapi.NewInlineKeyboardRow(
	//		tgbotapi.NewInlineKeyboardButtonData("1000"+"U", "click_launder_"+"1000"),
	//		tgbotapi.NewInlineKeyboardButtonData("2000"+"U", "click_launder_"+"2000"),
	//		tgbotapi.NewInlineKeyboardButtonData("5000"+"U", "click_launder_"+"4000"),
	//		tgbotapi.NewInlineKeyboardButtonData("10000"+"U", "click_launder_"+"8000"),
	//	),
	//	tgbotapi.NewInlineKeyboardRow(
	//		tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["other_blockchain_usdt_tips"], "click_launder_"+"1000"),
	//		tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["customize_usdt_amount_tips"], "click_launder_"+"2000"),
	//	),
	//)

	// ⚠️ 替换为你自己的本地 MP4 路径，例如 "./videos/demo.mp4"
	//videoPath := "../telegram_stars.mp4"
	//videoPath := "./static/telegram_premium.mp4"
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

	videoPath := "./static/fixedfloat.jpg"

	// 创建视频消息（从本地文件）
	videoMsg := tgbotapi.NewPhoto(_chatID, tgbotapi.FilePath(videoPath))
	videoMsg.Caption = global.Translations[_lang]["fixedfloat_rules"] + "\n\n" + global.Translations[_lang]["coin_laundering_tips"]
	// 创建视频消息（从本地文件）
	//videoMsg := tgbotapi.NewMessage(_chatID, global.Translations[_lang]["coin_laundering_tips"])
	videoMsg.ParseMode = "HTML"
	videoMsg.ReplyMarkup = inlineKeyboard
	//videoMsg.SupportsStreaming = true // 启用流式播放（推荐）

	// 发送视频
	if _, err := bot.Send(videoMsg); err != nil {
		log.Printf("发送视频失败: %v", err)
		//// 可选：给用户发错误提示
		//errorMsg := tgbotapi.NewMessage(callback.Message.Chat.ID, "❌ 视频发送失败，请稍后再试。")
		//bot.Send(errorMsg)
	}

}

func MenuNavigateForLaunder(cache cache.Cache, _lang string, db *gorm.DB, _chatID int64, username string, bot *tgbotapi.BotAPI, amount string) {
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		//tgbotapi.NewInlineKeyboardRow(
		//
		//	tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["member_telegram_menu"], "purchase_telegram_premium"),
		//	tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["telegram_id_menu"], "purchase_anonymous_mobile"),
		//),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["bsc_usdt_name"], "click_laundering_USDTBSC_"+amount),
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["ethereum_usdt_name"], "click_laundering_USDT_"+amount),
			//tgbotapi.NewInlineKeyboardButtonData("Poly-U", "click_laundering_polygon_"+amount),
			//tgbotapi.NewInlineKeyboardButtonData("Arb-U", "click_laundering_arbitrum_"+amount),
		),
		tgbotapi.NewInlineKeyboardRow(
			//tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["bsc_usdt_name"], "click_laundering_bsc_"+amount),
			//tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["ethereum_usdt_name"], "click_laundering_ethereum_"+amount),
			tgbotapi.NewInlineKeyboardButtonData("Polygon-USDT", "click_laundering_USDTMATIC_"+amount),
			tgbotapi.NewInlineKeyboardButtonData("Arbitrum-USDT", "click_laundering_USDTARBITRUM_"+amount),
		),

		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("BTC", "click_laundering_BTC_"+amount),
			tgbotapi.NewInlineKeyboardButtonData("ETH", "click_laundering_ETH_"+amount),
			tgbotapi.NewInlineKeyboardButtonData("BNB", "click_laundering_BSC_"+amount),
			//tgbotapi.NewInlineKeyboardButtonData("SOL", "click_laundering_sol_"+amount),
		),

		//tgbotapi.NewInlineKeyboardRow(
		//	tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["other_blockchain_usdt_tips"], "click_laundering_"+"1000"),
		//	tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["customize_usdt_amount_tips"], "click_laundering_"+"2000"),
		//),
	)

	videoMsg := tgbotapi.NewMessage(_chatID, global.Translations[_lang]["coin_laundering_choose_coin_tips"])

	videoMsg.ReplyMarkup = inlineKeyboard
	//videoMsg.SupportsStreaming = true // 启用流式播放（推荐）

	// 发送视频
	if _, err := bot.Send(videoMsg); err != nil {
		log.Printf("发送视频失败: %v", err)
		//// 可选：给用户发错误提示
		//errorMsg := tgbotapi.NewMessage(callback.Message.Chat.ID, "❌ 视频发送失败，请稍后再试。")
		//bot.Send(errorMsg)
	}

	expiration := 1 * time.Minute // 短时间缓存空值

	//设置用户状态
	cache.Set(strconv.FormatInt(_chatID, 10), "purchase_telegram_stars"+amount, expiration)

}

func Purchase(_lang string, cache cache.Cache, db *gorm.DB, bot *tgbotapi.BotAPI, username string, chatID int64, count string) {
	username = strings.ReplaceAll(username, "@", "")

	tips := global.Translations[_lang]["purchase_telegram_stars"]

	tips = strings.ReplaceAll(tips, "{stars}", count)
	tips = strings.ReplaceAll(tips, "{username}", username)

	firstName, lastName, _ := crawler.GetTelegramUserInfo(username)
	tips = strings.ReplaceAll(tips, "{nickname}", firstName+lastName)

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

	//来自于波场伴侣  //source  0代表充值、1代表智能托管、2代表检测、3代表预警、4代表VIP会员、5代表星星支付
	usdtDeposit.Source = 5
	_count, _ := strconv.ParseInt(count, 10, 64)
	usdtDeposit.BundleId = _count
	//
	//dictRepo := repositories.NewSysDictionariesRepo(db)
	_agent := os.Getenv("BOT_AGENT")
	//depositAddress, _ := dictRepo.GetDepositAddress(_agent)
	//_agent := os.Getenv("Agent")
	sysUserRepo := repositories.NewSysUsersRepository(db)
	_, depositAddress, _ := sysUserRepo.Find(context.Background(), _agent)
	usdtDeposit.Address = depositAddress

	dictDetailRepo := repositories.NewSysDictionariesRepo(db)
	star_unit, _ := dictDetailRepo.GetDictionaryDetail("star_unit")
	fmt.Printf("star_unit: %s\n", star_unit)
	unitPrice, _ := strconv.ParseFloat(star_unit, 64)
	price, _ := tools.StringMultiply(count, unitPrice)
	usdtDeposit.Amount = price
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

	fmt.Printf("无小数点：%s\n", price)
	fmt.Printf("有小数点：%s\n", tools.AddStringsAsFloats(price, usdtDeposit.Placeholder))
	fmt.Printf("小数点：%s\n", usdtDeposit.Placeholder)
	tips = strings.ReplaceAll(tips, "{amount}", tools.AddStringsAsFloats(price, usdtDeposit.Placeholder))

	//_agent := os.Getenv("Agent")
	//sysUserRepo := repositories.NewSysUsersRepository(db)
	//receiveAddress, _, _ := sysUserRepo.Find(context.Background(), _agent)

	tips = strings.ReplaceAll(tips, "{address}", depositAddress)

	msg := tgbotapi.NewMessage(chatID, tips)
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(

		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["balance_pay_order"], "launder_"+orderNO),
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["cancel_order"], "cancel_stars_"+orderNO),
		),
	)
	msg.ReplyMarkup = inlineKeyboard

	msg.ParseMode = "HTML"
	bot.Send(msg)

}
