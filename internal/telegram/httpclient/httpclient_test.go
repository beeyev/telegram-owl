package httpclient_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beeyev/telegram-owl/internal/telegram/httpclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("should return error if baseURL is empty", func(t *testing.T) {
		client, err := httpclient.New("", "token", "")
		assert.Nil(t, client)
		assert.EqualError(t, err, "apiBotURL value is not provided")
	})

	t.Run("should return error if token is empty", func(t *testing.T) {
		client, err := httpclient.New("http://example.com", "", "")
		assert.Nil(t, client)
		assert.EqualError(t, err, "token value is not provided")
	})

	t.Run("should return httpClient if baseURL and token are provided", func(t *testing.T) {
		client, err := httpclient.New("http://example.com", "token", "")
		assert.NotNil(t, client)
		assert.NoError(t, err)
	})

	t.Run("should return error if proxy is invalid", func(t *testing.T) {
		client, err := httpclient.New("http://example.com", "token", "://bad_proxy")
		assert.Nil(t, client)
		assert.EqualError(t, err, "invalid proxy URL: parse \"://bad_proxy\": missing protocol scheme")
	})
}

type capturedJSONRequest struct {
	body        string
	urlPath     string
	method      string
	contentType string
}

func TestSubmitJSON_Success(t *testing.T) {
	captured := capturedJSONRequest{}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		assert.NoError(t, err)

		captured.body = strings.TrimSpace(string(bodyBytes))
		captured.urlPath = r.URL.Path
		captured.method = r.Method
		captured.contentType = r.Header.Get("Content-Type")

		// Respond with a success JSON payload
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok": true}`))
	}))
	defer mockServer.Close()

	client, err := httpclient.New(mockServer.URL, "to:ken", "")
	require.NotNil(t, client)
	require.NoError(t, err)

	err = client.SubmitJSON(http.MethodPost, "method-a", map[string]string{
		"foo": "bar",
	})
	require.NoError(t, err, "SubmitJSON should succeed when the server returns ok=true")

	assert.JSONEq(t, `{"foo":"bar"}`, captured.body)
	assert.Exactly(t, `/botto:ken/method-a`, captured.urlPath)
	assert.Exactly(t, http.MethodPost, captured.method)
	assert.Exactly(t, "application/json", captured.contentType)
}

func TestSubmitMultipart_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse multipart form
		err := r.ParseMultipartForm(1 << 20) // 1MB
		assert.NoError(t, err)

		// fmt.Println(fmt.Sprint(r.MultipartForm.Value))

		// Check fields
		assert.Equal(t, "bar", r.FormValue("foo"), "foo field should have value 'bar'")

		// Check file
		file, fileHeader, err := r.FormFile("file_field")
		assert.NoError(t, err)
		assert.Equal(t, "test.txt", fileHeader.Filename)

		fileData, _ := io.ReadAll(file)
		assert.Equal(t, "hello world", string(fileData))

		// Respond with a success JSON payload
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok": true}`))
	}))
	defer mockServer.Close()

	client, err := httpclient.New(mockServer.URL, "to:ken", "")
	require.NotNil(t, client)
	require.NoError(t, err)

	fileReader := io.NopCloser(strings.NewReader("hello world"))
	defer fileReader.Close()

	multipartFile := httpclient.MultipartFile{
		FieldName:  "file_field",
		FileName:   "test.txt",
		FileReader: fileReader,
	}
	fields := map[string]string{"foo": "bar"}

	err = client.SubmitMultipart(http.MethodPost, "method-b", fields, []httpclient.MultipartFile{multipartFile})
	require.NoError(t, err)
}

func TestErrorHandling_NetworkTransportError(t *testing.T) {
	// If the server closes immediately or is unreachable, resty will return a transport error.
	// We'll just close the server before calling.
	mockServer := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	mockServer.Close() // close immediately

	client, err := httpclient.New(mockServer.URL, "to:ken", "")
	require.NotNil(t, client)
	require.NoError(t, err)

	err = client.SubmitJSON(http.MethodPost, "method-a", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "transport error:")
}

func TestErrorHandling_APIError(t *testing.T) {
	// This mock server returns an HTTP error code with a Telegram-style error response
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"ok": false, "error_code":400, "description":"Bad Request: something went wrong"}`))
	}))
	defer mockServer.Close()

	client, err := httpclient.New(mockServer.URL, "to:ken", "")
	require.NotNil(t, client)
	require.NoError(t, err)

	err = client.SubmitJSON(http.MethodPost, "/method-a", map[string]string{
		"foo": "bar",
	})
	require.Error(t, err)

	// Verify the error message is what we expect
	assert.Contains(t, err.Error(), "API error [/method-a] (HTTP 400): 400 - Bad Request: something went wrong")
}

func TestErrorHandling_EmptyResponse(t *testing.T) {
	// If the server returns an error status but no 'description', we parse the raw body
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer mockServer.Close()

	client, err := httpclient.New(mockServer.URL, "to:ken", "")
	require.NotNil(t, client)
	require.NoError(t, err)

	err = client.SubmitJSON(http.MethodPost, "/test-json", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected error (status=403): <empty response body>")
}

func TestExecuteRequest_UnexpectedError(t *testing.T) {
	responseBody := "some weird error"
	// If the server returns an error status but no 'description', we parse the raw body
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		// We'll return something not matching the expected JSON structure
		_, _ = w.Write([]byte(responseBody))
	}))
	defer mockServer.Close()

	client, err := httpclient.New(mockServer.URL, "to:ken", "")
	require.NotNil(t, client)
	require.NoError(t, err)

	err = client.SubmitJSON(http.MethodGet, "/test-json", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected error (status=403): some weird error")
}
