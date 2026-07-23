package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"ushield_bot/internal/bootstrap"
	"ushield_bot/internal/infrastructure/3rd/fixedfloat"
	"ushield_bot/internal/service/additional"
	"ushield_bot/internal/service/catfee"
	"ushield_bot/internal/service/command"
	"ushield_bot/internal/service/launder"
	"ushield_bot/internal/service/member"
	"ushield_bot/internal/service/yhb"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"

	"time"
	"ushield_bot/internal/global"
	trxfee "ushield_bot/internal/infrastructure/3rd"
	"ushield_bot/internal/service"

	"ushield_bot/internal/cache"
	"ushield_bot/internal/domain"
	"ushield_bot/internal/infrastructure/repositories"
	. "ushield_bot/internal/infrastructure/tools"
)

func main() {
	application, err := bootstrap.BuildApp()
	if err != nil {
		log.Fatalf("build app err: %v", err)
	}

	if err := application.Run(processUpdate); err != nil {
		log.Fatalf("run app err: %v", err)
	}
}

// 处理 /start 命令 - 显示永久键盘
func handleStartCommand(cache cache.Cache, bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	// 创建永久性回复键盘

	_lang, err := cache.Get("LANG_" + strconv.FormatInt(message.Chat.ID, 10))

	if err != nil {
		_lang = "zh"
	}
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(

			tgbotapi.NewKeyboardButton("⛽"+global.Translations[_lang]["tron_energy_menu"]),

			//tgbotapi.NewInlineKeyboardButtonData("⚡"+global.Translations[_lang]["energy_swap"], "click_energy_swap"),
			tgbotapi.NewKeyboardButton("🖊️"+global.Translations[_lang]["transaction_plans"]),
			//tgbotapi.NewInlineKeyboardButtonData("🤖"+global.Translations[_lang]["smart_transaction_plans"], "click_smart_transaction_plan"),
			tgbotapi.NewKeyboardButton("🤖"+global.Translations[_lang]["catfee_smart_transaction_menu"]),
			//tgbotapi.NewKeyboardButton("🤖"+global.Translations[_lang]["catfee_smart_transaction_menu"]),

			//tgbotapi.NewKeyboardButton("⚡"+global.Translations[_lang]["energy_swap"]),
			//tgbotapi.NewKeyboardButton("🖊️"+global.Translations[_lang]["transaction_plans"]),
			//tgbotapi.NewKeyboardButton("🤖"+global.Translations[_lang]["smart_transaction_plans"]),
		),

		tgbotapi.NewKeyboardButtonRow(
			//tgbotapi.NewKeyboardButton(global.Translations[_lang]["command_energy_menu"]),
			tgbotapi.NewKeyboardButton("✅"+global.Translations[_lang]["usdt_trx_swap"]),
			tgbotapi.NewKeyboardButton("🔍"+global.Translations[_lang]["address_check"]),
			tgbotapi.NewKeyboardButton("🚨"+global.Translations[_lang]["usdt_freeze_alert"]),
			//tgbotapi.NewKeyboardButton("🚨"+global.Translations[_lang]["usdt_freeze_alert"]),

			//tgbotapi.NewKeyboardButton("🧧"+global.Translations[_lang]["yhb_menu"]),
			//tgbotapi.NewKeyboardButton("👤"+global.Translations[_lang]["my_account"]),

		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(global.Translations[_lang]["member_telegram_menu"]),
			tgbotapi.NewKeyboardButton("🛒"+global.Translations[_lang]["ushield_additional_services_menu"]),
			tgbotapi.NewKeyboardButton("👤"+global.Translations[_lang]["my_account"]),
		),

		//tgbotapi.NewKeyboardButtonRow(
		//	tgbotapi.NewKeyboardButton("🥂"+global.Translations[_lang]["coin_laundering_menu"]),
		//	tgbotapi.NewKeyboardButton("🕸"+global.Translations[_lang]["address_trace_menu"]),
		//	tgbotapi.NewKeyboardButton("🔃"+global.Translations[_lang]["coin_swap_coin_menu"]),
		//),
		//tgbotapi.NewKeyboardButtonRow(
		//	tgbotapi.NewKeyboardButton("⚽️世界杯竞猜🏆"),
		//),
	)

	// 关键设置：确保键盘一直存在
	keyboard.OneTimeKeyboard = false
	keyboard.ResizeKeyboard = true
	keyboard.Selective = false
	originStr := global.Translations[_lang]["welcome_tips"]
	//msg := tgbotapi.NewMessage(message.Chat.ID, "🛡️U盾，做您链上资产的护盾！\n我们不仅关注低价能量，更专注于交易安全！\n让每一笔转账都更安心，让每一次链上交互都值得信任！\n🤖 三大实用功能，助您安全、高效地管理链上资产\n🔋 波场能量闪兑, 节省超过80%!\n🕵️ 地址风险检测, 让每一笔转账都更安心!\n🚨 USDT冻结预警,秒级响应，让您的U永不冻结！\n🎉新用户福利：每日一次免费地址风险查询")
	msg := tgbotapi.NewMessage(message.Chat.ID, originStr)
	msg.ReplyMarkup = keyboard
	msg.ParseMode = "HTML"
	bot.Send(msg)
}

// 处理 /hide 命令 - 隐藏键盘
func handleHideCommand(cache cache.Cache, bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	hideKeyboard := tgbotapi.NewRemoveKeyboard(true)
	msg := tgbotapi.NewMessage(message.Chat.ID, "键盘已隐藏，发送 /start 重新显示")
	msg.ReplyMarkup = hideKeyboard
	bot.Send(msg)
}

