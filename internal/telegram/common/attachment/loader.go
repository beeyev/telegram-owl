package attachment

import (
	"errors"
	"fmt"
	"path/filepath"
)

// Loader validates, classifies, and opens attachments according to configured
// limits. The caller owns returned files only when loading succeeds.
type Loader struct {
	FileOpener                  FileOpener
	IsEverythingDocument        bool
	MaxTotalAttachments         int
	MaxPhotoAttachmentSizeBytes int64
	MaxAttachmentSizeBytes      int64
	MaxTotalSizeBytes           int64
}

// LoadMultipleAttachments opens every path and returns a Telegram-compatible
// group. On failure it closes the current file and everything opened earlier,
// joining any cleanup failures with the primary error.
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
			return nil, errors.Join(err, attachments.Close())
		}

		totalSizeBytes += attachment.SizeBytes
		if totalSizeBytes > l.MaxTotalSizeBytes {
			limitErr := fmt.Errorf(
				"total attachments size exceeds the max allowed %d MB",
				bytesToMegabytes(l.MaxTotalSizeBytes),
			)
			return nil, errors.Join(limitErr, attachment.Close(), attachments.Close())
		}

		attachments = append(attachments, attachment)
		if !l.IsEverythingDocument {
			typesFound[attachment.AType] = true
		}
	}

	if !l.IsEverythingDocument && !isOnlyPhotoOrVideo(typesFound) {
		// A photo/video album may mix those two types. If any document or audio
		// is present, normalize the entire group to documents so Telegram accepts
		// one compatible media class.
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
		sizeErr := fmt.Errorf(
			"attachment %q: size %d MB exceeds the max allowed of %d MB",
			filePath,
			bytesToMegabytes(openedFile.SizeBytes),
			bytesToMegabytes(l.MaxAttachmentSizeBytes),
		)
		return nil, errors.Join(sizeErr, openedFile.File.Close())
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
		// Telegram accepts a larger file as a document even when the extension
		// would normally classify it as a photo.
		return Document
	}

	return attachmentType
}

// validateAttachments rejects collection-level errors before any file is
// opened, keeping these failure paths free of resource cleanup.
func (l *Loader) validateAttachments(filePaths []string) error {
	if len(filePaths) == 0 {
		return errors.New("no attachments provided")
	}
	if len(filePaths) > l.MaxTotalAttachments {
		return fmt.Errorf("too many attachments: max allowed is %d, but got %d", l.MaxTotalAttachments, len(filePaths))
	}
	return nil
}
