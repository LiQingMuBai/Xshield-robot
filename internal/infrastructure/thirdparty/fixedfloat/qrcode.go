package fixedfloat

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
	"ushield_bot/internal/infrastructure/tools"

	"github.com/skip2/go-qrcode"
)

// GenerateQRCodeWithTimestamp 生成带时间戳文件名的二维码
// content: 要编码的内容
// size: 二维码图像尺寸（像素）
// 返回生成的文件路径
func GenerateQRCodeWithTimestamp(content string, size int) (string, error) {
	// 生成时间戳文件名，格式：qrcode_20251224141503.png
	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("qrcode_%s.png", timestamp)
	outputPath := filepath.Join(tools.QRCodeOutputDir(), filename)

	if err := os.MkdirAll(tools.QRCodeOutputDir(), 0o755); err != nil {
		return "", err
	}

	err := qrcode.WriteFile(content, qrcode.Medium, size, outputPath)
	if err != nil {
		return "", err
	}

	return outputPath, nil
}
