package router

import (
	"strconv"
	"strings"
	"time"
	"ushield_bot/internal/global"
	"ushield_bot/internal/infrastructure/repositories"
	"ushield_bot/internal/service"
	"ushield_bot/internal/service/catfee"
	"ushield_bot/internal/service/launder"
	"ushield_bot/internal/service/member"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func handleCommerceCallback(lang string, callbackQuery *tgbotapi.CallbackQuery, ctx Context) bool {
	switch {
	case strings.HasPrefix(callbackQuery.Data, "ST_bundle_"):
		catfee.ST_BUNDLE_CHECK(lang, ctx.Cache, ctx.Bot, callbackQuery, ctx.DB)
		return true
	case callbackQuery.Data == "click_visa":
		sendAdditionalServiceDetail(ctx.Bot, callbackQuery.Message.Chat.ID, lang, ctx.DB, "ushield_additional_services_visa_desc")
		return true
	case callbackQuery.Data == "click_sim":
		sendAdditionalServiceDetail(ctx.Bot, callbackQuery.Message.Chat.ID, lang, ctx.DB, "ushield_additional_services_sim_desc")
		return true
	case callbackQuery.Data == "click_energy_financing":
		sendAdditionalServiceDetail(ctx.Bot, callbackQuery.Message.Chat.ID, lang, ctx.DB, "ushield_additional_services_energy_financing_desc")
		return true
	case callbackQuery.Data == "click_sns":
		sendAdditionalServiceDetail(ctx.Bot, callbackQuery.Message.Chat.ID, lang, ctx.DB, "ushield_additional_services_sns_desc")
		return true
	case callbackQuery.Data == "click_ecs":
		sendAdditionalServiceDetail(ctx.Bot, callbackQuery.Message.Chat.ID, lang, ctx.DB, "ushield_additional_services_ecs_desc")
		return true
	case strings.HasPrefix(callbackQuery.Data, "click_buy_month_"):
		month := strings.TrimPrefix(callbackQuery.Data, "click_buy_month_")
		member.MenuNavigateForMonth(ctx.Cache, lang, ctx.DB, callbackQuery.Message.Chat.ID, callbackQuery.From.UserName, ctx.Bot, month)
		return true
	case strings.HasPrefix(callbackQuery.Data, "activate_current_user_"):
		month := strings.TrimPrefix(callbackQuery.Data, "activate_current_user_")
		member.Rent(lang, ctx.Cache, ctx.DB, ctx.Bot, callbackQuery.Message.Chat.UserName, callbackQuery.Message.Chat.ID, month)
		return true
	case strings.HasPrefix(callbackQuery.Data, "pay_premium_order_"):
		orderNO := strings.TrimPrefix(callbackQuery.Data, "pay_premium_order_")
		service.PayPremiumOrder(ctx.Bot, callbackQuery.Message.Chat.ID, lang, ctx.DB, ctx.CatfeeClient, orderNO)
		return true
	case strings.HasPrefix(callbackQuery.Data, "cancel_premium_order_"):
		orderNO := strings.TrimPrefix(callbackQuery.Data, "cancel_premium_order_")
		service.CancelPremiumOrder(ctx.Bot, ctx.Cache, callbackQuery.Message.Chat.ID, lang, ctx.DB, orderNO)
		return true
	case strings.HasPrefix(callbackQuery.Data, "purchase_star_menu"):
		member.MenuStarNavigate(lang, ctx.DB, callbackQuery.Message.Chat.ID, ctx.Bot)
		return true
	case strings.HasPrefix(callbackQuery.Data, "purchase_telegram_premium"):
		member.MenuNavigate(lang, ctx.DB, callbackQuery.Message.Chat.ID, ctx.Bot)
		return true
	case strings.HasPrefix(callbackQuery.Data, "click_purchase_stars_"):
		count := strings.TrimPrefix(callbackQuery.Data, "click_purchase_stars_")
		member.MenuNavigateForStar(ctx.Cache, lang, ctx.DB, callbackQuery.Message.Chat.ID, callbackQuery.From.UserName, ctx.Bot, count)
		return true
	case strings.HasPrefix(callbackQuery.Data, "purchase_stars_current_user_"):
		count := strings.TrimPrefix(callbackQuery.Data, "purchase_stars_current_user_")
		member.Purchase(lang, ctx.Cache, ctx.DB, ctx.Bot, callbackQuery.Message.Chat.UserName, callbackQuery.Message.Chat.ID, count)
		return true
	case strings.HasPrefix(callbackQuery.Data, "purchase_stars_"):
		orderNO := strings.TrimPrefix(callbackQuery.Data, "purchase_stars_")
		service.PayStarsOrder(ctx.Bot, callbackQuery.Message.Chat.ID, lang, ctx.DB, orderNO)
		return true
	case strings.HasPrefix(callbackQuery.Data, "cancel_stars_"):
		orderNO := strings.TrimPrefix(callbackQuery.Data, "cancel_stars_")
		service.CancelStarsOrder(ctx.Bot, ctx.Cache, callbackQuery.Message.Chat.ID, lang, ctx.DB, orderNO)
		return true
	case strings.HasPrefix(callbackQuery.Data, "purchase_anonymous_mobile"):
		member.MenuMobileNavigate(lang, ctx.DB, callbackQuery.Message.Chat.ID, ctx.Bot)
		return true
	case strings.HasPrefix(callbackQuery.Data, "click_launder_"):
		amount := strings.TrimPrefix(callbackQuery.Data, "click_launder_")
		launder.MenuNavigateForLaunder(ctx.Cache, lang, ctx.DB, callbackQuery.Message.Chat.ID, callbackQuery.From.UserName, ctx.Bot, amount)
		return true
	case strings.HasPrefix(callbackQuery.Data, "click_laundering_"):
		content := strings.TrimPrefix(callbackQuery.Data, "click_laundering_")
		ctx.Cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), "click_laundering_"+content, time.Minute)
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "💬"+global.Translations[lang]["input_receive_address"]+"\n")
		msg.ParseMode = "HTML"
		ctx.Bot.Send(msg)
		return true
	}
	return false
}

func sendAdditionalServiceDetail(bot *tgbotapi.BotAPI, chatID int64, lang string, db *gorm.DB, descKey string) {
	dictRepo := repositories.NewSysDictionariesRepo(db)
	contact, _ := dictRepo.GetDictionaryDetail("ushield_additional_services_contact")
	wallet, _ := dictRepo.GetDictionaryDetail("ushield_additional_services_wallet")
	text := global.Translations[lang][descKey] + "\n" +
		strings.ReplaceAll(global.Translations[lang]["ushield_additional_services_contact"], "{ushield_additional_services_contact}", contact) +
		strings.ReplaceAll(global.Translations[lang]["ushield_additional_services_wallet"], "{ushield_additional_services_wallet}", wallet)
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	bot.Send(msg)
}
