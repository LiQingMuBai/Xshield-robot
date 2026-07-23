package router

import (
	"context"
	"strconv"
	"strings"
	"time"
	"ushield_bot/internal/cache"
	"ushield_bot/internal/global"
	"ushield_bot/internal/infrastructure/repositories"
	"ushield_bot/internal/service"
	"ushield_bot/internal/service/launder"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func handleGenericCallback(lang string, callbackQuery *tgbotapi.CallbackQuery, ctx Context) bool {
	switch {
	case callbackQuery.Data == "back_home":
		service.MenuNavigateHome(lang, ctx.Cache, ctx.DB, callbackQuery.Message, ctx.Bot)
		return true
	case callbackQuery.Data == "click_business_cooperation":
		service.ShowBusinessCooperation(lang, callbackQuery, ctx.Bot)
		return true
	case callbackQuery.Data == "click_offical_channel":
		service.ShowOfficialChannel(lang, callbackQuery, ctx.Bot)
		return true
	case callbackQuery.Data == "click_callcenter":
		service.ShowCallCenter(lang, callbackQuery, ctx.Bot)
		return true
	case callbackQuery.Data == "click_my_recepit":
		service.ShowReceiptSummary(lang, ctx.DB, callbackQuery, ctx.Bot)
		return true
	case callbackQuery.Data == "click_QA":
		service.ExtraQA(lang, ctx.Cache, ctx.Bot, callbackQuery)
		return true
	case callbackQuery.Data == "click_my_service":
		sendMyServiceOverview(ctx.Bot, ctx.Cache, callbackQuery.Message.Chat.ID, lang)
		return true
	case strings.HasPrefix(callbackQuery.Data, "set_lang_"):
		nextLang := strings.TrimPrefix(callbackQuery.Data, "set_lang_")
		ctx.Cache.Set("LANG_"+strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), nextLang, 24*time.Hour)
		userRepo := repositories.NewUserRepository(ctx.DB)
		userRepo.UpdateLang(nextLang, callbackQuery.Message.Chat.ID)
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[nextLang]["set_lang"]+"\n")
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔙"+global.Translations[nextLang]["back_home"], "back_home"),
			),
		)
		msg.ParseMode = "HTML"
		ctx.Bot.Send(msg)
		handleStartCommand(ctx.Cache, ctx.Bot, callbackQuery.Message)
		return true
	case strings.HasPrefix(callbackQuery.Data, "bundle_"):
		service.CheckBundlePackage(lang, ctx.Cache, ctx.Bot, callbackQuery, ctx.DB)
		return true
	case strings.HasPrefix(callbackQuery.Data, "deposit_usdt"):
		service.DepositPrevUSDTOrder(lang, ctx.Cache, ctx.Bot, callbackQuery, ctx.DB)
		return true
	case strings.HasPrefix(callbackQuery.Data, "deposit_trx"):
		service.DepositPrevOrder(lang, ctx.Cache, ctx.Bot, callbackQuery, ctx.DB)
		return true
	case callbackQuery.Data == "cancel_order":
		service.DepositCancelOrder(lang, ctx.Cache, ctx.Bot, callbackQuery, ctx.DB)
		return true
	case callbackQuery.Data == "cancel_catfee_order":
		service.DepositCancelOrder(lang, ctx.Cache, ctx.Bot, callbackQuery, ctx.DB)
		return true
	case callbackQuery.Data == "address_trace_add":
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[lang]["address_trace_add_tips"]+"\n")
		msg.ParseMode = "HTML"
		ctx.Bot.Send(msg)
		setShortState(ctx.Cache, callbackQuery.Message.Chat.ID, callbackQuery.Data)
		return true
	case callbackQuery.Data == "address_trace_delete":
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[lang]["address_trace_delete_tips"]+"\n")
		msg.ParseMode = "HTML"
		ctx.Bot.Send(msg)
		setShortState(ctx.Cache, callbackQuery.Message.Chat.ID, callbackQuery.Data)
		return true
	case callbackQuery.Data == "deposit_amount":
		service.ShowDepositOptions(lang, ctx.DB, callbackQuery, ctx.Bot)
		return true
	case callbackQuery.Data == "forward_deposit_usdt":
		sendForwardDepositUSDT(ctx.Bot, callbackQuery.Message.Chat.ID, lang, ctx.DB)
		return true
	case callbackQuery.Data == "coin_swap_coin":
		service.MenuNavigateCoin2CoinSwap2(lang, ctx.DB, callbackQuery.Message.Chat.ID, ctx.Bot)
		return true
	case callbackQuery.Data == "function_address_trace":
		service.MenuNavigateAddressTrace(lang, ctx.Cache, ctx.Bot, callbackQuery.Message.Chat.ID, ctx.DB)
		return true
	case callbackQuery.Data == "function_remove_label":
		launder.MenuLaunderNavigate(lang, ctx.DB, callbackQuery.Message.Chat.ID, ctx.Bot)
		return true
	}
	return false
}

