package fixedfloat

import (
	"fmt"
	"testing"
	"time"
)

func TestCcies(t *testing.T) {

	api := New("AtHmGIAucigijgkqaTiOvuGTArkBrm4pparh7V5E", "jDDzTJKmB8jfzhlxfZuXtdNnNQLrSjaGiKg2e4kf")

	// 示例：获取支持的币种
	ccies, err := api.Ccies()
	if err != nil {
		panic(err)
	}
	fmt.Println("Supported coins:", ccies)

	//// 示例：创建订单
	//params := map[string]interface{}{
	//	"from":      "btc",
	//	"to":        "eth",
	//	"type":      TypeFixed,
	//	"amount":    0.01,
	//	"addressTo": "0xYourEthAddressHere",
	//}
	//order, err := api.Create(params)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println("Order created:", order)

}

func TestPrice(t *testing.T) {

	api := New("AtHmGIAucigijgkqaTiOvuGTArkBrm4pparh7V5E", "jDDzTJKmB8jfzhlxfZuXtdNnNQLrSjaGiKg2e4kf")

	params := map[string]interface{}{
		"type":      TypeFloat,
		"fromCcy":   "USDTTRC",
		"toCcy":     "USDT",
		"amount":    100,
		"direction": "from",
	}
	order, err := api.Price(params)
	if err != nil {
		panic(err)
	}
	fmt.Println("Order price:", order)
}
func TestOrder(t *testing.T) {

	api := New("AtHmGIAucigijgkqaTiOvuGTArkBrm4pparh7V5E", "jDDzTJKmB8jfzhlxfZuXtdNnNQLrSjaGiKg2e4kf")

	// 示例：获取支持的币种
	//ccies, err := api.Ccies()
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println("Supported coins:", ccies)

	// 示例：创建订单
	params := map[string]interface{}{
		"fromCcy":   "USDTTRC",
		"toCcy":     "USDT",
		"type":      TypeFloat,
		"amount":    1234,
		"direction": "from",
		"toAddress": "0xF510e53EF8DA4e45FFA59EB554511a7410E5eFD3",
		"refcode":   "r8ck81xa",
	}
	rawMap, err := api.Create(params)
	if err != nil {
		panic(err)
	}
	fmt.Printf("order %v\n", rawMap)

	from, to, ok := ExtractFromAndTo(rawMap)
	if !ok {
		fmt.Println("Failed to extract from/to")
		return
	}

	fmt.Println("From Address:", from.Address)
	fmt.Println("From Amount:", *from.Amount)
	fmt.Println("To Address:", to.Address)
	fmt.Println("To Amount:", *to.Amount)

	timeInfo, ok := ExtractTime(rawMap)
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

	id, status, ok := ExtractIDAndStatus(rawMap)
	if !ok {
		fmt.Println("Failed to extract id or status")
		return
	}

	fmt.Printf("ID: %s\n", id)
	fmt.Printf("Status: %s\n", status)
}
