package logx

import (
	"os"

	"golang.org/x/exp/slog"
)

func GetLogger(name string) *Logger {
	l := &slog.LevelVar{}
	l.Set(slog.LevelDebug)
	return slog.New(slog.HandlerOptions{
		AddSource: true,
		Level:     l,
	}.NewJSONHandler(os.Stdout)).With("logger", name)
}
