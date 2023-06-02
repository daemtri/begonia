package filepathx

import (
	"os"
	"path/filepath"
)

func UserHomePath(path string) string {
	hp, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return filepath.Join(hp, path)
}
