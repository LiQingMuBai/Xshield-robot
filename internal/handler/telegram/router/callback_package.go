package router

import (
	"context"
	"strings"
	"ushield_bot/internal/global"
	"ushield_bot/internal/infrastructure/repositories"
	logger "ushield_bot/internal/logger"
	"ushield_bot/internal/service"
	"ushield_bot/internal/service/catfee"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handlePackageCallback(lang string, callbackQuery *tgbotapi.CallbackQuery, ctx Context) bool {
	switch {
	case strings.HasPrefix(callbackQuery.Data, "set_bundle_package_default_"):
		target := strings.TrimPrefix(callbackQuery.Data, "set_bundle_package_default_")
		userOperationPackageAddressesRepo := repositories.NewUserOperationPackageAddressesRepo(ctx.DB)
		if err := userOperationPackageAddressesRepo.SetDefaultByChatIDAndAddress(context.Background(), callbackQuery.Message.Chat.ID, target); err != nil {
			logger.Printf("set default bundle address err: %v", err)
			return true
		}
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "✅<b>设置默认地址成功 </b>\n")
		msg.ParseMode = "HTML"
		ctx.Bot.Send(msg)
		service.ShowBundlePackageAddressManagement(lang, ctx.Cache, ctx.Bot, callbackQuery.Message.Chat.ID, ctx.DB)
		return true
	case strings.HasPrefix(callbackQuery.Data, "remove_bundle_package_"):
		target := strings.TrimPrefix(callbackQuery.Data, "remove_bundle_package_")
		userOperationPackageAddressesRepo := repositories.NewUserOperationPackageAddressesRepo(ctx.DB)
		if err := userOperationPackageAddressesRepo.DeleteByChatIDAndAddress(context.Background(), callbackQuery.Message.Chat.ID, target); err != nil {
			logger.Printf("remove bundle address err: %v", err)
			return true
		}
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "✅<b>"+global.Translations[lang]["address_deleted_success"]+"</b>\n")
		msg.ParseMode = "HTML"
		ctx.Bot.Send(msg)
		msg2 := service.BuildBundlePackageAddressSummaryMessage(lang, ctx.DB, callbackQuery.Message.Chat.ID)
		ctx.Bot.Send(msg2)
		return true
	case strings.HasPrefix(callbackQuery.Data, "apply_bundle_package_"):
		target := strings.TrimPrefix(callbackQuery.Data, "apply_bundle_package_")
		service.ApplyBundlePackageAddress(lang, target, ctx.Cache, ctx.Bot, callbackQuery.Message, ctx.DB)
		return true
	case strings.HasPrefix(callbackQuery.Data, "config_bundle_package_address_"):
		target := strings.TrimPrefix(callbackQuery.Data, "config_bundle_package_address_")
		service.ShowBundlePackageAddressActions(lang, target, ctx.Cache, ctx.Bot, callbackQuery.Message, ctx.DB)
		return true
	case callbackQuery.Data == "click_switch_trx":
		service.MenuNavigateBundlePackage(lang, ctx.DB, callbackQuery.Message.Chat.ID, ctx.Bot, "TRX")
		return true
	case callbackQuery.Data == "click_switch_usdt":
		service.MenuNavigateBundlePackage(lang, ctx.DB, callbackQuery.Message.Chat.ID, ctx.Bot, "USDT")
		return true
	case callbackQuery.Data == "click_switch_trx_ST":
		service.ShowSmartTransactionBundlePackageMenu(lang, ctx.DB, callbackQuery.Message.Chat.ID, ctx.Bot, "TRX")
		return true
	case callbackQuery.Data == "click_switch_usdt_ST":
		service.ShowSmartTransactionBundlePackageMenu(lang, ctx.DB, callbackQuery.Message.Chat.ID, ctx.Bot, "USDT")
		return true
	case callbackQuery.Data == "back_bundle_package_ST":
		service.ShowSmartTransactionBundlePackageMenu(lang, ctx.DB, callbackQuery.Message.Chat.ID, ctx.Bot, "TRX")
		return true
	case callbackQuery.Data == "back_bundle_package":
		service.MenuNavigateBundlePackage(lang, ctx.DB, callbackQuery.Message.Chat.ID, ctx.Bot, "TRX")
		return true
	case callbackQuery.Data == "click_bundle_package_address_manager_config":
		service.ShowBundlePackageAddressConfigOptions(lang, ctx.Cache, ctx.Bot, callbackQuery.Message.Chat.ID, ctx.DB)
		return true
	case callbackQuery.Data == "click_bundle_package_address_manager_remove":
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[lang]["energy_address_remove_tips"]+"\n")
		msg.ParseMode = "HTML"
		ctx.Bot.Send(msg)
		setShortState(ctx.Cache, callbackQuery.Message.Chat.ID, callbackQuery.Data)
		return true
	case callbackQuery.Data == "click_bundle_package_address_manager_add":
		userOperationPackageAddressesRepo := repositories.NewUserOperationPackageAddressesRepo(ctx.DB)
		list, _ := userOperationPackageAddressesRepo.ListByChatID(context.Background(), callbackQuery.Message.Chat.ID)
		if len(list) >= 4 {
			msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "<b>"+global.Translations[lang]["energy_address_limit_tips"]+"</b>\n")
			msg.ParseMode = "HTML"
			ctx.Bot.Send(msg)
			return true
		}
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "<b>"+global.Translations[lang]["energy_address_limit"]+"</b>\n")
		msg.ParseMode = "HTML"
		ctx.Bot.Send(msg)
		setShortState(ctx.Cache, callbackQuery.Message.Chat.ID, callbackQuery.Data)
		return true
	case callbackQuery.Data == "click_bundle_package_address_stats":
		msg := service.BuildBundlePackageAddressSummaryMessage(lang, ctx.DB, callbackQuery.Message.Chat.ID)
		ctx.Bot.Send(msg)
		return true
	case callbackQuery.Data == "click_bundle_package_address_stats_ST":
		catfee.ShowSmartTransactionAddressStats(lang, ctx.Cache, ctx.DB, callbackQuery.Message.Chat.ID, ctx.Bot)
		return true
	case strings.HasPrefix(callbackQuery.Data, "custody_address_check_"):
		catfee.ToggleCustodyAddressOption(lang, ctx.DB, callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID, callbackQuery.Data, ctx.Bot, ctx.CatfeeClient)
		return true
	case callbackQuery.Data == "next_bundle_package_address_stats":
		service.ShowNextBundlePackageSubscriptionStatsPage(lang, callbackQuery, ctx.DB, ctx.Bot)
		return true
	case callbackQuery.Data == "prev_bundle_package_address_stats":
		service.ShowPrevBundlePackageSubscriptionStatsPage(lang, callbackQuery, ctx.DB, ctx.Bot)
		return true
	case callbackQuery.Data == "click_bundle_package_address_management":
		service.ShowBundlePackageAddressManagement(lang, ctx.Cache, ctx.Bot, callbackQuery.Message.Chat.ID, ctx.DB)
		return true
	case callbackQuery.Data == "click_bundle_package_cost_records":
		msg := service.ExtractBundlePackage(lang, ctx.DB, callbackQuery)
		ctx.Bot.Send(msg)
		return true
	case callbackQuery.Data == "click_bundle_package_cost_records_ST":
		msg := catfee.ExtractBundlePackageST(lang, ctx.DB, callbackQuery)
		ctx.Bot.Send(msg)
		return true
	case callbackQuery.Data == "prev_st_bundle_package_page":
		catfee.ShowPrevBundlePackagePage(lang, callbackQuery, ctx.DB, ctx.Bot)
		return true
	case callbackQuery.Data == "next_st_bundle_package_page":
		catfee.ShowNextBundlePackagePage(lang, callbackQuery, ctx.DB, ctx.Bot)
		return true
	case callbackQuery.Data == "catfee_add_address":
		catfee.PromptCustodyAddressAdd(lang, ctx.Cache, ctx.DB, ctx.Bot, callbackQuery)
		return true
	case callbackQuery.Data == "catfee_remove_address":
		catfee.PromptCustodyAddressRemove(lang, ctx.Cache, ctx.DB, ctx.Bot, callbackQuery)
		return true
	case callbackQuery.Data == "click_bundle_package_management":
		msg := service.ExtractBundlePackage(lang, ctx.DB, callbackQuery)
		ctx.Bot.Send(msg)
		return true
	case callbackQuery.Data == "click_deposit_usdt_records":
		service.ShowUSDTDepositRecords(lang, ctx.DB, callbackQuery, ctx.Bot)
		return true
	case callbackQuery.Data == "click_deposit_trx_records":
		service.ShowTRXDepositRecords(lang, ctx.DB, callbackQuery, ctx.Bot)
		return true
	case callbackQuery.Data == "next_address_detection_page":
		service.ShowNextAddressDetectionPage(lang, callbackQuery, ctx.DB, ctx.Bot)
		return true
	case callbackQuery.Data == "prev_address_detection_page":
		service.ShowPrevAddressDetectionPage(lang, callbackQuery, ctx.DB, ctx.Bot)
		return true
	case callbackQuery.Data == "prev_deposit_usdt_page":
		service.ShowPrevUSDTDepositPage(lang, callbackQuery, ctx.DB, ctx.Bot)
		return true
	case callbackQuery.Data == "prev_deposit_trx_page":
		service.ShowPrevTRXDepositPage(lang, callbackQuery, ctx.DB, ctx.Bot)
		return true
	case callbackQuery.Data == "prev_address_risk_page":
		service.ShowPrevAddressRiskPage(lang, callbackQuery, ctx.DB, ctx.Bot)
		return true
	case callbackQuery.Data == "next_address_risk_page":
		service.ShowNextAddressRiskPage(lang, callbackQuery, ctx.DB, ctx.Bot)
		return true
	case callbackQuery.Data == "next_deposit_usdt_page":
		service.ShowNextUSDTDepositPage(lang, callbackQuery, ctx.DB, ctx.Bot)
		return true
	case callbackQuery.Data == "next_deposit_trx_page":
		service.ShowNextTRXDepositPage(lang, callbackQuery, ctx.DB, ctx.Bot)
		return true
	case callbackQuery.Data == "prev_bundle_package_page":
		service.ShowPrevBundlePackagePage(lang, callbackQuery, ctx.DB, ctx.Bot)
		return true
	case callbackQuery.Data == "next_bundle_package_page":
		service.ShowNextBundlePackagePage(lang, callbackQuery, ctx.DB, ctx.Bot)
		return true
	}
	return false
}
