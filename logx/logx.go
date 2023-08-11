package logx

import (
	"log/slog"
)

type Logger = slog.Logger

func GetLogger(name string) *Logger {
	return slog.New(getLoggerHandler(name)).With("logger", name)
}
