package service

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
	"strconv"
	"time"
	"ushield_bot/internal/cache"
	"ushield_bot/internal/global"
)

func StartFreezeRiskInput(_lang string, cache cache.Cache, db *gorm.DB, callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI) {

	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[_lang]["enter_address_for_alert"])
	msg.ParseMode = "HTML"
	//inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
	//	tgbotapi.NewInlineKeyboardRow(
	//		tgbotapi.NewInlineKeyboardButtonData("✅ 确认开启", "start_freeze_risk_1"),
	//		tgbotapi.NewInlineKeyboardButtonData("❌ 取消操作", "back_risk_home"),
	//	),
	//	tgbotapi.NewInlineKeyboardRow(
	//		tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[_lang]["back_homepage"], "back_risk_home"),
	//	),
	//)
	//msg.ReplyMarkup = inlineKeyboard
	bot.Send(msg)
	expiration := 1 * time.Minute // 短时间缓存空值
	//设置用户状态
	cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), "usdt_risk_monitor", expiration)

	//userRepo := repositories.NewUserRepository(db)
	//user, _ := userRepo.GetByChatID(callbackQuery.Message.Chat.ID)
	//if IsEmpty(user.Amount) {
	//	user.Amount = "0"
	//}
	//
	//if IsEmpty(user.TronAmount) {
	//	user.TronAmount = "0"
	//}
	//
	//userAddressRepo := repositories.NewUserAddressMonitorRepo(db)
	//
	//addresses, _ := userAddressRepo.ListByChatID(context.Background(), callbackQuery.Message.Chat.ID)
	//
	//nums := len(addresses)
	//
	////if nums == 0 {
	////
	////	return
	////}
	////扣trx
	////var COST_FROM_TRX bool
	////var COST_FROM_USDT bool
	//
	//sysDictionariesRepo := repositories.NewSysDictionariesRepo(db)
	//
	//server_trx_price, _ := sysDictionariesRepo.GetDictionaryDetail("server_trx_price")
	//
	//server_usdt_price, _ := sysDictionariesRepo.GetDictionaryDetail("server_usdt_price")
	//
	//if CompareStringsWithFloat(user.TronAmount, server_trx_price, float64(nums)) || CompareStringsWithFloat(user.Amount, server_usdt_price, float64(nums)) {
	//
	//	var builder strings.Builder
	//	//- [6.29] +3000 TRX（订单 #TOPUP-92308）
	//
	//	for _, item := range addresses {
	//
	//		builder.WriteString(global.Translations[_lang]["address"]+"：")
	//		builder.WriteString(item.Address)
	//		builder.WriteString("\n")
	//	}
	//	// 去除最后一个空格
	//	result := strings.TrimSpace(builder.String())
	//	text := "余额充足\n\n\n\n🔍 即将为以下地址开启冻结预警：\n" +
	//		result +
	//		"🎯 服务开通后U盾将 24 小时不间断保护您的资产安全。\n" +
	//		"⏰ 系统将在冻结前启动预警机制，持续 10 分钟每分钟推送提醒，通知您及时转移资产。\n" +
	//		"📌 服务费用：" + server_trx_price + "TRX / 30 天 或 " + server_usdt_price + " USDT / 30 天\n是否确认开通该服务"

	//扣减

	//if CompareStringsWithFloat(user.TronAmount, server_trx_price, float64(nums)) {
	//	rest, _ := SubtractStringNumbers(user.TronAmount, server_trx_price, float64(nums))
	//
	//	user.TronAmount = rest
	//	userRepo.Save(context.Background(), &user)
	//	fmt.Printf("rest: %s", rest)
	//	COST_FROM_TRX = true
	//	//扣usdt
	//} else if CompareStringsWithFloat(user.Amount, server_usdt_price, float64(nums)) {
	//	rest, _ := SubtractStringNumbers(user.Amount, server_usdt_price, float64(nums))
	//	fmt.Printf("rest: %s", rest)
	//	user.Amount = rest
	//	userRepo.Save(context.Background(), &user)
	//	COST_FROM_USDT = true
	//}
	//
	////添加记录
	//userAddressEventRepo := repositories.NewUserAddressMonitorEventRepo(db)
	//
	//for _, address := range addresses {
	//	var event domain.UserAddressMonitorEvent
	//	event.ChatID = callbackQuery.Message.Chat.ID
	//	event.Status = 1
	//	event.Address = address.Address
	//	event.Network = address.Network
	//	event.Days = 1
	//	if COST_FROM_TRX {
	//		event.Amount = server_trx_price + " TRX"
	//	}
	//	if COST_FROM_USDT {
	//		event.Amount = server_usdt_price + " USDT"
	//	}
	//	userAddressEventRepo.Create(context.Background(), &event)
	//}
	////后台跟踪起来
	//user, _ := userRepo.GetByChatID(callbackQuery.Message.Chat.ID)
	//msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID,
	//	"💬"+"<b>"+"用户姓名: "+"</b>"+user.Username+"\n"+
	//		"👤"+"<b>"+"用户电报ID: "+"</b>"+user.Associates+"\n"+
	//		"💵"+"<b>"+"当前TRX余额:  "+"</b>"+user.TronAmount+" TRX"+"\n"+
	//		"💴"+"<b>"+"当前USDT余额:  "+"</b>"+user.Amount+" USDT")
	//msg.ParseMode = "HTML"
	//inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
	//	tgbotapi.NewInlineKeyboardRow(
	//		tgbotapi.NewInlineKeyboardButtonData("🔙️返回", "address_manager_return"),
	//	),
	//)

	//msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, text)
	//msg.ParseMode = "HTML"
	//inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
	//	tgbotapi.NewInlineKeyboardRow(
	//		tgbotapi.NewInlineKeyboardButtonData("✅ 确认开启", "start_freeze_risk_1"),
	//		tgbotapi.NewInlineKeyboardButtonData("❌ 取消操作", "back_risk_home"),
	//	),
	//	tgbotapi.NewInlineKeyboardRow(
	//		tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[_lang]["back_homepage"], "back_risk_home"),
	//	),
	//)
	//msg.ReplyMarkup = inlineKeyboard
	//bot.Send(msg)

	//feedback := "✅" + "USDT地址冻结预警扣款成功\n\n"
	//msg2 := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, feedback)
	//msg2.ParseMode = "HTML"
	//bot.Send(msg2)

	//} else {
	//
	//	//余额不足，需充值
	//	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID,
	//		//"💬"+"<b>"+"用户姓名: "+"</b>"+user.Username+"\n"+
	//		//	"👤"+"<b>"+"用户电报ID: "+"</b>"+user.Associates+"\n"+
	//		//	"💵"+"<b>"+"当前TRX余额:  "+"</b>"+user.TronAmount+" TRX"+"\n"+
	//		//	"💴"+"<b>"+"当前USDT余额:  "+"</b>"+user.Amount+" USDT")
	//
	//		"⚠️ 当前余额不足，无法开启冻结预警服务 "+"\n"+
	//			"🆔"+global.Translations[_lang]["user_id"]+": "+user.Associates+"\n"+
	//			"👤"+global.Translations[_lang]["username"]+": @"+user.Username+"\n"+
	//			"💰"+global.Translations[_lang]["balance"]+": "+"\n"+
	//			"- TRX：   "+user.TronAmount+"\n"+
	//			"-  USDT："+user.Amount)
	//	msg.ParseMode = "HTML"
	//	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
	//		tgbotapi.NewInlineKeyboardRow(
	//			tgbotapi.NewInlineKeyboardButtonData("💵"+global.Translations[_lang]["deposit"], "deposit_amount"),
	//		),
	//	)
	//
	//	msg.ReplyMarkup = inlineKeyboard
	//	bot.Send(msg)
}
