package sendrichmessage_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/beeyev/telegram-owl/internal/telegram/method/sendrichmessage"
	"github.com/beeyev/telegram-owl/internal/telegram/testutils"
)

func TestSend_ValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		options        sendrichmessage.Options
		expectedErrors []string
	}{
		{
			name:    "required fields",
			options: sendrichmessage.Options{},
			expectedErrors: []string{
				"chat ID is required",
				"message is required",
				"format must be rich-markdown or rich-html",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sender := sendrichmessage.New(testutils.NewMockHTTPDoer())
			err := sender.Send(t.Context(), &tt.options)
			require.Error(t, err)
			for _, expectedError := range tt.expectedErrors {
				assert.ErrorContains(t, err, expectedError)
			}
		})
	}
}

func TestSend_Success(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		options         sendrichmessage.Options
		expectedPayload string
	}{
		{
			name: "rich markdown",
			options: sendrichmessage.Options{
				ChatID: "123",
				Text:   "# Deployment\n\n**Passed**",
				Format: sendrichmessage.FormatMarkdown,
			},
			expectedPayload: `{"chat_id":"123","rich_message":{"markdown":"# Deployment\n\n**Passed**"}}`,
		},
		{
			name: "rich html with delivery options",
			options: sendrichmessage.Options{
				ChatID:              "123",
				MessageThreadID:     "456",
				Text:                "<b>Passed</b>",
				Format:              sendrichmessage.FormatHTML,
				DisableNotification: true,
				ProtectContent:      true,
			},
			expectedPayload: `{
				"chat_id":"123",
				"message_thread_id":"456",
				"rich_message":{"html":"<b>Passed</b>"},
				"disable_notification":true,
				"protect_content":true
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockHTTPClient := testutils.NewMockHTTPDoer()
			sender := sendrichmessage.New(mockHTTPClient)

			err := sender.Send(t.Context(), &tt.options)
			require.NoError(t, err)

			require.Len(t, mockHTTPClient.SubmitJSONResult, 1)
			request := mockHTTPClient.SubmitJSONResult[0]
			assert.Equal(t, http.MethodPost, request.Method)
			assert.Equal(t, "sendRichMessage", request.Endpoint)

			requestJSON, err := json.Marshal(request.Body)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expectedPayload, string(requestJSON))
		})
	}
}

func TestSend_DelegatesFormattedLengthValidation(t *testing.T) {
	t.Parallel()

	const maxParsedRichTextLength = 32768

	mockHTTPClient := testutils.NewMockHTTPDoer()
	sender := sendrichmessage.New(mockHTTPClient)
	options := &sendrichmessage.Options{
		ChatID: "123",
		Text:   "<b>" + strings.Repeat("a", maxParsedRichTextLength) + "</b>",
		Format: sendrichmessage.FormatHTML,
	}

	err := sender.Send(t.Context(), options)
	require.NoError(t, err)
	require.Len(t, mockHTTPClient.SubmitJSONResult, 1)
}
