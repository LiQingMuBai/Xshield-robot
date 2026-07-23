package service

import (
	"context"
	"fmt"
	"strings"
	"ushield_bot/internal/cache"
	"ushield_bot/internal/domain"
	"ushield_bot/internal/global"
	"ushield_bot/internal/handler"
	"ushield_bot/internal/infrastructure/repositories"
	. "ushield_bot/internal/infrastructure/tools"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func ExtractSlowMistRiskQuery(_lang string, cache cache.Cache, message *tgbotapi.Message, db *gorm.DB, _cookie string, bot *tgbotapi.BotAPI) {
	if IsValidAddress(message.Text) || IsValidEthereumAddress(message.Text) {
		userRepo := repositories.NewUserRepository(db)
		user, _ := userRepo.GetByUserID(message.Chat.ID)
		//if strings.Contains(message.Chat.UserName, "Ushield") {
		//	user.Times = 10000
		//}

		if user.Times >= 0 {
			dictRepo := repositories.NewSysDictionariesRepo(db)
			address_detection_cost_trx, _ := dictRepo.GetDictionaryDetail("address_detection_cost")
			address_detection_cost_usdt, _ := dictRepo.GetDictionaryDetail("address_detection_cost_usdt")
			feedback := ""
			//需要扣钱 4trx或者1u
			if CompareStringsWithFloat(user.Amount, address_detection_cost_usdt, 1) || CompareStringsWithFloat(user.TronAmount, address_detection_cost_trx, 1) {

				_text := ""
				if strings.HasPrefix(message.Text, "0x") && len(message.Text) == 42 {
					_symbol := "USDT-ERC20"
					_addressInfo, err := handler.GetAddressInfo(_symbol, message.Text, _cookie)

					fmt.Printf("addressInfo: %v\n", _addressInfo)
					if err != nil || !_addressInfo.Success {
						_text = "目前我们的AI大数据态势感知系统正在对区块链地址风险进行高强度的实时分析，导致“地址风险监测”功能出现短暂的“思考卡顿”（临时失效）。"
					} else {
						_text = handler.GetText(_lang, cache, _addressInfo)

						addressProfile := handler.GetAddressProfile(_symbol, message.Text, _cookie)
						_text7 := global.Translations[_lang]["balance"] + "：" + addressProfile.BalanceUsd + "\n"
						_text8 := global.Translations[_lang]["total_received"] + "：" + addressProfile.TotalReceivedUsd + "\n"
						_text9 := global.Translations[_lang]["total_spent"] + "：" + addressProfile.TotalSpentUsd + "\n"
						_text10 := global.Translations[_lang]["first_tx_time"] + "：" + addressProfile.FirstTxTime + "\n"
						_text11 := global.Translations[_lang]["last_tx_time"] + "：" + addressProfile.LastTxTime + "\n"
						_text12 := global.Translations[_lang]["tx_count"] + "：" + addressProfile.TxCount + "\n"
						_text99 := global.Translations[_lang]["counterparty_analysis"] + "：" + "\n"
						//_text5 := "📢更多查询请联系客服 @Ushield001\n"
						_text16 := global.Translations[_lang]["ushield_tips"] + "\n"
						_text100 := ""
						lableAddresList := handler.GetNotSafeAddress("ETH", message.Text, _cookie)
						if len(lableAddresList.GraphDic.NodeList) > 0 {
							for _, data := range lableAddresList.GraphDic.NodeList {
								if strings.Contains(data.Label, "huione") {
									_text100 = _text100 + data.Title[0:5] + "..." + data.Title[29:34] + global.Translations[_lang]["huione"] + "\n"
								}
								if strings.Contains(data.Label, "Theft") {
									_text100 = _text100 + data.Title[0:5] + "..." + data.Title[29:34] + global.Translations[_lang]["theft"] + "\n"
								}
								if strings.Contains(data.Label, "Drainer") {
									_text100 = _text100 + data.Title[0:5] + "..." + data.Title[29:34] + global.Translations[_lang]["scam"] + "\n"
								}
								if strings.Contains(data.Label, "Banned") {
									_text100 = _text100 + data.Title[0:5] + "..." + data.Title[29:34] + global.Translations[_lang]["sanctioned"] + "\n"
								}
							}
						}

						_text = _text + _text7 + _text8 + _text9 + _text10 + _text11 + _text12 + _text99 + _text100 + _text16

					}
				}
				if strings.HasPrefix(message.Text, "T") && len(message.Text) == 34 {
					_symbol := "USDT-TRC20"
					_addressInfo, err := handler.GetAddressInfo(_symbol, message.Text, _cookie)
					fmt.Printf("addressInfo: %v\n", _addressInfo)
					if err != nil || !_addressInfo.Success {
						_text = "目前我们的AI大数据态势感知系统正在对区块链地址风险进行高强度的实时分析，导致“地址风险监测”功能出现短暂的“思考卡顿”（临时失效）。"
					} else {
						_text = handler.GetText(_lang, cache, _addressInfo)

						addressProfile := handler.GetAddressProfile(_symbol, message.Text, _cookie)
						_text7 := global.Translations[_lang]["balance"] + "：" + addressProfile.BalanceUsd + "\n"
						_text8 := global.Translations[_lang]["total_received"] + "：" + addressProfile.TotalReceivedUsd + "\n"
						_text9 := global.Translations[_lang]["total_spent"] + "：" + addressProfile.TotalSpentUsd + "\n"
						_text10 := global.Translations[_lang]["first_tx_time"] + "：" + addressProfile.FirstTxTime + "\n"
						_text11 := global.Translations[_lang]["last_tx_time"] + "：" + addressProfile.LastTxTime + "\n"
						_text12 := global.Translations[_lang]["tx_count"] + "：" + addressProfile.TxCount + "\n"
						_text99 := global.Translations[_lang]["counterparty_analysis"] + "：" + "\n"
						lableAddresList := handler.GetNotSafeAddress(_symbol, message.Text, _cookie)

						_text100 := ""
						if len(lableAddresList.GraphDic.NodeList) > 0 {
							for _, data := range lableAddresList.GraphDic.NodeList {
								if strings.Contains(data.Label, "huione") {
									_text100 = _text100 + data.Title[0:5] + "..." + data.Title[29:34] + global.Translations[_lang]["huione"] + "\n"
								}
								if strings.Contains(data.Label, "Theft") {
									_text100 = _text100 + data.Title[0:5] + "..." + data.Title[29:34] + global.Translations[_lang]["theft"] + "\n"
								}
								if strings.Contains(data.Label, "Drainer") {
									_text100 = _text100 + data.Title[0:5] + "..." + data.Title[29:34] + global.Translations[_lang]["scam"] + "\n"
								}
								if strings.Contains(data.Label, "Banned") {
									_text100 = _text100 + data.Title[0:5] + "..." + data.Title[29:34] + global.Translations[_lang]["sanctioned"] + "\n"
								}
							}
						}
						//_text5 := "📢更多查询请联系客服 @Ushield001\n"
						_text16 := global.Translations[_lang]["ushield_tips"] + "\n"

						_text = _text + _text7 + _text8 + _text9 + _text10 + _text11 + _text12 + _text99 + _text100 + _text16

					}
				}
				msg := tgbotapi.NewMessage(message.Chat.ID, _text)
				inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("🔍"+global.Translations[_lang]["detect_again"], "back_address_detection_home"),
					),
				)
				msg.ReplyMarkup = inlineKeyboard
				msg.ParseMode = "HTML"
				bot.Send(msg)
				userRepo.UpdateTimesByChatID(1, message.Chat.ID)

				if !strings.Contains(_text, "AI大数据") {
					if CompareStringsWithFloat(user.TronAmount, address_detection_cost_trx, 1) {
						tronAmount, _ := SubtractStringNumbers(user.TronAmount, address_detection_cost_trx, 1)
						user.TronAmount = tronAmount
						err := userRepo.Update2(context.Background(), &user)
						if err != nil {
							fmt.Println("错误： ", err)
						}

						userAddressDetectionRepo := repositories.NewUserAddressDetectionRepository(db)
						var record domain.UserAddressDetection
						record.Status = 1
						record.Amount = address_detection_cost_trx
						record.ChatID = message.Chat.ID
						record.Address = message.Text
						userAddressDetectionRepo.Create(context.Background(), &record)

						feedback = "✅" + "🧾" + global.Translations[_lang]["address_detection_payment_tips"] + address_detection_cost_trx + " TRX \n\n"

					} else if CompareStringsWithFloat(user.Amount, address_detection_cost_usdt, 1) {
						amount, _ := SubtractStringNumbers(user.Amount, address_detection_cost_usdt, 1)
						user.Amount = amount
						err := userRepo.Update2(context.Background(), &user)
						if err != nil {
							fmt.Println("错误： ", err)
						}

						userAddressDetectionRepo := repositories.NewUserAddressDetectionRepository(db)

						var record domain.UserAddressDetection
						record.Status = 1
						record.Amount = address_detection_cost_usdt
						record.ChatID = message.Chat.ID
						record.Address = message.Text
						userAddressDetectionRepo.Create(context.Background(), &record)

						feedback = "✅" + "🧾" + global.Translations[_lang]["address_detection_payment_tips"] + address_detection_cost_usdt + " USDT \n\n"

					}
					msg2 := tgbotapi.NewMessage(message.Chat.ID, feedback)
					msg2.ParseMode = "HTML"
					bot.Send(msg2)
				}

			} else {
				//msg := tgbotapi.NewMessage(message.Chat.ID,
				//	"🔍普通用戶每日赠送 1 次地址风险查询\n"+
				//		"📞聯繫客服 @Ushield001\n")
				//msg.ReplyMarkup = inlineKeyboard

				msg := tgbotapi.NewMessage(message.Chat.ID,
					"<b>"+"🔍"+global.Translations[_lang]["daily_free_limit"]+"</b>"+"\n"+
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
					//tgbotapi.NewInlineKeyboardRow(
					//	tgbotapi.NewInlineKeyboardButtonData("🔙"+global.Translations[_lang]["back_home"], "back_home"),
					//),
				)

				msg.ReplyMarkup = inlineKeyboard
				//bot.Send(msg)

				msg.ParseMode = "HTML"
				bot.Send(msg)
			}
		} else {
			_text := ""
			if strings.HasPrefix(message.Text, "0x") && len(message.Text) == 42 {
				_symbol := "USDT-ERC20"
				_addressInfo, err := handler.GetAddressInfo(_symbol, message.Text, _cookie)
				fmt.Printf("addressInfo: %v\n", _addressInfo)
				if err != nil || !_addressInfo.Success {
					_text = "目前我们的AI大数据态势感知系统正在对区块链地址风险进行高强度的实时分析，导致“地址风险监测”功能出现短暂的“思考卡顿”（临时失效）。"
				} else {
					_text = handler.GetText(_lang, cache, _addressInfo)

					addressProfile := handler.GetAddressProfile(_symbol, message.Text, _cookie)
					_text7 := global.Translations[_lang]["balance"] + "：" + addressProfile.BalanceUsd + "\n"
					_text8 := global.Translations[_lang]["total_received"] + "：" + addressProfile.TotalReceivedUsd + "\n"
					_text9 := global.Translations[_lang]["total_spent"] + "：" + addressProfile.TotalSpentUsd + "\n"
					_text10 := global.Translations[_lang]["first_tx_time"] + "：" + addressProfile.FirstTxTime + "\n"
					_text11 := global.Translations[_lang]["last_tx_time"] + "：" + addressProfile.LastTxTime + "\n"
					_text12 := global.Translations[_lang]["tx_count"] + "：" + addressProfile.TxCount + "\n"
					_text99 := global.Translations[_lang]["counterparty_analysis"] + "：" + "\n"
					//_text5 := "📢更多查询请联系客服 @Ushield001\n"
					_text16 := global.Translations[_lang]["ushield_tips"] + "\n"
					_text100 := ""
					lableAddresList := handler.GetNotSafeAddress("ETH", message.Text, _cookie)
					if len(lableAddresList.GraphDic.NodeList) > 0 {
						for _, data := range lableAddresList.GraphDic.NodeList {
							if strings.Contains(data.Label, "huione") {
								_text100 = _text100 + data.Title[0:5] + "..." + data.Title[29:34] + global.Translations[_lang]["huione"] + "\n"
							}
							if strings.Contains(data.Label, "Theft") {
								_text100 = _text100 + data.Title[0:5] + "..." + data.Title[29:34] + global.Translations[_lang]["theft"] + "\n"
							}
							if strings.Contains(data.Label, "Drainer") {
								_text100 = _text100 + data.Title[0:5] + "..." + data.Title[29:34] + global.Translations[_lang]["scam"] + "\n"
							}
							if strings.Contains(data.Label, "Banned") {
								_text100 = _text100 + data.Title[0:5] + "..." + data.Title[29:34] + global.Translations[_lang]["sanctioned"] + "\n"
							}
						}
					}
					_text = _text + _text7 + _text8 + _text9 + _text10 + _text11 + _text12 + _text99 + _text100 + _text16

				}
			}
			if strings.HasPrefix(message.Text, "T") && len(message.Text) == 34 {
				_symbol := "USDT-TRC20"
				_addressInfo, err := handler.GetAddressInfo(_symbol, message.Text, _cookie)
				fmt.Printf("addressInfo: %v\n", _addressInfo)
				if err != nil || !_addressInfo.Success {
					_text = "目前我们的AI大数据态势感知系统正在对区块链地址风险进行高强度的实时分析，导致“地址风险监测”功能出现短暂的“思考卡顿”（临时失效）。"
				} else {
					_text = handler.GetText(_lang, cache, _addressInfo)

					addressProfile := handler.GetAddressProfile(_symbol, message.Text, _cookie)
					_text7 := global.Translations[_lang]["balance"] + "：" + addressProfile.BalanceUsd + "\n"
					_text8 := global.Translations[_lang]["total_received"] + "：" + addressProfile.TotalReceivedUsd + "\n"
					_text9 := global.Translations[_lang]["total_spent"] + "：" + addressProfile.TotalSpentUsd + "\n"
					_text10 := global.Translations[_lang]["first_tx_time"] + "：" + addressProfile.FirstTxTime + "\n"
					_text11 := global.Translations[_lang]["last_tx_time"] + "：" + addressProfile.LastTxTime + "\n"
					_text12 := global.Translations[_lang]["tx_count"] + "：" + addressProfile.TxCount + "\n"
					_text99 := global.Translations[_lang]["counterparty_analysis"] + "：" + "\n"
					lableAddresList := handler.GetNotSafeAddress(_symbol, message.Text, _cookie)

					_text100 := ""
					if len(lableAddresList.GraphDic.NodeList) > 0 {
						for _, data := range lableAddresList.GraphDic.NodeList {
							if strings.Contains(data.Label, "huione") {
								_text100 = _text100 + data.Title[0:5] + "..." + data.Title[29:34] + global.Translations[_lang]["huione"] + "\n"
							}
							if strings.Contains(data.Label, "Theft") {
								_text100 = _text100 + data.Title[0:5] + "..." + data.Title[29:34] + global.Translations[_lang]["theft"] + "\n"
							}
							if strings.Contains(data.Label, "Drainer") {
								_text100 = _text100 + data.Title[0:5] + "..." + data.Title[29:34] + global.Translations[_lang]["scam"] + "\n"
							}
							if strings.Contains(data.Label, "Banned") {
								_text100 = _text100 + data.Title[0:5] + "..." + data.Title[29:34] + global.Translations[_lang]["sanctioned"] + "\n"
							}
						}
					}
					//_text5 := "📢更多查询请联系客服 @Ushield001\n"
					_text16 := global.Translations[_lang]["ushield_tips"] + "\n"

					_text = _text + _text7 + _text8 + _text9 + _text10 + _text11 + _text12 + _text99 + _text100 + _text16

				}
			}
			msg := tgbotapi.NewMessage(message.Chat.ID, _text)
			inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("🔍"+global.Translations[_lang]["detect_again"], "back_address_detection_home"),
				),
			)
			msg.ReplyMarkup = inlineKeyboard
			msg.ParseMode = "HTML"
			bot.Send(msg)
			userRepo.UpdateTimesByChatID(1, message.Chat.ID)
		}

	} else {
		msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[_lang]["address_wrong_tips"]+"</b>"+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)
	}
}
