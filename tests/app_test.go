package tests_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/beeyev/telegram-owl/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestArgs(args []string) []string {
	return append(os.Args[0:1], args...)
}

func setupMockServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *bytes.Buffer) {
	mockServer := httptest.NewServer(handler)
	t.Cleanup(mockServer.Close)
	outputBuf := new(bytes.Buffer)
	return mockServer, outputBuf
}

func TestNoFlags(t *testing.T) {
	outputBuf := new(bytes.Buffer)
	ctx := t.Context()

	app := cli.NewApp("dummy")
	app.Writer = outputBuf

	err := app.Run(ctx, []string{})
	require.NoError(t, err)

	assert.Contains(t, outputBuf.String(), "GLOBAL OPTIONS")
}

func TestSendMessage_FromStdin(t *testing.T) {
	var capturedBody string

	mockServer, outputBuf := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		assert.NoError(t, err)

		capturedBody = strings.TrimSpace(string(bodyBytes))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok": true}`))
	})

	args := getTestArgs([]string{"--token=123:abc", "--chat=75757", "--stdin"})

	r, w, _ := os.Pipe()
	_, _ = w.WriteString("hello from stdin\n")
	_ = w.Close()
	//nolint:reassign // "reassigning variable Stdin in other package os"
	os.Stdin = r

	app := cli.NewApp(mockServer.URL)
	app.Writer = outputBuf

	err := app.Run(t.Context(), args)
	require.NoError(t, err)

	assert.JSONEq(t, `{"chat_id":"75757","text":"hello from stdin"}`, capturedBody)
	assert.Equal(t, "Message sent successfully. Chat ID: 75757", strings.TrimSpace(outputBuf.String()))
}

func TestSendMessage_Success(t *testing.T) {
	type capturedJSONRequest struct {
		body        string
		urlPath     string
		method      string
		contentType string
	}

	tests := []struct {
		name                string
		args                []string
		expectedJSONPayload string
	}{
		{
			name:                "Minimal required flags",
			args:                []string{"--token=123:abc", "--chat=75757", "--message=Hello"},
			expectedJSONPayload: `{"chat_id":"75757","text":"Hello"}`,
		},
		{
			name:                "format flag with markdown",
			args:                []string{"--token=123:abc", "--chat=75757", "--message=Hello", "--format=markdown"},
			expectedJSONPayload: `{"chat_id":"75757","text":"Hello","parse_mode":"MarkdownV2"}`,
		},
		{
			name:                "format flag with html",
			args:                []string{"--token=123:abc", "--chat=75757", "--message=Hello", "--format=html"},
			expectedJSONPayload: `{"chat_id":"75757","text":"Hello","parse_mode":"html"}`,
		},
		{
			name: "All flags",
			args: []string{
				"--token=123:abc",
				"--chat=75757",
				"--message=Hello",
				"--silent=true",
				"--spoiler=true",
				"--protect=true",
				"--no-link-preview=true",
				"--thread=1234",
			},
			//nolint:lll // Long line, but ok
			expectedJSONPayload: `{"chat_id":"75757","message_thread_id":"1234","text":"Hello","disable_notification":true,"protect_content":true,"link_preview_options":{"is_disabled":true}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			captured := capturedJSONRequest{}

			mockServer, outputBuf := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
				bodyBytes, err := io.ReadAll(r.Body)
				assert.NoError(t, err)

				captured.body = strings.TrimSpace(string(bodyBytes))
				captured.urlPath = r.URL.Path
				captured.method = r.Method
				captured.contentType = r.Header.Get("Content-Type")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"ok": true}`))
			})

			ctx := t.Context()
			args := getTestArgs(tt.args)

			app := cli.NewApp(mockServer.URL)
			app.Writer = outputBuf

			err := app.Run(ctx, args)
			require.NoError(t, err)

			assert.JSONEq(t, tt.expectedJSONPayload, captured.body)
			assert.Exactly(t, `/bot123:abc/sendMessage`, captured.urlPath)
			assert.Exactly(t, http.MethodPost, captured.method)
			assert.Exactly(t, "application/json", captured.contentType)
			assert.Equal(t, "Message sent successfully. Chat ID: 75757", strings.TrimSpace(outputBuf.String()))
		})
	}
}

func TestSendMediaGroup_Success(t *testing.T) {
	type capturedMultipartRequest struct {
		urlPath     string
		method      string
		contentType string
	}

	captured := capturedMultipartRequest{}

	mockServer, outputBuf := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		captured.urlPath = r.URL.Path
		captured.method = r.Method
		captured.contentType = r.Header.Get("Content-Type")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok": true}`))
	})

	app := cli.NewApp(mockServer.URL)
	app.Writer = outputBuf

	photoFile1, err := os.CreateTemp(t.TempDir(), "photo1.jpg")
	require.NoError(t, err)
	defer photoFile1.Close()

	args := getTestArgs([]string{
		"--token=123:abc",
		"--chat=75757",
		"--attach=" + photoFile1.Name(),
		"--message=Hello",
		"--as-document=true",
		"--silent=true",
		"--spoiler=true",
		"--protect=true",
		"--no-link-preview=true",
		"--thread=1234",
	})

	err = app.Run(t.Context(), args)
	require.NoError(t, err)

	assert.Exactly(t, `/bot123:abc/sendMediaGroup`, captured.urlPath)
	assert.Exactly(t, http.MethodPost, captured.method)
	assert.Contains(t, captured.contentType, "multipart/form-data")
	assert.Equal(t, "Message sent successfully. Chat ID: 75757", strings.TrimSpace(outputBuf.String()))
}

func Test_ErrorResponse(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "no token",
			args: []string{"--chat=whatever"},
			want: "missing required flag: --token",
		},
		{
			name: "no chat",
			args: []string{"--token=whatever"},
			want: "missing required flag: --chat",
		},
		{
			name: "no message and no attachments",
			args: []string{"--token=whatever", "--chat=whatever"},
			want: "nothing to send: provide a --message or --attach flag",
		},
		{
			name: "no message and no attachments",
			args: []string{"--token=whatever", "--chat=whatever", "--attach=does-not-exist.jpg"},
			want: "failed to load attachments",
		},
		{
			name: "format flag with incorrect value",
			args: []string{"--token=whatever", "--chat=whatever", "--message=hello", "--format=invalid"},
			want: "incorrect value for --format flag",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputBuf := new(bytes.Buffer)

			app := cli.NewApp("dummy")
			app.Writer = outputBuf

			err := app.Run(t.Context(), getTestArgs(tt.args))

			require.Error(t, err, "Run should fail when all conditions are met")
			require.ErrorContains(t, err, tt.want)
			assert.Empty(t, outputBuf.String())
		})
	}
}
