package member

import (
	"context"
	"strconv"
	"strings"
	"time"
	"ushield_bot/internal/cache"
	"ushield_bot/internal/global"
	"ushield_bot/internal/infrastructure/repositories"
	"ushield_bot/internal/infrastructure/tools"
	logger "ushield_bot/internal/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func MenuNavigate(lang string, db *gorm.DB, chatID int64, bot *tgbotapi.BotAPI) {

	premiumUserDB := repositories.NewTelegramPremiumConfigRepository(db)

	monthRecord_3m, _ := premiumUserDB.GetByEnName(context.Background(), "3_month_premium_fee")
	monthRecord_6m, _ := premiumUserDB.GetByEnName(context.Background(), "6_month_premium_fee")
	monthRecord_12m, _ := premiumUserDB.GetByEnName(context.Background(), "12_month_premium_fee")

	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(

			tgbotapi.NewInlineKeyboardButtonData(global.Translations[lang]["star_telegram_second_menu"], "purchase_star_menu"),
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[lang]["telegram_id_menu"], "purchase_anonymous_mobile"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[lang]["3_month_premium_fee"]+monthRecord_3m.Amount+"U", "click_buy_month_3"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[lang]["6_month_premium_fee"]+monthRecord_6m.Amount+"U", "click_buy_month_6"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[lang]["12_month_premium_fee"]+monthRecord_12m.Amount+"U", "click_buy_month_12"),
		),
	)

	videoPath := tools.StaticFile("telegram_premium.mp4")

	if err := sendVideoWithCache(
		bot,
		chatID,
		"telegram_premium.mp4",
		videoPath,
		global.Translations[lang]["purchase_telegram_products_tips"],
		inlineKeyboard,
	); err != nil {
		logger.Errorf("发送视频失败: %v", err)
	}

}

func MenuNavigateForMonth(cache cache.Cache, lang string, db *gorm.DB, chatID int64, username string, bot *tgbotapi.BotAPI, month string) {
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[lang]["activate_current_user"], "activate_current_user_"+month),
		),
	)

	tips := global.Translations[lang]["activate_current_user_tips"]

	tips = strings.ReplaceAll(tips, "{username}", username)
	tips = strings.ReplaceAll(tips, "{month}", month)
	premiumUserDB := repositories.NewTelegramPremiumConfigRepository(db)

	monthRecord, _ := premiumUserDB.GetByEnName(context.Background(), month+"_month_premium_fee")

	tips = strings.ReplaceAll(tips, "{price}", monthRecord.Amount)

	// 创建视频消息（从本地文件）
	videoMsg := tgbotapi.NewMessage(chatID, tips)
	videoMsg.ReplyMarkup = inlineKeyboard
	videoMsg.ParseMode = "HTML"

	// 发送视频
	if _, err := bot.Send(videoMsg); err != nil {
		logger.Errorf("发送视频失败: %v", err)
	}

	expiration := 1 * time.Minute // 短时间缓存空值

	//设置用户状态
	cache.Set(strconv.FormatInt(chatID, 10), "premium_user_rent_month"+month, expiration)

}
