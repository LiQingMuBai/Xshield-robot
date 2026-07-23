package command

import (
	"strings"
	"ushield_bot/internal/global"
	"ushield_bot/internal/infrastructure/repositories"
	"ushield_bot/internal/infrastructure/tools"
	logger "ushield_bot/internal/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func MenuNavigate(_lang string, db *gorm.DB, _chatID int64, bot *tgbotapi.BotAPI) {
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🧾"+global.Translations[_lang]["address_list"], "click_bundle_package_address_stats"),
			tgbotapi.NewInlineKeyboardButtonData("➕"+global.Translations[_lang]["add_address"], "address_manager_add"),
		),
	)

	dictDetailRepo := repositories.NewSysDictionariesRepo(db)

	energy_cost, _ := dictDetailRepo.GetDictionaryDetail("energy_cost")

	logger.Printf("energy_cost: %s\n", energy_cost)

	energy_cost_2x, _ := tools.StringMultiply(energy_cost, 2)
	energy_cost_3x, _ := tools.StringMultiply(energy_cost, 3)
	energy_cost_4x, _ := tools.StringMultiply(energy_cost, 4)
	energy_cost_5x, _ := tools.StringMultiply(energy_cost, 5)

	logger.Printf("energy_cost_2x: %s\n", energy_cost_2x)

	originStr := global.Translations[_lang]["command_energy_desc"]

	targetStr := strings.ReplaceAll(originStr, "{1_times_cost_trx}", energy_cost)
	targetStr = strings.ReplaceAll(targetStr, "{2_times_cost_trx}", energy_cost_2x)
	targetStr = strings.ReplaceAll(targetStr, "{3_times_cost_trx}", energy_cost_3x)
	targetStr = strings.ReplaceAll(targetStr, "{4_times_cost_trx}", energy_cost_4x)
	targetStr = strings.ReplaceAll(targetStr, "{5_times_cost_trx}", energy_cost_5x)

	msg := tgbotapi.NewMessage(_chatID, targetStr)
	msg.ReplyMarkup = inlineKeyboard
	msg.ParseMode = "HTML"

	bot.Send(msg)
}
