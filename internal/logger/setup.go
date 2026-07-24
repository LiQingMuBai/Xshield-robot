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

	encoder := zapcore.NewConsoleEncoder(encoderConfig)
	fileCore := zapcore.NewCore(
		encoder,
		zapcore.AddSync(writer),
		zapcore.DebugLevel,
	)
	stderrCore := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stderr),
		zapcore.ErrorLevel,
	)

	logger := zap.New(zapcore.NewTee(fileCore, stderrCore))
	zap.ReplaceGlobals(logger)
	_ = zap.RedirectStdLog(logger)

	log.SetFlags(0)
	log.SetPrefix("")
	return nil
}
