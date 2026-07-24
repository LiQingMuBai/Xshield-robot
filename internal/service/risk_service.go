package service

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
	"strconv"
	"time"
	"ushield_bot/internal/cache"
	"ushield_bot/internal/global"
)

func StartFreezeRiskInput(lang string, cache cache.Cache, db *gorm.DB, callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI) {

	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[lang]["enter_address_for_alert"])
	msg.ParseMode = "HTML"
	bot.Send(msg)
	expiration := 1 * time.Minute // 短时间缓存空值
	//设置用户状态
	cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), "usdt_risk_monitor", expiration)

	//扣减

}
