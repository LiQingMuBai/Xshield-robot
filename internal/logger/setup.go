package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Setup(logDir string) error {
	if logDir == "" {
		logDir = "."
	}

	logPath := filepath.Join(logDir, "bot.log")
	writer, err := NewDailyRotateWriter(logPath)
	if err != nil {
		return fmt.Errorf("create log writer: %w", err)
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "time"
	encoderConfig.LevelKey = "level"
	encoderConfig.MessageKey = "msg"
	encoderConfig.CallerKey = ""
	encoderConfig.StacktraceKey = ""
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000000")
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	syncer := zapcore.AddSync(&stdoutFileWriter{
		stdout: os.Stdout,
		file:   writer,
	})

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		syncer,
		zapcore.DebugLevel,
	)

	logger := zap.New(core)
	zap.ReplaceGlobals(logger)
	_ = zap.RedirectStdLog(logger)

	log.SetFlags(0)
	log.SetPrefix("")
	return nil
}

type stdoutFileWriter struct {
	stdout *os.File
	file   *DailyRotateWriter
}

func (w *stdoutFileWriter) Write(p []byte) (int, error) {
	if _, err := w.stdout.Write(p); err != nil {
		return 0, err
	}
	return w.file.Write(p)
}
