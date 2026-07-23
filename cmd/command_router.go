package main

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"strings"
	"time"
	"ushield_bot/internal/domain"
	"ushield_bot/internal/global"
	"ushield_bot/internal/infrastructure/repositories"
	"ushield_bot/internal/service"
)

func routeCommandUpdate(update tgbotapi.Update, deps updateDeps) {
	switch {
	case strings.HasPrefix(update.Message.Command(), "startDispatch"):
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "📢 功能开发中！想第一时间知道它上线吗？记得关注我们的官方频道：@ushield1 🔔\n\n")
		msg.ParseMode = "HTML"
		deps.bot.Send(msg)

	case strings.HasPrefix(update.Message.Command(), "open_ST"):
		subscribeBundleID := strings.ReplaceAll(update.Message.Command(), "open_ST", "")
		log.Println("subscribeBundleID : " + subscribeBundleID)
		log.Println(subscribeBundleID + "   open_ST command")

		dispatchService := service.NewEnergyDispatchService(deps.db, deps.trxfeeURL, deps.trxfeeAPIKey, deps.trxfeeSecret, deps.catfeeClient)
		lang, err := dispatchService.ToggleSmartDispatch(context.Background(), subscribeBundleID, update.Message.Chat.ID, true)
		if err != nil {
			if err == service.ErrSmartDispatchForbidden {
				log.Println("不是自己的权利")
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
		deps.bot.Send(msg)

	case strings.HasPrefix(update.Message.Command(), "close_ST"):
		subscribeBundleID := strings.ReplaceAll(update.Message.Command(), "close_ST", "")
		log.Println("subscribeBundleID : " + subscribeBundleID)
		log.Println(subscribeBundleID + "   closeST command")

		dispatchService := service.NewEnergyDispatchService(deps.db, deps.trxfeeURL, deps.trxfeeAPIKey, deps.trxfeeSecret, deps.catfeeClient)
		lang, err := dispatchService.ToggleSmartDispatch(context.Background(), subscribeBundleID, update.Message.Chat.ID, false)
		if err != nil {
			if err == service.ErrSmartDispatchForbidden {
				log.Println("不是自己的权利")
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
		deps.bot.Send(msg)

	case strings.HasPrefix(update.Message.Command(), "dispatchNow"):
		command := update.Message.Command()
		if !strings.Contains(command, "_") {
			command = command + "_1"
		}
		subscribeBundleIDStr := strings.Split(command, "_")[0]
		timesStr := strings.Split(command, "_")[1]

		log.Printf("times : %s\n", timesStr)
		times, _ := strconv.Atoi(timesStr)

		subscribeBundleID := strings.ReplaceAll(subscribeBundleIDStr, "dispatchNow", "")
		log.Println("subscribeBundleID : " + subscribeBundleID)
		log.Println(subscribeBundleID + "   dispatchNow command")

		dispatchService := service.NewEnergyDispatchService(deps.db, deps.trxfeeURL, deps.trxfeeAPIKey, deps.trxfeeSecret, deps.catfeeClient)
		result, err := dispatchService.DispatchFromPackageAddress(context.Background(), subscribeBundleID, update.Message.Chat.ID, times)
		if err != nil {
			switch err {
			case service.ErrDispatchForbidden:
				log.Println("不是自己的权利")
			case service.ErrDispatchInsufficientTimes:
				userRepo := repositories.NewUserRepository(deps.db)
				user, userErr := userRepo.GetByUserID(update.Message.Chat.ID)
				if userErr == nil {
					msg := service.CLICK_BUNDLE_PACKAGE_ADDRESS_STATS2(user.Lang, deps.db, update.Message.Chat.ID)
					deps.bot.Send(msg)
				}
			}
			return
		}
		sendDispatchSuccess(deps.bot, update.Message.Chat.ID, result)

	case strings.HasPrefix(update.Message.Command(), "stopDispatch"):
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "📢 功能开发中！想第一时间知道它上线吗？记得关注我们的官方频道：@ushield1 🔔\n\n")
		msg.ParseMode = "HTML"
		deps.bot.Send(msg)

	case strings.HasPrefix(update.Message.Command(), "dispatchOthers"):
		subscribeBundleID := strings.ReplaceAll(update.Message.Command(), "dispatchOthers", "")
		log.Println("subscribeBundleID :" + subscribeBundleID)
		log.Println(subscribeBundleID + "dispatchOthers command")

		userRepo := repositories.NewUserRepository(deps.db)
		record, _ := userRepo.GetByUserID(update.Message.Chat.ID)
		service.DispatchOthers(record.Lang, subscribeBundleID, deps.cache, deps.bot, update.Message.Chat.ID, deps.db)

	case update.Message.Command() == "start":
		log.Printf("1")
		log.Println("text: " + update.Message.Text)
		userRepo := repositories.NewUserRepository(deps.db)
		index := strings.LastIndex(update.Message.Text, " ")
		parentUID := ""
		if index > 0 {
			parentUIDStr := update.Message.Text
			parentUID = parentUIDStr[index+1:]
			fmt.Printf("parentUIDStr: %s\n", parentUID)

			record, err := userRepo.GetByUserIDStr(parentUID)
			if err != nil {
				parentUID = ""
			} else {
				parentUID = record.Associates
			}
		}

		record, err := userRepo.GetByUserID(update.Message.Chat.ID)
		if err != nil {
			var user domain.User
			user.Associates = strconv.FormatInt(update.Message.Chat.ID, 10)
			user.Username = update.Message.Chat.UserName
			user.Lang = "zh"
			user.CreatedAt = time.Now()
			user.BotName = deps.botName
			if len(parentUID) > 0 {
				user.ParentUserID = parentUID
			}
			err := userRepo.Create2(context.Background(), &user)

			expiration := 24 * time.Hour
			deps.cache.Set("LANG_"+strconv.FormatInt(update.Message.Chat.ID, 10), "zh", expiration)
			if err != nil {
				return
			}
		}

		if err == nil {
			record.Username = update.Message.From.UserName
			userRepo.UpdateUserNameByChatID(update.Message.From.UserName, update.Message.Chat.ID)

			log.Printf("UserName: %s", record.Username)
			log.Printf("Associates %s", record.Associates)
			expiration := 24 * time.Hour
			if len(record.Lang) > 0 {
				deps.cache.Set("LANG_"+strconv.FormatInt(update.Message.Chat.ID, 10), record.Lang, expiration)
			} else {
				deps.cache.Set("LANG_"+strconv.FormatInt(update.Message.Chat.ID, 10), "zh", expiration)
			}
		}

		handleStartCommand(deps.cache, deps.bot, update.Message)

	case update.Message.Command() == "hide":
		log.Printf("2")
		handleHideCommand(deps.cache, deps.bot, update.Message)
	}
}
