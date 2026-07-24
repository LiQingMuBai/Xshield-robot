package tools

import (
	"path/filepath"
	"testing"
)

func TestStaticFileDefaultsToStaticDir(t *testing.T) {
	t.Setenv(staticDirEnvKey, "")

	if got := StaticFile("Audi.png"); got != filepath.Join(defaultStaticDir, "Audi.png") {
		t.Fatalf("unexpected static file path: %s", got)
	}
}

func TestQRCodeOutputDirSupportsEnvOverride(t *testing.T) {
	t.Setenv(qrCodeOutputDirEnvKey, "./runtime/qrcode")

	if got := QRCodeOutputDir(); got != filepath.Clean("./runtime/qrcode") {
		t.Fatalf("unexpected qr code dir: %s", got)
	}
}

func TestTranslationsDirSupportsEnvOverride(t *testing.T) {
	t.Setenv(translationsDirEnvKey, "./i18n")

	if got := TranslationsDir(); got != filepath.Clean("./i18n") {
		t.Fatalf("unexpected translations dir: %s", got)
	}
}
