package router

import (
	"context"
	"strconv"
	"strings"
	"ushield_bot/internal/global"
	"ushield_bot/internal/service"
	"ushield_bot/internal/service/catfee"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleCallbackQuery(callbackQuery *tgbotapi.CallbackQuery, ctx Context) {
	lang := resolveUserLanguage(ctx.Cache, ctx.DB, callbackQuery.Message.Chat.ID)
	if handleRiskCallback(lang, callbackQuery, ctx) {
		return
	}
	if handlePackageCallback(lang, callbackQuery, ctx) {
		return
	}
	if handleGenericCallback(lang, callbackQuery, ctx) {
		return
	}
	if handleCommerceCallback(lang, callbackQuery, ctx) {
		return
	}

	switch {
	case callbackQuery.Data == "click_energy_swap":
		service.MenuNavigateEnergyExchange(lang, ctx.DB, callbackQuery.Message, ctx.Bot)
	case callbackQuery.Data == "click_transaction_plan":
		service.MenuNavigateBundlePackage(lang, ctx.DB, callbackQuery.Message.Chat.ID, ctx.Bot, "TRX")
	case callbackQuery.Data == "click_smart_transaction_plan":
		catfee.MenuNavigateCatfeeSmartTransactionPlans(lang, ctx.DB, callbackQuery.Message.Chat.ID, ctx.Bot, "TRX")
	case callbackQuery.Data == "click_language":
		service.MenuNavigateHome2(ctx.DB, callbackQuery.Message, ctx.Bot)
	case callbackQuery.Data == "dispatch_Now_Others":
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[lang]["enter_address"]+"\n\n")
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[lang]["back_homepage"], "back_bundle_package"),
			),
		)
		ctx.Bot.Send(msg)
		setShortState(ctx.Cache, callbackQuery.Message.Chat.ID, "dispatch_others")
	case strings.HasPrefix(callbackQuery.Data, "dispatch_others_"):
		bundleAddress := strings.ReplaceAll(callbackQuery.Data, "dispatch_others_", "")
		bundleID := strings.Split(bundleAddress, "_")[0]
		address := strings.Split(bundleAddress, "_")[1]

		dispatchService := service.NewEnergyDispatchService(ctx.DB, ctx.TrxfeeURL, ctx.TrxfeeAPIKey, ctx.TrxfeeSecret, ctx.CatfeeClient)
		result, dispatchErr := dispatchService.DispatchFromSubscription(context.Background(), bundleID, address, callbackQuery.Message.Chat.ID)
		if dispatchErr == nil {
			msg2 := service.BuildBundlePackageSubscriptionStatsMessage(lang, ctx.DB, callbackQuery.Message.Chat.ID)
			ctx.Bot.Send(msg2)
			sendDispatchSuccess(ctx.Bot, callbackQuery.Message.Chat.ID, result)
			setShortState(ctx.Cache, callbackQuery.Message.Chat.ID, "null_dispatch_others_")
		}
	}
}

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