// 处理普通消息（键盘按钮点击）
func handleRegularMessage(cache cache.Cache, bot *tgbotapi.BotAPI, message *tgbotapi.Message, db *gorm.DB, _cookie string, _trxfeeUrl, _trxfeeApiKey, _trxfeeSecret string, fixfloatedUrl string, catfeeClient *trxfee.CatfeeService) {
	_lang, err := cache.Get("LANG_" + strconv.FormatInt(message.Chat.ID, 10))
	if len(_lang) == 0 || err != nil {
		userRepo := repositories.NewUserRepository(db)
		record, _ := userRepo.GetByUserID(message.Chat.ID)
		expiration := 24 * time.Hour // 短时间缓存空值
		cache.Set("LANG_"+strconv.FormatInt(message.Chat.ID, 10), record.Lang, expiration)
		_lang = record.Lang
	}

	switch message.Text {
	case "⚽️世界杯竞猜🏆":
		msg := tgbotapi.NewMessage(message.Chat.ID, "点击跳转机器人 => 🤖@ushield_octopus_bot\n")
		msg.ParseMode = "HTML"
		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				//tgbotapi.NewInlineKeyboardButtonData("解绑地址", "free_monitor_address"),
				tgbotapi.NewInlineKeyboardButtonData("🔙"+global.Translations[_lang]["back_home"], "back_home"),
			),
		)
		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)

	case "🥂" + global.Translations[_lang]["coin_laundering_menu"]:

		fmt.Printf("洗U步骤开始\n")
		launder.MenuLaunderNavigate(_lang, db, message.Chat.ID, bot)
	case global.Translations[_lang]["member_telegram_menu"]:
		member.MenuNavigate(_lang, db, message.Chat.ID, bot)
	case global.Translations[_lang]["command_energy_menu"]:
		command.MenuNavigate(_lang, db, message.Chat.ID, bot)
	case "🛒" + global.Translations[_lang]["ushield_additional_services_menu"]:
		additional.MenuNavigate(_lang, db, message.Chat.ID, bot)
	case "🔃" + global.Translations[_lang]["coin_swap_coin_menu"]:
		service.MenuNavigateCoin2CoinSwap(_lang, db, message, bot, fixfloatedUrl)
	case "🧧" + global.Translations[_lang]["yhb_menu"]:
		yhb.MenuNavigateTronEnergy(_lang, db, message, bot)
	case "⛽" + global.Translations[_lang]["tron_energy_menu"]:
		service.MenuNavigateTronEnergy(_lang, db, message, bot)
	case "✅" + global.Translations[_lang]["usdt_trx_swap"]:
		service.MenuNavigateSwapExchange(_lang, db, message, bot)
	case "🕸" + global.Translations[_lang]["address_trace_menu"]:
		service.MenuNavigateAddressTrace(_lang, cache, bot, message.Chat.ID, db)
	case "🔍" + global.Translations[_lang]["address_check"]:
		service.MenuNavigateAddressDetection(_lang, cache, bot, message.Chat.ID, db)
	case "🚨" + global.Translations[_lang]["usdt_freeze_alert"]:
		service.MenuNavigateAddressFreeze(_lang, cache, bot, message.Chat.ID, db)
	case "🖊️" + global.Translations[_lang]["transaction_plans"]:
		service.MenuNavigateBundlePackage(_lang, db, message.Chat.ID, bot, "TRX")
	case "🤖" + global.Translations[_lang]["catfee_smart_transaction_menu"]:
		catfee.MenuNavigateCatfeeSmartTransactionPlans(_lang, db, message.Chat.ID, bot, "TRX")
	case "⚡" + global.Translations[_lang]["energy_swap"]:
		service.MenuNavigateEnergyExchange(_lang, db, message, bot)
	case "👤" + global.Translations[_lang]["my_account"]:
		service.MenuNavigateHome(_lang, cache, db, message, bot)
	case "🌍" + global.Translations[_lang]["language"]:
		service.MenuNavigateHome2(db, message, bot)
	default:
		status, _ := cache.Get(strconv.FormatInt(message.Chat.ID, 10))

		log.Printf("用户状态status %s", status)
		switch {
		case strings.HasPrefix(status, "user_backup_notify"):

			if service.ExtractBackup(message, bot, db) {
				return
			}
		case strings.HasPrefix(status, "start_freeze_risk"):
			freezeAlertService := service.NewFreezeAlertService(db)
			preview, previewErr := freezeAlertService.Preview(message.Text)
			if previewErr == service.ErrFreezeAlertInvalidAddress {
				msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[_lang]["address_wrong_tips"]+"</b>"+"\n")
				msg.ParseMode = "HTML"
				bot.Send(msg)
				return
			}
			if previewErr != nil {
				log.Printf("freeze alert preview err: %v", previewErr)
				return
			}

			sendFreezeAlertPreview(bot, message.Chat.ID, _lang, preview)
			expiration := 1 * time.Minute // 短时间缓存空值
			//设置用户状态
			cache.Set(strconv.FormatInt(message.Chat.ID, 10), "start_freeze_risk_status", expiration)

		case strings.HasPrefix(status, "address_list_trace"):

		case strings.HasPrefix(status, "address_manager_remove"):
			if IsValidAddress(message.Text) || IsValidEthereumAddress(message.Text) {
				userRepo := repositories.NewUserAddressMonitorRepo(db)
				err := userRepo.Remove(context.Background(), message.Chat.ID, message.Text)
				if err != nil {
				}
				msg := tgbotapi.NewMessage(message.Chat.ID, "✅ "+"<b>"+global.Translations[_lang]["address_deleted_success"]+"</b>"+"\n")
				msg.ParseMode = "HTML"
				bot.Send(msg)

				service.ADDRESS_MANAGER(_lang, cache, bot, message.Chat.ID, db)

			} else {
				msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[_lang]["invalid_address_tips"]+"</b>"+"\n")
				msg.ParseMode = "HTML"
				bot.Send(msg)
			}
		case strings.HasPrefix(status, "dispatch_others"):
			if IsValidAddress(message.Text) {
				dispatchService := service.NewEnergyDispatchService(db, _trxfeeUrl, _trxfeeApiKey, _trxfeeSecret, catfeeClient)
				result, dispatchErr := dispatchService.DispatchToManualAddress(context.Background(), message.Chat.ID, message.Text)
				if dispatchErr == nil {
					sendDispatchSuccess(bot, message.Chat.ID, result)
					resetDispatchState(cache, message.Chat.ID)
				} else if dispatchErr == service.ErrDispatchInsufficientTimes {
					service.MenuNavigateBundlePackage(_lang, db, message.Chat.ID, bot, "TRX")
				}

			} else {
				msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[_lang]["invalid_address_tips"]+"</b>"+"\n")
				msg.ParseMode = "HTML"
				bot.Send(msg)
			}
		case strings.HasPrefix(status, "DISPATCHOTHERS_"):
			if IsValidAddress(message.Text) {
				subscribeBundleID := strings.ReplaceAll(status, "DISPATCHOTHERS_", "")
				dispatchService := service.NewEnergyDispatchService(db, _trxfeeUrl, _trxfeeApiKey, _trxfeeSecret, catfeeClient)
				result, dispatchErr := dispatchService.DispatchFromSubscription(context.Background(), subscribeBundleID, message.Text, message.Chat.ID)
				if dispatchErr == nil {
					msg2 := service.CLICK_BUNDLE_PACKAGE_ADDRESS_STATS(_lang, db, message.Chat.ID)
					bot.Send(msg2)
					sendDispatchSuccess(bot, message.Chat.ID, result)
					resetDispatchState(cache, message.Chat.ID)
				}

			} else {
				msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[_lang]["invalid_address_tips"]+"</b>"+"\n")
				msg.ParseMode = "HTML"
				bot.Send(msg)
			}
		case strings.HasPrefix(status, "address_manager_add"):
			service.ExtractAddressManager(_lang, message, db, bot)

			service.ADDRESS_MANAGER(_lang, cache, bot, message.Chat.ID, db)

		case strings.HasPrefix(status, "bundle_"):
			fmt.Printf(">>>>>>>>>>>>>>>>>>>>bundle: %s", status)

			if service.ExtractBundleService(_lang, message, bot, db, status) {
				return
			}

		case strings.HasPrefix(status, "address_trace_add"):
			if !IsValidAddress(message.Text) {
				msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[_lang]["address_wrong_tips"]+"</b>"+"\n")
				msg.ParseMode = "HTML"
				bot.Send(msg)
				return
			}

			userRepo := repositories.NewUserAddressTraceRepo(db)

			model, err := userRepo.Find(context.Background(), message.Chat.ID, message.Text)

			if err != nil {
			}
			if model.Id > 0 {
				msg := tgbotapi.NewMessage(message.Chat.ID, "✅"+"<b>"+global.Translations[_lang]["address_trace_add_repeat_tips"]+"</b>"+"\n")
				inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(

					tgbotapi.NewInlineKeyboardRow(
						//tgbotapi.NewInlineKeyboardButtonData("解绑地址", "free_monitor_address"),
						tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[_lang]["back_homepage"], "back_user_address_trace"),
					),
				)
				msg.ReplyMarkup = inlineKeyboard
				msg.ParseMode = "HTML"
				bot.Send(msg)

				return
			}

			total, err := userRepo.Count(context.Background(), message.Chat.ID)
			if err != nil {
			}
			if total >= 4 {
				msg := tgbotapi.NewMessage(message.Chat.ID, "✅"+"<b>"+global.Translations[_lang]["address_trace_add_max_tips"]+"</b>"+"\n")
				inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(

					tgbotapi.NewInlineKeyboardRow(
						//tgbotapi.NewInlineKeyboardButtonData("解绑地址", "free_monitor_address"),
						tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[_lang]["back_homepage"], "back_user_address_trace"),
					),
				)
				msg.ReplyMarkup = inlineKeyboard
				msg.ParseMode = "HTML"
				bot.Send(msg)

				return
			}
			var record domain.UserAddressTrace
			record.ChatID = message.Chat.ID
			record.Address = message.Text
			record.Status = 1
			if IsValidAddress(message.Text) {
				record.Network = "tron"
			}
			if IsValidAddress(message.Text) {
				record.Network = "ethereum"
			}
			errsg := userRepo.Create(context.Background(), &record)
			if errsg != nil {
			}

			msg := tgbotapi.NewMessage(message.Chat.ID, "✅"+"<b>"+global.Translations[_lang]["address_added_success"]+"</b>"+"\n")
			inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(

				tgbotapi.NewInlineKeyboardRow(
					//tgbotapi.NewInlineKeyboardButtonData("解绑地址", "free_monitor_address"),
					tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[_lang]["back_homepage"], "back_user_address_trace"),
				),
			)
			msg.ReplyMarkup = inlineKeyboard
			msg.ParseMode = "HTML"
			bot.Send(msg)

		case strings.HasPrefix(status, "address_trace_delete"):
			if !IsValidAddress(message.Text) {
				msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[_lang]["address_wrong_tips"]+"</b>"+"\n")
				msg.ParseMode = "HTML"
				bot.Send(msg)
				return
			}
			userRepo := repositories.NewUserAddressTraceRepo(db)
			err := userRepo.Remove(context.Background(), message.Chat.ID, message.Text)
			if err != nil {
				return
			}
			msg := tgbotapi.NewMessage(message.Chat.ID, "✅ "+"<b>"+global.Translations[_lang]["address_deleted_success"]+"</b>"+"\n")
			inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(

				tgbotapi.NewInlineKeyboardRow(
					//tgbotapi.NewInlineKeyboardButtonData("解绑地址", "free_monitor_address"),
					tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[_lang]["back_homepage"], "back_user_address_trace"),
				),
			)
			msg.ReplyMarkup = inlineKeyboard

			msg.ParseMode = "HTML"
			bot.Send(msg)

		case strings.HasPrefix(status, "usdt_risk_monitor"):
			//fmt.Printf("bundle: %s", status)

			if !IsValidAddress(message.Text) {
				msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[_lang]["invalid_address_tips"]+"</b>"+"\n")
				msg.ParseMode = "HTML"
				bot.Send(msg)
			}

			msg := tgbotapi.NewMessage(message.Chat.ID, "")

			//msg.ReplyMarkup = inlineKeyboard
			msg.ParseMode = "HTML"

			bot.Send(msg)

		case strings.HasPrefix(status, "click_bundle_package_address_manager_remove"):
			if service.CLICK_BUNDLE_PACKAGE_ADDRESS_MANAGER_REMOVE(_lang, cache, bot, message, db) {
				return
			}

		case strings.HasPrefix(status, "click_bundle_package_address_manager_add"):
			if service.CLICK_BUNDLE_PACKAGE_ADDRESS_MANAGER_ADD(_lang, cache, bot, message, db) {
				return
			}

		case strings.HasPrefix(status, "apply_bundle_package_"):
			if service.APPLY_BUNDLE_PACKAGE(_lang, cache, bot, message, db, status) {
				return
			}
		case strings.HasPrefix(status, "apply_ST_bundle_package_"):
			trxfeeClient := trxfee.NewTrxfeeClient(_trxfeeUrl, _trxfeeApiKey, _trxfeeSecret)

			if service.APPLY_ST_BUNDLE_PACKAGE(trxfeeClient, _lang, cache, bot, message, db, status) {
				return
			}

		case strings.HasPrefix(status, "click_backup_account"):

			log.Printf("进入click_backup_account状态：%s\n", message.Text)
			if strings.Contains(message.Text, "@") {
				msg := tgbotapi.NewMessage(message.Chat.ID, "❌ "+global.Translations[_lang]["backup_account_tips"])
				msg.ParseMode = "HTML"
				bot.Send(msg)
				return
			}
			userName := strings.ReplaceAll(message.Text, "@", "")

			log.Printf("备份用户：%s\n", userName)
			userRepo := repositories.NewUserRepository(db)
			user, err := userRepo.GetByUsername(userName)

			if err != nil {
				log.Printf("访问失败 %s\n", err)
				msg := tgbotapi.NewMessage(message.Chat.ID, "❌"+global.Translations[_lang]["backup_account_tips2"])
				msg.ParseMode = "HTML"
				bot.Send(msg)
				return
			}

			if user.Id == 0 {
				log.Printf("无该用户 %s\n", userName)
				msg := tgbotapi.NewMessage(message.Chat.ID, "❌"+global.Translations[_lang]["backup_account_tips2"])
				msg.ParseMode = "HTML"
				bot.Send(msg)
				return
			}

			user.BackupChatID = userName

			err2 := userRepo.UpdateBackupChat(context.Background(), userName, message.Chat.ID)
			if err2 == nil {
				msg := tgbotapi.NewMessage(message.Chat.ID, "✅ "+global.Translations[_lang]["backup_account_tips3"]+message.Text)
				msg.ParseMode = "HTML"
				bot.Send(msg)
				//return true
			}

			service.MenuNavigateHome(_lang, cache, db, message, bot)

		case strings.HasPrefix(status, "usdt_risk_query"):
			//fmt.Printf("bundle: %s", status)
			service.ExtractSlowMistRiskQuery(_lang, cache, message, db, _cookie, bot)

		case strings.HasPrefix(status, "catfee_add_address"):
			trxfeeClient := trxfee.NewTrxfeeClient(_trxfeeUrl, _trxfeeApiKey, _trxfeeSecret)
			catfee.CustodyAddressAdd(_lang, cache, db, bot, message, trxfeeClient)

		case strings.HasPrefix(status, "catfee_remove_address"):
			catfee.CustodyAddressRemove(_lang, cache, db, bot, message, catfeeClient)

		case strings.HasPrefix(status, "premium_user_rent_month"):
			username := message.Text
			if len(username) < 4 || !strings.Contains(username, "@") {
				return
			}

			_month := strings.ReplaceAll(status, "premium_user_rent_month", "")

			fmt.Printf("message text: %s\n", message.Text)

			member.Rent(_lang, cache, db, bot, username, message.Chat.ID, _month)

		case strings.HasPrefix(status, "purchase_telegram_stars"):
			username := message.Text
			if len(username) < 4 || !strings.Contains(username, "@") {
				return
			}
			count := strings.ReplaceAll(status, "purchase_telegram_stars", "")

			fmt.Printf("message text: %s\n", message.Text)
			member.Purchase(_lang, cache, db, bot, message.Text, message.Chat.ID, count)

		case strings.HasPrefix(status, "click_laundering_"):

			content := strings.ReplaceAll(status, "click_laundering_", "")
			fmt.Printf("内容: %s\n", content)
			fmt.Printf("输入内容: %s\n", message.Text)
			token := strings.Split(content, "_")[0]
			amount := strings.Split(content, "_")[1]
			fmt.Printf("状态 代币: %s - 金额: %s\n", token, amount)

			if strings.ToUpper(token) != "BTC" && !IsValidEthereumAddress(message.Text) {
				msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[_lang]["address_wrong_tips"]+"</b>"+"\n")
				msg.ParseMode = "HTML"
				bot.Send(msg)
				return
			}

			valid := IsValidBitcoinAddress(message.Text)
			fmt.Printf("valid: %v\n", valid)
			if strings.ToUpper(token) == "BTC" && !IsValidBitcoinAddress(message.Text) {
				msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[_lang]["address_wrong_tips"]+"</b>"+"\n")
				msg.ParseMode = "HTML"
				bot.Send(msg)
				return
			}

			//fixedfloat 生成订单

			api := fixedfloat.New("AZxSXXl6VwqSgJkkC6HovFyxWib0ZPUNVBOO8Fkt", "vOGKrbXgFBepGBEze90EUUHHnsLzjQHC8197WtRC")

			params := map[string]interface{}{
				"fromCcy": "USDTTRC",
				//"toCcy":     "USDT",
				"type": fixedfloat.TypeFloat,
				//"amount":    1000,
				"direction": "from",
				//"toAddress": "0xF510e53EF8DA4e45FFA59EB554511a7410E5eFD3",
				"refcode": "r8ck81xa",
				"ref":     "r8ck81xa",
				"afftax":  1,
			}

			params["toCcy"] = token
			params["amount"] = amount
			params["toAddress"] = message.Text
			params["refcode"] = "r8ck81xa"
			params["ref"] = "r8ck81xa"
			params["afftax"] = 1
			rawMap, err := api.Create(params)

			if err != nil {
				return
			}

			from, to, ok := fixedfloat.ExtractFromAndTo(rawMap)
			if !ok {
				fmt.Println("Failed to extract from/to")
				return
			}

			fmt.Println("From Address:", from.Address)
			fmt.Println("From Amount:", *from.Amount)
			fmt.Println("To Address:", to.Address)
			fmt.Println("To Amount:", *to.Amount)

			timeInfo, ok := fixedfloat.ExtractTime(rawMap)
			if !ok {
				fmt.Println("Failed to extract time")
				return
			}

			fmt.Printf("Reg (Unix): %.0f\n", timeInfo.Reg)
			fmt.Printf("Expiration (Unix): %.0f\n", timeInfo.Expiration)
			fmt.Printf("Left: %.0f seconds\n", timeInfo.Left)

			// 转为 time.Time（可读时间）
			regTime := time.Unix(int64(timeInfo.Reg), 0)
			expTime := time.Unix(int64(timeInfo.Expiration), 0)
			fmt.Println("Reg Time:", regTime.UTC().Format("2006-01-02 15:04:05 UTC"))
			fmt.Println("Expire Time:", expTime.UTC().Format("2006-01-02 15:04:05 UTC"))

			id, status, ok := fixedfloat.ExtractIDAndStatus(rawMap)
			if !ok {
				fmt.Println("Failed to extract id or status")
				return
			}

			fmt.Printf("ID: %s\n", id)
			fmt.Printf("Status: %s\n", status)

			desc := global.Translations[_lang]["coin_laundering_order_desc"]

			if token == "BSC" {
				token = "BNB"
			}

			if token == "USDT" {
				token = "USDTERC20"
			}

			desc = strings.ReplaceAll(desc, "{RegTime}", regTime.UTC().Format("2006-01-02 15:04:05 UTC"))
			desc = strings.ReplaceAll(desc, "{ExpireTime}", expTime.UTC().Format("2006-01-02 15:04:05 UTC"))
			desc = strings.ReplaceAll(desc, "{token}", token)
			desc = strings.ReplaceAll(desc, "{amount1}", amount)
			desc = strings.ReplaceAll(desc, "{amount2}", *to.Amount)
			desc = strings.ReplaceAll(desc, "{from_address}", from.Address)
			desc = strings.ReplaceAll(desc, "{to_address}", to.Address)
			desc = strings.ReplaceAll(desc, "{orderNO}", id)

			fromAddress := from.Address
			size := 300

			filename, err := fixedfloat.GenerateQRCodeWithTimestamp(fromAddress, size)
			fmt.Printf("filename: %s\n", filename)

			//videoPath := "/Users/masion/Documents/GitHub/multi-lang-mist-bot/" + filename

			videoPath := "/root/ushield-telegram-bot/old/" + filename

			// 创建视频消息（从本地文件）
			msg := tgbotapi.NewPhoto(message.Chat.ID, tgbotapi.FilePath(videoPath))

			msg.Caption = "✅ " + desc

			//msg := tgbotapi.NewMessage(message.Chat.ID, "✅ "+desc)
			msg.ParseMode = "HTML"
			bot.Send(msg)

			orderDB := repositories.NewCoinLaunderingOrderRepo(db)

			var order domain.CoinLaunderingOrder
			order.OrderNO = id
			order.Amount = amount
			order.Token = token
			order.FromAddress = fromAddress
			order.ToAddress = to.Address
			order.ChatID = message.Chat.ID
			order.Status = 0
			order.CreatedAt = time.Now()

			orderDB.Create(context.Background(), &order)

		}
	}
}

