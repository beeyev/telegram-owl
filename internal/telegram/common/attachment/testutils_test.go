package attachment_test

import (
	"bytes"
	"errors"
	"sync"

	attach "github.com/beeyev/telegram-owl/internal/telegram/common/attachment"
)

// mockFileOpener is a mock implementation of the FileOpener interface.
type mockFileOpener struct {
	files map[string]*attach.OpenedFile
	err   error
}

func (m *mockFileOpener) Open(filePath string) (*attach.OpenedFile, error) {
	if m.err != nil {
		return nil, m.err
	}

	if file, ok := m.files[filePath]; ok {
		return file, nil
	}

	return nil, errors.New("file not found")
}

// mockReadCloser simulates a file that can be opened and closed.
type mockReadCloser struct {
	data   *bytes.Reader
	closed bool
	mu     sync.Mutex
}

// newMockReadCloser initializes the mock with test content.
func newMockReadCloser(content string) *mockReadCloser {
	return &mockReadCloser{data: bytes.NewReader([]byte(content))}
}

// Read implements io.Reader.
func (m *mockReadCloser) Read(p []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return 0, errors.New("attempted to read from closed file")
	}

	return m.data.Read(p)
}

// Close marks the file as closed.
func (m *mockReadCloser) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return errors.New("file already closed")
	}

	m.closed = true
	return nil
}

// IsClosed checks if the file has been closed.
func (m *mockReadCloser) IsClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}
