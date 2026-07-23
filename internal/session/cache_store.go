package session

import (
	"encoding/json"
	"fmt"
	"time"
	"ushield_bot/internal/cache"
)

const sessionKeyPrefix = "SESSION_"

type CacheStore struct {
	cache cache.Cache
	ttl   time.Duration
}

func NewCacheStore(cache cache.Cache, ttl time.Duration) *CacheStore {
	return &CacheStore{
		cache: cache,
		ttl:   ttl,
	}
}

func (s *CacheStore) Get(chatID int64) (*Session, error) {
	data, err := s.cache.Get(s.buildKey(chatID))
	if err != nil || data == "" {
		return nil, err
	}

	var session Session
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, err
	}

	return &session, nil
}

func (s *CacheStore) Set(chatID int64, session *Session) error {
	if session == nil {
		return s.Clear(chatID)
	}

	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	return s.cache.Set(s.buildKey(chatID), string(data), s.ttl)
}

func (s *CacheStore) Clear(chatID int64) error {
	return s.cache.Delete(s.buildKey(chatID))
}

func (s *CacheStore) buildKey(chatID int64) string {
	return fmt.Sprintf("%s%d", sessionKeyPrefix, chatID)
}
