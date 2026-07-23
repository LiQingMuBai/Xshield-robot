package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strconv"
	"strings"
	"ushield_bot/internal/global"
	. "ushield_bot/internal/infrastructure/tools"
	"ushield_bot/internal/service"
)

func sendFreezeAlertPromptAddress(bot *tgbotapi.BotAPI, chatID int64, lang string) {
	msg := tgbotapi.NewMessage(chatID, global.Translations[lang]["enter_address_for_alert"])
	msg.ParseMode = "HTML"
	bot.Send(msg)
}

func sendFreezeAlertInsufficientBalance(bot *tgbotapi.BotAPI, chatID int64, lang string) {
	msg := tgbotapi.NewMessage(chatID, "⚠️ "+global.Translations[lang]["freeze_alert_service_insufficient_balance"]+"\n\n")
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💵"+global.Translations[lang]["deposit"], "deposit_amount"),
		),
	)
	bot.Send(msg)
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

func sendFreezeAlertEnableSuccess(bot *tgbotapi.BotAPI, chatID int64, lang string, result *service.FreezeAlertConfirmResult) {
	msg := tgbotapi.NewMessage(
		chatID,
		"✅"+global.Translations[lang]["enable_freeze_alerts_success"]+"\n"+
			global.Translations[lang]["address"]+"："+result.Address+"\n"+
			global.Translations[lang]["network"]+result.Network,
	)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👁️‍🗨️ "+global.Translations[lang]["alert_monitoring_list"], "address_list_trace"),
			tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[lang]["back_homepage"], "back_risk_home"),
		),
	)
	bot.Send(msg)
}

func sendFreezeAlertStopList(bot *tgbotapi.BotAPI, chatID int64, lang string, items []service.FreezeAlertMonitorItem) {
	keyboard := make([][]tgbotapi.InlineKeyboardButton, 0, len(items)+1)
	for _, item := range items {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(TruncateString(item.Address), "close_freeze_risk_"+strconv.FormatInt(item.ID, 10)),
		))
	}
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[lang]["back_homepage"], "back_risk_home"),
	))

	msg := tgbotapi.NewMessage(chatID, global.Translations[lang]["monitoring_address_list"]+"\n\n")
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
	bot.Send(msg)
}

func sendFreezeAlertClosePreview(bot *tgbotapi.BotAPI, chatID int64, lang string, preview *service.FreezeAlertClosePreview) {
	msg := tgbotapi.NewMessage(
		chatID,
		global.Translations[lang]["confirm_stop_monitoring_address"]+"\n"+
			global.Translations[lang]["address"]+"："+preview.Address+"\n"+
			strings.ReplaceAll(global.Translations[lang]["confirm_stop_monitoring_address_tips"], "{days}", strconv.FormatInt(preview.RemainingDays, 10)),
	)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅"+global.Translations[lang]["confirm_stop_monitoring_address_yes"], "close_risk_"+strconv.FormatInt(preview.ID, 10)),
			tgbotapi.NewInlineKeyboardButtonData("❌"+global.Translations[lang]["cancel_freeze_alerts"], "back_risk_home"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[lang]["back_homepage"], "back_risk_home"),
		),
	)
	bot.Send(msg)
}

func sendFreezeAlertCloseSuccess(bot *tgbotapi.BotAPI, chatID int64, lang string) {
	msg := tgbotapi.NewMessage(chatID, global.Translations[lang]["confirm_stop_monitoring_address_success_tips"])
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👁️‍🗨️ "+global.Translations[lang]["alert_monitoring_list"], "address_list_trace"),
			tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[lang]["back_homepage"], "back_risk_home"),
		),
	)
	bot.Send(msg)
}
