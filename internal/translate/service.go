package translate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"ushield_bot/internal/config"
	"ushield_bot/internal/global"
)

type Service struct {
	defaultLang  string
	mutex        sync.RWMutex
	translations map[string]map[string]string
}

func New(cfg config.TranslationConfig) (*Service, error) {
	service := &Service{
		defaultLang:  cfg.DefaultLang,
		translations: make(map[string]map[string]string),
	}

	for _, lang := range cfg.SupportedLangs {
		filePath := filepath.Join(cfg.Dir, lang+".json")
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var langTranslations map[string]string
		if err := json.Unmarshal(data, &langTranslations); err != nil {
			return nil, fmt.Errorf("parse translation file %s: %w", filePath, err)
		}
		service.translations[lang] = langTranslations
	}

	if _, exists := service.translations[cfg.DefaultLang]; !exists {
		return nil, fmt.Errorf("default language %s not found in translations", cfg.DefaultLang)
	}

	global.Mutex.Lock()
	global.TranslationsDir = cfg.Dir
	global.SupportedLangs = append([]string(nil), cfg.SupportedLangs...)
	global.DefaultLang = cfg.DefaultLang
	global.Translations = service.translations
	global.Mutex.Unlock()

	return service, nil
}

func (s *Service) T(lang, key string) string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if langTranslations, exists := s.translations[lang]; exists {
		if value, exists := langTranslations[key]; exists {
			return value
		}
	}

	if lang != s.defaultLang {
		if value, exists := s.translations[s.defaultLang][key]; exists {
			return value
		}
	}

	return key
}

func (s *Service) TParam(lang, key string, params map[string]string) string {
	text := s.T(lang, key)
	for paramKey, value := range params {
		placeholder := "{" + paramKey + "}"
		text = strings.ReplaceAll(text, placeholder, value)
	}
	return text
}
