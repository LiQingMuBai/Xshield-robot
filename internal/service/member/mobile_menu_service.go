package member

import (
	"ushield_bot/internal/global"
	logger "ushield_bot/internal/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func MenuMobileNavigate(lang string, db *gorm.DB, chatID int64, bot *tgbotapi.BotAPI) {

	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[lang]["customer_service"], "click_callcenter"),
		),
	)

	// 创建视频消息（从本地文件）
	videoMsg := tgbotapi.NewMessage(chatID, global.Translations[lang]["anonymous_mobile"])
	videoMsg.ReplyMarkup = inlineKeyboard

	// 发送视频
	if _, err := bot.Send(videoMsg); err != nil {
		logger.Errorf("发送视频失败: %v", err)
	}

}
