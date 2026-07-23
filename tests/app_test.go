package tests_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/beeyev/telegram-owl/internal/cli"
)

func getTestArgs(args []string) []string {
	return append([]string{os.Args[0]}, args...)
}

func setupMockServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *bytes.Buffer) {
	mockServer := httptest.NewServer(handler)
	t.Cleanup(mockServer.Close)
	outputBuf := new(bytes.Buffer)
	return mockServer, outputBuf
}

func TestNoFlags(t *testing.T) {
	t.Parallel()

	outputBuf := new(bytes.Buffer)
	ctx := t.Context()

	app := cli.NewApp("dummy")
	app.Writer = outputBuf

	err := app.Run(ctx, []string{})
	require.NoError(t, err)

	assert.Contains(t, outputBuf.String(), "GLOBAL OPTIONS")
	assert.Contains(t, outputBuf.String(), "VERSION:")
	assert.Contains(t, outputBuf.String(), "--version")
}

func TestVersionOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		version string
		flag    string
	}{
		{name: "release version", version: "1.2.3", flag: "--version"},
		{name: "prefixed version with short flag", version: "v1.2.3", flag: "-v"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			outputBuf := new(bytes.Buffer)
			app := cli.NewApp("dummy")
			app.Version = tt.version
			app.Writer = outputBuf

			err := app.Run(t.Context(), getTestArgs([]string{tt.flag}))
			require.NoError(t, err)
			assert.Equal(
				t,
				"telegram-owl v1.2.3\nAlexander Tebiev - https://github.com/beeyev\n",
				outputBuf.String(),
			)
		})
	}
}

func TestSendMessage_FromStdin(t *testing.T) { //nolint:paralleltest // Reassigns process-global os.Stdin.
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
	_, _ = w.WriteString("    hello from stdin\n")
	_ = w.Close()
	originalStdin := os.Stdin
	t.Cleanup(func() {
		//nolint:reassign // Restore process-global stdin after this test.
		os.Stdin = originalStdin
	})
	//nolint:reassign // "reassigning variable Stdin in other package os"
	os.Stdin = r

	app := cli.NewApp(mockServer.URL)
	app.Writer = outputBuf

	err := app.Run(t.Context(), args)
	require.NoError(t, err)

	assert.JSONEq(t, `{"chat_id":"75757","text":"    hello from stdin"}`, capturedBody)
	assert.Empty(t, outputBuf.String())
}

func TestSendRichMessage_FromStdin(t *testing.T) { //nolint:paralleltest // Reassigns process-global os.Stdin.
	var capturedBody string
	var capturedPath string

	mockServer, outputBuf := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		assert.NoError(t, err)

		capturedBody = strings.TrimSpace(string(bodyBytes))
		capturedPath = r.URL.Path

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	stdinReader, stdinWriter, err := os.Pipe()
	require.NoError(t, err)
	_, err = stdinWriter.WriteString("  # Deployment\n\n- passed\n")
	require.NoError(t, err)
	require.NoError(t, stdinWriter.Close())
	t.Cleanup(func() {
		require.NoError(t, stdinReader.Close())
	})

	originalStdin := os.Stdin
	t.Cleanup(func() {
		//nolint:reassign // Restore process-global stdin after this test.
		os.Stdin = originalStdin
	})
	//nolint:reassign // Exercise rich message input from a pipe.
	os.Stdin = stdinReader

	app := cli.NewApp(mockServer.URL)
	app.Writer = outputBuf
	err = app.Run(t.Context(), getTestArgs([]string{
		"--token=123:abc",
		"--chat=75757",
		"--stdin",
		"--format=rich-markdown",
	}))
	require.NoError(t, err)

	assert.Equal(t, `/bot123:abc/sendRichMessage`, capturedPath)
	assert.JSONEq(t, `{
		"chat_id":"75757",
		"rich_message":{"markdown":"  # Deployment\n\n- passed"}
	}`, capturedBody)
	assert.Empty(t, outputBuf.String())
}

