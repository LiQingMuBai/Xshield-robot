package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"ushield_bot/internal/app"
	telegramrouter "ushield_bot/internal/handler/telegram/router"
)

func processUpdate(update tgbotapi.Update, c *app.Container) {
	handlerCtx := telegramrouter.NewContext(c)

	switch {
	case update.Message != nil && update.Message.IsCommand():
		telegramrouter.RouteCommandUpdate(update, handlerCtx)
	case update.Message != nil:
		telegramrouter.HandleMessageUpdate(update.Message, handlerCtx)
	case update.CallbackQuery != nil:
		telegramrouter.HandleCallbackQuery(update.CallbackQuery, handlerCtx)
	}
}
