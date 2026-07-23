package bootstrap

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
	"ushield_bot/internal/app"
	"ushield_bot/internal/cache"
	"ushield_bot/internal/config"
	trxfee "ushield_bot/internal/infrastructure/thirdparty"
	"ushield_bot/internal/session"
	"ushield_bot/internal/translate"
)

func BuildApp() (*app.App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	translator, err := translate.New(cfg.Translation)
	if err != nil {
		return nil, fmt.Errorf("load translations: %w", err)
	}

	db, err := gorm.Open(mysql.Open(cfg.DB.DSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	bot, err := tgbotapi.NewBotAPI(cfg.Telegram.Token)
	if err != nil {
		return nil, fmt.Errorf("init telegram bot: %w", err)
	}
	bot.Debug = cfg.Telegram.Debug

	catfeeClient, err := trxfee.NewCatfeeService(
		cfg.Catfee.APIKey,
		cfg.Catfee.APISecret,
		cfg.Catfee.BaseURL,
	)
	if err != nil {
		return nil, fmt.Errorf("init catfee client: %w", err)
	}

	cacheStore := cache.NewMemoryCache()
	sessionStore := session.NewCacheStore(cacheStore, 24*time.Hour)

	container := &app.Container{
		Config:       cfg,
		DB:           db,
		Bot:          bot,
		Cache:        cacheStore,
		Session:      sessionStore,
		Translator:   translator,
		Catfee:       catfeeClient,
		RandomCookie: cfg.Bot.MistCookie,
	}

	return app.New(container), nil
}
