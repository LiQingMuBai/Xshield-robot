package router

import (
	"ushield_bot/internal/global"
	"ushield_bot/internal/service"
	"ushield_bot/internal/service/additional"
	"ushield_bot/internal/service/catfee"
	"ushield_bot/internal/service/command"
	"ushield_bot/internal/service/launder"
	"ushield_bot/internal/service/member"
	"ushield_bot/internal/service/yhb"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleMenuMessage(message *tgbotapi.Message, ctx Context, lang string) bool {
	switch message.Text {
	case "⚽️世界杯竞猜🏆":
		msg := tgbotapi.NewMessage(message.Chat.ID, "点击跳转机器人 => 🤖@ushield_octopus_bot\n")
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔙"+global.Translations[lang]["back_home"], "back_home"),
			),
		)
		ctx.Bot.Send(msg)
		return true
	case "🥂" + global.Translations[lang]["coin_laundering_menu"]:
		launder.MenuLaunderNavigate(lang, ctx.DB, message.Chat.ID, ctx.Bot)
		return true
	case global.Translations[lang]["member_telegram_menu"]:
		member.MenuNavigate(lang, ctx.DB, message.Chat.ID, ctx.Bot)
		return true
	case global.Translations[lang]["command_energy_menu"]:
		command.MenuNavigate(lang, ctx.DB, message.Chat.ID, ctx.Bot)
		return true
	case "🛒" + global.Translations[lang]["ushield_additional_services_menu"]:
		additional.MenuNavigate(lang, ctx.DB, message.Chat.ID, ctx.Bot)
		return true
	case "🔃" + global.Translations[lang]["coin_swap_coin_menu"]:
		service.ShowCoinToCoinSwapMenu(lang, ctx.DB, message, ctx.Bot)
		return true
	case "🧧" + global.Translations[lang]["yhb_menu"]:
		yhb.MenuNavigateTronEnergy(lang, ctx.DB, message, ctx.Bot)
		return true
	case "⛽" + global.Translations[lang]["tron_energy_menu"]:
		service.MenuNavigateTronEnergy(lang, ctx.DB, message, ctx.Bot)
		return true
	case "✅" + global.Translations[lang]["usdt_trx_swap"]:
		service.MenuNavigateSwapExchange(lang, ctx.DB, message, ctx.Bot)
		return true
	case "🕸" + global.Translations[lang]["address_trace_menu"]:
		service.MenuNavigateAddressTrace(lang, ctx.Cache, ctx.Bot, message.Chat.ID, ctx.DB)
		return true
	case "🔍" + global.Translations[lang]["address_check"]:
		service.MenuNavigateAddressDetection(lang, ctx.Cache, ctx.Bot, message.Chat.ID, ctx.DB)
		return true
	case "🚨" + global.Translations[lang]["usdt_freeze_alert"]:
		service.MenuNavigateAddressFreeze(lang, ctx.Cache, ctx.Bot, message.Chat.ID, ctx.DB)
		return true
	case "🖊️" + global.Translations[lang]["transaction_plans"]:
		service.MenuNavigateBundlePackage(lang, ctx.DB, message.Chat.ID, ctx.Bot, "TRX")
		return true
	case "🤖" + global.Translations[lang]["catfee_smart_transaction_menu"]:
		catfee.MenuNavigateCatfeeSmartTransactionPlans(lang, ctx.DB, message.Chat.ID, ctx.Bot, "TRX")
		return true
	case "⚡" + global.Translations[lang]["energy_swap"]:
		service.MenuNavigateEnergyExchange(lang, ctx.DB, message, ctx.Bot)
		return true
	case "👤" + global.Translations[lang]["my_account"]:
		service.MenuNavigateHome(lang, ctx.Cache, ctx.DB, message, ctx.Bot)
		return true
	case "🌍" + global.Translations[lang]["language"]:
		service.MenuNavigateHome2(ctx.DB, message, ctx.Bot)
		return true
	default:
		return false
	}
}
