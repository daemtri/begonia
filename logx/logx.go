package logx

import (
	"golang.org/x/exp/slog"
)

type Logger = slog.Logger

func GetLogger(name string) *Logger {
	return slog.New(getLoggerHandler(name)).With("logger", name)
}
