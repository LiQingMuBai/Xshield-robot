package service

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"
	"ushield_bot/internal/cache"
	"ushield_bot/internal/domain"
	"ushield_bot/internal/global"
	"ushield_bot/internal/infrastructure/repositories"
	. "ushield_bot/internal/infrastructure/tools"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

type bundleMenuOption struct {
	id   int64
	name string
}

func buildKeyboardRows(buttons []tgbotapi.InlineKeyboardButton, rowSize int) [][]tgbotapi.InlineKeyboardButton {
	if rowSize <= 0 {
		return nil
	}

	var keyboard [][]tgbotapi.InlineKeyboardButton
	for i := 0; i < len(buttons); i += rowSize {
		end := i + rowSize
		if end > len(buttons) {
			end = len(buttons)
		}
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(buttons[i:end]...))
	}

	return keyboard
}

func loadMenuUser(db *gorm.DB, chatID int64) domain.User {
	userRepo := repositories.NewUserRepository(db)
	user, _ := userRepo.GetByChatID(chatID)
	if IsEmpty(user.Amount) {
		user.Amount = "0"
	}
	if IsEmpty(user.TronAmount) {
		user.TronAmount = "0"
	}
	return user
}

func getBotReceiveAddress(db *gorm.DB) string {
	agent := os.Getenv("BOT_AGENT")
	sysUserRepo := repositories.NewSysUsersRepository(db)
	receiveAddress, _, _ := sysUserRepo.GetAddressesByUsername(context.Background(), agent)
	return receiveAddress
}

func sendFixedFloatMenu(lang string, chatID int64, db *gorm.DB, bot *tgbotapi.BotAPI) {
	dictRepo := repositories.NewSysDictionariesRepo(db)
	fixedFloatURL, _ := dictRepo.GetDictionaryDetail("ff_ref_url")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(global.Translations[lang]["coin_swap_coin_menu"], fixedFloatURL),
			tgbotapi.NewInlineKeyboardButtonURL("FixedFloat rules", "https://ff.io/terms-of-service"),
		),
	)

	msg := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath("./static/fixedfloat.jpg"))
	msg.Caption = global.Translations[lang]["fixedfloat_rules"] + "\n\n" + global.Translations[lang]["coin_swap_coin_tips"]
	msg.ReplyMarkup = keyboard
	msg.ParseMode = "HTML"
	bot.Send(msg)
}

func buildBundleMenuKeyboard(bundleButtons []tgbotapi.InlineKeyboardButton, switchButtons []tgbotapi.InlineKeyboardButton, extraButtons []tgbotapi.InlineKeyboardButton) tgbotapi.InlineKeyboardMarkup {
	var keyboard [][]tgbotapi.InlineKeyboardButton
	keyboard = append(keyboard, buildKeyboardRows(bundleButtons, 2)...)
	keyboard = append(keyboard, buildKeyboardRows(switchButtons, 1)...)
	keyboard = append(keyboard, buildKeyboardRows(extraButtons, 2)...)
	return tgbotapi.NewInlineKeyboardMarkup(keyboard...)
}

func renderMenuTemplate(template string, replacements map[string]string) string {
	replacerArgs := make([]string, 0, len(replacements)*2)
	for placeholder, value := range replacements {
		replacerArgs = append(replacerArgs, placeholder, value)
	}
	if len(replacerArgs) == 0 {
		return template
	}
	return strings.NewReplacer(replacerArgs...).Replace(template)
}

func buildEnergySwapText(lang string, db *gorm.DB) string {
	dictRepo := repositories.NewSysDictionariesRepo(db)
	energyCost, _ := dictRepo.GetDictionaryDetail("energy_cost")

	energyCost2x, _ := StringMultiply(energyCost, 2)
	energyCost10x, _ := StringMultiply(energyCost, 10)

	return renderMenuTemplate(global.Translations[lang]["energy_swap_tips"], map[string]string{
		"{energy_cost}":     energyCost,
		"{energy_cost_2x}":  energyCost2x,
		"{energy_cost_10x}": energyCost10x,
		"{receiveAddress}":  getBotReceiveAddress(db),
	})
}

