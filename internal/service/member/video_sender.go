package member

import (
	"encoding/json"
	"os"
	"sync"
	"ushield_bot/internal/infrastructure/tools"
	logger "ushield_bot/internal/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var videoFileIDCache sync.Map
var videoCacheFilePath = tools.VideoCacheFilePath()

func init() {
	loadVideoFileIDCacheFromDisk()
}

func sendVideoWithCache(
	bot *tgbotapi.BotAPI,
	chatID int64,
	cacheKey string,
	videoPath string,
	caption string,
	replyMarkup tgbotapi.InlineKeyboardMarkup,
) error {
	if cachedFileID, ok := loadVideoFileID(cacheKey); ok {
		cfg := tgbotapi.NewVideo(chatID, tgbotapi.FileID(cachedFileID))
		cfg.Caption = caption
		cfg.ReplyMarkup = replyMarkup
		cfg.SupportsStreaming = true
		if _, err := bot.Send(cfg); err == nil {
			return nil
		} else {
			logger.Errorf("send cached video failed, fallback upload: %v", err)
		}
	}

	cfg := tgbotapi.NewVideo(chatID, tgbotapi.FilePath(videoPath))
	cfg.Caption = caption
	cfg.ReplyMarkup = replyMarkup
	cfg.SupportsStreaming = true

	sent, err := bot.Send(cfg)
	if err != nil {
		return err
	}
	if sent.Video != nil && sent.Video.FileID != "" {
		videoFileIDCache.Store(cacheKey, sent.Video.FileID)
		persistVideoFileIDCache()
	}
	return nil
}

func loadVideoFileID(cacheKey string) (string, bool) {
	value, ok := videoFileIDCache.Load(cacheKey)
	if !ok {
		return "", false
	}
	fileID, ok := value.(string)
	if !ok || fileID == "" {
		return "", false
	}
	return fileID, true
}

func loadVideoFileIDCacheFromDisk() {
	raw, err := os.ReadFile(videoCacheFilePath)
	if err != nil {
		return
	}

	cache := make(map[string]string)
	if err := json.Unmarshal(raw, &cache); err != nil {
		logger.Errorf("load video file id cache failed: %v", err)
		return
	}

	for key, value := range cache {
		if value == "" {
			continue
		}
		videoFileIDCache.Store(key, value)
	}
}

func persistVideoFileIDCache() {
	cache := make(map[string]string)
	videoFileIDCache.Range(func(key, value any) bool {
		cacheKey, ok := key.(string)
		if !ok || cacheKey == "" {
			return true
		}
		fileID, ok := value.(string)
		if !ok || fileID == "" {
			return true
		}
		cache[cacheKey] = fileID
		return true
	})

	raw, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		logger.Errorf("marshal video file id cache failed: %v", err)
		return
	}

	if err := os.WriteFile(videoCacheFilePath, raw, 0o644); err != nil {
		logger.Errorf("persist video file id cache failed: %v", err)
	}
}
