package router

import (
	"context"
	"strconv"
	"strings"
	"time"
	"ushield_bot/internal/domain"
	"ushield_bot/internal/global"
	"ushield_bot/internal/infrastructure/repositories"
	trxfee "ushield_bot/internal/infrastructure/thirdparty"
	"ushield_bot/internal/infrastructure/thirdparty/fixedfloat"
	. "ushield_bot/internal/infrastructure/tools"
	logger "ushield_bot/internal/logger"
	"ushield_bot/internal/service"
	"ushield_bot/internal/service/additional"
	"ushield_bot/internal/service/catfee"
	"ushield_bot/internal/service/command"
	"ushield_bot/internal/service/launder"
	"ushield_bot/internal/service/member"
	"ushield_bot/internal/service/yhb"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleMessageUpdate(message *tgbotapi.Message, ctx Context) {
	lang, err := ctx.Cache.Get("LANG_" + strconv.FormatInt(message.Chat.ID, 10))
	if len(lang) == 0 || err != nil {
		userRepo := repositories.NewUserRepository(ctx.DB)
		record, _ := userRepo.GetByChatID(message.Chat.ID)
		ctx.Cache.Set("LANG_"+strconv.FormatInt(message.Chat.ID, 10), record.Lang, 24*time.Hour)
		lang = record.Lang
	}

	switch message.Text {
	case "⚽️世界杯竞猜🏆":
		msg := tgbotapi.NewMessage(message.Chat.ID, "点击跳转机器人 => 🤖@ushield_octopus_bot\n")
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔙"+global.Translations[lang]["back_home"], "back_home"),
			),
		)
		ctx.Bot.Send(msg)

	case "🥂" + global.Translations[lang]["coin_laundering_menu"]:
		logger.Printf("洗U步骤开始\n")
		launder.MenuLaunderNavigate(lang, ctx.DB, message.Chat.ID, ctx.Bot)
	case global.Translations[lang]["member_telegram_menu"]:
		member.MenuNavigate(lang, ctx.DB, message.Chat.ID, ctx.Bot)
	case global.Translations[lang]["command_energy_menu"]:
		command.MenuNavigate(lang, ctx.DB, message.Chat.ID, ctx.Bot)
	case "🛒" + global.Translations[lang]["ushield_additional_services_menu"]:
		additional.MenuNavigate(lang, ctx.DB, message.Chat.ID, ctx.Bot)
	case "🔃" + global.Translations[lang]["coin_swap_coin_menu"]:
		service.MenuNavigateCoin2CoinSwap(lang, ctx.DB, message, ctx.Bot, ctx.FixedfloatURL)
	case "🧧" + global.Translations[lang]["yhb_menu"]:
		yhb.MenuNavigateTronEnergy(lang, ctx.DB, message, ctx.Bot)
	case "⛽" + global.Translations[lang]["tron_energy_menu"]:
		service.MenuNavigateTronEnergy(lang, ctx.DB, message, ctx.Bot)
	case "✅" + global.Translations[lang]["usdt_trx_swap"]:
		service.MenuNavigateSwapExchange(lang, ctx.DB, message, ctx.Bot)
	case "🕸" + global.Translations[lang]["address_trace_menu"]:
		service.MenuNavigateAddressTrace(lang, ctx.Cache, ctx.Bot, message.Chat.ID, ctx.DB)
	case "🔍" + global.Translations[lang]["address_check"]:
		service.MenuNavigateAddressDetection(lang, ctx.Cache, ctx.Bot, message.Chat.ID, ctx.DB)
	case "🚨" + global.Translations[lang]["usdt_freeze_alert"]:
		service.MenuNavigateAddressFreeze(lang, ctx.Cache, ctx.Bot, message.Chat.ID, ctx.DB)
	case "🖊️" + global.Translations[lang]["transaction_plans"]:
		service.MenuNavigateBundlePackage(lang, ctx.DB, message.Chat.ID, ctx.Bot, "TRX")
	case "🤖" + global.Translations[lang]["catfee_smart_transaction_menu"]:
		catfee.MenuNavigateCatfeeSmartTransactionPlans(lang, ctx.DB, message.Chat.ID, ctx.Bot, "TRX")
	case "⚡" + global.Translations[lang]["energy_swap"]:
		service.MenuNavigateEnergyExchange(lang, ctx.DB, message, ctx.Bot)
	case "👤" + global.Translations[lang]["my_account"]:
		service.MenuNavigateHome(lang, ctx.Cache, ctx.DB, message, ctx.Bot)
	case "🌍" + global.Translations[lang]["language"]:
		service.MenuNavigateHome2(ctx.DB, message, ctx.Bot)
	default:
		status, _ := ctx.Cache.Get(strconv.FormatInt(message.Chat.ID, 10))
		logger.Printf("用户状态status %s", status)

		switch {
		case strings.HasPrefix(status, "user_backup_notify"):
			if service.ExtractBackup(message, ctx.Bot, ctx.DB) {
				return
			}
		case strings.HasPrefix(status, "start_freeze_risk"):
			freezeAlertService := service.NewFreezeAlertService(ctx.DB)
			preview, previewErr := freezeAlertService.Preview(message.Text)
			if previewErr == service.ErrFreezeAlertInvalidAddress {
				msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[lang]["address_wrong_tips"]+"</b>"+"\n")
				msg.ParseMode = "HTML"
				ctx.Bot.Send(msg)
				return
			}
			if previewErr != nil {
				logger.Printf("freeze alert preview err: %v", previewErr)
				return
			}

			sendFreezeAlertPreview(ctx.Bot, message.Chat.ID, lang, preview)
			setShortState(ctx.Cache, message.Chat.ID, "start_freeze_risk_status")

		case strings.HasPrefix(status, "address_list_trace"):
		case strings.HasPrefix(status, "address_manager_remove"):
			if IsValidAddress(message.Text) || IsValidEthereumAddress(message.Text) {
				userRepo := repositories.NewUserAddressMonitorRepo(ctx.DB)
				_ = userRepo.Remove(context.Background(), message.Chat.ID, message.Text)
				msg := tgbotapi.NewMessage(message.Chat.ID, "✅ "+"<b>"+global.Translations[lang]["address_deleted_success"]+"</b>"+"\n")
				msg.ParseMode = "HTML"
				ctx.Bot.Send(msg)
				service.ShowAddressManager(lang, ctx.Cache, ctx.Bot, message.Chat.ID, ctx.DB)
			} else {
				msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[lang]["invalid_address_tips"]+"</b>"+"\n")
				msg.ParseMode = "HTML"
				ctx.Bot.Send(msg)
			}
		case strings.HasPrefix(status, "dispatch_others"):
			if IsValidAddress(message.Text) {
				dispatchService := service.NewEnergyDispatchService(ctx.DB, ctx.TrxfeeURL, ctx.TrxfeeAPIKey, ctx.TrxfeeSecret, ctx.CatfeeClient)
				result, dispatchErr := dispatchService.DispatchToManualAddress(context.Background(), message.Chat.ID, message.Text)
				if dispatchErr == nil {
					sendDispatchSuccess(ctx.Bot, message.Chat.ID, result)
					setShortState(ctx.Cache, message.Chat.ID, "null_dispatch_others_")
				} else if dispatchErr == service.ErrDispatchInsufficientTimes {
					service.MenuNavigateBundlePackage(lang, ctx.DB, message.Chat.ID, ctx.Bot, "TRX")
				}
			} else {
				msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[lang]["invalid_address_tips"]+"</b>"+"\n")
				msg.ParseMode = "HTML"
				ctx.Bot.Send(msg)
			}
		case strings.HasPrefix(status, "DISPATCHOTHERS_"):
			if IsValidAddress(message.Text) {
				subscribeBundleID := strings.ReplaceAll(status, "DISPATCHOTHERS_", "")
				dispatchService := service.NewEnergyDispatchService(ctx.DB, ctx.TrxfeeURL, ctx.TrxfeeAPIKey, ctx.TrxfeeSecret, ctx.CatfeeClient)
				result, dispatchErr := dispatchService.DispatchFromSubscription(context.Background(), subscribeBundleID, message.Text, message.Chat.ID)
				if dispatchErr == nil {
					msg2 := service.BuildBundlePackageSubscriptionStatsMessage(lang, ctx.DB, message.Chat.ID)
					ctx.Bot.Send(msg2)
					sendDispatchSuccess(ctx.Bot, message.Chat.ID, result)
					setShortState(ctx.Cache, message.Chat.ID, "null_dispatch_others_")
				}
			} else {
				msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[lang]["invalid_address_tips"]+"</b>"+"\n")
				msg.ParseMode = "HTML"
				ctx.Bot.Send(msg)
			}
		case strings.HasPrefix(status, "address_manager_add"):
			service.AddManagedAddress(lang, message, ctx.DB, ctx.Bot)
			service.ShowAddressManager(lang, ctx.Cache, ctx.Bot, message.Chat.ID, ctx.DB)
		case strings.HasPrefix(status, "bundle_"):
			logger.Printf(">>>>>>>>>>>>>>>>>>>>bundle: %s", status)
			if service.ExtractBundleService(lang, message, ctx.Bot, ctx.DB, status) {
				return
			}
		case strings.HasPrefix(status, "address_trace_add"):
			if !IsValidAddress(message.Text) {
				msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[lang]["address_wrong_tips"]+"</b>"+"\n")
				msg.ParseMode = "HTML"
				ctx.Bot.Send(msg)
				return
			}

			userRepo := repositories.NewUserAddressTraceRepo(ctx.DB)
			model, _ := userRepo.GetByChatIDAndAddress(context.Background(), message.Chat.ID, message.Text)
			if model.Id > 0 {
				msg := tgbotapi.NewMessage(message.Chat.ID, "✅"+"<b>"+global.Translations[lang]["address_trace_add_repeat_tips"]+"</b>"+"\n")
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[lang]["back_homepage"], "back_user_address_trace"),
					),
				)
				msg.ParseMode = "HTML"
				ctx.Bot.Send(msg)
				return
			}

			total, _ := userRepo.Count(context.Background(), message.Chat.ID)
			if total >= 4 {
				msg := tgbotapi.NewMessage(message.Chat.ID, "✅"+"<b>"+global.Translations[lang]["address_trace_add_max_tips"]+"</b>"+"\n")
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[lang]["back_homepage"], "back_user_address_trace"),
					),
				)
				msg.ParseMode = "HTML"
				ctx.Bot.Send(msg)
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
			_ = userRepo.Create(context.Background(), &record)

			msg := tgbotapi.NewMessage(message.Chat.ID, "✅"+"<b>"+global.Translations[lang]["address_added_success"]+"</b>"+"\n")
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[lang]["back_homepage"], "back_user_address_trace"),
				),
			)
			msg.ParseMode = "HTML"
			ctx.Bot.Send(msg)

		case strings.HasPrefix(status, "address_trace_delete"):
			if !IsValidAddress(message.Text) {
				msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[lang]["address_wrong_tips"]+"</b>"+"\n")
				msg.ParseMode = "HTML"
				ctx.Bot.Send(msg)
				return
			}
			userRepo := repositories.NewUserAddressTraceRepo(ctx.DB)
			if err := userRepo.Remove(context.Background(), message.Chat.ID, message.Text); err != nil {
				return
			}
			msg := tgbotapi.NewMessage(message.Chat.ID, "✅ "+"<b>"+global.Translations[lang]["address_deleted_success"]+"</b>"+"\n")
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[lang]["back_homepage"], "back_user_address_trace"),
				),
			)
			msg.ParseMode = "HTML"
			ctx.Bot.Send(msg)

		case strings.HasPrefix(status, "usdt_risk_monitor"):
			if !IsValidAddress(message.Text) {
				msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[lang]["invalid_address_tips"]+"</b>"+"\n")
				msg.ParseMode = "HTML"
				ctx.Bot.Send(msg)
			}

			msg := tgbotapi.NewMessage(message.Chat.ID, "")
			msg.ParseMode = "HTML"
			ctx.Bot.Send(msg)

		case strings.HasPrefix(status, "click_bundle_package_address_manager_remove"):
			if service.RemoveBundlePackageAddress(lang, ctx.Cache, ctx.Bot, message, ctx.DB) {
				return
			}
		case strings.HasPrefix(status, "click_bundle_package_address_manager_add"):
			if service.AddBundlePackageAddress(lang, ctx.Cache, ctx.Bot, message, ctx.DB) {
				return
			}
		case strings.HasPrefix(status, "apply_bundle_package_"):
			if service.ApplyBundlePackage(lang, ctx.Cache, ctx.Bot, message, ctx.DB, status) {
				return
			}
		case strings.HasPrefix(status, "apply_ST_bundle_package_"):
			trxfeeClient := trxfee.NewTrxfeeClient(ctx.TrxfeeURL, ctx.TrxfeeAPIKey, ctx.TrxfeeSecret)
			if service.APPLY_ST_BUNDLE_PACKAGE(trxfeeClient, lang, ctx.Cache, ctx.Bot, message, ctx.DB, status) {
				return
			}
		case strings.HasPrefix(status, "click_backup_account"):
			logger.Printf("进入click_backup_account状态：%s\n", message.Text)
			if strings.Contains(message.Text, "@") {
				msg := tgbotapi.NewMessage(message.Chat.ID, "❌ "+global.Translations[lang]["backup_account_tips"])
				msg.ParseMode = "HTML"
				ctx.Bot.Send(msg)
				return
			}
			userName := strings.ReplaceAll(message.Text, "@", "")
			logger.Printf("备份用户：%s\n", userName)
			userRepo := repositories.NewUserRepository(ctx.DB)
			user, err := userRepo.GetByUsername(userName)
			if err != nil || user.Id == 0 {
				logger.Printf("访问失败 %v\n", err)
				msg := tgbotapi.NewMessage(message.Chat.ID, "❌"+global.Translations[lang]["backup_account_tips2"])
				msg.ParseMode = "HTML"
				ctx.Bot.Send(msg)
				return
			}

			user.BackupChatID = userName
			if err := userRepo.UpdateBackupChat(context.Background(), userName, message.Chat.ID); err == nil {
				msg := tgbotapi.NewMessage(message.Chat.ID, "✅ "+global.Translations[lang]["backup_account_tips3"]+message.Text)
				msg.ParseMode = "HTML"
				ctx.Bot.Send(msg)
			}

			service.MenuNavigateHome(lang, ctx.Cache, ctx.DB, message, ctx.Bot)

		case strings.HasPrefix(status, "usdt_risk_query"):
			service.ExtractSlowMistRiskQuery(lang, ctx.Cache, message, ctx.DB, ctx.RandomCookie, ctx.Bot)

		case strings.HasPrefix(status, "catfee_add_address"):
			trxfeeClient := trxfee.NewTrxfeeClient(ctx.TrxfeeURL, ctx.TrxfeeAPIKey, ctx.TrxfeeSecret)
			catfee.AddCustodyAddress(lang, ctx.Cache, ctx.DB, ctx.Bot, message, trxfeeClient)

		case strings.HasPrefix(status, "catfee_remove_address"):
			catfee.RemoveCustodyAddress(lang, ctx.Cache, ctx.DB, ctx.Bot, message, ctx.CatfeeClient)

		case strings.HasPrefix(status, "premium_user_rent_month"):
			username := message.Text
			if len(username) < 4 || !strings.Contains(username, "@") {
				return
			}
			month := strings.ReplaceAll(status, "premium_user_rent_month", "")
			logger.Printf("message text: %s\n", message.Text)
			member.Rent(lang, ctx.Cache, ctx.DB, ctx.Bot, username, message.Chat.ID, month)

		case strings.HasPrefix(status, "purchase_telegram_stars"):
			username := message.Text
			if len(username) < 4 || !strings.Contains(username, "@") {
				return
			}
			count := strings.ReplaceAll(status, "purchase_telegram_stars", "")
			logger.Printf("message text: %s\n", message.Text)
			member.Purchase(lang, ctx.Cache, ctx.DB, ctx.Bot, message.Text, message.Chat.ID, count)

		case strings.HasPrefix(status, "click_laundering_"):
			content := strings.ReplaceAll(status, "click_laundering_", "")
			logger.Printf("内容: %s\n", content)
			logger.Printf("输入内容: %s\n", message.Text)
			token := strings.Split(content, "_")[0]
			amount := strings.Split(content, "_")[1]
			logger.Printf("状态 代币: %s - 金额: %s\n", token, amount)

			if strings.ToUpper(token) != "BTC" && !IsValidEthereumAddress(message.Text) {
				msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[lang]["address_wrong_tips"]+"</b>"+"\n")
				msg.ParseMode = "HTML"
				ctx.Bot.Send(msg)
				return
			}
			if strings.ToUpper(token) == "BTC" && !IsValidBitcoinAddress(message.Text) {
				msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[lang]["address_wrong_tips"]+"</b>"+"\n")
				msg.ParseMode = "HTML"
				ctx.Bot.Send(msg)
				return
			}

			api := fixedfloat.New("AZxSXXl6VwqSgJkkC6HovFyxWib0ZPUNVBOO8Fkt", "vOGKrbXgFBepGBEze90EUUHHnsLzjQHC8197WtRC")
			params := map[string]interface{}{
				"fromCcy":   "USDTTRC",
				"type":      fixedfloat.TypeFloat,
				"direction": "from",
				"refcode":   "r8ck81xa",
				"ref":       "r8ck81xa",
				"afftax":    1,
			}
			params["toCcy"] = token
			params["amount"] = amount
			params["toAddress"] = message.Text
			rawMap, err := api.Create(params)
			if err != nil {
				return
			}

			from, to, ok := fixedfloat.ExtractFromAndTo(rawMap)
			if !ok {
				logger.Println("Failed to extract from/to")
				return
			}

			timeInfo, ok := fixedfloat.ExtractTime(rawMap)
			if !ok {
				logger.Println("Failed to extract time")
				return
			}

			regTime := time.Unix(int64(timeInfo.Reg), 0)
			expTime := time.Unix(int64(timeInfo.Expiration), 0)
			id, _, ok := fixedfloat.ExtractIDAndStatus(rawMap)
			if !ok {
				logger.Println("Failed to extract id or status")
				return
			}

			desc := global.Translations[lang]["coin_laundering_order_desc"]
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
			filename, err := fixedfloat.GenerateQRCodeWithTimestamp(fromAddress, 300)
			if err != nil {
				return
			}

			videoPath := "/root/ushield-telegram-bot/old/" + filename
			msg := tgbotapi.NewPhoto(message.Chat.ID, tgbotapi.FilePath(videoPath))
			msg.Caption = "✅ " + desc
			msg.ParseMode = "HTML"
			ctx.Bot.Send(msg)

			orderDB := repositories.NewCoinLaunderingOrderRepo(ctx.DB)
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

func sendFreezeAlertPreview(bot *tgbotapi.BotAPI, chatID int64, lang string, preview *service.FreezeAlertPreview) {
	tips := strings.ReplaceAll(
		strings.ReplaceAll(global.Translations[lang]["enable_freeze_alerts_tips"], "{server_usdt_price}", preview.USDTPrice),
		"{server_trx_price}",
		preview.TRXPrice,
	)

	msg := tgbotapi.NewMessage(
		chatID,
		global.Translations[lang]["enable_freeze_alerts_tips_suffix"]+"\n"+
			global.Translations[lang]["address"]+"："+preview.Address+"\n\n"+
			tips,
	)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅"+global.Translations[lang]["confirm_freeze_alerts"], "confirm_freeze_risk_"+preview.Address),
			tgbotapi.NewInlineKeyboardButtonData("❌"+global.Translations[lang]["cancel_freeze_alerts"], "back_risk_home"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[lang]["back_homepage"], "back_risk_home"),
		),
	)
	bot.Send(msg)
}
