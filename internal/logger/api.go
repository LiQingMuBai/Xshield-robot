package logger

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
)

func Print(args ...any) {
	zap.S().Info(fmt.Sprint(args...))
}

func Printf(format string, args ...any) {
	zap.S().Infof(format, args...)
}

func Println(args ...any) {
	zap.S().Info(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

func Error(args ...any) {
	zap.S().Error(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

func Errorf(format string, args ...any) {
	zap.S().Errorf(format, args...)
}

func Fatal(args ...any) {
	Error(args...)
	os.Exit(1)
}

func Fatalf(format string, args ...any) {
	Errorf(format, args...)
	os.Exit(1)
}

func Panic(args ...any) {
	Error(args...)
}

func Panicf(format string, args ...any) {
	Errorf(format, args...)
}
