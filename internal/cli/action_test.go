package cli //nolint:testpackage // Verify resource ownership at the unexported action boundary.

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/beeyev/telegram-owl/internal/telegram"
	"github.com/beeyev/telegram-owl/internal/telegram/common/attachment"
)

type singleCloseReader struct {
	*strings.Reader

	closeCalls int
	firstErr   error
	secondErr  error
}

func (r *singleCloseReader) Close() error {
	r.closeCalls++
	if r.closeCalls == 1 {
		return r.firstErr
	}

	return r.secondErr
}

type fixedFileOpener struct {
	file io.ReadCloser
	size int64
}

func (o fixedFileOpener) Open(string) (*attachment.OpenedFile, error) {
	return &attachment.OpenedFile{
		File:      o.file,
		SizeBytes: o.size,
	}, nil
}

func TestActionSendMediaGroup_ClosesAttachmentOnce(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.Copy(io.Discard, r.Body)
		assert.NoError(t, err)

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write([]byte(`{"ok": true}`))
		assert.NoError(t, err)
	}))
	t.Cleanup(server.Close)

	client, err := telegram.NewClient(server.URL, "token", "")
	require.NoError(t, err)

	file := &singleCloseReader{
		Reader:    strings.NewReader("attachment content"),
		secondErr: errors.New("attachment closed twice"),
	}
	size := int64(file.Len())
	a := &action{
		ctx:    t.Context(),
		client: client,
		attachLoader: &attachment.Loader{
			FileOpener:                  fixedFileOpener{file: file, size: size},
			MaxTotalAttachments:         1,
			MaxPhotoAttachmentSizeBytes: size,
			MaxAttachmentSizeBytes:      size,
			MaxTotalSizeBytes:           size,
		},
		chatID:           "123",
		attachmentsPaths: []string{"attachment.txt"},
	}

	err = a.sendMediaGroup("")
	require.NoError(t, err)
	assert.Equal(t, 1, file.closeCalls)
}

func TestActionExecute_ContinuesAfterPostSendCloseError(t *testing.T) {
	t.Parallel()

	var requestPaths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPaths = append(requestPaths, r.URL.Path)
		_, err := io.Copy(io.Discard, r.Body)
		assert.NoError(t, err)

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write([]byte(`{"ok": true}`))
		assert.NoError(t, err)
	}))
	t.Cleanup(server.Close)

	client, err := telegram.NewClient(server.URL, "token", "")
	require.NoError(t, err)

	closeErr := errors.New("close failed")
	file := &singleCloseReader{
		Reader:   strings.NewReader("attachment content"),
		firstErr: closeErr,
	}
	var warning bytes.Buffer
	size := int64(file.Len())
	a := &action{
		ctx:           t.Context(),
		client:        client,
		warningWriter: &warning,
		attachLoader: &attachment.Loader{
			FileOpener:                  fixedFileOpener{file: file, size: size},
			MaxTotalAttachments:         1,
			MaxPhotoAttachmentSizeBytes: size,
			MaxAttachmentSizeBytes:      size,
			MaxTotalSizeBytes:           size,
		},
		chatID:           "123",
		message:          strings.Repeat("a", 1025),
		attachmentsPaths: []string{"attachment.txt"},
	}

	err = a.execute()
	require.NoError(t, err)
	assert.Equal(t, []string{"/bottoken/sendMediaGroup", "/bottoken/sendMessage"}, requestPaths)
	assert.Equal(t, 1, file.closeCalls)
	assert.Contains(t, warning.String(), "warning: attachments sent, but cleanup failed")
	assert.Contains(t, warning.String(), closeErr.Error())
}
