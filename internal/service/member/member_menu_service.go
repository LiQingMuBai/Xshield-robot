package member

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"
	"ushield_bot/internal/cache"
	"ushield_bot/internal/global"
	"ushield_bot/internal/infrastructure/repositories"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func MenuNavigate(_lang string, db *gorm.DB, _chatID int64, bot *tgbotapi.BotAPI) {

	premiumUserDB := repositories.NewTelegramPremiumConfigRepository(db)

	monthRecord_3m, _ := premiumUserDB.Query(context.Background(), "3_month_premium_fee")
	monthRecord_6m, _ := premiumUserDB.Query(context.Background(), "6_month_premium_fee")
	monthRecord_12m, _ := premiumUserDB.Query(context.Background(), "12_month_premium_fee")

	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(

			tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["star_telegram_second_menu"], "purchase_star_menu"),
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["telegram_id_menu"], "purchase_anonymous_mobile"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["3_month_premium_fee"]+monthRecord_3m.Amount+"U", "click_buy_month_3"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["6_month_premium_fee"]+monthRecord_6m.Amount+"U", "click_buy_month_6"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["12_month_premium_fee"]+monthRecord_12m.Amount+"U", "click_buy_month_12"),
		),

		//tgbotapi.NewInlineKeyboardRow(
		//	tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["balance_pay_order"], "click_bundle_package_address_stats"),
		//	tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["cancel_order"], "click_bundle_package_address_stats"),
		//),
	)

	// ⚠️ 替换为你自己的本地 MP4 路径，例如 "./videos/demo.mp4"
	//videoPath := "../telegram_stars.mp4"
	videoPath := "./static/telegram_premium.mp4"

	// 创建视频消息（从本地文件）
	videoMsg := tgbotapi.NewVideo(_chatID, tgbotapi.FilePath(videoPath))

	videoMsg.Caption = global.Translations[_lang]["purchase_telegram_products_tips"]
	videoMsg.ReplyMarkup = inlineKeyboard
	videoMsg.SupportsStreaming = true // 启用流式播放（推荐）

	// 发送视频
	if _, err := bot.Send(videoMsg); err != nil {
		log.Printf("发送视频失败: %v", err)
		//// 可选：给用户发错误提示
		//errorMsg := tgbotapi.NewMessage(callback.Message.Chat.ID, "❌ 视频发送失败，请稍后再试。")
		//bot.Send(errorMsg)
	}

	//dictDetailRepo := repositories.NewSysDictionariesRepo(db)
	//
	//energy_cost, _ := dictDetailRepo.GetDictionaryDetail("energy_cost")
	//
	//fmt.Printf("energy_cost: %s\n", energy_cost)
	//
	//energy_cost_2x, _ := tools.StringMultiply(energy_cost, 2)
	//energy_cost_3x, _ := tools.StringMultiply(energy_cost, 3)
	//energy_cost_4x, _ := tools.StringMultiply(energy_cost, 4)
	//energy_cost_5x, _ := tools.StringMultiply(energy_cost, 5)
	//
	//fmt.Printf("energy_cost_2x: %s\n", energy_cost_2x)
	//
	//originStr := global.Translations[_lang]["member_telegram_desc"]
	//
	//targetStr := strings.ReplaceAll(originStr, "{1_times_cost_trx}", energy_cost)
	//targetStr = strings.ReplaceAll(targetStr, "{2_times_cost_trx}", energy_cost_2x)
	//targetStr = strings.ReplaceAll(targetStr, "{3_times_cost_trx}", energy_cost_3x)
	//targetStr = strings.ReplaceAll(targetStr, "{4_times_cost_trx}", energy_cost_4x)
	//targetStr = strings.ReplaceAll(targetStr, "{5_times_cost_trx}", energy_cost_5x)
	//
	//msg := tgbotapi.NewMessage(_chatID, global.Translations[_lang]["member_telegram_desc"])
	//msg.ReplyMarkup = inlineKeyboard
	//msg.ParseMode = "HTML"
	//
	//bot.Send(msg)
}

func MenuNavigateForMonth(cache cache.Cache, _lang string, db *gorm.DB, _chatID int64, username string, bot *tgbotapi.BotAPI, _month string) {
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["activate_current_user"], "activate_current_user_"+_month),
		),
	)

	tips := global.Translations[_lang]["activate_current_user_tips"]

	tips = strings.ReplaceAll(tips, "{username}", username)
	tips = strings.ReplaceAll(tips, "{month}", _month)
	premiumUserDB := repositories.NewTelegramPremiumConfigRepository(db)

	monthRecord, _ := premiumUserDB.Query(context.Background(), _month+"_month_premium_fee")

	tips = strings.ReplaceAll(tips, "{price}", monthRecord.Amount)

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
	cache.Set(strconv.FormatInt(_chatID, 10), "premium_user_rent_month"+_month, expiration)

}
