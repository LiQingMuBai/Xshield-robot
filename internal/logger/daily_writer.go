package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type DailyRotateWriter struct {
	mu       sync.Mutex
	filePath string
	file     *os.File
	day      string
}

func NewDailyRotateWriter(filePath string) (*DailyRotateWriter, error) {
	writer := &DailyRotateWriter{
		filePath: filePath,
	}
	if err := writer.rotateIfNeeded(time.Now()); err != nil {
		return nil, err
	}
	return writer, nil
}

func (w *DailyRotateWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.rotateIfNeeded(time.Now()); err != nil {
		return 0, err
	}
	return w.file.Write(p)
}

func (w *DailyRotateWriter) rotateIfNeeded(now time.Time) error {
	currentDay := now.Format("2006-01-02")
	if w.file != nil && w.day == currentDay {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(w.filePath), 0o755); err != nil {
		return err
	}

	if w.file != nil {
		if err := w.file.Close(); err != nil {
			return err
		}
		archivedPath := archivedLogPath(w.filePath, w.day)
		if _, err := os.Stat(w.filePath); err == nil {
			if renameErr := os.Rename(w.filePath, archivedPath); renameErr != nil {
				return renameErr
			}
		}
	}

	file, err := os.OpenFile(w.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}

	w.file = file
	w.day = currentDay
	return nil
}

func archivedLogPath(filePath, day string) string {
	ext := filepath.Ext(filePath)
	name := filePath[:len(filePath)-len(ext)]
	return fmt.Sprintf("%s-%s%s", name, day, ext)
}
