package router

import (
	"strconv"
	"strings"
	"ushield_bot/internal/global"
	logger "ushield_bot/internal/logger"
	"ushield_bot/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleMessageUpdate(message *tgbotapi.Message, ctx Context) {
	lang := resolveUserLanguage(ctx.Cache, ctx.DB, message.Chat.ID)

	if handleMenuMessage(message, ctx, lang) {
		return
	}

	status, _ := ctx.Cache.Get(strconv.FormatInt(message.Chat.ID, 10))
	logger.Printf("用户状态status %s", status)
	handleStateMessage(message, ctx, lang, status)
}

func sendFreezeAlertPreview(bot *tgbotapi.BotAPI, chatID int64, lang string, preview *service.FreezeAlertPreview) {
	tips := strings.ReplaceAll(
		strings.ReplaceAll(global.Translations[lang]["enable_freeze_alerts_tips"], "{server_usdt_price}", preview.USDTPrice),
		"{server_trx_price}",
		preview.TRXPrice,
	)

	msg := tgbotapi.NewMessage(
		chatID,
		global.Translations[lang]["enable_freeze_alerts_tips_suffix"]+"\n"+
			global.Translations[lang]["address"]+"："+preview.Address+"\n\n"+
			tips,
	)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅"+global.Translations[lang]["confirm_freeze_alerts"], "confirm_freeze_risk_"+preview.Address),
			tgbotapi.NewInlineKeyboardButtonData("❌"+global.Translations[lang]["cancel_freeze_alerts"], "back_risk_home"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[lang]["back_homepage"], "back_risk_home"),
		),
	)
	bot.Send(msg)
}