func buildSwapExchangeText(lang string, db *gorm.DB) string {
	dictRepo := repositories.NewSysDictionariesRepo(db)

	return renderMenuTemplate(global.Translations[lang]["usdt_trx_swap_head"], map[string]string{
		"{trx_amount}":   getDictionaryDetail(dictRepo, "usdt_swap_trx_amount_real"),
		"{min_amount}":   getDictionaryDetail(dictRepo, "usdt_swap_trx_min_amount"),
		"{max_amount}":   getDictionaryDetail(dictRepo, "usdt_swap_trx_max_amount"),
		"{swap_address}": getDictionaryDetail(dictRepo, "usdt_swap_trx_swap_address"),
	})
}

func getDictionaryDetail(repo *repositories.SysDictionariesRepo, key string) string {
	value, _ := repo.GetDictionaryDetail(key)
	return value
}

func buildBundleOptionButtons(lang string, options []bundleMenuOption, callbackPrefix string) []tgbotapi.InlineKeyboardButton {
	buttons := make([]tgbotapi.InlineKeyboardButton, 0, len(options))
	for _, option := range options {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(
			strings.ReplaceAll(option.name, "笔", global.Translations[lang]["笔"]),
			CombineInt64AndString(callbackPrefix, option.id),
		))
	}
	return buttons
}

func buildBundleSwitchButtons(lang string, token string, usdtCallback string, trxCallback string) []tgbotapi.InlineKeyboardButton {
	switch token {
	case "TRX":
		return []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData("🔁"+global.Translations[lang]["transaction_plans_usdt_payment"], usdtCallback),
		}
	case "USDT":
		return []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData("🔁"+global.Translations[lang]["transaction_plans_trx_payment"], trxCallback),
		}
	default:
		return nil
	}
}

func buildBundleExtraButtons(lang string, addressTextKey string, addressCallback string, billingCallback string) []tgbotapi.InlineKeyboardButton {
	return []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("🔢"+global.Translations[lang][addressTextKey], addressCallback),
		tgbotapi.NewInlineKeyboardButtonData("📜"+global.Translations[lang]["billing"], billingCallback),
	}
}

func buildSecondaryContactText(lang string, backupChatID string) string {
	if len(backupChatID) == 0 {
		return global.Translations[lang]["secondary_contact_none"]
	}
	return "🔗 " + global.Translations[lang]["secondary_contact"] + "：  @" + backupChatID
}

func buildHomeSummaryText(lang string, user domain.User, chatID int64) string {
	return "🆔 " + global.Translations[lang]["user_id"] + "：<code>" + user.Associates + "</code>\n\n👤 " + global.Translations[lang]["username"] + "：@" + user.Username + "\n\n" +
		buildSecondaryContactText(lang, user.BackupChatID) + "\n\n💰" +
		global.Translations[lang]["balance"] + "：\n" +
		"- TRX：" + user.TronAmount + "\n" +
		"- USDT：" + user.Amount + "\n" +
		"- " + global.Translations[lang]["promotion_income"] + "：" + user.PromotionIncome + " USDT" + "\n\n" +
		global.Translations[lang]["promotion_link"] + ":" + "<code>" + "https://t.me/ushield_bot?start=" + strconv.FormatInt(chatID, 10) + "</code>"
}

func ShowCoinToCoinSwapMenuByChatID(lang string, db *gorm.DB, chatID int64, bot *tgbotapi.BotAPI) {
	sendFixedFloatMenu(lang, chatID, db, bot)
}

func ShowCoinToCoinSwapMenu(lang string, db *gorm.DB, message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	sendFixedFloatMenu(lang, message.Chat.ID, db, bot)
}

