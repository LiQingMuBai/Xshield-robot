package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"ushield_bot/internal/cache"
	"ushield_bot/internal/global"
	logger "ushield_bot/internal/logger"
)

func ListRiskAddresses(coin string, address, cookie string) LabeledAddressList {
	url := "https://dashboard.misttrack.io/api/v1/address_graph_analysis?coin=" + coin + "&address=" + address + "&time_filter="
	req, _ := http.NewRequest("GET", url, nil)
	//https://dashboard.misttrack.io/api/v1/address_graph_analysis?coin=ETH&address=0xf510e53ef8da4e45ffa59eb554511a7410e5efd3&time_filter=
	req.Header.Add("accept", "application/json, text/plain, */*")

	req.Header.Add("cookie", cookie)
	req.Header.Add("language", "EN")

	req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36")

	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	var labeledAddressList LabeledAddressList
	if err := json.Unmarshal(body, &labeledAddressList); err != nil { // Parse []byte to go struct pointer
		logger.Error("Can not unmarshal JSON")
	}
	return labeledAddressList
}

type LabeledAddressList struct {
	Success  bool   `json:"success"`
	Msg      string `json:"msg"`
	GraphDic struct {
		NodeList []struct {
			ID    string `json:"id"`
			Label string `json:"label"`
			Title string `json:"title"`
			Layer int    `json:"layer"`
			Addr  string `json:"addr"`
			Track string `json:"track"`
			//Pid       string `json:"pid"`
			Color     string `json:"color,omitempty"`
			Shape     string `json:"shape,omitempty"`
			Expanded  bool   `json:"expanded"`
			Malicious int    `json:"malicious,omitempty"`
			Dex       int    `json:"dex"`
		} `json:"node_list"`
		EdgeList []struct {
			From       string   `json:"from"`
			To         string   `json:"to"`
			Label      string   `json:"label"`
			Val        float64  `json:"val"`
			TxHashList []string `json:"tx_hash_list"`
			TxTime     string   `json:"tx_time"`
			Color      struct {
				Color     string `json:"color"`
				Highlight string `json:"highlight"`
			} `json:"color"`
		} `json:"edge_list"`
		TxCount                 int    `json:"tx_count"`
		FirstTxDatetime         string `json:"first_tx_datetime"`
		LatestTxDatetime        string `json:"latest_tx_datetime"`
		AddressFirstTxDatetime  string `json:"address_first_tx_datetime"`
		AddressLatestTxDatetime string `json:"address_latest_tx_datetime"`
	} `json:"graph_dic"`
	AddressFirstTxDatetime  string `json:"address_first_tx_datetime"`
	AddressLatestTxDatetime string `json:"address_latest_tx_datetime"`
}

type AddressProfile struct {
	Success          bool   `json:"success"`
	Msg              string `json:"msg"`
	Balance          string `json:"balance"`
	TxCount          string `json:"tx_count"`
	FirstTxTime      string `json:"first_tx_time"`
	LastTxTime       string `json:"last_tx_time"`
	TotalReceived    string `json:"total_received"`
	TotalSpent       string `json:"total_spent"`
	ReceivedCount    string `json:"received_count"`
	SpentCount       string `json:"spent_count"`
	TotalReceivedUsd string `json:"total_received_usd"`
	TotalSpentUsd    string `json:"total_spent_usd"`
	BalanceUsd       string `json:"balance_usd"`
}

func GetAddressInfo(symbol string, address, cookie string) (SlowMistAddressInfo, error) {
	url := "https://dashboard.misttrack.io/api/v1/address_risk_analysis?coin=" + symbol + "&address=" + address
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("accept", "application/json, text/plain, */*")
	req.Header.Add("cookie", cookie)
	req.Header.Add("language", "EN")

	req.Header.Add("referer", "https://dashboard.misttrack.io/address/"+symbol+"/"+address)
	req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36")

	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	var addressInfo SlowMistAddressInfo
	if err := json.Unmarshal(body, &addressInfo); err != nil { // Parse []byte to go struct pointer
		logger.Error("Can not unmarshal JSON")
		return addressInfo, err
	}
	return addressInfo, nil
}

func BuildRiskSummaryText(lang string, cache cache.Cache, addressInfo SlowMistAddressInfo) string {
	_item0 := addressInfo.RiskDic.TriangleLevel[0]
	_item1 := addressInfo.RiskDic.TriangleLevel[1]
	_item2 := addressInfo.RiskDic.TriangleLevel[2]

	_text0 := "🔍" + global.Translations[lang]["risk_score"] + ":" + strconv.Itoa(addressInfo.RiskDic.Score)

	if addressInfo.RiskDic.Score <= 30 {
		_text0 += " 🟢" + "\n"
	}
	if addressInfo.RiskDic.Score > 30 && addressInfo.RiskDic.Score <= 70 {
		_text0 += " 🟡" + "\n"
	}
	if addressInfo.RiskDic.Score > 70 && addressInfo.RiskDic.Score <= 90 {
		_text0 += " 🟠" + "\n"
	}
	if addressInfo.RiskDic.Score > 90 {
		_text0 += " 🔴" + "\n"
	}
	_text1 := ""
	_text2 := ""
	_text3 := ""
	_text4 := ""
	if _item0 > 1 {
		_text1 = "⚠️" + global.Translations[lang]["suspected_malicious_address_contact"] + "\n"
	}
	if _item1 > 1 {
		_text2 = "⚠️️" + global.Translations[lang]["confirmed_malicious_address_contact"] + "\n"
	}
	if _item2 > 1 {
		_text3 = "⚠️" + global.Translations[lang]["high_risk_address_contact"] + "\n"
	}

	_banned_item := addressInfo.RiskDic.HackingEvent

	if _banned_item != "" {
		_text4 = "⚠️️" + global.Translations[lang]["sanctioned_entity_association"] + "\n"
	}

	_text6 := "📊 " + global.Translations[lang]["address_overview"] + "\n"

	text := _text0 + _text1 + _text2 + _text3 + _text4 + _text6
	return text
}

type SlowMistAddressInfo struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
	RiskDic struct {
		Score         int    `json:"score"`
		RiskList      []any  `json:"risk_list"`
		TriangleLevel []int  `json:"triangle_level"`
		HackingEvent  string `json:"hacking_event"`
		RiskDetail    []any  `json:"risk_detail"`
		ChkPhishDn    int    `json:"chk_phish_dn"`
		Upgrade       int    `json:"upgrade"`
	} `json:"risk_dic"`
}

func GetAddressProfile(coin string, address, cookie string) AddressProfile {
	url := "https://dashboard.misttrack.io/api/v1/address_overview?coin=" + coin + "&address=" + address
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("accept", "application/json, text/plain, */*")

	req.Header.Add("cookie", cookie)
	req.Header.Add("language", "EN")

	req.Header.Add("referer", "https://dashboard.misttrack.io/address/"+coin+"/"+address)
	req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36")

	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	var addressProfile AddressProfile
	if err := json.Unmarshal(body, &addressProfile); err != nil { // Parse []byte to go struct pointer
		logger.Error("Can not unmarshal JSON")
	}
	return addressProfile
}
