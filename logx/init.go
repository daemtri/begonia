package logx

import (
	"context"
	"fmt"

	"github.com/daemtri/begonia/di/box"
)

func Recover(l *Logger) {
	if r := recover(); r != nil {
		l.Error("发生了panic", "err", r)
	}
}

// Init 日志初始化
func init() {
	box.Provide[Config](Config{}, box.WithFlags("log"), box.WithOverride())
	box.UseInit(func(ctx context.Context) error {
		cfg := box.Invoke[Config](ctx)
		err := InitConfig(cfg)
		if err != nil {
			return fmt.Errorf("初始化日志出错: %w", err)
		}
		return nil
	})
}
