package attachment

import (
	"errors"
	"fmt"
	"io"
	"os"
)

type FileOpener interface {
	Open(string) (*OpenedFile, error)
}

type OpenedFile struct {
	File      io.ReadCloser
	SizeBytes int64
}

type OSFileOpener struct{}

func NewOSFileOpener() *OSFileOpener {
	return &OSFileOpener{}
}

func (o *OSFileOpener) Open(path string) (*OpenedFile, error) {
	if path == "" {
		return nil, errors.New("file path cannot be empty")
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	// Get file info
	info, err := file.Stat()
	if err != nil {
		file.Close() // Ensure the file is closed on error
		return nil, fmt.Errorf("stat %q: %w", path, err)
	}

	return &OpenedFile{File: file, SizeBytes: info.Size()}, nil
}
