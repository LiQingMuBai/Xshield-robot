package fixedfloat

import "testing"

func TestExtractTimeReturnsFalseWhenRequiredFieldMissing(t *testing.T) {
	input := map[string]interface{}{
		"data": map[string]interface{}{
			"time": map[string]interface{}{
				"expiration": 1.0,
				"left":       2.0,
				"reg":        3.0,
			},
		},
	}

	if _, ok := ExtractTime(input); ok {
		t.Fatalf("expected ExtractTime to reject missing update field")
	}
}

func TestExtractFromAndToAcceptsMissingOptionalAmount(t *testing.T) {
	input := map[string]interface{}{
		"data": map[string]interface{}{
			"from": buildAddressInfoMap("FROM", "12.5"),
			"to":   buildAddressInfoMap("TO", nil),
		},
	}

	from, to, ok := ExtractFromAndTo(input)
	if !ok {
		t.Fatalf("expected ExtractFromAndTo to succeed for optional amount")
	}
	if from.Amount == nil || *from.Amount != "12.5" {
		t.Fatalf("expected from amount to be parsed")
	}
	if to.Amount != nil {
		t.Fatalf("expected missing amount to stay nil")
	}
}

func TestMapToResponseReturnsErrorOnInvalidShape(t *testing.T) {
	_, err := MapToResponse(map[string]interface{}{
		"code": 0.0,
		"msg":  "ok",
		"data": map[string]interface{}{},
	})
	if err == nil {
		t.Fatalf("expected MapToResponse to reject incomplete payload")
	}
}

func buildAddressInfoMap(address string, amount interface{}) map[string]interface{} {
	return map[string]interface{}{
		"address": address,
		"alias":   "",
		"amount":  amount,
		"code":    "USDTTRC",
		"coin":    "USDT",
		"name":    "USDT",
		"network": "TRON",
		"tag":     "",
		"tx": map[string]interface{}{
			"amount":        nil,
			"ccyfee":        nil,
			"confirmations": nil,
			"fee":           nil,
			"id":            nil,
			"timeBlock":     nil,
			"timeReg":       nil,
		},
	}
}
