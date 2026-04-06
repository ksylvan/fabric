package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/danielmiessler/fabric/internal/domain"
	debuglog "github.com/danielmiessler/fabric/internal/log"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	args := []string{"--copy"}
	expectedFlags := &Flags{Copy: true}
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = append([]string{"cmd"}, args...)

	flags, err := Init()
	assert.NoError(t, err)
	assert.Equal(t, expectedFlags.Copy, flags.Copy)
}

func TestInit_DefaultServerDefaultsAreLocalAndDebugOff(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd"}

	flags, err := Init()
	assert.NoError(t, err)
	assert.Equal(t, "127.0.0.1:8080", flags.ServeAddress)
	assert.Equal(t, 0, flags.Debug)
}

func TestReadStdin(t *testing.T) {
	input := "test input"
	stdin := io.NopCloser(strings.NewReader(input))
	// No need to cast stdin to *os.File, pass it as io.ReadCloser directly
	content, err := ReadStdin(stdin)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != input {
		t.Fatalf("expected %q, got %q", input, content)
	}
}

// ReadStdin function assuming it's part of `cli` package
func ReadStdin(reader io.ReadCloser) (string, error) {
	defer reader.Close()
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(reader)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func TestBuildChatOptions(t *testing.T) {
	flags := &Flags{
		Temperature:      0.8,
		TopP:             0.9,
		PresencePenalty:  0.1,
		FrequencyPenalty: 0.2,
		Seed:             1,
	}

	expectedOptions := &domain.ChatOptions{
		Temperature:      0.8,
		TopP:             0.9,
		PresencePenalty:  0.1,
		FrequencyPenalty: 0.2,
		Raw:              false,
		Seed:             1,
		Thinking:         domain.ThinkingLevel(""),
		SuppressThink:    false,
		ThinkStartTag:    "<think>",
		ThinkEndTag:      "</think>",
	}
	options, err := flags.BuildChatOptions()
	assert.NoError(t, err)
	assert.Equal(t, expectedOptions, options)
}

func TestBuildChatOptionsDefaultSeed(t *testing.T) {
	flags := &Flags{
		Temperature:      0.8,
		TopP:             0.9,
		PresencePenalty:  0.1,
		FrequencyPenalty: 0.2,
	}

	expectedOptions := &domain.ChatOptions{
		Temperature:      0.8,
		TopP:             0.9,
		PresencePenalty:  0.1,
		FrequencyPenalty: 0.2,
		Raw:              false,
		Seed:             0,
		Thinking:         domain.ThinkingLevel(""),
		SuppressThink:    false,
		ThinkStartTag:    "<think>",
		ThinkEndTag:      "</think>",
	}
	options, err := flags.BuildChatOptions()
	assert.NoError(t, err)
	assert.Equal(t, expectedOptions, options)
}

func TestBuildChatOptionsSuppressThink(t *testing.T) {
	flags := &Flags{
		SuppressThink: true,
		ThinkStartTag: "[[t]]",
		ThinkEndTag:   "[[/t]]",
	}

	options, err := flags.BuildChatOptions()
	assert.NoError(t, err)
	assert.True(t, options.SuppressThink)
	assert.Equal(t, "[[t]]", options.ThinkStartTag)
	assert.Equal(t, "[[/t]]", options.ThinkEndTag)
}

func TestInitWithYAMLConfig(t *testing.T) {
	// Create a temporary YAML config file
	configContent := `
temperature: 0.9
model: gpt-4
pattern: analyze
stream: true
visual: true
visualSensitivity: 0.25
visualFPS: 3
`
	tmpfile, err := os.CreateTemp("", "config.*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(configContent)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Test 1: Basic YAML loading
	t.Run("Load YAML config", func(t *testing.T) {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		os.Args = []string{"cmd", "--config", tmpfile.Name()}

		flags, err := Init()
		assert.NoError(t, err)
		assert.Equal(t, 0.9, flags.Temperature)
		assert.Equal(t, "gpt-4", flags.Model)
		assert.Equal(t, "analyze", flags.Pattern)
		assert.True(t, flags.Stream)
		assert.True(t, flags.YouTubeVisual)
		assert.Equal(t, 0.25, flags.YouTubeVisualSensitivity)
		assert.Equal(t, 3, flags.YouTubeVisualFPS)
	})

	// Test 2: CLI overrides YAML
	t.Run("CLI overrides YAML", func(t *testing.T) {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		os.Args = []string{
			"cmd", "--config", tmpfile.Name(),
			"--temperature", "0.7",
			"--model", "gpt-3.5-turbo",
			"--visual-sensitivity", "0.6",
			"--visual-fps", "5",
		}

		flags, err := Init()
		assert.NoError(t, err)
		assert.Equal(t, 0.7, flags.Temperature)
		assert.Equal(t, "gpt-3.5-turbo", flags.Model)
		assert.Equal(t, "analyze", flags.Pattern) // unchanged from YAML
		assert.True(t, flags.Stream)              // unchanged from YAML
		assert.True(t, flags.YouTubeVisual)       // unchanged from YAML
		assert.Equal(t, 0.6, flags.YouTubeVisualSensitivity)
		assert.Equal(t, 5, flags.YouTubeVisualFPS)
	})

	// Test 3: Invalid YAML config
	t.Run("Invalid YAML config", func(t *testing.T) {
		badConfig := "temperature: \"not a float\"\nmodel: 123  # should be string\n"
		badfile, err := os.CreateTemp("", "bad-config.*.yaml")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(badfile.Name())

		if _, err := badfile.Write([]byte(badConfig)); err != nil {
			t.Fatal(err)
		}
		if err := badfile.Close(); err != nil {
			t.Fatal(err)
		}

		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		os.Args = []string{"cmd", "--config", badfile.Name()}

		_, err = Init()
		assert.Error(t, err)
	})

	// Test 4: Unknown YAML keys are rejected
	t.Run("Unknown YAML key", func(t *testing.T) {
		unknownKeyConfig := "model: gpt-4\nnotARealFlag: true\n"
		unknownFile, err := os.CreateTemp("", "unknown-config.*.yaml")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(unknownFile.Name())

		if _, err := unknownFile.Write([]byte(unknownKeyConfig)); err != nil {
			t.Fatal(err)
		}
		if err := unknownFile.Close(); err != nil {
			t.Fatal(err)
		}

		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		os.Args = []string{"cmd", "--config", unknownFile.Name()}

		_, err = Init()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "notARealFlag")
		assert.Contains(t, err.Error(), "not found in type cli.Flags")
	})
}

func TestLoadYAMLConfigDebugOutputRedactsValues(t *testing.T) {
	configContent := `
model: gpt-4
temperature: 0.4
notificationCommand: super-secret-server-key
`

	tmpfile, err := os.CreateTemp("", "config-debug.*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(configContent)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	oldLevel := debuglog.GetLevel()
	defer debuglog.SetLevel(oldLevel)
	debuglog.SetLevel(debuglog.Detailed)

	var buf bytes.Buffer
	debuglog.SetOutput(&buf)
	defer debuglog.SetOutput(os.Stderr)

	cfg, err := loadYAMLConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("loadYAMLConfig returned error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected config to be loaded")
	}

	logged := buf.String()
	if strings.Contains(logged, "super-secret-server-key") {
		t.Fatalf("expected debug output to redact config values, got %q", logged)
	}
	if strings.Contains(logged, tmpfile.Name()) {
		t.Fatalf("expected debug output to omit absolute config path, got %q", logged)
	}
	if !strings.Contains(logged, "Loaded YAML config") {
		t.Fatalf("expected debug output to mention the config load, got %q", logged)
	}
	if !strings.Contains(logged, "model") || !strings.Contains(logged, "temperature") {
		t.Fatalf("expected debug output to summarize configured yaml keys, got %q", logged)
	}
	if !strings.Contains(logged, "notificationCommand") {
		t.Fatalf("expected debug output to summarize notificationCommand, got %q", logged)
	}
}

func TestLoadYAMLConfigRejectsMultipleDocuments(t *testing.T) {
	configContent := `
model: gpt-4
---
temperature: 0.4
`

	tmpfile, err := os.CreateTemp("", "config-multidoc.*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(configContent)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	_, err = loadYAMLConfig(tmpfile.Name())
	if err == nil {
		t.Fatal("expected multiple YAML documents to be rejected")
	}
	if !strings.Contains(err.Error(), "multiple YAML documents are not supported") {
		t.Fatalf("expected multi-document error, got %q", err.Error())
	}
}

func TestInitDebugOutputSummarizesAppliedYAMLKeysWithoutValues(t *testing.T) {
	configContent := `
model: gpt-4
notificationCommand: super-secret-server-key
visual: true
`

	tmpfile, err := os.CreateTemp("", "config-init-debug.*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(configContent)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "--debug", "2", "--config", tmpfile.Name()}

	oldLevel := debuglog.GetLevel()
	defer debuglog.SetLevel(oldLevel)

	var buf bytes.Buffer
	debuglog.SetOutput(&buf)
	defer debuglog.SetOutput(os.Stderr)

	flags, err := Init()
	if err != nil {
		t.Fatalf("Init returned error: %v", err)
	}
	if flags.NotificationCommand != "super-secret-server-key" {
		t.Fatalf("expected notification command to be loaded from YAML, got %q", flags.NotificationCommand)
	}
	if !flags.YouTubeVisual {
		t.Fatal("expected visual flag to be loaded from YAML")
	}

	logged := buf.String()
	if strings.Contains(logged, "super-secret-server-key") {
		t.Fatalf("expected Init debug output to omit YAML values, got %q", logged)
	}
	if !strings.Contains(logged, "Applied YAML config keys:") {
		t.Fatalf("expected Init debug output to summarize applied keys, got %q", logged)
	}
	for _, key := range []string{"model", "notificationCommand", "visual"} {
		if !strings.Contains(logged, key) {
			t.Fatalf("expected Init debug output to include %q, got %q", key, logged)
		}
	}
}

func TestValidateImageFile(t *testing.T) {
	t.Run("Empty path should be valid", func(t *testing.T) {
		err := validateImageFile("")
		assert.NoError(t, err)
	})

	t.Run("Valid extensions should pass", func(t *testing.T) {
		validExtensions := []string{".png", ".jpeg", ".jpg", ".webp"}
		for _, ext := range validExtensions {
			filename := "/tmp/test" + ext
			err := validateImageFile(filename)
			assert.NoError(t, err, "Extension %s should be valid", ext)
		}
	})

	t.Run("Invalid extensions should fail", func(t *testing.T) {
		invalidExtensions := []string{".gif", ".bmp", ".tiff", ".svg", ".txt", ""}
		for _, ext := range invalidExtensions {
			filename := "/tmp/test" + ext
			err := validateImageFile(filename)
			assert.Error(t, err, "Extension %s should be invalid", ext)
			assert.Contains(t, err.Error(), "invalid image file extension")
		}
	})

	t.Run("Existing file should fail", func(t *testing.T) {
		// Create a temporary file
		tempFile, err := os.CreateTemp("", "test*.png")
		assert.NoError(t, err)
		defer os.Remove(tempFile.Name())
		tempFile.Close()

		// Validation should fail because file exists
		err = validateImageFile(tempFile.Name())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "image file already exists")
	})

	t.Run("Non-existing file with valid extension should pass", func(t *testing.T) {
		nonExistentFile := filepath.Join(os.TempDir(), "non_existent_file.png")
		// Make sure the file doesn't exist
		os.Remove(nonExistentFile)

		err := validateImageFile(nonExistentFile)
		assert.NoError(t, err)
	})
}

func TestBuildChatOptionsWithImageFileValidation(t *testing.T) {
	t.Run("Valid image file should pass", func(t *testing.T) {
		flags := &Flags{
			ImageFile: "/tmp/output.png",
		}

		options, err := flags.BuildChatOptions()
		assert.NoError(t, err)
		assert.Equal(t, "/tmp/output.png", options.ImageFile)
	})

	t.Run("Invalid extension should fail", func(t *testing.T) {
		flags := &Flags{
			ImageFile: "/tmp/output.gif",
		}

		options, err := flags.BuildChatOptions()
		assert.Error(t, err)
		assert.Nil(t, options)
		assert.Contains(t, err.Error(), "invalid image file extension")
	})

	t.Run("Existing file should fail", func(t *testing.T) {
		// Create a temporary file
		tempFile, err := os.CreateTemp("", "existing*.png")
		assert.NoError(t, err)
		defer os.Remove(tempFile.Name())
		tempFile.Close()

		flags := &Flags{
			ImageFile: tempFile.Name(),
		}

		options, err := flags.BuildChatOptions()
		assert.Error(t, err)
		assert.Nil(t, options)
		assert.Contains(t, err.Error(), "image file already exists")
	})
}

func TestValidateImageParameters(t *testing.T) {
	t.Run("No image file and no parameters should pass", func(t *testing.T) {
		err := validateImageParameters("", "", "", "", 0)
		assert.NoError(t, err)
	})

	t.Run("Image parameters without image file should fail", func(t *testing.T) {
		// Test each parameter individually
		err := validateImageParameters("", "1024x1024", "", "", 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "image parameters")
		assert.Contains(t, err.Error(), "can only be used with --image-file")

		err = validateImageParameters("", "", "high", "", 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "image parameters")

		err = validateImageParameters("", "", "", "transparent", 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "image parameters")

		err = validateImageParameters("", "", "", "", 50)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "image parameters")

		// Test multiple parameters
		err = validateImageParameters("", "1024x1024", "high", "transparent", 50)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "image parameters")
	})

	t.Run("Valid size values should pass", func(t *testing.T) {
		validSizes := []string{"1024x1024", "1536x1024", "1024x1536", "auto"}
		for _, size := range validSizes {
			err := validateImageParameters("/tmp/test.png", size, "", "", 0)
			assert.NoError(t, err, "Size %s should be valid", size)
		}
	})

	t.Run("Invalid size should fail", func(t *testing.T) {
		err := validateImageParameters("/tmp/test.png", "invalid", "", "", 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid image size")
	})

	t.Run("Valid quality values should pass", func(t *testing.T) {
		validQualities := []string{"low", "medium", "high", "auto"}
		for _, quality := range validQualities {
			err := validateImageParameters("/tmp/test.png", "", quality, "", 0)
			assert.NoError(t, err, "Quality %s should be valid", quality)
		}
	})

	t.Run("Invalid quality should fail", func(t *testing.T) {
		err := validateImageParameters("/tmp/test.png", "", "invalid", "", 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid image quality")
	})

	t.Run("Valid background values should pass", func(t *testing.T) {
		validBackgrounds := []string{"opaque", "transparent"}
		for _, background := range validBackgrounds {
			err := validateImageParameters("/tmp/test.png", "", "", background, 0)
			assert.NoError(t, err, "Background %s should be valid", background)
		}
	})

	t.Run("Invalid background should fail", func(t *testing.T) {
		err := validateImageParameters("/tmp/test.png", "", "", "invalid", 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid image background")
	})

	t.Run("Compression for JPEG should pass", func(t *testing.T) {
		err := validateImageParameters("/tmp/test.jpg", "", "", "", 75)
		assert.NoError(t, err)
	})

	t.Run("Compression for WebP should pass", func(t *testing.T) {
		err := validateImageParameters("/tmp/test.webp", "", "", "", 50)
		assert.NoError(t, err)
	})

	t.Run("Compression for PNG should fail", func(t *testing.T) {
		err := validateImageParameters("/tmp/test.png", "", "", "", 75)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "image compression can only be used with JPEG and WebP formats")
	})

	t.Run("Invalid compression range should fail", func(t *testing.T) {
		err := validateImageParameters("/tmp/test.jpg", "", "", "", 150)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "image compression must be between 0 and 100")

		err = validateImageParameters("/tmp/test.jpg", "", "", "", -10)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "image compression must be between 0 and 100")
	})

	t.Run("Transparent background for PNG should pass", func(t *testing.T) {
		err := validateImageParameters("/tmp/test.png", "", "", "transparent", 0)
		assert.NoError(t, err)
	})

	t.Run("Transparent background for WebP should pass", func(t *testing.T) {
		err := validateImageParameters("/tmp/test.webp", "", "", "transparent", 0)
		assert.NoError(t, err)
	})

	t.Run("Transparent background for JPEG should fail", func(t *testing.T) {
		err := validateImageParameters("/tmp/test.jpg", "", "", "transparent", 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "transparent background can only be used with PNG and WebP formats")
	})
}