// 处理内联键盘回调
func handleCallbackQuery(cache cache.Cache, bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *gorm.DB, _trxfeeUrl, _trxfeeApiKey, _trxfeeSecret string, catfeeClient *trxfee.CatfeeService) {
	// 先应答回调

	//log.Println("已选择: " + callbackQuery.Data)
	//callback := tgbotapi.NewCallback(callbackQuery.ID, "已选择: "+callbackQuery.Data)
	//if _, err := bot.Request(callback); err != nil {
	//	log.Printf("Error answering callback: %v", err)
	//}
	_lang, err := cache.Get("LANG_" + strconv.FormatInt(callbackQuery.Message.Chat.ID, 10))

	if err != nil {
		_lang = "zh"
	}
	// 根据回调数据执行不同操作
	var responseText string
	switch {

	//case "🖊️" + global.Translations[_lang]["transaction_plans"]:
	//	service.MenuNavigateBundlePackage(_lang, db, message.Chat.ID, bot, "TRX")
	//
	//case "🤖" + global.Translations[_lang]["smart_transaction_plans"]:
	//	service.MenuNavigateSmartTransactionPlans(_lang, db, message.Chat.ID, bot, "TRX")
	//
	//case "⚡" + global.Translations[_lang]["energy_swap"]:
	//	service.MenuNavigateEnergyExchange(_lang, db, message, bot)
	case callbackQuery.Data == "click_energy_swap":
		service.MenuNavigateEnergyExchange(_lang, db, callbackQuery.Message, bot)
	case callbackQuery.Data == "click_transaction_plan":
		service.MenuNavigateBundlePackage(_lang, db, callbackQuery.Message.Chat.ID, bot, "TRX")
	case callbackQuery.Data == "click_smart_transaction_plan":
		//service.MenuNavigateSmartTransactionPlans(_lang, db, callbackQuery.Message.Chat.ID, bot, "TRX")
		catfee.MenuNavigateCatfeeSmartTransactionPlans(_lang, db, callbackQuery.Message.Chat.ID, bot, "TRX")

	case callbackQuery.Data == "click_language":
		service.MenuNavigateHome2(db, callbackQuery.Message, bot)
	case callbackQuery.Data == "dispatch_Now_Others":
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[_lang]["enter_address"]+"\n\n")
		msg.ParseMode = "HTML"

		var allButtons []tgbotapi.InlineKeyboardButton
		var extraButtons []tgbotapi.InlineKeyboardButton
		var keyboard [][]tgbotapi.InlineKeyboardButton
		//for _, item := range addresses {
		//	allButtons = append(allButtons, tgbotapi.NewInlineKeyboardButtonData(TruncateString(item.Address), "dispatch_others_"+bundleID+"_"+item.Address))
		//}

		extraButtons = append(extraButtons, tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[_lang]["back_homepage"], "back_bundle_package"))

		for i := 0; i < len(allButtons); i += 1 {
			end := i + 1
			if end > len(allButtons) {
				end = len(allButtons)
			}
			row := allButtons[i:end]
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(row...))
		}

		for i := 0; i < len(extraButtons); i += 1 {
			end := i + 1
			if end > len(extraButtons) {
				end = len(extraButtons)
			}
			row := extraButtons[i:end]
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(row...))
		}

		// 3. 创建键盘标记
		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(keyboard...)

		msg.ReplyMarkup = inlineKeyboard

		bot.Send(msg)

		expiration := 1 * time.Minute // 短时间缓存空值

		//设置用户状态
		cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), "dispatch_others", expiration)

	case callbackQuery.Data == "back_address_detection_home":

		service.MenuNavigateAddressDetection(_lang, cache, bot, callbackQuery.Message.Chat.ID, db)

	case strings.HasPrefix(callbackQuery.Data, "dispatch_others_"):
		bundleAddress := strings.ReplaceAll(callbackQuery.Data, "dispatch_others_", "")

		bundleID := strings.Split(bundleAddress, "_")[0]
		address := strings.Split(bundleAddress, "_")[1]

		fmt.Printf("bundleID %s\n", bundleID)
		fmt.Printf("address %s\n", address)

		dispatchService := service.NewEnergyDispatchService(db, _trxfeeUrl, _trxfeeApiKey, _trxfeeSecret, catfeeClient)
		result, dispatchErr := dispatchService.DispatchFromSubscription(context.Background(), bundleID, address, callbackQuery.Message.Chat.ID)
		if dispatchErr == nil {
			msg2 := service.CLICK_BUNDLE_PACKAGE_ADDRESS_STATS(_lang, db, callbackQuery.Message.Chat.ID)
			bot.Send(msg2)
			sendDispatchSuccess(bot, callbackQuery.Message.Chat.ID, result)
			resetDispatchState(cache, callbackQuery.Message.Chat.ID)
		}

		//msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "📢【✅"+global.Translations[_lang]["UShield_sent_transaction_energy"]+"】\n\n"+
		//	global.Translations[_lang]["to_address"]+address+"\n\n"+
		//	global.Translations[_lang]["remaining_transactions"]+strconv.FormatInt(restTimes, 10)+"\n\n")
		//msg.ParseMode = "HTML"
		//bot.Send(msg)

	case strings.HasPrefix(callbackQuery.Data, "confirm_freeze_risk_"):
		address := strings.ReplaceAll(callbackQuery.Data, "confirm_freeze_risk_", "")
		freezeAlertService := service.NewFreezeAlertService(db)
		result, confirmErr := freezeAlertService.Confirm(context.Background(), callbackQuery.Message.Chat.ID, address)
		if confirmErr == service.ErrFreezeAlertInsufficientBalance {
			sendFreezeAlertInsufficientBalance(bot, callbackQuery.Message.Chat.ID, _lang)
			return
		}
		if confirmErr != nil {
			log.Printf("freeze alert confirm err: %v", confirmErr)
			return
		}
		sendFreezeAlertEnableSuccess(bot, callbackQuery.Message.Chat.ID, _lang, result)

	case strings.HasPrefix(callbackQuery.Data, "set_bundle_package_default_"):
		target := strings.ReplaceAll(callbackQuery.Data, "set_bundle_package_default_", "")
		userOperationPackageAddressesRepo := repositories.NewUserOperationPackageAddressesRepo(db)

		errsg := userOperationPackageAddressesRepo.Update(context.Background(), callbackQuery.Message.Chat.ID, target)
		if errsg != nil {
			log.Printf("errsg: %s", errsg)
			return
		}
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "✅"+"<b>"+"设置默认地址成功 "+"</b>"+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)
		service.CLICK_BUNDLE_PACKAGE_ADDRESS_MANAGEMENT(_lang, cache, bot, callbackQuery.Message.Chat.ID, db)

	case strings.HasPrefix(callbackQuery.Data, "remove_bundle_package_"):
		target := strings.ReplaceAll(callbackQuery.Data, "remove_bundle_package_", "")
		userOperationPackageAddressesRepo := repositories.NewUserOperationPackageAddressesRepo(db)

		var record domain.UserOperationPackageAddresses
		record.Status = 0
		record.Address = target
		record.ChatID = callbackQuery.Message.Chat.ID

		errsg := userOperationPackageAddressesRepo.Remove(context.Background(), callbackQuery.Message.Chat.ID, target)
		if errsg != nil {
			log.Printf("errsg: %s", errsg)
			return
		}
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "✅"+"<b>"+global.Translations[_lang]["address_deleted_success"]+"</b>"+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)
		//service.CLICK_BUNDLE_PACKAGE_ADDRESS_MANAGEMENT(cache, bot, callbackQuery.Message.Chat.ID, db)
		msg2 := service.CLICK_BUNDLE_PACKAGE_ADDRESS_STATS2(_lang, db, callbackQuery.Message.Chat.ID)
		bot.Send(msg2)
	case strings.HasPrefix(callbackQuery.Data, "close_freeze_risk_"):
		target := strings.ReplaceAll(callbackQuery.Data, "close_freeze_risk_", "")
		freezeAlertService := service.NewFreezeAlertService(db)
		preview, previewErr := freezeAlertService.PreviewClose(callbackQuery.Message.Chat.ID, target)
		if previewErr != nil {
			log.Printf("freeze alert close preview err: %v", previewErr)
			return
		}
		sendFreezeAlertClosePreview(bot, callbackQuery.Message.Chat.ID, _lang, preview)

	case strings.HasPrefix(callbackQuery.Data, "close_risk_"):
		target := strings.ReplaceAll(callbackQuery.Data, "close_risk_", "")
		freezeAlertService := service.NewFreezeAlertService(db)
		if err := freezeAlertService.Close(context.Background(), callbackQuery.Message.Chat.ID, target); err != nil {
			log.Printf("freeze alert close err: %v", err)
			return
		}
		sendFreezeAlertCloseSuccess(bot, callbackQuery.Message.Chat.ID, _lang)
	case strings.HasPrefix(callbackQuery.Data, "apply_bundle_package_"):

		target := strings.ReplaceAll(callbackQuery.Data, "apply_bundle_package_", "")
		service.APPLY_BUNDLE_PACKAGE_ADDRESS(_lang, target, cache, bot, callbackQuery.Message, db)

	case strings.HasPrefix(callbackQuery.Data, "config_bundle_package_address_"):

		target := strings.ReplaceAll(callbackQuery.Data, "config_bundle_package_address_", "")
		service.CONFIG_BUNDLE_PACKAGE_ADDRESS(_lang, target, cache, bot, callbackQuery.Message, db)
	case callbackQuery.Data == "click_backup_account":

		//msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "👥欢迎使用第二通知人服务"+"\n"+
		//	"为确保实时接收预警信息，您可绑定一个第二通知人TG帐号。"+"\n"+
		//	"绑定前请确保第二通知人已与本机器人互动，绑定后该账号将同步接收预警信息，第二通知人替换请重复绑定步骤，系统将自动替换。请输入的第二通知人TG帐号@用户名 👇")
		//
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[_lang]["secondary_contact_tips"])
		msg.ParseMode = "HTML"

		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[_lang]["back_homepage"], "back_home"),
				//tgbotapi.NewInlineKeyboardButtonData("第二紧急通知", ""),
			),
		)
		msg.ReplyMarkup = inlineKeyboard

		bot.Send(msg)

		expiration := 1 * time.Minute // 短时间缓存空值

		//设置用户状态
		cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), "click_backup_account", expiration)

	case callbackQuery.Data == "back_user_address_trace":
		service.MenuNavigateAddressTrace(_lang, cache, bot, callbackQuery.Message.Chat.ID, db)
	case callbackQuery.Data == "back_risk_home":
		service.MenuNavigateAddressFreeze(_lang, cache, bot, callbackQuery.Message.Chat.ID, db)
	case callbackQuery.Data == "click_switch_trx":
		service.MenuNavigateBundlePackage(_lang, db, callbackQuery.Message.Chat.ID, bot, "TRX")
	case callbackQuery.Data == "click_switch_usdt":
		service.MenuNavigateBundlePackage(_lang, db, callbackQuery.Message.Chat.ID, bot, "USDT")

	case callbackQuery.Data == "click_switch_trx_ST":
		service.MenuNavigateSTBundlePackage(_lang, db, callbackQuery.Message.Chat.ID, bot, "TRX")
	case callbackQuery.Data == "click_switch_usdt_ST":
		service.MenuNavigateSTBundlePackage(_lang, db, callbackQuery.Message.Chat.ID, bot, "USDT")

	case callbackQuery.Data == "back_bundle_package_ST":
		service.MenuNavigateSTBundlePackage(_lang, db, callbackQuery.Message.Chat.ID, bot, "TRX")

	case callbackQuery.Data == "back_bundle_package":
		service.MenuNavigateBundlePackage(_lang, db, callbackQuery.Message.Chat.ID, bot, "TRX")
	case callbackQuery.Data == "click_bundle_package_address_manager_config":
		service.CLICK_BUNDLE_PACKAGE_ADDRESS_MANAGER_CONFIG(_lang, cache, bot, callbackQuery.Message.Chat.ID, db)
	case callbackQuery.Data == "click_bundle_package_address_manager_remove":
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[_lang]["energy_address_remove_tips"]+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)

		expiration := 1 * time.Minute // 短时间缓存空值

		//设置用户状态
		cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), callbackQuery.Data, expiration)

	case callbackQuery.Data == "click_bundle_package_address_manager_add":

		userOperationPackageAddressesRepo := repositories.NewUserOperationPackageAddressesRepo(db)

		list, _ := userOperationPackageAddressesRepo.Query(context.Background(), callbackQuery.Message.Chat.ID)
		if len(list) >= 4 {
			msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "<b>"+global.Translations[_lang]["energy_address_limit_tips"]+"</b>"+"\n")
			msg.ParseMode = "HTML"
			bot.Send(msg)
			return
		}
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "<b>"+global.Translations[_lang]["energy_address_limit"]+"</b>"+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)

		expiration := 1 * time.Minute // 短时间缓存空值

		//设置用户状态
		cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), callbackQuery.Data, expiration)
		//笔数套餐地址列表
	case callbackQuery.Data == "click_bundle_package_address_stats":
		msg := service.CLICK_BUNDLE_PACKAGE_ADDRESS_STATS2(_lang, db, callbackQuery.Message.Chat.ID)
		bot.Send(msg)

	case callbackQuery.Data == "click_bundle_package_address_stats_ST":
		//msg := service.CLICK_BUNDLE_PACKAGE_ADDRESS_STATS_ST(_lang, db, callbackQuery.Message.Chat.ID)
		//bot.Send(msg)
		catfee.CLICK_BUNDLE_PACKAGE_ADDRESS_STATS_ST(_lang, cache, db, callbackQuery.Message.Chat.ID, bot)

	case strings.HasPrefix(callbackQuery.Data, "custody_address_check_"):
		{
			catfee.CheckOption(_lang, db, callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID, callbackQuery.Data, bot, catfeeClient)
		}
	case callbackQuery.Data == "next_bundle_package_address_stats":
		if service.NEXT_BUNDLE_PACKAGE_ADDRESS_STATS(_lang, callbackQuery, db, bot) {
			return
		}
	case callbackQuery.Data == "prev_bundle_package_address_stats":
		state, done := service.PREV_BUNDLE_PACKAGE_ADDRESS_STATS(_lang, callbackQuery, db, bot)
		if done {
			return
		}
		fmt.Printf("state: %v\n", state)

	case callbackQuery.Data == "click_bundle_package_address_management":
		service.CLICK_BUNDLE_PACKAGE_ADDRESS_MANAGEMENT(_lang, cache, bot, callbackQuery.Message.Chat.ID, db)
	case callbackQuery.Data == "address_list_trace":
		service.ADDRESS_LIST_TRACE(_lang, cache, bot, callbackQuery, db)
	case callbackQuery.Data == "back_home":
		service.MenuNavigateHome(_lang, cache, db, callbackQuery.Message, bot)
	case callbackQuery.Data == "click_business_cooperation":
		service.ClickBusinessCooperation(_lang, callbackQuery, bot)
	case callbackQuery.Data == "click_offical_channel":
		service.ClickOfficalChannel(_lang, callbackQuery, bot)
	case callbackQuery.Data == "click_callcenter":
		service.ClickCallCenter(_lang, callbackQuery, bot)
	case callbackQuery.Data == "click_my_recepit":
		service.CLICK_MY_RECEPIT(_lang, db, callbackQuery, bot)
	case callbackQuery.Data == "address_freeze_risk_records":
		msg := service.ExtractAddressRiskQuery(_lang, db, callbackQuery)
		bot.Send(msg)
	case callbackQuery.Data == "user_detection_cost_records":
		msg := service.ExtractAddressDetection(_lang, cache, db, callbackQuery)
		bot.Send(msg)
	case callbackQuery.Data == "click_bundle_package_cost_records":
		msg := service.ExtractBundlePackage(_lang, db, callbackQuery)
		bot.Send(msg)

	case callbackQuery.Data == "click_bundle_package_cost_records_ST":
		//msg := service.ExtractBundlePackageST(_lang, db, callbackQuery)
		//bot.Send(msg)
		msg := catfee.ExtractBundlePackageST(_lang, db, callbackQuery)
		bot.Send(msg)

	case callbackQuery.Data == "prev_st_bundle_package_page":
		state, done := catfee.EXTRACT_PREV_BUNDLE_PACKAGE_PAGE(_lang, callbackQuery, db, bot)
		if done {
			return
		}
		fmt.Printf("state: %v\n", state)

	case callbackQuery.Data == "next_st_bundle_package_page":
		if catfee.EXTRACT_NEXT_BUNDLE_PACKAGE_PAGE(_lang, callbackQuery, db, bot) {
			return
		}

	case callbackQuery.Data == "catfee_add_address":
		catfee.CustodyAddressCond(_lang, cache, db, bot, callbackQuery)

	case callbackQuery.Data == "catfee_remove_address":
		catfee.CustodyRemoveAddressCond(_lang, cache, db, bot, callbackQuery)

	case callbackQuery.Data == "click_bundle_package_management":
		msg := service.ExtractBundlePackage(_lang, db, callbackQuery)
		bot.Send(msg)
	case callbackQuery.Data == "click_deposit_usdt_records":
		service.CLICK_DEPOSIT_USDT_RECORDS(_lang, db, callbackQuery, bot)
	case callbackQuery.Data == "click_deposit_trx_records":
		service.CLICK_DEPOSIT_TRX_RECORDS(_lang, db, callbackQuery, bot)
	case callbackQuery.Data == "next_address_detection_page":
		if service.EXTRACT_NEXT_ADDRESS_DETECTION_PAGE(_lang, callbackQuery, db, bot) {
			return
		}
	case callbackQuery.Data == "prev_address_detection_page":
		state, done := service.EXTRACT_PREV_ADDRESS_DETECTION_PAGE(_lang, callbackQuery, db, bot)
		if done {
			return
		}
		fmt.Printf("state: %v\n", state)
	case callbackQuery.Data == "prev_deposit_usdt_page":
		state, done := service.EXTRACT_PREV_DEPOSIT_USDT_PAGE(_lang, callbackQuery, db, bot)
		if done {
			return
		}
		fmt.Printf("state: %v\n", state)
	case callbackQuery.Data == "prev_deposit_trx_page":
		state, done := service.EXTRACT_PREV_DEPOSIT_TRX_PAGE(_lang, callbackQuery, db, bot)
		if done {
			return
		}
		fmt.Printf("state: %v\n", state)
	case callbackQuery.Data == "prev_address_risk_page":
		state, done := service.EXTRACT_PREV_ADDRESS_RISK_PAGE(_lang, callbackQuery, db, bot)
		if done {
			return
		}
		fmt.Printf("state: %v\n", state)

	case callbackQuery.Data == "next_address_risk_page":
		if service.ExtraNextAddressRiskPage(_lang, callbackQuery, db, bot) {
			return
		}
	case callbackQuery.Data == "next_deposit_usdt_page":
		if service.ExtraNextDepositUSDTPage(_lang, callbackQuery, db, bot) {
			return
		}
	case callbackQuery.Data == "next_deposit_trx_page":
		if service.ExtracNextDepositTrxPage(_lang, callbackQuery, db, bot) {
			return
		}

	case callbackQuery.Data == "prev_bundle_package_page":
		state, done := service.EXTRACT_PREV_BUNDLE_PACKAGE_PAGE(_lang, callbackQuery, db, bot)
		if done {
			return
		}
		fmt.Printf("state: %v\n", state)

	case callbackQuery.Data == "next_bundle_package_page":
		if service.EXTRACT_NEXT_BUNDLE_PACKAGE_PAGE(_lang, callbackQuery, db, bot) {
			return
		}

	case callbackQuery.Data == "click_QA":
		service.ExtraQA(_lang, cache, bot, callbackQuery)

	case callbackQuery.Data == "user_backup_notify":
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "💬"+"<b>"+"请输入需添加的第二紧急通知用户电报ID: "+"</b>"+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)

		expiration := 1 * time.Minute // 短时间缓存空值

		//设置用户状态
		cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), callbackQuery.Data, expiration)
	case callbackQuery.Data == "start_freeze_risk_1":
		//查看余额
		service.START_FREEZE_RISK_1(_lang, cache, db, callbackQuery, bot)

	case callbackQuery.Data == "click_my_service":
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "🛡 当前服务状态：\n\n🔋 能量闪兑\n\n- 剩余笔数：12\n- 自动补能：关闭 /开启\n\n➡️ /闪兑\n\n➡️ /笔数套餐\n\n➡️ /手动发能（1笔）\n\n➡️ /开启/关闭自动发能\n\n📍 地址风险检测\n\n- 今日免费次数：已用完\n\n➡️ /地址风险检测\n\n🚨 USDT冻结预警\n\n- 地址1：TX8kY...5a9rP（剩余12天）✅\n- 地址2：TEw9Q...iS6Ht（剩余28天）✅")
		msg.ParseMode = "HTML"

		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("👁️‍🗨️ "+global.Translations[_lang]["alert_monitoring_list"], "address_list_trace"),
				//	tgbotapi.NewInlineKeyboardButtonData("地址管理", "address_manager"),
			),
		)
		msg.ReplyMarkup = inlineKeyboard

		bot.Send(msg)

		expiration := 1 * time.Minute // 短时间缓存空值

		//设置用户状态
		cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), "usdt_risk_monitor", expiration)

	case callbackQuery.Data == "stop_freeze_risk_1":

		//删除event表里面
		userAddressEventRepo := repositories.NewUserAddressMonitorEventRepo(db)

		userAddressEventRepo.RemoveAll(context.Background(), callbackQuery.Message.Chat.ID)

		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "已经暂停所有监控")
		msg.ParseMode = "HTML"

		bot.Send(msg)

		expiration := 1 * time.Minute // 短时间缓存空值

		//设置用户状态
		cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), "reset", expiration)

	case callbackQuery.Data == "start_freeze_risk_0":
		service.MenuNavigateAddressFreeze(_lang, cache, bot, callbackQuery.Message.Chat.ID, db)
	case callbackQuery.Data == "stop_freeze_risk":
		freezeAlertService := service.NewFreezeAlertService(db)
		items, listErr := freezeAlertService.ListActive(callbackQuery.Message.Chat.ID)
		if listErr != nil {
			log.Printf("freeze alert list err: %v", listErr)
			return
		}
		sendFreezeAlertStopList(bot, callbackQuery.Message.Chat.ID, _lang, items)

	case callbackQuery.Data == "start_freeze_risk":
		freezeAlertService := service.NewFreezeAlertService(db)
		if err := freezeAlertService.Start(callbackQuery.Message.Chat.ID); err == service.ErrFreezeAlertInsufficientBalance {
			sendFreezeAlertInsufficientBalance(bot, callbackQuery.Message.Chat.ID, _lang)
			return
		} else if err != nil {
			log.Printf("freeze alert start err: %v", err)
			return
		}

		sendFreezeAlertPromptAddress(bot, callbackQuery.Message.Chat.ID, _lang)
		expiration := 1 * time.Minute // 短时间缓存空值

		//设置用户状态
		cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), "start_freeze_risk", expiration)

	case callbackQuery.Data == "address_manager_return":
		service.MenuNavigateAddressFreeze(_lang, cache, bot, callbackQuery.Message.Chat.ID, db)

	case callbackQuery.Data == "address_manager_add":
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "💬"+"<b>"+"请输入需添加的地址: "+"</b>"+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)

		expiration := 1 * time.Minute // 短时间缓存空值

		//设置用户状态
		cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), callbackQuery.Data, expiration)
	case callbackQuery.Data == "address_manager_remove":
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "💬"+"<b>"+"请输入需删除的地址: "+"</b>"+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)

		expiration := 1 * time.Minute // 短时间缓存空值

		//设置用户状态
		cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), callbackQuery.Data, expiration)
	case callbackQuery.Data == "address_manager":
		service.ADDRESS_MANAGER(_lang, cache, bot, callbackQuery.Message.Chat.ID, db)

	case strings.HasPrefix(callbackQuery.Data, "set_lang_"):
		lang := strings.ReplaceAll(callbackQuery.Data, "set_lang_", "")
		expiration := 24 * time.Hour // 短时间缓存空值
		cache.Set("LANG_"+strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), lang, expiration)
		//数据库设置用户的默认选项语言

		userRepo := repositories.NewUserRepository(db)
		userRepo.UpdateLang(lang, callbackQuery.Message.Chat.ID)

		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[lang]["set_lang"]+"\n")

		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔙"+global.Translations[lang]["back_home"], "back_home"),
			),
		)
		msg.ReplyMarkup = inlineKeyboard
		msg.ParseMode = "HTML"
		bot.Send(msg)

		handleStartCommand(cache, bot, callbackQuery.Message)

	case strings.HasPrefix(callbackQuery.Data, "ST_bundle_"):
		//service.ST_BUNDLE_CHECK(_lang, cache, bot, callbackQuery, db)

		catfee.ST_BUNDLE_CHECK(_lang, cache, bot, callbackQuery, db)

	case strings.HasPrefix(callbackQuery.Data, "bundle_"):
		service.BUNDLE_CHECK2(_lang, cache, bot, callbackQuery, db)
		//调用trxfee接口进行笔数扣款
	case strings.HasPrefix(callbackQuery.Data, "deposit_usdt"):
		service.DepositPrevUSDTOrder(_lang, cache, bot, callbackQuery, db)
		//responseText = "你选择了选项 A"
	case strings.HasPrefix(callbackQuery.Data, "deposit_trx"):
		service.DepositPrevOrder(_lang, cache, bot, callbackQuery, db)
	case callbackQuery.Data == "cancel_order":
		service.DepositCancelOrder(_lang, cache, bot, callbackQuery, db)

	case callbackQuery.Data == "cancel_catfee_order":
		service.DepositCancelOrder(_lang, cache, bot, callbackQuery, db)

	case callbackQuery.Data == "address_trace_add":
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[_lang]["address_trace_add_tips"]+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)

		expiration := 1 * time.Minute // 短时间缓存空值

		//设置用户状态
		cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), callbackQuery.Data, expiration)
	case callbackQuery.Data == "address_trace_delete":
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[_lang]["address_trace_delete_tips"]+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)

		expiration := 1 * time.Minute // 短时间缓存空值

		//设置用户状态
		cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), callbackQuery.Data, expiration)
	case callbackQuery.Data == "deposit_amount":

		service.DEPOSIT_AMOUNT(_lang, db, callbackQuery, bot)

	case callbackQuery.Data == "forward_deposit_usdt":

		fmt.Printf("\nforward_deposit_usdt\n")
		usdtSubscriptionsRepo := repositories.NewUserUsdtSubscriptionsRepository(db)

		usdtlist, err := usdtSubscriptionsRepo.ListAll(context.Background())

		if err != nil {

		}
		var allButtons []tgbotapi.InlineKeyboardButton
		var extraButtons []tgbotapi.InlineKeyboardButton
		var keyboard [][]tgbotapi.InlineKeyboardButton
		for _, usdtRecord := range usdtlist {
			allButtons = append(allButtons, tgbotapi.NewInlineKeyboardButtonData("💰"+usdtRecord.Name, "deposit_usdt_"+usdtRecord.Amount))
		}

		extraButtons = append(extraButtons, tgbotapi.NewInlineKeyboardButtonData("🔁"+global.Translations[_lang]["switch_to_trx_deposit"], "deposit_amount"), tgbotapi.NewInlineKeyboardButtonData("🔙"+global.Translations[_lang]["back_home"], "back_home"))

		for i := 0; i < len(allButtons); i += 2 {
			end := i + 2
			if end > len(allButtons) {
				end = len(allButtons)
			}
			row := allButtons[i:end]
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(row...))
		}

		for i := 0; i < len(extraButtons); i += 1 {
			end := i + 1
			if end > len(extraButtons) {
				end = len(extraButtons)
			}
			row := extraButtons[i:end]
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(row...))
		}

		// 3. 创建键盘标记
		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(keyboard...)

		userRepo := repositories.NewUserRepository(db)

		user, _ := userRepo.GetByUserID(callbackQuery.Message.Chat.ID)
		if IsEmpty(user.Amount) {
			user.Amount = "0"
		}

		if IsEmpty(user.TronAmount) {
			user.TronAmount = "0"
		}

		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID,
			"🆔"+global.Translations[_lang]["user_id"]+": <code>"+user.Associates+"</code>\n"+
				"👤"+global.Translations[_lang]["username"]+": @"+user.Username+"\n"+
				"💰"+global.Translations[_lang]["balance"]+": "+"\n"+
				"- TRX：   "+user.TronAmount+"\n"+
				"-  USDT："+user.Amount)

		msg.ReplyMarkup = inlineKeyboard
		msg.ParseMode = "HTML"

		bot.Send(msg)

	case callbackQuery.Data == "click_visa":

		dictRepo := repositories.NewSysDictionariesRepo(db)
		ushield_additional_services_contact, _ := dictRepo.GetDictionaryDetail("ushield_additional_services_contact")
		ushield_additional_services_wallet, _ := dictRepo.GetDictionaryDetail("ushield_additional_services_wallet")

		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[_lang]["ushield_additional_services_visa_desc"]+"\n"+strings.ReplaceAll(global.Translations[_lang]["ushield_additional_services_contact"], "{ushield_additional_services_contact}", ushield_additional_services_contact)+strings.ReplaceAll(global.Translations[_lang]["ushield_additional_services_wallet"], "{ushield_additional_services_wallet}", ushield_additional_services_wallet))
		msg.ParseMode = "HTML"
		bot.Send(msg)
	case callbackQuery.Data == "click_sim":
		dictRepo := repositories.NewSysDictionariesRepo(db)
		ushield_additional_services_contact, _ := dictRepo.GetDictionaryDetail("ushield_additional_services_contact")
		ushield_additional_services_wallet, _ := dictRepo.GetDictionaryDetail("ushield_additional_services_wallet")

		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[_lang]["ushield_additional_services_sim_desc"]+"\n"+strings.ReplaceAll(global.Translations[_lang]["ushield_additional_services_contact"], "{ushield_additional_services_contact}", ushield_additional_services_contact)+strings.ReplaceAll(global.Translations[_lang]["ushield_additional_services_wallet"], "{ushield_additional_services_wallet}", ushield_additional_services_wallet))
		msg.ParseMode = "HTML"
		bot.Send(msg)

	case callbackQuery.Data == "click_energy_financing":
		dictRepo := repositories.NewSysDictionariesRepo(db)
		ushield_additional_services_contact, _ := dictRepo.GetDictionaryDetail("ushield_additional_services_contact")
		ushield_additional_services_wallet, _ := dictRepo.GetDictionaryDetail("ushield_additional_services_wallet")

		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[_lang]["ushield_additional_services_energy_financing_desc"]+"\n"+strings.ReplaceAll(global.Translations[_lang]["ushield_additional_services_contact"], "{ushield_additional_services_contact}", ushield_additional_services_contact)+strings.ReplaceAll(global.Translations[_lang]["ushield_additional_services_wallet"], "{ushield_additional_services_wallet}", ushield_additional_services_wallet))
		msg.ParseMode = "HTML"
		bot.Send(msg)
	case callbackQuery.Data == "click_sns":
		dictRepo := repositories.NewSysDictionariesRepo(db)
		ushield_additional_services_contact, _ := dictRepo.GetDictionaryDetail("ushield_additional_services_contact")
		ushield_additional_services_wallet, _ := dictRepo.GetDictionaryDetail("ushield_additional_services_wallet")

		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[_lang]["ushield_additional_services_sns_desc"]+"\n"+strings.ReplaceAll(global.Translations[_lang]["ushield_additional_services_contact"], "{ushield_additional_services_contact}", ushield_additional_services_contact)+strings.ReplaceAll(global.Translations[_lang]["ushield_additional_services_wallet"], "{ushield_additional_services_wallet}", ushield_additional_services_wallet))
		msg.ParseMode = "HTML"
		bot.Send(msg)
	case callbackQuery.Data == "click_ecs":
		dictRepo := repositories.NewSysDictionariesRepo(db)
		ushield_additional_services_contact, _ := dictRepo.GetDictionaryDetail("ushield_additional_services_contact")
		ushield_additional_services_wallet, _ := dictRepo.GetDictionaryDetail("ushield_additional_services_wallet")

		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[_lang]["ushield_additional_services_ecs_desc"]+"\n"+strings.ReplaceAll(global.Translations[_lang]["ushield_additional_services_contact"], "{ushield_additional_services_contact}", ushield_additional_services_contact)+strings.ReplaceAll(global.Translations[_lang]["ushield_additional_services_wallet"], "{ushield_additional_services_wallet}", ushield_additional_services_wallet))
		msg.ParseMode = "HTML"
		bot.Send(msg)
	case strings.HasPrefix(callbackQuery.Data, "click_buy_month_"):
		month := strings.ReplaceAll(callbackQuery.Data, "click_buy_month_", "")
		fmt.Printf("month: %s\n", month)
		member.MenuNavigateForMonth(cache, _lang, db, callbackQuery.Message.Chat.ID, callbackQuery.From.UserName, bot, month)
		//default:
		//	responseText = "未知选项"

	case strings.HasPrefix(callbackQuery.Data, "activate_current_user_"):
		month := strings.ReplaceAll(callbackQuery.Data, "activate_current_user_", "")
		fmt.Printf("month: %s\n", month)
		//fmt.Printf("username: %s\n", callbackQuery.Message.From.UserName)
		fmt.Printf("username: %s\n", callbackQuery.Message.Chat.UserName)
		//fmt.Printf("chatid: %s\n", callbackQuery.Message.From.)
		//member.MenuNavigateForMonth(cache, _lang, db, callbackQuery.Message.Chat.ID, callbackQuery.From.UserName, bot, month)
		member.Rent(_lang, cache, db, bot, callbackQuery.Message.Chat.UserName, callbackQuery.Message.Chat.ID, month)

	case strings.HasPrefix(callbackQuery.Data, "pay_premium_order_"):
		orderNO := strings.ReplaceAll(callbackQuery.Data, "pay_premium_order_", "")
		fmt.Printf("支付订单 %s\n", orderNO)
		usdtDepositRepo := repositories.NewUserUSDTDepositsRepository(db)
		record, _ := usdtDepositRepo.Query(context.Background(), orderNO)

		userRepo := repositories.NewUserRepository(db)
		user, _ := userRepo.GetByUserID(callbackQuery.Message.Chat.ID)
		if IsEmpty(user.Amount) {
			user.Amount = "0"
		}
		if IsEmpty(user.TronAmount) {
			user.TronAmount = "0"
		}

		if flag, _ := CompareNumberStrings(user.Amount, record.Amount); flag < 0 {
			msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID,
				"<b>"+"🔍"+global.Translations[_lang]["insufficient_balance_tips"]+"</b>"+"\n"+
					"🆔"+global.Translations[_lang]["user_id"]+": <code>"+user.Associates+"</code>\n"+
					"👤"+global.Translations[_lang]["username"]+": @"+user.Username+"\n"+
					"💰"+global.Translations[_lang]["balance"]+"\n"+
					"- TRX：   "+user.TronAmount+"\n"+
					"-  USDT："+user.Amount)
			msg.ParseMode = "HTML"
			inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("💵"+global.Translations[_lang]["deposit"], "deposit_amount"),
				),
			)

			msg.ReplyMarkup = inlineKeyboard
			msg.ParseMode = "HTML"
			bot.Send(msg)
			return
		}

		balance, _ := SubtractStringNumbers(user.Amount, record.Amount, 1)
		fmt.Printf("USDT balance %s", balance)
		user.Amount = balance
		err = userRepo.Update2(context.Background(), &user)
		if err != nil {
			fmt.Println("支付失败")
		}

		//调用catfee支付会员

		tgOrderDB := repositories.NewTelegramPremiumOrderRepository(db)
		orderRecord, _ := tgOrderDB.Query(context.Background(), orderNO)

		tgOrderDB.Update(context.Background(), orderNO, 1)

		catfeeClient.Premium(orderRecord.TGUsername, orderRecord.Month)

		//设置用户状态
		//orderNO, _ := cache.Get(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10) + "_order_no")
		tips := global.Translations[_lang]["successfully_purchased_telegram"]
		tips = strings.ReplaceAll(tips, "{month_package}", global.Translations[_lang][orderRecord.Month+"_month_premium"])
		msg_order := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID,
			global.Translations[_lang]["order_id"]+"：TOPUP-"+orderNO+" , "+tips)
		msg_order.ParseMode = "HTML"
		//msg.DisableWebPagePreview = true
		bot.Send(msg_order)

		//修改placeholder 0
		//修改 depositorder 2
		usdtPlaceholderRepo := repositories.NewUserUsdtPlaceholdersRepository(db)
		usdtPlaceholderRepo.UpdateByPlaceholder(context.Background(), record.Placeholder, 0)
		fmt.Printf("placeholder重置 %s\n", record.Placeholder)
		usdtDepositRepo.Update(context.Background(), orderNO, 2)

	case strings.HasPrefix(callbackQuery.Data, "cancel_premium_order_"):
		orderNO := strings.ReplaceAll(callbackQuery.Data, "cancel_premium_order_", "")
		fmt.Printf("取消支付订单 %s\n", orderNO)

		usdtDepositRepo := repositories.NewUserUSDTDepositsRepository(db)
		record, _ := usdtDepositRepo.Query(context.Background(), orderNO)

		usdtPlaceholderRepo := repositories.NewUserUsdtPlaceholdersRepository(db)
		usdtPlaceholderRepo.UpdateByPlaceholder(context.Background(), record.Placeholder, 0)
		fmt.Printf("placeholder重置 %s\n", record.Placeholder)
		usdtDepositRepo.Update(context.Background(), orderNO, 2)

		//设置用户状态
		//orderNO, _ := cache.Get(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10) + "_order_no")
		msg_order := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID,
			global.Translations[_lang]["order_id"]+"：TOPUP-"+orderNO+" , "+global.Translations[_lang]["cancel_order_tips"])
		msg_order.ParseMode = "HTML"
		//msg.DisableWebPagePreview = true
		bot.Send(msg_order)

		prevMessageIDStr, _ := cache.Get(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10) + "_order")

		prevMessageID, err := strconv.Atoi(prevMessageIDStr)
		if err != nil {
			fmt.Println("转换失败:", err)
			//return
		}

		bot.Request(tgbotapi.DeleteMessageConfig{ChatID: callbackQuery.Message.Chat.ID, MessageID: prevMessageID})

	case strings.HasPrefix(callbackQuery.Data, "purchase_star_menu"):
		member.MenuStarNavigate(_lang, db, callbackQuery.Message.Chat.ID, bot)
	case strings.HasPrefix(callbackQuery.Data, "purchase_telegram_premium"):
		member.MenuNavigate(_lang, db, callbackQuery.Message.Chat.ID, bot)

	case strings.HasPrefix(callbackQuery.Data, "click_purchase_stars_"):
		count := strings.ReplaceAll(callbackQuery.Data, "click_purchase_stars_", "")
		fmt.Printf("count: %s\n", count)
		member.MenuNavigateForStar(cache, _lang, db, callbackQuery.Message.Chat.ID, callbackQuery.From.UserName, bot, count)

	case strings.HasPrefix(callbackQuery.Data, "purchase_stars_current_user_"):
		count := strings.ReplaceAll(callbackQuery.Data, "purchase_stars_current_user_", "")
		fmt.Printf("数量: %s\n", count)
		//fmt.Printf("username: %s\n", callbackQuery.Message.From.UserName)
		fmt.Printf("username: %s\n", callbackQuery.Message.Chat.UserName)
		//fmt.Printf("chatid: %s\n", callbackQuery.Message.From.)
		//member.MenuNavigateForMonth(cache, _lang, db, callbackQuery.Message.Chat.ID, callbackQuery.From.UserName, bot, month)
		member.Purchase(_lang, cache, db, bot, callbackQuery.Message.Chat.UserName, callbackQuery.Message.Chat.ID, count)

	case strings.HasPrefix(callbackQuery.Data, "purchase_stars_"):
		orderNO := strings.ReplaceAll(callbackQuery.Data, "purchase_stars_", "")
		fmt.Printf("支付订单 %s\n", orderNO)
		usdtDepositRepo := repositories.NewUserUSDTDepositsRepository(db)
		record, _ := usdtDepositRepo.Query(context.Background(), orderNO)

		userRepo := repositories.NewUserRepository(db)
		user, _ := userRepo.GetByUserID(callbackQuery.Message.Chat.ID)
		if IsEmpty(user.Amount) {
			user.Amount = "0"
		}
		if IsEmpty(user.TronAmount) {
			user.TronAmount = "0"
		}

		if flag, _ := CompareNumberStrings(user.Amount, record.Amount); flag < 0 {
			msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID,
				"<b>"+"🔍"+global.Translations[_lang]["insufficient_balance_tips"]+"</b>"+"\n"+
					"🆔"+global.Translations[_lang]["user_id"]+": <code>"+user.Associates+"</code>\n"+
					"👤"+global.Translations[_lang]["username"]+": @"+user.Username+"\n"+
					"💰"+global.Translations[_lang]["balance"]+"\n"+
					"- TRX：   "+user.TronAmount+"\n"+
					"-  USDT："+user.Amount)
			msg.ParseMode = "HTML"
			inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("💵"+global.Translations[_lang]["deposit"], "deposit_amount"),
				),
			)

			msg.ReplyMarkup = inlineKeyboard
			msg.ParseMode = "HTML"
			bot.Send(msg)
			return
		}

		balance, _ := SubtractStringNumbers(user.Amount, record.Amount, 1)
		fmt.Printf("USDT balance %s", balance)
		user.Amount = balance
		err = userRepo.Update2(context.Background(), &user)
		if err != nil {
			fmt.Println("支付失败")
		}

		//调用catfee支付会员

		tgOrderDB := repositories.NewTelegramStarsOrderRepository(db)
		orderRecord, _ := tgOrderDB.Query(context.Background(), orderNO)

		tgOrderDB.Update(context.Background(), orderNO, 1)

		//catfeeClient.Premium(orderRecord.TGUsername, orderRecord.Month)

		//设置用户状态
		//orderNO, _ := cache.Get(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10) + "_order_no")
		tips := global.Translations[_lang]["successfully_purchased_stars"]
		tips = strings.ReplaceAll(tips, "{count}", orderRecord.Stars)
		msg_order := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID,
			global.Translations[_lang]["order_id"]+"：TOPUP-"+orderNO+" , "+tips)
		msg_order.ParseMode = "HTML"
		//msg.DisableWebPagePreview = true
		bot.Send(msg_order)

		//修改placeholder 0
		//修改 depositorder 2
		usdtPlaceholderRepo := repositories.NewUserUsdtPlaceholdersRepository(db)
		usdtPlaceholderRepo.UpdateByPlaceholder(context.Background(), record.Placeholder, 0)
		fmt.Printf("placeholder重置 %s\n", record.Placeholder)
		usdtDepositRepo.Update(context.Background(), orderNO, 2)

	case strings.HasPrefix(callbackQuery.Data, "cancel_stars_"):
		orderNO := strings.ReplaceAll(callbackQuery.Data, "cancel_stars_", "")
		fmt.Printf("取消支付订单 %s\n", orderNO)

		usdtDepositRepo := repositories.NewUserUSDTDepositsRepository(db)
		record, _ := usdtDepositRepo.Query(context.Background(), orderNO)

		usdtPlaceholderRepo := repositories.NewUserUsdtPlaceholdersRepository(db)
		usdtPlaceholderRepo.UpdateByPlaceholder(context.Background(), record.Placeholder, 0)
		fmt.Printf("placeholder重置 %s\n", record.Placeholder)
		usdtDepositRepo.Update(context.Background(), orderNO, 2)

		//设置用户状态
		//orderNO, _ := cache.Get(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10) + "_order_no")
		msg_order := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID,
			global.Translations[_lang]["order_id"]+"：TOPUP-"+orderNO+" , "+global.Translations[_lang]["cancel_order_tips"])
		msg_order.ParseMode = "HTML"
		//msg.DisableWebPagePreview = true
		bot.Send(msg_order)

		prevMessageIDStr, _ := cache.Get(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10) + "_order")

		prevMessageID, err := strconv.Atoi(prevMessageIDStr)
		if err != nil {
			fmt.Println("转换失败:", err)
			//return
		}

		bot.Request(tgbotapi.DeleteMessageConfig{ChatID: callbackQuery.Message.Chat.ID, MessageID: prevMessageID})

		//case strings.HasPrefix(callbackQuery.Data, "cancel_stars_"):
		//	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[_lang]["ushield_additional_services_ecs_desc"]+"\n"+strings.ReplaceAll(global.Translations[_lang]["ushield_additional_services_contact"], "{ushield_additional_services_contact}", ushield_additional_services_contact)+strings.ReplaceAll(global.Translations[_lang]["ushield_additional_services_wallet"], "{ushield_additional_services_wallet}", ushield_additional_services_wallet))
		//	msg.ParseMode = "HTML"
		//	bot.Send(msg)

	case strings.HasPrefix(callbackQuery.Data, "purchase_anonymous_mobile"):

		member.MenuMobileNavigate(_lang, db, callbackQuery.Message.Chat.ID, bot)

	case strings.HasPrefix(callbackQuery.Data, "click_launder_"):
		amount := strings.ReplaceAll(callbackQuery.Data, "click_launder_", "")
		fmt.Printf("金额amount: %s\n", amount)
		launder.MenuNavigateForLaunder(cache, _lang, db, callbackQuery.Message.Chat.ID, callbackQuery.From.UserName, bot, amount)

	case strings.HasPrefix(callbackQuery.Data, "click_laundering_"):
		content := strings.ReplaceAll(callbackQuery.Data, "click_laundering_", "")
		fmt.Printf("内容: %s\n", content)
		amount := strings.Split(content, "_")[1]
		token := strings.Split(content, "_")[0]
		fmt.Printf("代币: %s - 金额: %s\n", token, amount)

		expiration := 1 * time.Minute // 短时间缓存空值

		//设置用户状态
		cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), "click_laundering_"+content, expiration)

		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "💬"+global.Translations[_lang]["input_receive_address"]+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)

	//launder.MenuNavigateForLaunder(cache, _lang, db, callbackQuery.Message.Chat.ID, callbackQuery.From.UserName, bot, amount)

	//api := fixedfloat.New("AtHmGIAucigijgkqaTiOvuGTArkBrm4pparh7V5E", "jDDzTJKmB8jfzhlxfZuXtdNnNQLrSjaGiKg2e4kf")

	//params := map[string]interface{}{
	//	"fromCcy":   "USDTTRC",
	//	"toCcy":     "USDT",
	//	"type":      fixedfloat.TypeFloat,
	//	"amount":    1000,
	//	"direction": "from",
	//	"toAddress": "0xF510e53EF8DA4e45FFA59EB554511a7410E5eFD3",
	//}
	//order, err := api.Create(params)

	case callbackQuery.Data == "coin_swap_coin":
		service.MenuNavigateCoin2CoinSwap2(_lang, db, callbackQuery.Message.Chat.ID, bot)

	case callbackQuery.Data == "function_address_trace":
		service.MenuNavigateAddressTrace(_lang, cache, bot, callbackQuery.Message.Chat.ID, db)

	case callbackQuery.Data == "function_remove_label":
		launder.MenuLaunderNavigate(_lang, db, callbackQuery.Message.Chat.ID, bot)

	}

	// 发送新消息作为响应
	bot.Send(tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, responseText))
}
