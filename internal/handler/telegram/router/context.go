package router

import (
	"ushield_bot/internal/app"
	"ushield_bot/internal/cache"
	trxfee "ushield_bot/internal/infrastructure/thirdparty"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

type Context struct {
	DB                  *gorm.DB
	Bot                 *tgbotapi.BotAPI
	Cache               cache.Cache
	TrxfeeURL           string
	TrxfeeAPIKey        string
	TrxfeeSecret        string
	FixedfloatURL       string
	FixedFloatAPIKey    string
	FixedFloatAPISecret string
	BotName             string
	CatfeeClient        *trxfee.CatfeeService
	RandomCookie        string
}

func NewContext(container *app.Container) Context {
	return Context{
		DB:                  container.DB,
		Bot:                 container.Bot,
		Cache:               container.Cache,
		TrxfeeURL:           container.Config.Trxfee.BaseURL,
		TrxfeeAPIKey:        container.Config.Trxfee.APIKey,
		TrxfeeSecret:        container.Config.Trxfee.APISecret,
		FixedfloatURL:       container.Config.FixedFloat.RefURL,
		FixedFloatAPIKey:    container.Config.FixedFloat.APIKey,
		FixedFloatAPISecret: container.Config.FixedFloat.APISecret,
		BotName:             container.Config.Bot.Name,
		CatfeeClient:        container.Catfee,
		RandomCookie:        container.RandomCookie,
	}
}
