package sendmessage_test

import (
	"encoding/json"
	"testing"

	"github.com/beeyev/telegram-owl/internal/telegram/method/sendmessage"
	"github.com/beeyev/telegram-owl/internal/telegram/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSend_ValidationErrors(t *testing.T) {
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
			sender := sendmessage.New(testutils.NewMockHTTPDoer())
			err := sender.Send(&tt.options)
			require.Error(t, err)
			for _, expectedError := range tt.expectedErrors {
				assert.Containsf(t, err.Error(), expectedError, "expected error not found")
			}
		})
	}
}

func TestSend_Success(t *testing.T) {
	mockHTTPClient := testutils.NewMockHTTPDoer()
	sender := sendmessage.New(mockHTTPClient)

	options := &sendmessage.Options{
		ChatID: "123",
		Text:   "Hello, world!",
	}

	err := sender.Send(options)
	require.NoError(t, err)

	require.Len(t, mockHTTPClient.SubmitJSONResult, 1)
	requestJSON, err := json.Marshal(mockHTTPClient.SubmitJSONResult[0].Body)
	require.NoError(t, err, "Marshal should succeed")
	assert.JSONEq(t, `{"chat_id":"123","text":"Hello, world!"}`, string(requestJSON))
}
