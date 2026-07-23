package app

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
	"ushield_bot/internal/cache"
	"ushield_bot/internal/config"
	trxfee "ushield_bot/internal/infrastructure/thirdparty"
	"ushield_bot/internal/session"
	"ushield_bot/internal/translate"
)

type Container struct {
	Config       *config.Config
	DB           *gorm.DB
	Bot          *tgbotapi.BotAPI
	Cache        cache.Cache
	Session      session.Store
	Translator   *translate.Service
	Catfee       *trxfee.CatfeeService
	RandomCookie string
}
