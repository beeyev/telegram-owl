package attachment

import (
	"errors"
	"fmt"
	"io"
)

type Attachment struct {
	AType     AType
	FileName  string
	SizeBytes int64
	File      io.ReadCloser // Owned by Attachment until Close is called.
}

// Close releases the underlying file. Nil attachments are accepted so callers
// can clean partially assembled collections without special cases.
func (a *Attachment) Close() error {
	if a == nil || a.File == nil {
		return nil
	}

	return a.File.Close()
}

// Attachments owns a set of files returned by Loader.
type Attachments []*Attachment

// Close attempts every close and joins failures with their file names. It does
// not stop at the first error because every remaining file still needs cleanup.
func (a Attachments) Close() error {
	closeErrors := make([]error, 0, len(a))

	for _, attach := range a {
		if err := attach.Close(); err != nil {
			fileName := "<unknown>"
			if attach != nil && attach.FileName != "" {
				fileName = attach.FileName
			}
			closeErrors = append(closeErrors, fmt.Errorf("close %q: %w", fileName, err))
		}
	}

	return errors.Join(closeErrors...)
}