func sendMyServiceOverview(bot *tgbotapi.BotAPI, cacheStore cache.Cache, chatID int64, lang string) {
	msg := tgbotapi.NewMessage(chatID, "🛡 当前服务状态：\n\n🔋 能量闪兑\n\n- 剩余笔数：12\n- 自动补能：关闭 /开启\n\n➡️ /闪兑\n\n➡️ /笔数套餐\n\n➡️ /手动发能（1笔）\n\n➡️ /开启/关闭自动发能\n\n📍 地址风险检测\n\n- 今日免费次数：已用完\n\n➡️ /地址风险检测\n\n🚨 USDT冻结预警\n\n- 地址1：TX8kY...5a9rP（剩余12天）✅\n- 地址2：TEw9Q...iS6Ht（剩余28天）✅")
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👁️‍🗨️ "+global.Translations[lang]["alert_monitoring_list"], "address_list_trace"),
		),
	)
	bot.Send(msg)
	cacheStore.Set(strconv.FormatInt(chatID, 10), "usdt_risk_monitor", time.Minute)
}

func sendForwardDepositUSDT(bot *tgbotapi.BotAPI, chatID int64, lang string, db *gorm.DB) {
	usdtSubscriptionsRepo := repositories.NewUserUsdtSubscriptionsRepository(db)
	usdtList, _ := usdtSubscriptionsRepo.ListAll(context.Background())
	var allButtons []tgbotapi.InlineKeyboardButton
	var extraButtons []tgbotapi.InlineKeyboardButton
	var keyboard [][]tgbotapi.InlineKeyboardButton
	for _, record := range usdtList {
		allButtons = append(allButtons, tgbotapi.NewInlineKeyboardButtonData("💰"+record.Name, "deposit_usdt_"+record.Amount))
	}
	extraButtons = append(extraButtons,
		tgbotapi.NewInlineKeyboardButtonData("🔁"+global.Translations[lang]["switch_to_trx_deposit"], "deposit_amount"),
		tgbotapi.NewInlineKeyboardButtonData("🔙"+global.Translations[lang]["back_home"], "back_home"),
	)
	for i := 0; i < len(allButtons); i += 2 {
		end := i + 2
		if end > len(allButtons) {
			end = len(allButtons)
		}
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(allButtons[i:end]...))
	}
	for i := 0; i < len(extraButtons); i += 1 {
		end := i + 1
		if end > len(extraButtons) {
			end = len(extraButtons)
		}
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(extraButtons[i:end]...))
	}
	userRepo := repositories.NewUserRepository(db)
	user, _ := userRepo.GetByChatID(chatID)
	if user.Amount == "" {
		user.Amount = "0"
	}
	if user.TronAmount == "" {
		user.TronAmount = "0"
	}
	msg := tgbotapi.NewMessage(
		chatID,
		"🆔"+global.Translations[lang]["user_id"]+": <code>"+user.Associates+"</code>\n"+
			"👤"+global.Translations[lang]["username"]+": @"+user.Username+"\n"+
			"💰"+global.Translations[lang]["balance"]+": \n"+
			"- TRX：   "+user.TronAmount+"\n"+
			"-  USDT："+user.Amount,
	)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
	msg.ParseMode = "HTML"
	bot.Send(msg)
}
