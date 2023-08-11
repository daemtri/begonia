package cfgtable

import (
	"errors"
	"sync"

	"log/slog"
)

var (
	// backupFilesPath 配置备份文件路径
	backupFilesPath string
	// 保护文件备份
	mux sync.Mutex

	ErrKeyLength = errors.New("a 16 or 24 or 32 length secret key is required")

	log = slog.Default()

	// defaultSecret 默认文件密钥
	defaultSecret string
)
