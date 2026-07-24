package app

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	logger "ushield_bot/internal/logger"
)

type UpdateProcessor func(update tgbotapi.Update, c *Container)

type App struct {
	container *Container
}

func New(container *Container) *App {
	return &App{container: container}
}

func (a *App) Run(processor UpdateProcessor) error {
	if processor == nil {
		return fmt.Errorf("update processor is nil")
	}

	printStartupBanner(a.container)
	logger.Print(a.container.Translator.T("zh", "start"))
	logger.Print(a.container.Translator.T("en", "start"))

	_, err := a.container.Bot.Request(tgbotapi.NewSetMyCommands(
		tgbotapi.BotCommand{Command: "start", Description: "start"},
		tgbotapi.BotCommand{Command: "hide", Description: "hide"},
	))
	if err != nil {
		logger.Errorf("error setting commands: %v", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = a.container.Config.Telegram.Timeout
	updates := a.container.Bot.GetUpdatesChan(u)

	for update := range updates {
		processor(update, a.container)
	}

	return nil
}
