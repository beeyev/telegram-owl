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
	File      io.ReadCloser // todo rename to FileReadCloser
}

func (a *Attachment) Close() error {
	if a == nil || a.File == nil {
		return nil
	}

	return a.File.Close()
}

type Attachments []*Attachment

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
