package fixedfloat

import (
	"fmt"
	"log"
	"testing"
)

func TestGenerateQRCodeWithTimestamp(t *testing.T) {
	content := "TAPH2hzc29WZPpnsfVjnFVGc1YDJs2Audi"
	size := 300

	filename, err := GenerateQRCodeWithTimestamp(content, size)
	if err != nil {
		log.Fatal("生成二维码失败:", err)
	}

	fmt.Printf("✅ 二维码已生成，文件名：%s\n", filename)

}