//nolint:paralleltest // Reassigns process-global os.Stdin.
func TestSendAttachment_WithStdinFromCharacterDevice(
	t *testing.T,
) {
	var capturedPath string
	mockServer, outputBuf := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		_, err := io.Copy(io.Discard, r.Body)
		assert.NoError(t, err)

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write([]byte(`{"ok": true}`))
		assert.NoError(t, err)
	})

	attachmentFile, err := os.CreateTemp(t.TempDir(), "attachment.txt")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, attachmentFile.Close())
	})

	stdin, err := os.Open(os.DevNull)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, stdin.Close())
	})
	stdinInfo, err := stdin.Stat()
	require.NoError(t, err)
	require.NotZero(t, stdinInfo.Mode()&os.ModeCharDevice)

	originalStdin := os.Stdin
	t.Cleanup(func() {
		//nolint:reassign // Restore process-global stdin after this test.
		os.Stdin = originalStdin
	})
	//nolint:reassign // Exercise --stdin with a character device.
	os.Stdin = stdin

	app := cli.NewApp(mockServer.URL)
	app.Writer = outputBuf
	err = app.Run(t.Context(), getTestArgs([]string{
		"--token=123:abc",
		"--chat=75757",
		"--attach=" + attachmentFile.Name(),
		"--as-document=true",
		"--stdin",
	}))

	require.NoError(t, err)
	assert.Equal(t, `/bot123:abc/sendMediaGroup`, capturedPath)
	assert.Empty(t, outputBuf.String())
}

func TestSendMessage_Success(t *testing.T) {
	t.Parallel()

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
		expectedOutput      string
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
			name: "verbose success output",
			args: []string{
				"--token=123:abc",
				"--chat=75757",
				"--message=Hello",
				"--format=html",
				"--thread=1234",
				"--verbose=true",
			},
			expectedJSONPayload: `{"chat_id":"75757","message_thread_id":"1234","text":"Hello","parse_mode":"html"}`,
			expectedOutput:      "Sending Telegram message: chat=75757, message=yes, attachments=0, thread=1234, format=html\nMessage sent successfully. Chat ID: 75757. Duration: ",
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
			t.Parallel()

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
			if tt.expectedOutput == "" {
				assert.Empty(t, outputBuf.String())
			} else {
				assert.Contains(t, outputBuf.String(), tt.expectedOutput)
			}
		})
	}
}

func TestSendRichMessage_Success(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		args                []string
		expectedJSONPayload string
	}{
		{
			name: "rich markdown",
			args: []string{
				"--token=123:abc",
				"--chat=75757",
				"--message=# Deployment\n\n**Passed**",
				"--format=rich-markdown",
			},
			expectedJSONPayload: `{
				"chat_id":"75757",
				"rich_message":{"markdown":"# Deployment\n\n**Passed**"}
			}`,
		},
		{
			name: "rich html with delivery options",
			args: []string{
				"--token=123:abc",
				"--chat=75757",
				"--message=<b>Passed</b>",
				"--format=rich-html",
				"--thread=1234",
				"--silent=true",
				"--protect=true",
			},
			expectedJSONPayload: `{
				"chat_id":"75757",
				"message_thread_id":"1234",
				"rich_message":{"html":"<b>Passed</b>"},
				"disable_notification":true,
				"protect_content":true
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var capturedBody string
			var capturedPath string
			mockServer, outputBuf := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
				bodyBytes, err := io.ReadAll(r.Body)
				assert.NoError(t, err)
				capturedBody = strings.TrimSpace(string(bodyBytes))
				capturedPath = r.URL.Path

				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"ok":true}`))
			})

			app := cli.NewApp(mockServer.URL)
			app.Writer = outputBuf

			err := app.Run(t.Context(), getTestArgs(tt.args))
			require.NoError(t, err)

			assert.Equal(t, `/bot123:abc/sendRichMessage`, capturedPath)
			assert.JSONEq(t, tt.expectedJSONPayload, capturedBody)
			assert.Empty(t, outputBuf.String())
		})
	}
}

func TestSendMediaGroup_Success(t *testing.T) {
	t.Parallel()

	type capturedMultipartRequest struct {
		urlPath     string
		method      string
		contentType string
		media       string
	}

	captured := capturedMultipartRequest{}

	mockServer, outputBuf := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		captured.urlPath = r.URL.Path
		captured.method = r.Method
		captured.contentType = r.Header.Get("Content-Type")
		if !assert.NoError(t, r.ParseMultipartForm(32<<20)) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		captured.media = r.FormValue("media")

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
		"--message=*Hello*",
		"--format=markdown",
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
	expectedMedia := `[{
		"type":"document",
		"media":"attach://file0",
		"caption":"*Hello*",
		"parse_mode":"MarkdownV2",
		"has_spoiler":true
	}]`
	assert.JSONEq(t, expectedMedia, captured.media)
	assert.Empty(t, outputBuf.String())
}

