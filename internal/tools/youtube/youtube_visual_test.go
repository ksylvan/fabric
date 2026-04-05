package youtube

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGrabVisualRedactsFFmpegOutput(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix shell fixture")
	}

	binDir := t.TempDir()
	writeExecutable(t, binDir, "yt-dlp", "#!/bin/sh\nprintf '%s\\n' 'https://cdn.example.com/video?token=super-secret-token'\n")
	writeExecutable(t, binDir, "ffmpeg", "#!/bin/sh\nprintf '%s\\n' 'ffmpeg saw https://cdn.example.com/video?token=super-secret-token' >&2\nexit 1\n")
	writeExecutable(t, binDir, "tesseract", "#!/bin/sh\nexit 0\n")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	yt := NewYouTube()
	_, err := yt.GrabVisual("video123", "en", "", 0.4, 0)
	if err == nil {
		t.Fatal("expected ffmpeg failure")
	}
	if !strings.Contains(err.Error(), "ffmpeg frame extraction failed") {
		t.Fatalf("expected ffmpeg context, got %q", err.Error())
	}
	if strings.Contains(err.Error(), "super-secret-token") {
		t.Fatalf("expected ffmpeg error to redact sensitive output, got %q", err.Error())
	}
}

func TestGrabVisualRedactsTesseractStderr(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix shell fixture")
	}

	binDir := t.TempDir()
	writeExecutable(t, binDir, "yt-dlp", "#!/bin/sh\nprintf '%s\\n' 'https://cdn.example.com/video?token=super-secret-token'\n")
	writeExecutable(t, binDir, "ffmpeg", "#!/bin/sh\nlast=''\nfor arg in \"$@\"; do\n\tlast=\"$arg\"\ndone\nframe=$(printf '%s' \"$last\" | sed 's/%04d/0001/')\n: > \"$frame\"\n")
	writeExecutable(t, binDir, "tesseract", "#!/bin/sh\nprintf '%s\\n' 'tesseract secret stderr super-secret-token' >&2\nexit 1\n")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	yt := NewYouTube()
	_, err := yt.GrabVisual("video123", "en", "", 0.4, 0)
	if err == nil {
		t.Fatal("expected tesseract failure")
	}
	if !strings.Contains(err.Error(), "tesseract failed on frame 0") {
		t.Fatalf("expected tesseract context, got %q", err.Error())
	}
	if strings.Contains(err.Error(), "super-secret-token") {
		t.Fatalf("expected tesseract error to redact sensitive stderr, got %q", err.Error())
	}
}

func writeExecutable(t *testing.T, dir, name, content string) {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
