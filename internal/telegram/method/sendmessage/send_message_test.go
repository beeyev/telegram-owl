package sendmessage_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/beeyev/telegram-owl/internal/telegram/method/sendmessage"
	"github.com/beeyev/telegram-owl/internal/telegram/testutils"
)

func TestSend_ValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		options        sendmessage.Options
		expectedErrors []string
	}{
		{
			name:    "chat ID and message are required",
			options: sendmessage.Options{},
			expectedErrors: []string{
				"chat ID is required",
				"message is required",
			},
		},
		{
			name: "message is too long",
			options: sendmessage.Options{
				ChatID: "123",
				Text:   string(make([]byte, sendmessage.MaxTextLength+1)),
			},
			expectedErrors: []string{"message is too long"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sender := sendmessage.New(testutils.NewMockHTTPDoer())
			err := sender.Send(t.Context(), &tt.options)
			require.Error(t, err)
			for _, expectedError := range tt.expectedErrors {
				assert.Containsf(t, err.Error(), expectedError, "expected error not found")
			}
		})
	}
}

func TestSend_Success(t *testing.T) {
	t.Parallel()

	mockHTTPClient := testutils.NewMockHTTPDoer()
	sender := sendmessage.New(mockHTTPClient)

	options := &sendmessage.Options{
		ChatID: "123",
		Text:   "Hello, world!",
	}

	err := sender.Send(t.Context(), options)
	require.NoError(t, err)

	require.Len(t, mockHTTPClient.SubmitJSONResult, 1)
	requestJSON, err := json.Marshal(mockHTTPClient.SubmitJSONResult[0].Body)
	require.NoError(t, err, "Marshal should succeed")
	assert.JSONEq(t, `{"chat_id":"123","text":"Hello, world!"}`, string(requestJSON))
}

func TestSend_DelegatesFormattedLengthValidation(t *testing.T) {
	t.Parallel()

	mockHTTPClient := testutils.NewMockHTTPDoer()
	sender := sendmessage.New(mockHTTPClient)

	options := &sendmessage.Options{
		ChatID:    "123",
		Text:      strings.Repeat("a", sendmessage.MaxTextLength+1),
		ParseMode: "html",
	}

	err := sender.Send(t.Context(), options)
	require.NoError(t, err)
	require.Len(t, mockHTTPClient.SubmitJSONResult, 1)
}
