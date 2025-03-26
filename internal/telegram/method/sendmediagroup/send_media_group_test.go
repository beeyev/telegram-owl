package sendmediagroup_test

import (
	"os"
	"testing"

	"github.com/beeyev/telegram-owl/internal/telegram/common/attachment"
	"github.com/beeyev/telegram-owl/internal/telegram/method/sendmediagroup"
	"github.com/beeyev/telegram-owl/internal/telegram/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSend_ValidationErrors(t *testing.T) {
	tests := []struct {
		name           string
		options        sendmediagroup.Options
		expectedErrors []string
	}{
		{
			name:    "chat ID and message are required",
			options: sendmediagroup.Options{},
			expectedErrors: []string{
				"chat ID is required",
				"at least one attachment required"},
		},
		{
			name: "message is too long",
			options: sendmediagroup.Options{
				ChatID:  "123",
				Caption: string(make([]byte, sendmediagroup.MaxCaptionLength+1)),
			},
			expectedErrors: []string{
				"at least one attachment required",
				"message is too long",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender := sendmediagroup.New(testutils.NewMockHTTPDoer())
			err := sender.Send(&tt.options)
			require.Error(t, err)
			for _, expectedError := range tt.expectedErrors {
				assert.Containsf(t, err.Error(), expectedError, "expected error not found")
			}
		})
	}
}

func TestSend_Success(t *testing.T) {
	attachments := attachment.Attachments{
		{
			AType:     attachment.Photo,
			FileName:  "file1.jpg",
			SizeBytes: 1024,
			File:      &os.File{},
		},
		{
			AType:     attachment.Photo,
			FileName:  "file2.jpg",
			SizeBytes: 1024,
			File:      &os.File{},
		},
	}

	options := &sendmediagroup.Options{
		ChatID:      "123",
		Attachments: attachments,
	}

	mockHTTPClient := testutils.NewMockHTTPDoer()
	sender := sendmediagroup.New(mockHTTPClient)

	err := sender.Send(options)
	require.NoError(t, err)
	require.Len(t, mockHTTPClient.SubmitMultipartResult, 1)

	expected := map[string]string{
		"chat_id": "123",
		"media":   `[{"type":"photo","media":"attach://file0"},{"type":"photo","media":"attach://file1"}]`,
	}
	assert.Exactly(t, expected, mockHTTPClient.SubmitMultipartResult[0].Fields, "unexpected request payload")
}

func TestSend_Success2(t *testing.T) {
	dummyFile := &os.File{}

	tests := []struct {
		name     string
		options  *sendmediagroup.Options
		expected map[string]string
	}{
		{
			name: "required fields only",
			options: &sendmediagroup.Options{
				ChatID: "123",
				Attachments: attachment.Attachments{
					{
						AType:     attachment.Photo,
						FileName:  "file1.jpg",
						SizeBytes: 1024,
						File:      dummyFile,
					},
				},
			},
			expected: map[string]string{
				"chat_id": "123",
				"media":   `[{"type":"photo","media":"attach://file0"}]`,
			},
		},
		{
			name: "all fields",
			options: &sendmediagroup.Options{
				ChatID:              "123",
				MessageThreadID:     "456",
				Caption:             "hello",
				HasSpoiler:          true,
				DisableNotification: true,
				ProtectContent:      true,
				Attachments: attachment.Attachments{
					{
						AType:     attachment.Document,
						FileName:  "file1.jpg",
						SizeBytes: 1024,
						File:      dummyFile,
					},
					{
						AType:     attachment.Audio,
						FileName:  "file1.mp3",
						SizeBytes: 1024,
						File:      dummyFile,
					},
				},
			},
			expected: map[string]string{
				"chat_id":              "123",
				"message_thread_id":    "456",
				"disable_notification": "1",
				"protect_content":      "1",
				//nolint:lll // Long line, but ok
				"media": `[{"type":"document","media":"attach://file0","has_spoiler":true},{"type":"audio","media":"attach://file1","caption":"hello","has_spoiler":true}]`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHTTPClient := testutils.NewMockHTTPDoer()
			sender := sendmediagroup.New(mockHTTPClient)

			err := sender.Send(tt.options)
			require.NoError(t, err)
			require.Len(t, mockHTTPClient.SubmitMultipartResult, 1)

			assert.Equal(t, tt.expected, mockHTTPClient.SubmitMultipartResult[0].Fields, "unexpected request payload")

			require.Len(t, mockHTTPClient.SubmitMultipartResult[0].Files, len(tt.options.Attachments))
		})
	}
}
