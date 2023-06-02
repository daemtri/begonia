package cfgtable

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func backupFilename(name string) string {
	return filepath.Join(backupFilesPath, name+".json")
}

func loadFromBackupFile(name string) ([]byte, error) {
	mux.Lock()
	defer mux.Unlock()
	if backupFilesPath == "" {
		return nil, os.ErrNotExist
	}
	filename := backupFilename(name)
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

func saveToBackupFile(name string, rd []byte) error {
	if backupFilesPath == "" {
		panic(fmt.Errorf("backup file path is empty"))
	}
	filename := backupFilename(name)
	return os.WriteFile(filename, rd, 0666)
}
