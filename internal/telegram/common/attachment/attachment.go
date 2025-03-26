package attachment

import (
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
	if len(a) == 0 {
		return nil
	}

	for _, attach := range a {
		if err := attach.Close(); err != nil {
			return err
		}
	}

	return nil
}