func MenuNavigateTronEnergy(lang string, db *gorm.DB, message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		//tgbotapi.NewInlineKeyboardRow(
		//	tgbotapi.NewInlineKeyboardButtonData("🆔我的账户", "click_my_account"),
		//
		//),

		//tgbotapi.NewKeyboardButton("⚡"+global.Translations[lang]["energy_swap"]),
		//tgbotapi.NewKeyboardButton("🖊️"+global.Translations[lang]["transaction_plans"]),
		//tgbotapi.NewKeyboardButton("🤖"+global.Translations[lang]["smart_transaction_plans"]),

		tgbotapi.NewInlineKeyboardRow(
			//tgbotapi.NewInlineKeyboardButtonData("⚡"+global.Translations[lang]["energy_swap"], "click_energy_swap"),
			tgbotapi.NewInlineKeyboardButtonData("🖊️"+global.Translations[lang]["transaction_plans"], "click_transaction_plan"),
			//tgbotapi.NewInlineKeyboardButtonData("🤖"+global.Translations[lang]["smart_transaction_plans"], "click_smart_transaction_plan"),
			tgbotapi.NewInlineKeyboardButtonData("🤖"+global.Translations[lang]["catfee_smart_transaction_menu"], "click_smart_transaction_plan"),
		),
	)

	videoPath := "./static/Dior.png"

	// 创建视频消息（从本地文件）
	msg := tgbotapi.NewPhoto(message.Chat.ID, tgbotapi.FilePath(videoPath))

	msg.Caption = buildEnergySwapText(lang, db)
	msg.ReplyMarkup = inlineKeyboard
	//msg.SupportsStreaming = true // 启用流式播放（推荐）

	//msg := tgbotapi.new(message.Chat.ID, targetStr)
	//msg.ReplyMarkup = inlineKeyboard
	msg.ParseMode = "HTML"
	bot.Send(msg)
	//str := ""
	//if len(user.BackupChatID) > 0 {
	//	//id, _ := strconv.ParseInt(user.BackupChatID, 10, 64)
	//	//backup_user, _ := userRepo.GetByChatID(id)
	//	str = "🔗 " + global.Translations[lang]["secondary_contact"] + "：  " + "@" + user.BackupChatID
	//} else {
	//	str = global.Translations[lang]["secondary_contact_none"]
	//}

	//msg := tgbotapi.NewMessage(message.Chat.ID, "🆔 "+global.Translations[lang]["user_id"]+"："+user.Associates+"\n👤 "+global.Translations[lang]["username"]+"：@"+user.Username+"\n"+
	//	str+"\n💰"+
	//	global.Translations[lang]["balance"]+"：\n"+
	//	"- TRX："+user.TronAmount+"\n"+
	//	"- USDT："+user.Amount)
	//msg.ReplyMarkup = inlineKeyboard
	//msg.ParseMode = "HTML"
	//bot.Send(msg)

	//msg := tgbotapi.NewMessage(message.Chat.ID, "🆔 ID："+user.Associates+"\n👤：@"+user.Username+"\n\n")
	//msg.ReplyMarkup = inlineKeyboard
	//msg.ParseMode = "HTML"
	//bot.Send(msg)
}

func MenuNavigateSwapExchange(lang string, db *gorm.DB, message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	// 当点击"按钮 1"时显示内联键盘
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔃"+global.Translations[lang]["coin_swap_coin_menu"], "coin_swap_coin"),
		),
	)
	//agent := os.Getenv("Agent")
	//sysUserRepo := repositories.NewSysUsersRepository(db)
	//receiveAddress, _, _ := sysUserRepo.GetAddressesByUsername(context.Background(), agent)

	//dictRepo := repositories.NewSysDictionariesRepo(db)
	//receiveAddress, _ := dictRepo.GetReceiveAddress(agent)

	videoPath := "./static/Prada.png"

	// 创建视频消息（从本地文件）
	msg := tgbotapi.NewPhoto(message.Chat.ID, tgbotapi.FilePath(videoPath))

	msg.Caption = buildSwapExchangeText(lang, db)

	//msg := tgbotapi.NewMessage(message.Chat.ID, targetStr)
	msg.ReplyMarkup = inlineKeyboard
	msg.ParseMode = "HTML"
	//msg.DisableWebPagePreview = true
	bot.Send(msg)
}