func TestSendMediaGroup_LongMessageSendsTextSeparately(t *testing.T) {
	t.Parallel()

	type capturedRequest struct {
		urlPath string
		body    string
		media   string
	}

	var captured []capturedRequest

	mockServer, outputBuf := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		req := capturedRequest{urlPath: r.URL.Path}
		if strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
			if !assert.NoError(t, r.ParseMultipartForm(32<<20)) {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			req.media = r.FormValue("media")
		} else {
			bodyBytes, err := io.ReadAll(r.Body)
			if !assert.NoError(t, err) {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			req.body = strings.TrimSpace(string(bodyBytes))
		}
		captured = append(captured, req)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok": true}`))
	})

	app := cli.NewApp(mockServer.URL)
	app.Writer = outputBuf

	photoFile1, err := os.CreateTemp(t.TempDir(), "photo1.jpg")
	require.NoError(t, err)
	defer photoFile1.Close()

	message := strings.Repeat("a", 1025)
	args := getTestArgs([]string{
		"--token=123:abc",
		"--chat=75757",
		"--attach=" + photoFile1.Name(),
		"--message=" + message,
		"--as-document=true",
	})

	err = app.Run(t.Context(), args)
	require.NoError(t, err)

	require.Len(t, captured, 2)
	assert.Exactly(t, `/bot123:abc/sendMediaGroup`, captured[0].urlPath)
	assert.JSONEq(t, `[{"type":"document","media":"attach://file0"}]`, captured[0].media)
	assert.Exactly(t, `/bot123:abc/sendMessage`, captured[1].urlPath)
	assert.JSONEq(t, `{"chat_id":"75757","text":"`+message+`"}`, captured[1].body)
	assert.Empty(t, outputBuf.String())
}

func TestSendMediaGroup_RichMessageSendsTextSeparately(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                    string
		format                  string
		message                 string
		expectedRichMessageJSON string
	}{
		{
			name:                    "rich markdown",
			format:                  "rich-markdown",
			message:                 "**Deployment passed**",
			expectedRichMessageJSON: `{"markdown":"**Deployment passed**"}`,
		},
		{
			name:                    "rich html",
			format:                  "rich-html",
			message:                 "<strong>Deployment passed</strong>",
			expectedRichMessageJSON: `{"html":"<strong>Deployment passed</strong>"}`,
		},
		{
			name:   "attachments only",
			format: "rich-markdown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testRichAttachmentRouting(
				t,
				tt.format,
				tt.message,
				tt.expectedRichMessageJSON,
			)
		})
	}
}

func testRichAttachmentRouting(t *testing.T, format, message, expectedRichMessageJSON string) {
	t.Helper()

	type capturedRequest struct {
		urlPath string
		body    string
		media   string
	}

	var captured []capturedRequest
	server, outputBuf := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		request := capturedRequest{urlPath: r.URL.Path}
		if strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
			if !assert.NoError(t, r.ParseMultipartForm(32<<20)) {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			request.media = r.FormValue("media")
		} else {
			bodyBytes, err := io.ReadAll(r.Body)
			if !assert.NoError(t, err) {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			request.body = strings.TrimSpace(string(bodyBytes))
		}
		captured = append(captured, request)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	attachmentFile, err := os.CreateTemp(t.TempDir(), "report.txt")
	require.NoError(t, err)
	defer attachmentFile.Close()

	args := []string{
		"--token=123:abc",
		"--chat=75757",
		"--attach=" + attachmentFile.Name(),
		"--as-document=true",
		"--format=" + format,
	}
	if message != "" {
		args = append(args, "--message="+message)
	}

	app := cli.NewApp(server.URL)
	app.Writer = outputBuf
	err = app.Run(t.Context(), getTestArgs(args))
	require.NoError(t, err)

	expectedRequestCount := 1
	if expectedRichMessageJSON != "" {
		expectedRequestCount = 2
	}
	require.Len(t, captured, expectedRequestCount)
	assert.Equal(t, `/bot123:abc/sendMediaGroup`, captured[0].urlPath)
	assert.JSONEq(t, `[{"type":"document","media":"attach://file0"}]`, captured[0].media)

	if expectedRichMessageJSON != "" {
		assert.Equal(t, `/bot123:abc/sendRichMessage`, captured[1].urlPath)
		assert.JSONEq(t, `{
			"chat_id":"75757",
			"rich_message":`+expectedRichMessageJSON+`
		}`, captured[1].body)
	}
	assert.Empty(t, outputBuf.String())
}

func TestSendMediaGroup_RichMessageReportsPartialDelivery(t *testing.T) {
	t.Parallel()

	var capturedPaths []string
	server, outputBuf := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPaths = append(capturedPaths, r.URL.Path)
		_, err := io.Copy(io.Discard, r.Body)
		assert.NoError(t, err)

		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/sendRichMessage") {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"ok":false,"error_code":400,"description":"rich message rejected"}`))
			return
		}

		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	attachmentFile, err := os.CreateTemp(t.TempDir(), "report.txt")
	require.NoError(t, err)
	defer attachmentFile.Close()

	app := cli.NewApp(server.URL)
	app.Writer = outputBuf
	err = app.Run(t.Context(), getTestArgs([]string{
		"--token=123:abc",
		"--chat=75757",
		"--attach=" + attachmentFile.Name(),
		"--as-document=true",
		"--message=**Deployment passed**",
		"--format=rich-markdown",
	}))

	require.ErrorContains(t, err, "attachments sent, but rich message failed")
	assert.Equal(t, []string{
		`/bot123:abc/sendMediaGroup`,
		`/bot123:abc/sendRichMessage`,
	}, capturedPaths)
	assert.Empty(t, outputBuf.String())
}

