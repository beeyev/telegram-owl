package attachment

import (
	"errors"
	"fmt"
	"io"
	"os"
)

// FileOpener abstracts file access for Loader. A successful Open transfers the
// returned file to the caller; an error must not leave an open file behind.
type FileOpener interface {
	Open(string) (*OpenedFile, error)
}

// OpenedFile keeps the stream and its stat result together so Loader does not
// need to reopen or reread a file to enforce size limits.
type OpenedFile struct {
	File      io.ReadCloser
	SizeBytes int64
}

// OSFileOpener reads attachments from the local filesystem.
type OSFileOpener struct{}

// NewOSFileOpener returns a local filesystem opener.
func NewOSFileOpener() *OSFileOpener {
	return &OSFileOpener{}
}

// Open opens and stats path. The caller owns File only when Open succeeds.
func (o *OSFileOpener) Open(path string) (*OpenedFile, error) {
	if path == "" {
		return nil, errors.New("file path cannot be empty")
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		// Preserve both errors: the stat failure is the root cause, while a close
		// failure means cleanup was incomplete.
		return nil, errors.Join(fmt.Errorf("stat %q: %w", path, err), file.Close())
	}

	return &OpenedFile{File: file, SizeBytes: info.Size()}, nil
}
