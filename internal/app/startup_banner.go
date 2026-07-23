package app

import (
	"fmt"
	"strings"
	logger "ushield_bot/internal/logger"
)

const bannerWidth = 68

func printStartupBanner(c *Container) {
	lines := []string{
		"USHIELD ROBOT STARTUP",
		"",
		"              _ooOoo_",
		"             o8888888o",
		"             88\" . \"88",
		"             (| -_- |)",
		"             O\\  =  /O",
		"          ____/`---'\\____",
		"        .'  \\\\|     |//  `.",
		"       /  \\\\|||  :  |||//  \\",
		"      /  _||||| -:- |||||-  \\",
		"      |   | \\\\\\  -  /// |   |",
		"      | \\_|  ''\\---/''  |   |",
		"      \\  .-\\__  `-`  ___/-. /",
		"    ___`. .'  /--.--\\  `. . __",
		" .\"\" '<  `.___\\_<|>_/___.'  >'\"\".",
		" | | :  `- \\`.;`\\ _ /`;.`/ - ` : | |",
		" \\  \\ `-.   \\_ __\\ /__ _/   .-` /  /",
		"====== BUDDHA BLESS | NO BUG | NO DOWN ======",
		"",
		fmt.Sprintf("bot_name   : %s", displayOrDash(c.Config.Bot.Name)),
		fmt.Sprintf("telegram   : @%s", displayOrDash(c.Bot.Self.UserName)),
		fmt.Sprintf("debug_mode : %t", c.Config.Telegram.Debug),
		fmt.Sprintf("timeout    : %ds", c.Config.Telegram.Timeout),
		fmt.Sprintf("lang       : %s", displayOrDash(c.Config.Translation.DefaultLang)),
	}

	border := strings.Repeat("=", bannerWidth)
	logger.Print(border)
	for _, line := range lines {
		logger.Printf("|| %-*s ||", bannerWidth-6, line)
	}
	logger.Print(border)
}

func displayOrDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}
