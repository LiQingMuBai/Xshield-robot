package router

import (
	"strconv"
	"time"
	"ushield_bot/internal/cache"
)

func setShortState(cacheStore cache.Cache, chatID int64, state string) {
	_ = cacheStore.Set(strconv.FormatInt(chatID, 10), state, time.Minute)
}
