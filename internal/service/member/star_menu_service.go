package member

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

func MenuStarNavigate(_lang string, db *gorm.DB, _chatID int64, bot *tgbotapi.BotAPI) {

	dictDetailRepo := repositories.NewSysDictionariesRepo(db)
	star_unit, _ := dictDetailRepo.GetDictionaryDetail("star_unit")
	fmt.Printf("star_unit: %s\n", star_unit)
	unitPrice, _ := strconv.ParseFloat(star_unit, 64)

	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(

			tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["member_telegram_menu"], "purchase_telegram_premium"),
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["telegram_id_menu"], "purchase_anonymous_mobile"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("100"+" "+tools.StringMultiply2("100", unitPrice)+"U", "click_purchase_stars_"+"100"),
			tgbotapi.NewInlineKeyboardButtonData("200"+" "+tools.StringMultiply2("200", unitPrice)+"U", "click_purchase_stars_"+"200"),
			tgbotapi.NewInlineKeyboardButtonData("400"+" "+tools.StringMultiply2("400", unitPrice)+"U", "click_purchase_stars_"+"400"),
			tgbotapi.NewInlineKeyboardButtonData("800"+" "+tools.StringMultiply2("800", unitPrice)+"U", "click_purchase_stars_"+"800"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("1000"+" "+tools.StringMultiply2("1000", unitPrice)+"U", "click_purchase_stars_"+"1000"),
			tgbotapi.NewInlineKeyboardButtonData("2000"+" "+tools.StringMultiply2("2000", unitPrice)+"U", "click_purchase_stars_"+"2000"),
			tgbotapi.NewInlineKeyboardButtonData("4000"+" "+tools.StringMultiply2("4000", unitPrice)+"U", "click_purchase_stars_"+"4000"),
			tgbotapi.NewInlineKeyboardButtonData("8000"+" "+tools.StringMultiply2("8000", unitPrice)+"U", "click_purchase_stars_"+"8000"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("10000"+" "+tools.StringMultiply2("10000", unitPrice)+"U", "click_purchase_stars_"+"10000"),
			tgbotapi.NewInlineKeyboardButtonData("20000"+" "+tools.StringMultiply2("20000", unitPrice)+"U", "click_purchase_stars_"+"20000"),
			tgbotapi.NewInlineKeyboardButtonData("40000"+" "+tools.StringMultiply2("40000", unitPrice)+"U", "click_purchase_stars_"+"40000"),
			tgbotapi.NewInlineKeyboardButtonData("80000"+" "+tools.StringMultiply2("80000", unitPrice)+"U", "click_purchase_stars_"+"80000"),
		),
	)

	// ⚠️ 替换为你自己的本地 MP4 路径，例如 "./videos/demo.mp4"
	//videoPath := "../telegram_stars.mp4"
	videoPath := "./static/telegram_stars.mp4"

	// 创建视频消息（从本地文件）
	videoMsg := tgbotapi.NewVideo(_chatID, tgbotapi.FilePath(videoPath))
	videoMsg.Caption = global.Translations[_lang]["purchase_telegram_stars_tips"]
	videoMsg.ReplyMarkup = inlineKeyboard
	videoMsg.SupportsStreaming = true // 启用流式播放（推荐）

	// 发送视频
	if _, err := bot.Send(videoMsg); err != nil {
		log.Printf("发送视频失败: %v", err)
		//// 可选：给用户发错误提示
		//errorMsg := tgbotapi.NewMessage(callback.Message.Chat.ID, "❌ 视频发送失败，请稍后再试。")
		//bot.Send(errorMsg)
	}

}

func MenuNavigateForStar(cache cache.Cache, _lang string, db *gorm.DB, _chatID int64, username string, bot *tgbotapi.BotAPI, count string) {
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["purchase_stars_current_user"], "purchase_stars_current_user_"+count),
		),
	)

	tips := global.Translations[_lang]["purchase_stars_tips"]

	tips = strings.ReplaceAll(tips, "{username}", username)
	tips = strings.ReplaceAll(tips, "{count}", count)

	dictDetailRepo := repositories.NewSysDictionariesRepo(db)
	star_unit, _ := dictDetailRepo.GetDictionaryDetail("star_unit")
	fmt.Printf("star_unit: %s\n", star_unit)

	unitPrice, _ := strconv.ParseFloat(star_unit, 64)

	price, _ := tools.StringMultiply(count, unitPrice)

	tips = strings.ReplaceAll(tips, "{price}", price)

	// 创建视频消息（从本地文件）
	videoMsg := tgbotapi.NewMessage(_chatID, tips)
	videoMsg.ReplyMarkup = inlineKeyboard
	videoMsg.ParseMode = "HTML"

	// 发送视频
	if _, err := bot.Send(videoMsg); err != nil {
		log.Printf("发送视频失败: %v", err)
	}

	expiration := 1 * time.Minute // 短时间缓存空值

	//设置用户状态
	cache.Set(strconv.FormatInt(_chatID, 10), "purchase_telegram_stars"+count, expiration)

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
	_agent := os.Getenv("Agent")
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
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["balance_pay_order"], "purchase_stars_"+orderNO),
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["cancel_order"], "cancel_stars_"+orderNO),
		),
	)
	msg.ReplyMarkup = inlineKeyboard

	msg.ParseMode = "HTML"
	bot.Send(msg)

}
