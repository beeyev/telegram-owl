package attachment_test

import (
	"testing"

	"github.com/beeyev/telegram-owl/internal/telegram/common/attachment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadMultipleAttachments_EmptyFiles(t *testing.T) {
	loader := &attachment.Loader{}
	_, err := loader.LoadMultipleAttachments(nil)
	require.Error(t, err)
	assert.Equal(t, "no attachments provided", err.Error())
}

func TestLoadMultipleAttachments_ExceedsMaxCount(t *testing.T) {
	loader := &attachment.Loader{
		MaxTotalAttachments: 2,
	}
	filePaths := []string{"a", "b", "c"}
	_, err := loader.LoadMultipleAttachments(filePaths)
	require.Error(t, err)
	assert.ErrorContains(t, err, "too many attachments: max allowed is 2, but got 3")
}

func TestLoadMultipleAttachments_ExceedsMaxSize(t *testing.T) {
	file1 := "abc1/file1.jpg"
	file2 := "abc2/file2.jpg"
	filePaths := []string{file1, file2}

	mockOpener := &mockFileOpener{
		files: map[string]*attachment.OpenedFile{
			file1: {File: newMockReadCloser("dummy content"), SizeBytes: 2 * attachment.BytesPerMegabyte},
			file2: {File: newMockReadCloser("dummy content"), SizeBytes: 4 * attachment.BytesPerMegabyte},
		},
	}

	loader := &attachment.Loader{
		FileOpener:                  mockOpener,
		MaxTotalAttachments:         2,
		MaxPhotoAttachmentSizeBytes: 3 * attachment.BytesPerMegabyte,
		MaxAttachmentSizeBytes:      5 * attachment.BytesPerMegabyte,
		MaxTotalSizeBytes:           5 * attachment.BytesPerMegabyte,
	}
	_, err := loader.LoadMultipleAttachments(filePaths)
	require.Error(t, err)
	require.ErrorContains(t, err, "total attachments size exceeds the max allowed")

	for s, file := range mockOpener.files {
		fileMock := file.File.(*mockReadCloser)
		assert.Truef(t, fileMock.IsClosed(), "file %s should be closed", s)
	}
}

func TestLoadMultipleAttachments_AttachmentExceedsMaxAllowedSize(t *testing.T) {
	file1 := "abc1/file1.jpg"
	file2 := "abc2/file2.jpg"
	filePaths := []string{file1, file2}

	mockOpener := &mockFileOpener{
		files: map[string]*attachment.OpenedFile{
			file1: {File: newMockReadCloser("dummy content"), SizeBytes: 2 * attachment.BytesPerMegabyte},
			file2: {File: newMockReadCloser("dummy content"), SizeBytes: 4 * attachment.BytesPerMegabyte},
		},
	}
	loader := &attachment.Loader{
		FileOpener:                  mockOpener,
		MaxTotalAttachments:         2,
		MaxPhotoAttachmentSizeBytes: 2 * attachment.BytesPerMegabyte,
		MaxAttachmentSizeBytes:      2 * attachment.BytesPerMegabyte,
		MaxTotalSizeBytes:           4 * attachment.BytesPerMegabyte,
	}
	_, err := loader.LoadMultipleAttachments(filePaths)
	require.Error(t, err)
	require.ErrorContains(t, err, "exceeds the max allowed of")

	for s, file := range mockOpener.files {
		fileMock := file.File.(*mockReadCloser)
		assert.Truef(t, fileMock.IsClosed(), "file %s should be closed", s)
	}
}

