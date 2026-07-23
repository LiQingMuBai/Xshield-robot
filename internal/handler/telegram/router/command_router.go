package router

import (
	"context"
	"strconv"
	"strings"
	"time"
	"ushield_bot/internal/cache"
	"ushield_bot/internal/domain"
	"ushield_bot/internal/global"
	"ushield_bot/internal/infrastructure/repositories"
	logger "ushield_bot/internal/logger"
	"ushield_bot/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func RouteCommandUpdate(update tgbotapi.Update, ctx Context) {
	switch {
	case strings.HasPrefix(update.Message.Command(), "startDispatch"):
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "📢 功能开发中！想第一时间知道它上线吗？记得关注我们的官方频道：@ushield1 🔔\n\n")
		msg.ParseMode = "HTML"
		ctx.Bot.Send(msg)

	case strings.HasPrefix(update.Message.Command(), "open_ST"):
		subscribeBundleID := strings.ReplaceAll(update.Message.Command(), "open_ST", "")
		dispatchService := service.NewEnergyDispatchService(ctx.DB, ctx.TrxfeeURL, ctx.TrxfeeAPIKey, ctx.TrxfeeSecret, ctx.CatfeeClient)
		lang, err := dispatchService.ToggleSmartDispatch(context.Background(), subscribeBundleID, update.Message.Chat.ID, true)
		if err != nil {
			if err == service.ErrSmartDispatchForbidden {
				logger.Println("不是自己的权利")
			}
			return
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "✅"+global.Translations[lang]["smart_transaction_auto_dispatch_open_successfully"]+"\n")
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[lang]["back_homepage"], "back_bundle_package_ST"),
			),
		)
		ctx.Bot.Send(msg)

	case strings.HasPrefix(update.Message.Command(), "close_ST"):
		subscribeBundleID := strings.ReplaceAll(update.Message.Command(), "close_ST", "")
		dispatchService := service.NewEnergyDispatchService(ctx.DB, ctx.TrxfeeURL, ctx.TrxfeeAPIKey, ctx.TrxfeeSecret, ctx.CatfeeClient)
		lang, err := dispatchService.ToggleSmartDispatch(context.Background(), subscribeBundleID, update.Message.Chat.ID, false)
		if err != nil {
			if err == service.ErrSmartDispatchForbidden {
				logger.Println("不是自己的权利")
			}
			return
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "✅"+global.Translations[lang]["smart_transaction_auto_dispatch_close_successfully"]+"\n")
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[lang]["back_homepage"], "back_bundle_package_ST"),
			),
		)
		ctx.Bot.Send(msg)

	case strings.HasPrefix(update.Message.Command(), "dispatchNow"):
		command := update.Message.Command()
		if !strings.Contains(command, "_") {
			command = command + "_1"
		}
		subscribeBundleIDStr := strings.Split(command, "_")[0]
		timesStr := strings.Split(command, "_")[1]
		times, _ := strconv.Atoi(timesStr)
		subscribeBundleID := strings.ReplaceAll(subscribeBundleIDStr, "dispatchNow", "")

		dispatchService := service.NewEnergyDispatchService(ctx.DB, ctx.TrxfeeURL, ctx.TrxfeeAPIKey, ctx.TrxfeeSecret, ctx.CatfeeClient)
		result, err := dispatchService.DispatchFromPackageAddress(context.Background(), subscribeBundleID, update.Message.Chat.ID, times)
		if err != nil {
			switch err {
			case service.ErrDispatchForbidden:
				logger.Println("不是自己的权利")
			case service.ErrDispatchInsufficientTimes:
				userRepo := repositories.NewUserRepository(ctx.DB)
				user, userErr := userRepo.GetByChatID(update.Message.Chat.ID)
				if userErr == nil {
					msg := service.BuildBundlePackageAddressSummaryMessage(user.Lang, ctx.DB, update.Message.Chat.ID)
					ctx.Bot.Send(msg)
				}
			}
			return
		}
		sendDispatchSuccess(ctx.Bot, update.Message.Chat.ID, result)

	case strings.HasPrefix(update.Message.Command(), "stopDispatch"):
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "📢 功能开发中！想第一时间知道它上线吗？记得关注我们的官方频道：@ushield1 🔔\n\n")
		msg.ParseMode = "HTML"
		ctx.Bot.Send(msg)

	case strings.HasPrefix(update.Message.Command(), "dispatchOthers"):
		subscribeBundleID := strings.ReplaceAll(update.Message.Command(), "dispatchOthers", "")
		userRepo := repositories.NewUserRepository(ctx.DB)
		record, _ := userRepo.GetByChatID(update.Message.Chat.ID)
		service.DispatchOthers(record.Lang, subscribeBundleID, ctx.Cache, ctx.Bot, update.Message.Chat.ID, ctx.DB)

	case update.Message.Command() == "start":
		handleStartBootstrap(update.Message, ctx)

	case update.Message.Command() == "hide":
		handleHideCommand(ctx.Bot, update.Message)
	}
}

