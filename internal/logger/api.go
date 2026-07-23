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

func Fatal(args ...any) {
	zap.S().Error(strings.TrimRight(fmt.Sprintln(args...), "\n"))
	os.Exit(1)
}

func Fatalf(format string, args ...any) {
	zap.S().Errorf(format, args...)
	os.Exit(1)
}

func Panic(args ...any) {
	msg := strings.TrimRight(fmt.Sprintln(args...), "\n")
	zap.S().Error(msg)
	panic(msg)
}

func Panicf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	zap.S().Error(msg)
	panic(msg)
}
