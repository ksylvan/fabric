package openai

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/danielmiessler/fabric/internal/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTranscribeFile_ValidationErrorsAreLocalized(t *testing.T) {
	_, err := i18n.Init("en")
	require.NoError(t, err)

	client := &Client{}

	audioFile, err := os.CreateTemp("", "transcribe-valid-*.mp3")
	require.NoError(t, err)
	require.NoError(t, audioFile.Close())
	t.Cleanup(func() { _ = os.Remove(audioFile.Name()) })

	_, err = client.TranscribeFile(context.Background(), audioFile.Name(), "not-a-model", false)
	require.Error(t, err)
	assert.Equal(t,
		fmt.Sprintf(i18n.T("openai_audio_model_not_supported_for_transcription"), "not-a-model"),
		err.Error(),
	)

	unsupportedFile, err := os.CreateTemp("", "transcribe-invalid-*.txt")
	require.NoError(t, err)
	require.NoError(t, unsupportedFile.Close())
	t.Cleanup(func() { _ = os.Remove(unsupportedFile.Name()) })

	_, err = client.TranscribeFile(context.Background(), unsupportedFile.Name(), AllowedTranscriptionModels[0], false)
	require.Error(t, err)
	assert.Equal(t,
		fmt.Sprintf(i18n.T("openai_audio_unsupported_audio_format"), filepath.Ext(unsupportedFile.Name())),
		err.Error(),
	)
}

func TestTranscribeFile_FileSizeLimitErrorIsLocalized(t *testing.T) {
	_, err := i18n.Init("en")
	require.NoError(t, err)

	client := &Client{}
	largeFile, err := os.CreateTemp("", "transcribe-large-*.mp3")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Remove(largeFile.Name()) })

	require.NoError(t, largeFile.Truncate(MaxAudioFileSize+1))
	require.NoError(t, largeFile.Close())

	_, err = client.TranscribeFile(context.Background(), largeFile.Name(), AllowedTranscriptionModels[0], false)
	require.Error(t, err)
	assert.Equal(t,
		fmt.Sprintf(i18n.T("openai_audio_file_exceeds_limit_enable_split"), largeFile.Name()),
		err.Error(),
	)
}