func handleStartBootstrap(message *tgbotapi.Message, ctx Context) {
	userRepo := repositories.NewUserRepository(ctx.DB)
	index := strings.LastIndex(message.Text, " ")
	parentUID := ""
	if index > 0 {
		parentUIDStr := message.Text
		parentUID = parentUIDStr[index+1:]
		record, err := userRepo.GetByChatIDString(parentUID)
		if err != nil {
			parentUID = ""
		} else {
			parentUID = record.Associates
		}
	}

	record, err := userRepo.GetByChatID(message.Chat.ID)
	if err != nil {
		var user domain.User
		user.Associates = strconv.FormatInt(message.Chat.ID, 10)
		user.Username = message.Chat.UserName
		user.Lang = "zh"
		user.CreatedAt = time.Now()
		user.BotName = ctx.BotName
		if len(parentUID) > 0 {
			user.ParentUserID = parentUID
		}
		if createErr := userRepo.CreateWithContext(context.Background(), &user); createErr != nil {
			return
		}
		ctx.Cache.Set("LANG_"+strconv.FormatInt(message.Chat.ID, 10), "zh", 24*time.Hour)
	} else {
		record.Username = message.From.UserName
		userRepo.UpdateUsernameByChatID(message.From.UserName, message.Chat.ID)
		if len(record.Lang) > 0 {
			ctx.Cache.Set("LANG_"+strconv.FormatInt(message.Chat.ID, 10), record.Lang, 24*time.Hour)
		} else {
			ctx.Cache.Set("LANG_"+strconv.FormatInt(message.Chat.ID, 10), "zh", 24*time.Hour)
		}
	}

	handleStartCommand(ctx.Cache, ctx.Bot, message)
}

func handleStartCommand(cacheStore cache.Cache, bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	lang, err := cacheStore.Get("LANG_" + strconv.FormatInt(message.Chat.ID, 10))
	if err != nil {
		lang = "zh"
	}
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("⛽"+global.Translations[lang]["tron_energy_menu"]),
			tgbotapi.NewKeyboardButton("🖊️"+global.Translations[lang]["transaction_plans"]),
			tgbotapi.NewKeyboardButton("🤖"+global.Translations[lang]["catfee_smart_transaction_menu"]),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("✅"+global.Translations[lang]["usdt_trx_swap"]),
			tgbotapi.NewKeyboardButton("🔍"+global.Translations[lang]["address_check"]),
			tgbotapi.NewKeyboardButton("🚨"+global.Translations[lang]["usdt_freeze_alert"]),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(global.Translations[lang]["member_telegram_menu"]),
			tgbotapi.NewKeyboardButton("🛒"+global.Translations[lang]["ushield_additional_services_menu"]),
			tgbotapi.NewKeyboardButton("👤"+global.Translations[lang]["my_account"]),
		),
	)
	keyboard.OneTimeKeyboard = false
	keyboard.ResizeKeyboard = true
	keyboard.Selective = false
	msg := tgbotapi.NewMessage(message.Chat.ID, global.Translations[lang]["welcome_tips"])
	msg.ReplyMarkup = keyboard
	msg.ParseMode = "HTML"
	bot.Send(msg)
}

func handleHideCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	hideKeyboard := tgbotapi.NewRemoveKeyboard(true)
	msg := tgbotapi.NewMessage(message.Chat.ID, "键盘已隐藏，发送 /start 重新显示")
	msg.ReplyMarkup = hideKeyboard
	bot.Send(msg)
}
