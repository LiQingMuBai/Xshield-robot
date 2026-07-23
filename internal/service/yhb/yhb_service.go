package yhb

import (
	"fmt"
	"strings"
	"ushield_bot/internal/global"
	"ushield_bot/internal/infrastructure/repositories"
	. "ushield_bot/internal/infrastructure/tools"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func MenuNavigateTronEnergy(_lang string, db *gorm.DB, message *tgbotapi.Message, bot *tgbotapi.BotAPI) {

	dictDetailRepo := repositories.NewSysDictionariesRepo(db)

	yhb_energy_cost, _ := dictDetailRepo.GetDictionaryDetail("yhb_energy_unit_price")
	yhb_receive_address, _ := dictDetailRepo.GetDictionaryDetail("yhb_receive_address")
	yhb_tron_receive_address, _ := dictDetailRepo.GetDictionaryDetail("yhb_tron_receive_address")
	fmt.Printf("元红包能量单位价: %s\n", yhb_energy_cost)

	result := yhb_energy_cost[:len(yhb_energy_cost)-18]
	fmt.Printf("元红包能量单位价: %s\n", result)
	fmt.Printf("币安链收款地址: %s\n", yhb_receive_address)
	fmt.Printf("波场链收款地址: %s\n", yhb_tron_receive_address)

	energy_cost_2x, _ := StringMultiply(result, 2)
	energy_cost_10x, _ := StringMultiply(result, 10)

	fmt.Printf("energy_cost_2x: %s\n", energy_cost_2x)
	fmt.Printf("energy_cost_10x: %s\n", energy_cost_10x)

	originStr := global.Translations[_lang]["yhb_desc"]

	targetStr := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(originStr, "{energy_cost}", result), "{energy_cost_2x}", energy_cost_2x), "{receiveAddress}", yhb_tron_receive_address), "{energy_cost_10x}", energy_cost_10x)

	targetStr = strings.ReplaceAll(targetStr, "{yhb_tron_address}", yhb_tron_receive_address)
	targetStr = strings.ReplaceAll(targetStr, "{yhb_bsc_address}", yhb_receive_address)
	msg := tgbotapi.NewMessage(message.Chat.ID, targetStr)
	msg.ParseMode = "HTML"
	bot.Send(msg)

}
