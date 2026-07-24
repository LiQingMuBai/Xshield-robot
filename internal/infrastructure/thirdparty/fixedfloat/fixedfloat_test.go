package fixedfloat

import (
	"fmt"
	"testing"
	"time"
	logger "ushield_bot/internal/logger"
)

func TestCcies(t *testing.T) {

	api := New("AtHmGIAucigijgkqaTiOvuGTArkBrm4pparh7V5E", "jDDzTJKmB8jfzhlxfZuXtdNnNQLrSjaGiKg2e4kf")

	// 示例：获取支持的币种
	ccies, err := api.Ccies()
	if err != nil {
		panic(err)
	}
	logger.Println("Supported coins:", ccies)

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
	logger.Println("Order price:", order)
}
func TestOrder(t *testing.T) {

	api := New("AtHmGIAucigijgkqaTiOvuGTArkBrm4pparh7V5E", "jDDzTJKmB8jfzhlxfZuXtdNnNQLrSjaGiKg2e4kf")

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
	logger.Printf("order %v\n", rawMap)

	from, to, ok := ExtractFromAndTo(rawMap)
	if !ok {
		logger.Println("Failed to extract from/to")
		return
	}

	logger.Println("From Address:", from.Address)
	logger.Println("From Amount:", *from.Amount)
	logger.Println("To Address:", to.Address)
	logger.Println("To Amount:", *to.Amount)

	timeInfo, ok := ExtractTime(rawMap)
	if !ok {
		logger.Println("Failed to extract time")
		return
	}

	logger.Printf("Reg (Unix): %.0f\n", timeInfo.Reg)
	logger.Printf("Expiration (Unix): %.0f\n", timeInfo.Expiration)
	logger.Printf("Left: %.0f seconds\n", timeInfo.Left)

	// 转为 time.Time（可读时间）
	regTime := time.Unix(int64(timeInfo.Reg), 0)
	expTime := time.Unix(int64(timeInfo.Expiration), 0)
	logger.Println("Reg Time:", regTime.UTC().Format("2006-01-02 15:04:05 UTC"))
	logger.Println("Expire Time:", expTime.UTC().Format("2006-01-02 15:04:05 UTC"))

	id, status, ok := ExtractIDAndStatus(rawMap)
	if !ok {
		logger.Println("Failed to extract id or status")
		return
	}

	logger.Printf("ID: %s\n", id)
	logger.Printf("Status: %s\n", status)
}
