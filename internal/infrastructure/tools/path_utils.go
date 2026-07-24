package tools

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	staticDirEnvKey        = "STATIC_DIR"
	qrCodeOutputDirEnvKey  = "QRCODE_OUTPUT_DIR"
	videoCacheFileEnvKey   = "VIDEO_CACHE_FILE_PATH"
	translationsDirEnvKey  = "TRANSLATIONS_DIR"
	defaultStaticDir       = "static"
	defaultTranslationsDir = "translations"
	defaultVideoCacheFile  = "video_file_ids.json"
	qrCodeDirName          = "qrcode"
)

func StaticDir() string {
	return configPathOrDefault(staticDirEnvKey, defaultStaticDir)
}

func StaticFile(name string) string {
	return filepath.Join(StaticDir(), name)
}

func QRCodeOutputDir() string {
	return configPathOrDefault(qrCodeOutputDirEnvKey, filepath.Join(StaticDir(), qrCodeDirName))
}

func VideoCacheFilePath() string {
	return configPathOrDefault(videoCacheFileEnvKey, defaultVideoCacheFile)
}

func TranslationsDir() string {
	return configPathOrDefault(translationsDirEnvKey, defaultTranslationsDir)
}

func configPathOrDefault(envKey string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(envKey))
	if value == "" {
		return fallback
	}
	return filepath.Clean(value)
}
