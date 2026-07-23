package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strconv"
	"time"
	"ushield_bot/internal/cache"
	"ushield_bot/internal/global"
	"ushield_bot/internal/service"
)

func sendDispatchSuccess(bot *tgbotapi.BotAPI, chatID int64, result *service.DispatchResult) {
	msg := tgbotapi.NewMessage(chatID, "📢【✅"+global.Translations[result.UserLang]["UShield_sent_transaction_energy"]+"】\n\n"+
		global.Translations[result.UserLang]["to_address"]+result.Address+"\n\n"+
		global.Translations[result.UserLang]["remaining_transactions"]+strconv.FormatInt(result.RemainingTimes, 10)+"\n\n")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚡️"+global.Translations[result.UserLang]["dispatch_again"], "click_bundle_package_address_stats"),
		),
	)
	msg.ParseMode = "HTML"
	bot.Send(msg)
}

func resetDispatchState(cacheStore cache.Cache, chatID int64) {
	expiration := 1 * time.Minute
	cacheStore.Set(strconv.FormatInt(chatID, 10), "null_dispatch_others_", expiration)
}
