package config

import (
	"github.com/joho/godotenv"
	"os"
)

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	cfg := &Config{
		DB: DBConfig{
			DSN: os.Getenv("MYSQL_DSN"),
		},
		Telegram: TelegramConfig{
			Token:   os.Getenv("TELEGRAM_BOT_TOKEN"),
			Debug:   true,
			Timeout: 60,
		},
		Trxfee: TrxfeeConfig{
			BaseURL:   os.Getenv("TRXFEE_BASE_URL"),
			APIKey:    os.Getenv("TRXFEE_API_KEY"),
			APISecret: os.Getenv("TRXFEE_API_SECRET"),
		},
		Catfee: CatfeeConfig{
			BaseURL:   os.Getenv("CATFEE_BASE_URL"),
			APIKey:    os.Getenv("CATFEE_API_KEY"),
			APISecret: os.Getenv("CATFEE_API_SECRET"),
		},
		FixedFloat: FixedFloatConfig{
			RefURL: os.Getenv("FIXEDFLOAT_REF_URL"),
		},
		Bot: BotConfig{
			Name:       os.Getenv("BOT_NAME"),
			MistCookie: os.Getenv("MIST_COOKIE"),
			Agent:      os.Getenv("BOT_AGENT"),
		},
		Translation: TranslationConfig{
			Dir:            "translations",
			SupportedLangs: []string{"en", "zh", "ar", "es", "pt", "ko", "th", "ja", "vi", "ch", "ru", "fa"},
			DefaultLang:    "zh",
		},
	}

	return cfg, nil
}