func MenuNavigateAddressTrace(lang string, cache cache.Cache, bot *tgbotapi.BotAPI, chatID int64, db *gorm.DB) {

	originStr := global.Translations[lang]["address_trace_head_tips"]
	userRepo := repositories.NewUserAddressTraceRepo(db)
	orderlist, _ := userRepo.ListByChatID(context.Background(), chatID)

	var builder strings.Builder
	if len(orderlist) > 0 {

		builder.WriteString("\n")
		//builder.WriteString("\n") // 添加分隔符
		//- [6.29] +3000 TRX（订单 #TOPUP-92308）
		for _, order := range orderlist {
			builder.WriteString("\n") // 添加分隔符
			builder.WriteString("<code>" + order.Address + "</code>")
			builder.WriteString("\n")
			// 添加分隔符
		}

	}

	// 去除最后一个空格
	result := strings.TrimSpace(builder.String())

	//msg := tgbotapi.NewMessage(chatID, "🧾"+global.Translations[lang]["package_address_list"]+"\n"+
	//	result+"\n")

	msg := tgbotapi.NewMessage(chatID, originStr+"\n"+
		result+"\n")
	msg.ParseMode = "HTML"

	// 当点击"按钮 1"时显示内联键盘
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕"+global.Translations[lang]["address_trace_add"], "address_trace_add"),
			tgbotapi.NewInlineKeyboardButtonData("➖"+global.Translations[lang]["address_trace_delete"], "address_trace_delete"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard

	bot.Send(msg)

	expiration := 1 * time.Minute // 短时间缓存空值

	//设置用户状态
	cache.Set(strconv.FormatInt(chatID, 10), "usdt_address_trace", expiration)
}

func MenuNavigateAddressFreeze(lang string, cache cache.Cache, bot *tgbotapi.BotAPI, chatID int64, db *gorm.DB) {

	userRepo := repositories.NewSysDictionariesRepo(db)

	server_trx_price, _ := userRepo.GetDictionaryDetail("server_trx_price")

	server_usdt_price, _ := userRepo.GetDictionaryDetail("server_usdt_price")

	//msg := tgbotapi.NewMessage(chatID, "欢迎使用U盾 USDT冻结预警服务\n"+
	//	"🛡️ U盾，做您链上资产的护盾！\n"+
	//	"地址一旦被链上风控冻，资产将难以追回，损失巨大！\n"+
	//	"每天都有数百个 USDT 钱包地址被冻结锁定，风险就在身边！\n"+
	//	"✅ 适用于经常收付款 / 被制裁地址感染/与诈骗地址交互\n"+
	//	"✅ 支持TRON/ETH网络的USDT 钱包地址\n"+
	//	"📌 服务价格（每地址）：\n • "+server_trx_price+" TRX / 30天\n • "+
	//	" 或 "+server_usdt_price+" USDT / 30天\n"+
	//	"🎯 服务开启后U盾将24 小时不间断保护您的资产安全。\n"+
	//	"⏰ 系统将在冻结前启动预警机制，持续 10 分钟每分钟推送提醒，通知您及时转移资产。\n"+
	//	"📩 所有预警信息将通过 Telegram 实时推送")

	originStr := global.Translations[lang]["usdt_freeze_alert_tips"]

	targetStr := renderMenuTemplate(originStr, map[string]string{
		"{server_usdt_price}": server_usdt_price,
		"{server_trx_price}":  server_trx_price,
	})

	msg := tgbotapi.NewMessage(chatID, targetStr)
	msg.ParseMode = "HTML"

	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚠️"+global.Translations[lang]["enable_freeze_alert"], "start_freeze_risk"),
			//tgbotapi.NewInlineKeyboardButtonData("地址管理", "address_manager"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👁️‍🗨️ "+global.Translations[lang]["alert_monitoring_list"], "address_list_trace"),
			tgbotapi.NewInlineKeyboardButtonData("📜"+global.Translations[lang]["freeze_alert_deduction_record"], "address_freeze_risk_records"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🕸"+global.Translations[lang]["address_trace_menu"], "function_address_trace"),
			tgbotapi.NewInlineKeyboardButtonData("🧽"+global.Translations[lang]["coin_laundering_menu"], "function_remove_label"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard

	bot.Send(msg)

	expiration := 1 * time.Minute // 短时间缓存空值

	//设置用户状态
	cache.Set(strconv.FormatInt(chatID, 10), "usdt_risk_monitor", expiration)
}

func MenuNavigateAddressDetection(lang string, cache cache.Cache, bot *tgbotapi.BotAPI, chatID int64, db *gorm.DB) {
	user := loadMenuUser(db, chatID)

	dictRepo := repositories.NewSysDictionariesRepo(db)

	address_detection_cost, _ := dictRepo.GetDictionaryDetail("address_detection_cost")
	address_detection_cost_usdt, _ := dictRepo.GetDictionaryDetail("address_detection_cost_usdt")

	originStr := global.Translations[lang]["address_check_tips"]

	targetStr := renderMenuTemplate(originStr, map[string]string{
		"{address_detection_cost}":      address_detection_cost,
		"{address_detection_cost_usdt}": address_detection_cost_usdt,
		"{tron_amount}":                 user.TronAmount,
		"{amount}":                      user.Amount,
	})

	msg := tgbotapi.NewMessage(chatID, targetStr)
	//msg := tgbotapi.NewMessage(chatID, " 欢迎使用 U盾地址风险检测\n"+
	//	"✅ 支持 TRON 或 ETH 网络任意地址查询\n"+
	//	"✅ 系统将基于链上行为、风险标签、关联实体进行评分与分析\n📊 风险等级说明：\n"+
	//	"🟢低风险(0–30):无异常交易，未关联已知风险实体\n"+
	//	"🟡中风险(31–70):存在少量高风险交互，对手方不明\n"+
	//	"🟠高风险(71–90):频繁异常转账，或与恶意地址有关\n"+
	//	"🔴极高风险(91–100):涉及诈骗、制裁、黑客、洗钱等高风险行为\n\n"+
	//	"📌 每位用户每天可免费检测 1 次\n"+
	//	"📌 超出后每次扣除 "+address_detection_cost+"TRX 或 "+address_detection_cost_usdt+"USDT（系统将优先扣除 TRX）\n"+
	//	"💰当前余额：\n"+
	//	"- TRX："+user.TronAmount+"\n"+"- USDT："+user.Amount+"\n"+
	//	//"\n🔋 快速充值：\n➡️ 充值TRX\n➡️ 充值USDT\n\n请输入要检测的地址 👇")
	//	"请输入要检测的地址 👇")
	msg.ParseMode = "HTML"
	// 当点击"按钮 1"时显示内联键盘
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💵"+global.Translations[lang]["deposit"], "deposit_amount"),
			tgbotapi.NewInlineKeyboardButtonData("📜"+global.Translations[lang]["address_detection_payment_history"], "user_detection_cost_records"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard

	bot.Send(msg)

	expiration := 1 * time.Minute // 短时间缓存空值

	//设置用户状态
	cache.Set(strconv.FormatInt(chatID, 10), "usdt_risk_query", expiration)
}

func MenuNavigateEnergyExchange(lang string, db *gorm.DB, message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	// 当点击"按钮 1"时显示内联键盘
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🖊️"+global.Translations[lang]["transaction_plans"], "back_bundle_package"),
		),
	)
	msg := tgbotapi.NewMessage(message.Chat.ID, buildEnergySwapText(lang, db))
	msg.ReplyMarkup = inlineKeyboard
	msg.ParseMode = "HTML"
	//msg.DisableWebPagePreview = true
	bot.Send(msg)
}
func MenuNavigateBundlePackage(lang string, db *gorm.DB, chatID int64, bot *tgbotapi.BotAPI, token string) {
	bundlesRepo := repositories.NewUserOperationBundlesRepository(db)

	trxlist, err := bundlesRepo.ListActiveByToken(context.Background(), token)

	if err != nil {

	}

	options := make([]bundleMenuOption, 0, len(trxlist))
	for _, trx := range trxlist {
		options = append(options, bundleMenuOption{id: trx.Id, name: trx.Name})
	}

	inlineKeyboard := buildBundleMenuKeyboard(
		buildBundleOptionButtons(lang, options, "bundle_"),
		buildBundleSwitchButtons(lang, token, "click_switch_usdt", "click_switch_trx"),
		buildBundleExtraButtons(lang, "address_list", "click_bundle_package_address_stats", "click_bundle_package_cost_records"),
	)
	//
	//msg := tgbotapi.NewMessage(chatID,
	//	"欢迎使用 U盾能量笔数套餐\n"+
	//		"一次购买/多地址使用/一键发能/快捷高效\n"+
	//		"⚙️ 功能介绍\n"+
	//		"📍 地址列表\n"+
	//		"    最多可同时管理 4 个接收地址。\n"+
	//		"⚡️ 发能管理\n"+
	//		"自动发能开启后系统会自动检测地址能量余量，不足时自动补充（每次消耗 1 笔），默认关闭，可在“地址列表”中开启/关闭。\n "+
	//		"一键发能：可向地址列表中任意地址或自定义地址快速发放 1 笔能量\n"+
	//		"⏳ 能量有效期 1 小时，过期将自动回收并扣除笔数。\n"+
	//		"🆔"+global.Translations[lang]["user_id"]+": "+user.Associates+"\n"+
	//		"👤"+global.Translations[lang]["username"]+": @"+user.Username+"\n"+
	//		"💰"+global.Translations[lang]["balance"]+": "+"- TRX：   "+user.TronAmount+"   - USDT："+user.Amount)
	//

	//msg := tgbotapi.NewMessage(chatID,
	//	"欢迎使用 U盾能量笔数套餐\n"+
	//		"一次购买/多地址使用/一键发能\n"+
	//		"⚙️ 功能介绍\n"+
	//		"➕添加地址：可添加4个常用地址\n"+
	//		"📍地址列表：可向4个常用地址或向其他地址快速发能\n"+
	//		"⏳ 能量有效期 1 小时，过期将自动回收")

	msg := tgbotapi.NewMessage(chatID, global.Translations[lang]["transaction_plans_tips"])
	msg.ReplyMarkup = inlineKeyboard
	msg.ParseMode = "HTML"

	bot.Send(msg)
}

func MenuNavigateHome(lang string, cache cache.Cache, db *gorm.DB, message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		//tgbotapi.NewInlineKeyboardRow(
		//	tgbotapi.NewInlineKeyboardButtonData("🆔我的账户", "click_my_account"),
		//
		//),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💳"+global.Translations[lang]["deposit"], "deposit_amount"),
			//tgbotapi.NewInlineKeyboardButtonData("🔗第二通知人", "click_backup_account"),
			tgbotapi.NewInlineKeyboardButtonData("📜"+global.Translations[lang]["billing"], "click_my_recepit"),
			tgbotapi.NewInlineKeyboardButtonData("🛎️"+global.Translations[lang]["support"], "click_callcenter"),
			//tgbotapi.NewInlineKeyboardButtonData("🛠️我的服务", "click_my_service"),
		),
		tgbotapi.NewInlineKeyboardRow(
			//	//tgbotapi.NewInlineKeyboardButtonData("🔗绑定备用帐号", "click_backup_account"),
			tgbotapi.NewInlineKeyboardButtonData("👥"+global.Translations[lang]["business"], "click_business_cooperation"),
			tgbotapi.NewInlineKeyboardButtonData("💬"+global.Translations[lang]["channel"], "click_offical_channel"),

			tgbotapi.NewInlineKeyboardButtonData("❓"+global.Translations[lang]["tutorials"], "click_QA"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🌍"+global.Translations[lang]["language"], "click_language"),
		),
		//tgbotapi.NewInlineKeyboardRow(),
	)

	user := loadMenuUser(db, message.Chat.ID)

	msg := tgbotapi.NewMessage(message.Chat.ID, buildHomeSummaryText(lang, user, message.Chat.ID))

	msg.ReplyMarkup = inlineKeyboard
	msg.ParseMode = "HTML"
	bot.Send(msg)
}
func MenuNavigateHome2(db *gorm.DB, message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		//tgbotapi.NewInlineKeyboardRow(
		//	tgbotapi.NewInlineKeyboardButtonData("🆔我的账户", "click_my_account"),
		//
		//),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("中文", "set_lang_"+"zh"),
			tgbotapi.NewInlineKeyboardButtonData("English", "set_lang_"+"en"),
			tgbotapi.NewInlineKeyboardButtonData("粵語", "set_lang_"+"ch"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("한국어", "set_lang_"+"ko"),
			tgbotapi.NewInlineKeyboardButtonData("ภาษาไทย", "set_lang_"+"th"),
			tgbotapi.NewInlineKeyboardButtonData("日本語", "set_lang_"+"ja"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Русский язык", "set_lang_"+"ru"),
			tgbotapi.NewInlineKeyboardButtonData("فارسی", "set_lang_"+"fa"),
			tgbotapi.NewInlineKeyboardButtonData("Español", "set_lang_"+"es"),
		),
	)

	user := loadMenuUser(db, message.Chat.ID)

	msg := tgbotapi.NewMessage(message.Chat.ID, "🆔 ID："+user.Associates+"\n👤：@"+user.Username+"\n\n")
	msg.ReplyMarkup = inlineKeyboard
	msg.ParseMode = "HTML"
	bot.Send(msg)
}

func ShowSmartTransactionBundlePackageMenu(lang string, db *gorm.DB, chatID int64, bot *tgbotapi.BotAPI, token string) {
	//bundlesRepo := repositories.NewUserOperationBundlesRepository(db)
	bundlesRepo := repositories.NewUserSmartTransactionBundlesRepository(db)

	trxlist, err := bundlesRepo.ListActiveByToken(context.Background(), token)

	if err != nil {

	}

	options := make([]bundleMenuOption, 0, len(trxlist))
	for _, trx := range trxlist {
		options = append(options, bundleMenuOption{id: trx.Id, name: trx.Name})
	}

	inlineKeyboard := buildBundleMenuKeyboard(
		buildBundleOptionButtons(lang, options, "ST_bundle_"),
		buildBundleSwitchButtons(lang, token, "click_switch_usdt_ST", "click_switch_trx_ST"),
		buildBundleExtraButtons(lang, "smart_transaction_address_list", "click_bundle_package_address_stats_ST", "click_bundle_package_cost_records_ST"),
	)

	msg := tgbotapi.NewMessage(chatID, "<b>"+global.Translations[lang]["smart_transaction_plans_head"]+"</b>"+"\n\n"+global.Translations[lang]["smart_transaction_plans_tips"])
	msg.ReplyMarkup = inlineKeyboard
	msg.ParseMode = "HTML"

	bot.Send(msg)
}
