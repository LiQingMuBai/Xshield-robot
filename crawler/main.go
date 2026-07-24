package crawler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

func GetTelegramUserInfo(username string) (string, string, error) {
	// 构建 Telegram Web 链接
	url := fmt.Sprintf("https://t.me/%s", username)

	// 发送 HTTP 请求
	resp, err := http.Get(url)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	// 读取 HTML 内容
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	htmlContent := string(body)

	// 从 meta 标签中提取信息
	titleRegex := regexp.MustCompile(`<meta property="og:title" content="([^"]+)"`)
	titleMatch := titleRegex.FindStringSubmatch(htmlContent)

	if len(titleMatch) < 2 {
		return "", "", fmt.Errorf("未找到用户信息")
	}

	// 解析标题格式: "FirstName LastName (@username)"
	fullTitle := titleMatch[1]

	// 移除用户名部分
	namePart := strings.Split(fullTitle, " (@")[0]

	// 分割姓和名
	names := strings.SplitN(namePart, " ", 2)

	var firstName, lastName string
	if len(names) >= 1 {
		firstName = names[0]
	}
	if len(names) >= 2 {
		lastName = names[1]
	}

	return firstName, lastName, nil
}
