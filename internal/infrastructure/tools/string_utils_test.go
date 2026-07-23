package tools

import (
	"fmt"
	"testing"
	logger "ushield_bot/internal/logger"
)

func TestExtractNumberBeforeBi(t *testing.T) {

	nums, err := ExtractNumberBeforeBi("5笔（15TRX）")
	if err != nil {
		t.Fatal(err)
	}
	logger.Println(nums)
}
