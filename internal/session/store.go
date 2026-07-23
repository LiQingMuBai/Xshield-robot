package session

import "time"

type Session struct {
	Flow      string
	Step      string
	Lang      string
	Page      int
	Payload   map[string]string
	ExpiredAt time.Time
}

type Store interface {
	Get(chatID int64) (*Session, error)
	Set(chatID int64, session *Session) error
	Clear(chatID int64) error
}
