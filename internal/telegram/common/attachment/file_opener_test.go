package attachment_test

import (
	"io"
	"os"
	"testing"

	"github.com/beeyev/telegram-owl/internal/telegram/common/attachment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOSFileOpener_Open(t *testing.T) {
	opener := attachment.OSFileOpener{}

	t.Run("Valid File", func(t *testing.T) {
		// Create a temporary file for testing
		tempFile, err := os.CreateTemp(t.TempDir(), "testfile")
		require.NoError(t, err)
		defer os.Remove(tempFile.Name()) // Cleanup after test

		// Write some test data
		_, err = tempFile.WriteString("test data")
		require.NoError(t, err)
		tempFile.Close() // Close to avoid file lock issues

		// Open the file using OSFileOpener
		openedFile, err := opener.Open(tempFile.Name())
		require.NoError(t, err)
		assert.NotNil(t, openedFile)
		assert.Equal(t, int64(9), openedFile.SizeBytes)

		// Ensure the file can be read
		data, err := io.ReadAll(openedFile.File)
		require.NoError(t, err)
		assert.Equal(t, "test data", string(data))

		// Close the file
		err = openedFile.File.Close()
		assert.NoError(t, err)
	})

	t.Run("Non-Existent File", func(t *testing.T) {
		fileName := "non_existent_file.txt"
		openedFile, err := opener.Open(fileName)
		require.Error(t, err)
		assert.Nil(t, openedFile)
		assert.Contains(t, err.Error(), fileName)
	})

	t.Run("Empty File Path", func(t *testing.T) {
		openedFile, err := opener.Open("")
		require.Error(t, err)
		assert.Nil(t, openedFile)
		assert.Equal(t, "file path cannot be empty", err.Error())
	})
}
