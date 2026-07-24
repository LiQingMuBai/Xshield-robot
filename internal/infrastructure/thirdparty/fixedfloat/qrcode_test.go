package fixedfloat

import (
	"os"
	"testing"
	logger "ushield_bot/internal/logger"
)

func TestGenerateQRCodeWithTimestamp(t *testing.T) {
	content := "TAPH2hzc29WZPpnsfVjnFVGc1YDJs2Audi"
	size := 300

	filename, err := GenerateQRCodeWithTimestamp(content, size)
	if err != nil {
		logger.Fatal("生成二维码失败:", err)
	}
	defer os.Remove(filename)

	if _, statErr := os.Stat(filename); statErr != nil {
		t.Fatalf("二维码文件未生成: %v", statErr)
	}

	logger.Printf("✅ 二维码已生成，文件名：%s\n", filename)

}
