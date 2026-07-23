package main

import (
	"ushield_bot/internal/app"
	"ushield_bot/internal/bootstrap"
	telegramrouter "ushield_bot/internal/handler/telegram/router"
	logger "ushield_bot/internal/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	if err := logger.Setup("."); err != nil {
		logger.Fatalf("setup logger err: %v", err)
	}

	application, err := bootstrap.BuildApp()
	if err != nil {
		logger.Fatalf("build app err: %v", err)
	}

	if err := application.Run(processUpdate); err != nil {
		logger.Fatalf("run app err: %v", err)
	}
}

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
