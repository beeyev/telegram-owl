package attachment

import (
	"errors"
	"fmt"
	"path/filepath"
)

// Loader validates and loads attachments according to configured limits.
type Loader struct {
	FileOpener                  FileOpener
	IsEverythingDocument        bool
	MaxTotalAttachments         int
	MaxPhotoAttachmentSizeBytes int64
	MaxAttachmentSizeBytes      int64
	MaxTotalSizeBytes           int64
}

func (l *Loader) LoadMultipleAttachments(filePaths []string) (Attachments, error) {
	if err := l.validateAttachments(filePaths); err != nil {
		return nil, err
	}

	attachments := make(Attachments, 0, len(filePaths))
	typesFound := make(map[AType]bool)
	var totalSizeBytes int64

	for _, path := range filePaths {
		attachment, err := l.loadAttachment(path)
		if err != nil {
			attachments.Close()
			return nil, err
		}

		totalSizeBytes += attachment.SizeBytes
		if totalSizeBytes > l.MaxTotalSizeBytes {
			// TODO find better way to close attachments
			attachment.Close()
			attachments.Close()
			return nil, fmt.Errorf("total attachments size exceeds the max allowed %d MB", bytesToMegabytes(l.MaxTotalSizeBytes))
		}

		attachments = append(attachments, attachment)
		if !l.IsEverythingDocument {
			typesFound[attachment.AType] = true
		}
	}

	if !l.IsEverythingDocument && !isOnlyPhotoOrVideo(typesFound) {
		for _, attach := range attachments {
			attach.AType = Document
		}
	}

	return attachments, nil
}

func (l *Loader) loadAttachment(filePath string) (*Attachment, error) {
	openedFile, err := l.FileOpener.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read attachment: %w", err)
	}

	attachmentType := l.determineAttachmentType(filePath, openedFile.SizeBytes)

	if openedFile.SizeBytes > l.MaxAttachmentSizeBytes {
		openedFile.File.Close()
		return nil, fmt.Errorf("attachment %q: size %d MB exceeds the max allowed of %d MB",
			filePath, bytesToMegabytes(openedFile.SizeBytes), bytesToMegabytes(l.MaxAttachmentSizeBytes))
	}

	return &Attachment{
		AType:     attachmentType,
		FileName:  filepath.Base(filePath),
		SizeBytes: openedFile.SizeBytes,
		File:      openedFile.File,
	}, nil
}

func (l *Loader) determineAttachmentType(filePath string, sizeBytes int64) AType {
	if l.IsEverythingDocument {
		return Document
	}

	attachmentType := DetectType(filePath)
	if attachmentType == Photo && sizeBytes > l.MaxPhotoAttachmentSizeBytes {
		return Document
	}

	return attachmentType
}

// validateAttachments checks limits before processing files.
func (l *Loader) validateAttachments(filePaths []string) error {
	if len(filePaths) == 0 {
		return errors.New("no attachments provided")
	}
	if len(filePaths) > l.MaxTotalAttachments {
		return fmt.Errorf("too many attachments: max allowed is %d, but got %d", l.MaxTotalAttachments, len(filePaths))
	}
	return nil
}
