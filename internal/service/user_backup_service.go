package service

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
	"strconv"
	"ushield_bot/internal/infrastructure/repositories"
)

func ExtractBackup(message *tgbotapi.Message, bot *tgbotapi.BotAPI, db *gorm.DB) bool {
	chat_ID, err := strconv.ParseInt(message.Text, 10, 64)
	if err != nil {

		msg := tgbotapi.NewMessage(message.Chat.ID, "请输入正确的对方👤用户电报ID？")
		msg.ParseMode = "HTML"
		bot.Send(msg)
		return true
	}

	//用户电报ID
	userRepo := repositories.NewUserRepository(db)
	backupUser, queryErr := userRepo.GetByChatID(chat_ID)
	if queryErr != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "👤用户电报ID未在机器人发现，请让对方用户电报登录机器人")
		msg.ParseMode = "HTML"
		bot.Send(msg)
		return true
	}
	user, _ := userRepo.GetByChatID(message.Chat.ID)
	user.BackupChatID = backupUser.Associates
	err2 := userRepo.Save(context.Background(), &user)
	if err2 == nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "✅ 成功绑定第二紧急联系人: "+backupUser.Associates)
		msg.ParseMode = "HTML"
		bot.Send(msg)
		return true
	}
	return false
}