func TestSendMediaGroup_RichMessageStopsAfterAttachmentFailure(t *testing.T) {
	t.Parallel()

	var capturedPaths []string
	server, outputBuf := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPaths = append(capturedPaths, r.URL.Path)
		_, err := io.Copy(io.Discard, r.Body)
		assert.NoError(t, err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"ok":false,"error_code":400,"description":"attachment rejected"}`))
	})

	attachmentFile, err := os.CreateTemp(t.TempDir(), "report.txt")
	require.NoError(t, err)
	defer attachmentFile.Close()

	app := cli.NewApp(server.URL)
	app.Writer = outputBuf
	err = app.Run(t.Context(), getTestArgs([]string{
		"--token=123:abc",
		"--chat=75757",
		"--attach=" + attachmentFile.Name(),
		"--as-document=true",
		"--message=**Deployment passed**",
		"--format=rich-markdown",
	}))

	require.ErrorContains(t, err, "send attachments")
	assert.Equal(t, []string{`/bot123:abc/sendMediaGroup`}, capturedPaths)
	assert.Empty(t, outputBuf.String())
}

func TestSendMediaGroup_FormattedCaptionDelegatesLengthValidation(t *testing.T) {
	t.Parallel()

	var capturedPath string
	var capturedMedia string

	mockServer, outputBuf := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		if !assert.NoError(t, r.ParseMultipartForm(32<<20)) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		capturedMedia = r.FormValue("media")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok": true}`))
	})

	app := cli.NewApp(mockServer.URL)
	app.Writer = outputBuf

	photoFile, err := os.CreateTemp(t.TempDir(), "photo.jpg")
	require.NoError(t, err)
	defer photoFile.Close()

	message := "<b>" + strings.Repeat("a", 1018) + "</b>"
	require.Len(t, []rune(message), 1025)

	err = app.Run(t.Context(), getTestArgs([]string{
		"--token=123:abc",
		"--chat=75757",
		"--attach=" + photoFile.Name(),
		"--message=" + message,
		"--format=html",
		"--as-document=true",
	}))
	require.NoError(t, err)

	assert.Exactly(t, `/bot123:abc/sendMediaGroup`, capturedPath)
	expectedMedia := `[{"type":"document","media":"attach://file0","caption":"` +
		message + `","parse_mode":"html"}]`
	assert.JSONEq(t, expectedMedia, capturedMedia)
	assert.Empty(t, outputBuf.String())
}

func Test_ErrorResponse(t *testing.T) {
	t.Parallel()

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
			name: "attachment does not exist",
			args: []string{"--token=whatever", "--chat=whatever", "--attach=does-not-exist.jpg"},
			want: "failed to load attachments",
		},
		{
			name: "format flag with incorrect value",
			args: []string{"--token=whatever", "--chat=whatever", "--message=hello", "--format=invalid"},
			want: "incorrect value for --format flag",
		},
		{
			name: "link preview option with rich format",
			args: []string{
				"--token=whatever",
				"--chat=whatever",
				"--message=hello",
				"--format=rich-markdown",
				"--no-link-preview",
			},
			want: "--no-link-preview is not supported with rich message formats",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

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
