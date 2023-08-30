package cfgtable

import (
	"context"
	"os"

	"github.com/daemtri/begonia/di/box"
)

type Context struct {
	changes map[ConfigInterface]bool
}

func (ctx Context) IsChanged(c ConfigInterface) bool {
	return ctx.changes[c]
}

type Aggregation interface {
	OnLoad(Context)
}

func Init[T Aggregation]() {
	defaultDackupDir := "./cfgtable-files"
	stats, err := os.Stat("./configtables-backups")
	if err == nil && stats.IsDir() {
		defaultDackupDir = "./configtables-backups"
	}
	box.FlagSet("cfgtbale").StringVar(&backupFilesPath, "backup_dir", defaultDackupDir, "配置文件目录")
	builder, err := NewRemoteBuilder[T]()
	if err != nil {
		panic(err)
	}
	box.Provide[T](builder, box.WithFlags("cfgtable"))
	box.UseInit(func(ctx context.Context) error {
		_ = os.MkdirAll(backupFilesPath, os.ModePerm)
		_ = box.Invoke[T](ctx)
		return nil
	})
}
