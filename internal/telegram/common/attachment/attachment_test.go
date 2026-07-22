package attachment_test

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/beeyev/telegram-owl/internal/telegram/common/attachment"
)

type trackingReadCloser struct {
	closeErr error
	closed   bool
}

func (c *trackingReadCloser) Read([]byte) (int, error) {
	return 0, io.EOF
}

func (c *trackingReadCloser) Close() error {
	c.closed = true
	return c.closeErr
}

func TestAttachmentsClose_AttemptsEveryClose(t *testing.T) {
	t.Parallel()

	firstErr := errors.New("first close failed")
	secondErr := errors.New("second close failed")
	first := &trackingReadCloser{closeErr: firstErr}
	second := &trackingReadCloser{closeErr: secondErr}
	items := attachment.Attachments{
		{FileName: "first.txt", File: first},
		nil,
		{FileName: "second.txt", File: second},
	}

	err := items.Close()
	require.ErrorIs(t, err, firstErr)
	require.ErrorIs(t, err, secondErr)
	require.ErrorContains(t, err, `close "first.txt"`)
	require.ErrorContains(t, err, `close "second.txt"`)
	assert.True(t, first.closed)
	assert.True(t, second.closed)
}
