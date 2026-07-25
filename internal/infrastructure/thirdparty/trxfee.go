package thirdparty

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	jsoniter "github.com/json-iterator/go"
)

type TrxfeeClient struct {
	APIKey    string
	APISecret string
	URL       string
}

func NewTrxfeeClient(url, apiKey, apiSecret string) *TrxfeeClient {
	return &TrxfeeClient{
		URL:       url,
		APIKey:    apiKey,
		APISecret: apiSecret,
	}
}

type Data struct {
	EnergyAmount   int    `json:"energy_amount"`
	Period         string `json:"period"`
	ReceiveAddress string `json:"receive_address"`
	CallbackURL    string `json:"callback_url"`
	OutTradeNo     string `json:"out_trade_no"`
}

type AccountDataResp struct {
	Msg  string `json:"msg"`
	Code int    `json:"code"`
	Data struct {
		Balance      float64 `json:"balance"`
		UsdtBalance  float64 `json:"usdtBalance"`
		RechargeAddr string  `json:"rechargeAddr"`
	} `json:"data"`
}

func (c *TrxfeeClient) Account() (resp *AccountDataResp, err error) {
	url := c.URL + "/v1/account"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create trxfee account request failed: %w", err)
	}

	req.Header.Add("API-KEY", c.APIKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send trxfee account request failed: %w", err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read trxfee account response failed: %w", err)
	}

	var accountResp AccountDataResp

	if err := json.Unmarshal(body, &accountResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal account response: %w", err)
	}
	return &accountResp, nil

}

func (c *TrxfeeClient) Order(outTradeNo, receiveAddress string, energyAmount int) error {
	time.Sleep(1 * time.Second)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	energyAmount = 65001

	data := Data{
		EnergyAmount:   energyAmount,
		Period:         "1H",
		ReceiveAddress: receiveAddress,
		CallbackURL:    "",
		OutTradeNo:     outTradeNo,
	}

	ordered_data := map[string]interface{}{
		"energy_amount":   data.EnergyAmount,
		"period":          data.Period,
		"receive_address": data.ReceiveAddress,
		"callback_url":    data.CallbackURL,
		"out_trade_no":    data.OutTradeNo,
	}

	json := jsoniter.ConfigCompatibleWithStandardLibrary
	b, err := json.Marshal(ordered_data)
	if err != nil {
		return fmt.Errorf("marshal trxfee order payload failed: %w", err)
	}
	json_data := string(b)

	message := timestamp + "&" + json_data
	signature := createHmac(message, c.APISecret)

	client := &http.Client{}
	req, err := http.NewRequest("POST", c.URL+"/v1/api", bytes.NewBuffer([]byte(json_data)))
	if err != nil {
		return fmt.Errorf("create trxfee order request failed: %w", err)
	}

	req.Header.Set("API-KEY", c.APIKey)
	req.Header.Set("TIMESTAMP", timestamp)
	req.Header.Set("SIGNATURE", signature)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send trxfee order request failed: %w", err)
	}
	defer resp.Body.Close()

	if _, err := ioutil.ReadAll(resp.Body); err != nil {
		return fmt.Errorf("read trxfee order response failed: %w", err)
	}

	fmt.Printf("trxfee order response: %s\n", resp.Status)
	return nil
}

type TimeOrderData struct {
	RentTimes         int    `json:"rentTimes"`
	RecvAddr          string `json:"recvAddr"`
	FreePause         int    `json:"freePause"`
	ResourceReplenish string `json:"resourceReplenish"`
}

func (c *TrxfeeClient) TimesOrder(receiveAddress string, times int) error {
	time.Sleep(1 * time.Second)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	data := TimeOrderData{
		RentTimes:         times,
		RecvAddr:          receiveAddress,
		FreePause:         2,
		ResourceReplenish: "1",
	}

	ordered_data := map[string]interface{}{
		"rentTimes":         data.RentTimes,
		"recvAddr":          data.RecvAddr,
		"freePause":         data.FreePause,
		"resourceReplenish": data.ResourceReplenish,
	}

	json := jsoniter.ConfigCompatibleWithStandardLibrary
	b, err := json.Marshal(ordered_data)
	if err != nil {
		return fmt.Errorf("marshal trxfee times order payload failed: %w", err)
	}
	json_data := string(b)

	message := timestamp + "&" + json_data
	signature := createHmac(message, c.APISecret)

	client := &http.Client{}
	req, err := http.NewRequest("POST", c.URL+"/v1/timesOrder", bytes.NewBuffer([]byte(json_data)))
	if err != nil {
		return fmt.Errorf("create trxfee times order request failed: %w", err)
	}

	req.Header.Set("API-KEY", c.APIKey)
	req.Header.Set("TIMESTAMP", timestamp)
	req.Header.Set("SIGNATURE", signature)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send trxfee times order request failed: %w", err)
	}
	defer resp.Body.Close()

	if _, err := ioutil.ReadAll(resp.Body); err != nil {
		return fmt.Errorf("read trxfee times order response failed: %w", err)
	}
	return nil
}

func (c *TrxfeeClient) EnableTimesOrder(receiveAddress string) error {
	time.Sleep(1 * time.Second)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	ordered_data := map[string]interface{}{
		"recvAddr": receiveAddress,
	}

	json := jsoniter.ConfigCompatibleWithStandardLibrary
	b, err := json.Marshal(ordered_data)
	if err != nil {
		return fmt.Errorf("marshal trxfee enable times order payload failed: %w", err)
	}
	json_data := string(b)

	message := timestamp + "&" + json_data
	signature := createHmac(message, c.APISecret)

	client := &http.Client{}
	req, err := http.NewRequest("POST", c.URL+"/v1/enableTimesOrder", bytes.NewBuffer([]byte(json_data)))
	if err != nil {
		return fmt.Errorf("create trxfee enable times order request failed: %w", err)
	}

	req.Header.Set("API-KEY", c.APIKey)
	req.Header.Set("TIMESTAMP", timestamp)
	req.Header.Set("SIGNATURE", signature)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send trxfee enable times order request failed: %w", err)
	}
	defer resp.Body.Close()

	if _, err := ioutil.ReadAll(resp.Body); err != nil {
		return fmt.Errorf("read trxfee enable times order response failed: %w", err)
	}
	return nil
}

func createHmac(message string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

func (c *TrxfeeClient) Activation(receiveAddress string) error {
	time.Sleep(1 * time.Second)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	ordered_data := map[string]interface{}{
		"receive_address": receiveAddress,
	}

	json := jsoniter.ConfigCompatibleWithStandardLibrary
	b, err := json.Marshal(ordered_data)
	if err != nil {
		return fmt.Errorf("marshal trxfee activation payload failed: %w", err)
	}
	json_data := string(b)

	message := timestamp + "&" + json_data
	signature := createHmac(message, c.APISecret)

	client := &http.Client{}
	req, err := http.NewRequest("POST", c.URL+"/v1/activation", bytes.NewBuffer([]byte(json_data)))
	if err != nil {
		return fmt.Errorf("create trxfee activation request failed: %w", err)
	}

	req.Header.Set("API-KEY", c.APIKey)
	req.Header.Set("TIMESTAMP", timestamp)
	req.Header.Set("SIGNATURE", signature)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send trxfee activation request failed: %w", err)
	}
	defer resp.Body.Close()

	if _, err := ioutil.ReadAll(resp.Body); err != nil {
		return fmt.Errorf("read trxfee activation response failed: %w", err)
	}
	return nil
}
