package fs

import "fmt"

type MockFS struct {
	DirListing  string
	WrittenFile string
	WrittenData []byte
}

func (m *MockFS) ListDirectory(path string) (string, error) {
	if m.DirListing == "" {
		return "", fmt.Errorf("no directory listing available")
	}
	return m.DirListing, nil
}

func (m *MockFS) WriteFile(path string, data []byte) error {
	m.WrittenFile = path
	m.WrittenData = data
	return nil
}

func (m *MockFS) ReadFile(path string) ([]byte, error) {
	return nil, nil
}