func TestLoadMultipleAttachments_Success(t *testing.T) {
	filePath1 := "abc1/file1.jpg"
	filePath2 := "abc2/file2.jpg"
	filePaths := []string{filePath1, filePath2}

	file1 := newMockReadCloser("dummy content 1")
	file2 := newMockReadCloser("dummy content 2")

	mockOpener := &mockFileOpener{
		files: map[string]*attachment.OpenedFile{
			filePath1: {File: file1, SizeBytes: 1024},
			filePath2: {File: file2, SizeBytes: 2048},
		},
	}

	loader := &attachment.Loader{
		FileOpener:                  mockOpener,
		MaxTotalAttachments:         2,
		MaxPhotoAttachmentSizeBytes: 2048,
		MaxAttachmentSizeBytes:      4096,
		MaxTotalSizeBytes:           8192,
	}
	attachments, err := loader.LoadMultipleAttachments(filePaths)
	require.NoError(t, err)
	assert.Len(t, attachments, 2)

	for s, file := range mockOpener.files {
		fileMock := file.File.(*mockReadCloser)
		assert.Falsef(t, fileMock.IsClosed(), "file %s should not be closed", s)
	}

	expectedAttachments := attachment.Attachments{
		{
			AType:     attachment.Photo,
			FileName:  "file1.jpg",
			SizeBytes: 1024,
			File:      file1,
		},
		{
			AType:     attachment.Photo,
			FileName:  "file2.jpg",
			SizeBytes: 2048,
			File:      file2,
		},
	}
	assert.Exactly(t, expectedAttachments, attachments)
}

func TestLoadMultipleAttachments_PhotoSwitchedToDocument(t *testing.T) {
	filePath1 := "abc1/file1.jpg"
	filePaths := []string{filePath1}

	file := newMockReadCloser("dummy content")

	mockOpener := &mockFileOpener{
		files: map[string]*attachment.OpenedFile{
			filePath1: {File: file, SizeBytes: 2024},
		},
	}
	loader := &attachment.Loader{
		FileOpener:                  mockOpener,
		MaxTotalAttachments:         2,
		MaxPhotoAttachmentSizeBytes: 1048,
		MaxAttachmentSizeBytes:      4096,
		MaxTotalSizeBytes:           8192,
	}
	attachments, err := loader.LoadMultipleAttachments(filePaths)
	require.NoError(t, err)

	expectedAttachment := attachment.Attachments{
		0: {
			AType:     attachment.Document,
			FileName:  "file1.jpg",
			SizeBytes: 2024,
			File:      file,
		},
	}
	assert.Equal(t, expectedAttachment, attachments)
}

func TestLoadMultipleAttachments_MixedAttachmentTypes(t *testing.T) {
	filePath1 := "abc1/file1.jpg"
	filePath2 := "abc1/file2.mp3"
	filePaths := []string{filePath1, filePath2}

	mockOpener := &mockFileOpener{
		files: map[string]*attachment.OpenedFile{
			filePath1: {File: newMockReadCloser("dummy content"), SizeBytes: 1024},
			filePath2: {File: newMockReadCloser("dummy content"), SizeBytes: 2048},
		},
	}
	loader := &attachment.Loader{
		FileOpener:                  mockOpener,
		MaxTotalAttachments:         2,
		MaxPhotoAttachmentSizeBytes: 2048,
		MaxAttachmentSizeBytes:      4096,
		MaxTotalSizeBytes:           8192,
		IsEverythingDocument:        false,
	}

	attachments, err := loader.LoadMultipleAttachments(filePaths)
	require.NoError(t, err)
	assert.Len(t, attachments, 2)
	assert.Equal(t, attachment.Document, attachments[0].AType)
	assert.Equal(t, attachment.Document, attachments[1].AType)
}

func TestLoadMultipleAttachments_AllAttachmentsAsDocuments(t *testing.T) {
	filePath1 := "abc1/file1.jpg"
	filePath2 := "abc1/file2.jpg"
	filePaths := []string{filePath1, filePath2}

	mockOpener := &mockFileOpener{
		files: map[string]*attachment.OpenedFile{
			filePath1: {File: newMockReadCloser("dummy content"), SizeBytes: 1024},
			filePath2: {File: newMockReadCloser("dummy content"), SizeBytes: 2048},
		},
	}

	loader := &attachment.Loader{
		FileOpener:                  mockOpener,
		MaxTotalAttachments:         2,
		MaxPhotoAttachmentSizeBytes: 2048,
		MaxAttachmentSizeBytes:      4096,
		MaxTotalSizeBytes:           8192,
		IsEverythingDocument:        true,
	}

	attachments, err := loader.LoadMultipleAttachments(filePaths)
	require.NoError(t, err)
	assert.Len(t, attachments, 2)
	assert.Equal(t, attachment.Document, attachments[0].AType)
	assert.Equal(t, attachment.Document, attachments[1].AType)
}
