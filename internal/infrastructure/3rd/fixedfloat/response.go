package fixedfloat

// Response 是顶层结构
type Response struct {
	Code float64 `json:"code"`
	Msg  string  `json:"msg"`
	Data Data    `json:"data"`
}

type Data struct {
	Back      AddressInfo `json:"back"`
	Email     interface{} `json:"email"` // 未使用，保留
	Emergency Emergency   `json:"emergency"`
	From      AddressInfo `json:"from"`
	ID        string      `json:"id"`
	Status    string      `json:"status"`
	Time      TimeInfo    `json:"time"`
	To        AddressInfo `json:"to"`
	Token     string      `json:"token"`
	Type      string      `json:"type"`
}

type AddressInfo struct {
	Address          string      `json:"address"`
	AddressAlt       interface{} `json:"addressAlt"`
	Alias            string      `json:"alias"`
	Amount           *string     `json:"amount"` // 字符串形式的数字，避免精度问题
	Code             string      `json:"code"`
	Coin             string      `json:"coin"`
	MaxConfirmations int         `json:"maxConfirmations,omitempty"`
	Name             string      `json:"name"`
	Network          string      `json:"network"`
	ReqConfirmations int         `json:"reqConfirmations,omitempty"`
	Tag              string      `json:"tag"`
	TagName          interface{} `json:"tagName"`
	Tx               TxInfo      `json:"tx"`
}

type TxInfo struct {
	Amount        interface{} `json:"amount"`
	CcyFee        interface{} `json:"ccyfee"`
	Confirmations interface{} `json:"confirmations"`
	Fee           interface{} `json:"fee"`
	ID            interface{} `json:"id"`
	TimeBlock     interface{} `json:"timeBlock"`
	TimeReg       interface{} `json:"timeReg"`
}

type Emergency struct {
	Choice string        `json:"choice"`
	Repeat int           `json:"repeat"`
	Status []interface{} `json:"status"`
}

type TimeInfo struct {
	Expiration float64  `json:"expiration"`
	Finish     *float64 `json:"finish"`
	Left       float64  `json:"left"`
	Reg        float64  `json:"reg"`
	Start      *float64 `json:"start"`
	Update     float64  `json:"update"`
}
