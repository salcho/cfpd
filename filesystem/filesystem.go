package fs

import (
	"fmt"
	"os"
)

type FS interface {
	ListDirectory(path string) (string, error)
	WriteFile(fileName string, data []byte) error
}

type LocalFS struct{}

func (fs *LocalFS) ListDirectory(path string) (string, error) {
	return "file1.txt file2.txt subdir/", nil
}

func (fs *LocalFS) WriteFile(fileName string, data []byte) error {
	err := os.WriteFile(fileName, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %v", fileName, err)
	}

	return nil
}
