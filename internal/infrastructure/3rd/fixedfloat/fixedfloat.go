package fixedfloat

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	APIBaseURL = "https://ff.io/api/v2/"
	RespOK     = 0
	TypeFixed  = "fixed"
	TypeFloat  = "float"
)

type FixedFloatAPI struct {
	Key    string
	Secret string
}

// New 创建一个新的 FixedFloatAPI 实例
func New(key, secret string) *FixedFloatAPI {
	return &FixedFloatAPI{
		Key:    key,
		Secret: secret,
	}
}

// sign 对数据进行 HMAC-SHA256 签名（输入为 JSON 字节）
func (api *FixedFloatAPI) sign(data []byte) string {
	mac := hmac.New(sha256.New, []byte(api.Secret))
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}

// req 发送 POST 请求到指定方法
func (api *FixedFloatAPI) req(method string, payload interface{}) (map[string]interface{}, error) {
	url := APIBaseURL + method

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	signature := api.sign(body)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-KEY", api.Key)
	req.Header.Set("X-API-SIGN", signature)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	code, ok := result["code"].(float64) // JSON numbers are float64 in Go
	if !ok {
		return nil, fmt.Errorf("invalid response format: missing or non-numeric 'code'")
	}

	if int(code) != RespOK {
		msg, _ := result["msg"].(string)
		return nil, fmt.Errorf("api error [%d]: %s", int(code), msg)
	}

	fmt.Printf("result: %v\n", result)

	//response, err := MapToResponse(result)
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("'data' field is not an object")
	}

	return data, nil
}

// ccies 获取支持的币种列表
func (api *FixedFloatAPI) Ccies() (map[string]interface{}, error) {
	return api.req("ccies", map[string]interface{}{})
}

// Price 获取报价
func (api *FixedFloatAPI) Price(params map[string]interface{}) (map[string]interface{}, error) {
	return api.req("price", params)
}

// Create 创建订单
func (api *FixedFloatAPI) Create(params map[string]interface{}) (map[string]interface{}, error) {
	return api.req2("create", params)
}

// Order 查询订单状态
func (api *FixedFloatAPI) Order(params map[string]interface{}) (map[string]interface{}, error) {
	return api.req("order", params)
}

// Emergency 紧急取消或处理订单
func (api *FixedFloatAPI) Emergency(params map[string]interface{}) (map[string]interface{}, error) {
	return api.req("emergency", params)
}

// SetEmail 设置订单关联邮箱
func (api *FixedFloatAPI) SetEmail(params map[string]interface{}) (map[string]interface{}, error) {
	return api.req("setEmail", params)
}

// QR 获取订单二维码信息
func (api *FixedFloatAPI) QR(params map[string]interface{}) (map[string]interface{}, error) {
	return api.req("qr", params)
}

// req 发送 POST 请求到指定方法
func (api *FixedFloatAPI) req2(method string, payload interface{}) (map[string]interface{}, error) {
	url := APIBaseURL + method

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	signature := api.sign(body)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-KEY", api.Key)
	req.Header.Set("X-API-SIGN", signature)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	code, ok := result["code"].(float64) // JSON numbers are float64 in Go
	if !ok {
		return nil, fmt.Errorf("invalid response format: missing or non-numeric 'code'")
	}

	if int(code) != RespOK {
		msg, _ := result["msg"].(string)
		return nil, fmt.Errorf("api error [%d]: %s", int(code), msg)
	}

	fmt.Printf("result: %v\n", result)

	//response, err := MapToResponse(result)
	////data, ok := result["data"].(map[string]interface{})
	//if err != nil {
	//	return nil, fmt.Errorf("'data' field is not an object")
	//}

	return result, nil
}
