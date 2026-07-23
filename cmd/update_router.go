package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
	"ushield_bot/internal/app"
	"ushield_bot/internal/cache"
	trxfee "ushield_bot/internal/infrastructure/3rd"
)

type updateDeps struct {
	db            *gorm.DB
	bot           *tgbotapi.BotAPI
	cache         cache.Cache
	trxfeeURL     string
	trxfeeAPIKey  string
	trxfeeSecret  string
	fixedfloatURL string
	botName       string
	catfeeClient  *trxfee.CatfeeService
	randomCookie  string
}

func processUpdate(update tgbotapi.Update, c *app.Container) {
	deps := updateDeps{
		db:            c.DB,
		bot:           c.Bot,
		cache:         c.Cache,
		trxfeeURL:     c.Config.Trxfee.BaseURL,
		trxfeeAPIKey:  c.Config.Trxfee.APIKey,
		trxfeeSecret:  c.Config.Trxfee.APISecret,
		fixedfloatURL: c.Config.FixedFloat.RefURL,
		botName:       c.Config.Bot.Name,
		catfeeClient:  c.Catfee,
		randomCookie:  c.RandomCookie,
	}

	switch {
	case update.Message != nil && update.Message.IsCommand():
		routeCommandUpdate(update, deps)
	case update.Message != nil:
		routeMessageUpdate(update, deps)
	case update.CallbackQuery != nil:
		routeCallbackUpdate(update, deps)
	}
}

func routeMessageUpdate(update tgbotapi.Update, deps updateDeps) {
	handleRegularMessage(
		deps.cache,
		deps.bot,
		update.Message,
		deps.db,
		deps.randomCookie,
		deps.trxfeeURL,
		deps.trxfeeAPIKey,
		deps.trxfeeSecret,
		deps.fixedfloatURL,
		deps.catfeeClient,
	)
}

func routeCallbackUpdate(update tgbotapi.Update, deps updateDeps) {
	handleCallbackQuery(
		deps.cache,
		deps.bot,
		update.CallbackQuery,
		deps.db,
		deps.trxfeeURL,
		deps.trxfeeAPIKey,
		deps.trxfeeSecret,
		deps.catfeeClient,
	)
}
