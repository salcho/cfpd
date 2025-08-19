package fs

type FS interface {
	ListDirectory(path string) (string, error)
}

type LocalFS struct{}

func (fs *LocalFS) ListDirectory(path string) (string, error) {
	return "file1.txt file2.txt subdir/", nil
}