func TestBuildChatOptionsWithImageParameters(t *testing.T) {
	t.Run("Valid image parameters should pass", func(t *testing.T) {
		flags := &Flags{
			ImageFile:        "/tmp/test.png",
			ImageSize:        "1024x1024",
			ImageQuality:     "high",
			ImageBackground:  "transparent",
			ImageCompression: 0, // Not set for PNG
		}

		options, err := flags.BuildChatOptions()
		assert.NoError(t, err)
		assert.NotNil(t, options)
		assert.Equal(t, "/tmp/test.png", options.ImageFile)
		assert.Equal(t, "1024x1024", options.ImageSize)
		assert.Equal(t, "high", options.ImageQuality)
		assert.Equal(t, "transparent", options.ImageBackground)
		assert.Equal(t, 0, options.ImageCompression)
	})

	t.Run("Invalid image parameters should fail", func(t *testing.T) {
		flags := &Flags{
			ImageFile:       "/tmp/test.png",
			ImageSize:       "invalid",
			ImageQuality:    "high",
			ImageBackground: "transparent",
		}

		options, err := flags.BuildChatOptions()
		assert.Error(t, err)
		assert.Nil(t, options)
		assert.Contains(t, err.Error(), "invalid image size")
	})

	t.Run("JPEG with compression should pass", func(t *testing.T) {
		flags := &Flags{
			ImageFile:        "/tmp/test.jpg",
			ImageSize:        "1536x1024",
			ImageQuality:     "medium",
			ImageBackground:  "opaque",
			ImageCompression: 80,
		}

		options, err := flags.BuildChatOptions()
		assert.NoError(t, err)
		assert.NotNil(t, options)
		assert.Equal(t, 80, options.ImageCompression)
	})

	t.Run("Image parameters without image file should fail in BuildChatOptions", func(t *testing.T) {
		flags := &Flags{
			ImageSize: "1024x1024", // Image parameter without ImageFile
		}

		options, err := flags.BuildChatOptions()
		assert.Error(t, err)
		assert.Nil(t, options)
		assert.Contains(t, err.Error(), "image parameters")
		assert.Contains(t, err.Error(), "can only be used with --image-file")
	})
}

func TestExtractFlag(t *testing.T) {
	tests := []struct {
		name     string
		arg      string
		expected string
	}{
		// Unix-style flags
		{"long flag", "--help", "help"},
		{"long flag with value", "--pattern=analyze", "pattern"},
		{"short flag", "-h", "h"},
		{"short flag with value", "-p=test", "p"},
		{"single dash", "-", ""},
		{"double dash only", "--", ""},

		// Non-flags
		{"regular arg", "analyze", ""},
		{"path arg", "./file.txt", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFlag(tt.arg)
			assert.Equal(t, tt.expected, result)
		})
	}
}
