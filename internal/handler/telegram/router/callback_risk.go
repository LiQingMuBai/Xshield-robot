package router

import (
	"context"
	"strconv"
	"strings"
	"ushield_bot/internal/global"
	"ushield_bot/internal/infrastructure/repositories"
	. "ushield_bot/internal/infrastructure/tools"
	logger "ushield_bot/internal/logger"
	"ushield_bot/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleRiskCallback(lang string, callbackQuery *tgbotapi.CallbackQuery, ctx Context) bool {
	switch {
	case callbackQuery.Data == "back_address_detection_home":
		service.MenuNavigateAddressDetection(lang, ctx.Cache, ctx.Bot, callbackQuery.Message.Chat.ID, ctx.DB)
		return true
	case strings.HasPrefix(callbackQuery.Data, "confirm_freeze_risk_"):
		address := strings.TrimPrefix(callbackQuery.Data, "confirm_freeze_risk_")
		freezeAlertService := service.NewFreezeAlertService(ctx.DB)
		result, err := freezeAlertService.Confirm(context.Background(), callbackQuery.Message.Chat.ID, address)
		if err == service.ErrFreezeAlertInsufficientBalance {
			sendFreezeAlertInsufficientBalance(ctx.Bot, callbackQuery.Message.Chat.ID, lang)
			return true
		}
		if err != nil {
			logger.Printf("freeze alert confirm err: %v", err)
			return true
		}
		sendFreezeAlertEnableSuccess(ctx.Bot, callbackQuery.Message.Chat.ID, lang, result)
		return true
	case strings.HasPrefix(callbackQuery.Data, "close_freeze_risk_"):
		target := strings.TrimPrefix(callbackQuery.Data, "close_freeze_risk_")
		freezeAlertService := service.NewFreezeAlertService(ctx.DB)
		preview, err := freezeAlertService.PreviewClose(callbackQuery.Message.Chat.ID, target)
		if err != nil {
			logger.Printf("freeze alert close preview err: %v", err)
			return true
		}
		sendFreezeAlertClosePreview(ctx.Bot, callbackQuery.Message.Chat.ID, lang, preview)
		return true
	case strings.HasPrefix(callbackQuery.Data, "close_risk_"):
		target := strings.TrimPrefix(callbackQuery.Data, "close_risk_")
		freezeAlertService := service.NewFreezeAlertService(ctx.DB)
		if err := freezeAlertService.Close(context.Background(), callbackQuery.Message.Chat.ID, target); err != nil {
			logger.Printf("freeze alert close err: %v", err)
			return true
		}
		sendFreezeAlertCloseSuccess(ctx.Bot, callbackQuery.Message.Chat.ID, lang)
		return true
	case callbackQuery.Data == "click_backup_account":
		sendBackupAccountPrompt(ctx.Bot, callbackQuery.Message.Chat.ID, lang)
		setShortState(ctx.Cache, callbackQuery.Message.Chat.ID, "click_backup_account")
		return true
	case callbackQuery.Data == "back_user_address_trace":
		service.MenuNavigateAddressTrace(lang, ctx.Cache, ctx.Bot, callbackQuery.Message.Chat.ID, ctx.DB)
		return true
	case callbackQuery.Data == "back_risk_home":
		service.MenuNavigateAddressFreeze(lang, ctx.Cache, ctx.Bot, callbackQuery.Message.Chat.ID, ctx.DB)
		return true
	case callbackQuery.Data == "address_list_trace":
		service.ShowAddressTraceList(lang, ctx.Cache, ctx.Bot, callbackQuery, ctx.DB)
		return true
	case callbackQuery.Data == "address_freeze_risk_records":
		msg := service.ExtractAddressRiskQuery(lang, ctx.DB, callbackQuery)
		ctx.Bot.Send(msg)
		return true
	case callbackQuery.Data == "user_detection_cost_records":
		msg := service.ExtractAddressDetection(lang, ctx.Cache, ctx.DB, callbackQuery)
		ctx.Bot.Send(msg)
		return true
	case callbackQuery.Data == "user_backup_notify":
		sendBackupNotifyPrompt(ctx.Bot, callbackQuery.Message.Chat.ID)
		setShortState(ctx.Cache, callbackQuery.Message.Chat.ID, callbackQuery.Data)
		return true
	case callbackQuery.Data == "start_freeze_risk_1":
		service.StartFreezeRiskInput(lang, ctx.Cache, ctx.DB, callbackQuery, ctx.Bot)
		return true
	case callbackQuery.Data == "stop_freeze_risk_1":
		userAddressEventRepo := repositories.NewUserAddressMonitorEventRepo(ctx.DB)
		userAddressEventRepo.RemoveAll(context.Background(), callbackQuery.Message.Chat.ID)
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "已经暂停所有监控")
		msg.ParseMode = "HTML"
		ctx.Bot.Send(msg)
		setShortState(ctx.Cache, callbackQuery.Message.Chat.ID, "reset")
		return true
	case callbackQuery.Data == "start_freeze_risk_0":
		service.MenuNavigateAddressFreeze(lang, ctx.Cache, ctx.Bot, callbackQuery.Message.Chat.ID, ctx.DB)
		return true
	case callbackQuery.Data == "stop_freeze_risk":
		freezeAlertService := service.NewFreezeAlertService(ctx.DB)
		items, err := freezeAlertService.ListActive(callbackQuery.Message.Chat.ID)
		if err != nil {
			logger.Printf("freeze alert list err: %v", err)
			return true
		}
		sendFreezeAlertStopList(ctx.Bot, callbackQuery.Message.Chat.ID, lang, items)
		return true
	case callbackQuery.Data == "start_freeze_risk":
		freezeAlertService := service.NewFreezeAlertService(ctx.DB)
		if err := freezeAlertService.Start(callbackQuery.Message.Chat.ID); err == service.ErrFreezeAlertInsufficientBalance {
			sendFreezeAlertInsufficientBalance(ctx.Bot, callbackQuery.Message.Chat.ID, lang)
			return true
		} else if err != nil {
			logger.Printf("freeze alert start err: %v", err)
			return true
		}
		sendFreezeAlertPromptAddress(ctx.Bot, callbackQuery.Message.Chat.ID, lang)
		setShortState(ctx.Cache, callbackQuery.Message.Chat.ID, "start_freeze_risk")
		return true
	case callbackQuery.Data == "address_manager_return":
		service.MenuNavigateAddressFreeze(lang, ctx.Cache, ctx.Bot, callbackQuery.Message.Chat.ID, ctx.DB)
		return true
	case callbackQuery.Data == "address_manager_add":
		sendAddressManagerPrompt(ctx.Bot, callbackQuery.Message.Chat.ID, "💬<b>请输入需添加的地址: </b>\n")
		setShortState(ctx.Cache, callbackQuery.Message.Chat.ID, callbackQuery.Data)
		return true
	case callbackQuery.Data == "address_manager_remove":
		sendAddressManagerPrompt(ctx.Bot, callbackQuery.Message.Chat.ID, "💬<b>请输入需删除的地址: </b>\n")
		setShortState(ctx.Cache, callbackQuery.Message.Chat.ID, callbackQuery.Data)
		return true
	case callbackQuery.Data == "address_manager":
		service.ShowAddressManager(lang, ctx.Cache, ctx.Bot, callbackQuery.Message.Chat.ID, ctx.DB)
		return true
	}
	return false
}

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

func sendBackupAccountPrompt(bot *tgbotapi.BotAPI, chatID int64, lang string) {
	msg := tgbotapi.NewMessage(chatID, global.Translations[lang]["secondary_contact_tips"])
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[lang]["back_homepage"], "back_home"),
		),
	)
	bot.Send(msg)
}

func sendBackupNotifyPrompt(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "💬<b>请输入需添加的第二紧急通知用户电报ID: </b>\n")
	msg.ParseMode = "HTML"
	bot.Send(msg)
}

func sendAddressManagerPrompt(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	bot.Send(msg)
}
