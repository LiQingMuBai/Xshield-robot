package config

type Config struct {
	DB          DBConfig
	Telegram    TelegramConfig
	Trxfee      TrxfeeConfig
	Catfee      CatfeeConfig
	FixedFloat  FixedFloatConfig
	Tron        TronConfig
	Bot         BotConfig
	Translation TranslationConfig
}

type DBConfig struct {
	DSN string
}

type TelegramConfig struct {
	Token   string
	Debug   bool
	Timeout int
}

type TrxfeeConfig struct {
	BaseURL   string
	APIKey    string
	APISecret string
}

type CatfeeConfig struct {
	BaseURL   string
	APIKey    string
	APISecret string
}

type FixedFloatConfig struct {
	RefURL    string
	APIKey    string
	APISecret string
}

type TronConfig struct {
	Mnemonic string
}

type BotConfig struct {
	Name       string
	MistCookie string
	Agent      string
}

type TranslationConfig struct {
	Dir            string
	SupportedLangs []string
	DefaultLang    string
}
