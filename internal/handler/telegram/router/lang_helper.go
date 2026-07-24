package router

import (
	"strconv"
	"time"
	"ushield_bot/internal/cache"
	"ushield_bot/internal/infrastructure/repositories"

	"gorm.io/gorm"
)

const defaultUserLang = "zh"

func languageCacheKey(chatID int64) string {
	return "LANG_" + strconv.FormatInt(chatID, 10)
}

func cacheUserLanguage(cacheStore cache.Cache, chatID int64, lang string) string {
	if len(lang) == 0 {
		lang = defaultUserLang
	}
	_ = cacheStore.Set(languageCacheKey(chatID), lang, 24*time.Hour)
	return lang
}

func resolveUserLanguage(cacheStore cache.Cache, db *gorm.DB, chatID int64) string {
	lang, _ := cacheStore.Get(languageCacheKey(chatID))
	if len(lang) > 0 {
		return lang
	}

	userRepo := repositories.NewUserRepository(db)
	record, err := userRepo.GetByChatID(chatID)
	if err == nil && len(record.Lang) > 0 {
		return cacheUserLanguage(cacheStore, chatID, record.Lang)
	}

	return cacheUserLanguage(cacheStore, chatID, defaultUserLang)
}
