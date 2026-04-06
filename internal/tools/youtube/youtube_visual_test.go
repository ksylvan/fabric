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

func TestGrabVisualParallelOCRPreservesFrameOrder(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix shell fixture")
	}

	binDir := t.TempDir()
	writeExecutable(t, binDir, "yt-dlp", "#!/bin/sh\nprintf '%s\\n' 'https://cdn.example.com/video'\n")
	writeExecutable(t, binDir, "ffmpeg", "#!/bin/sh\nlast=''\nfor arg in \"$@\"; do\n\tlast=\"$arg\"\ndone\nfor n in 1 2 3; do\n\tframe=$(printf '%s' \"$last\" | sed \"s/%04d/$(printf '%04d' \"$n\")/\")\n\t: > \"$frame\"\ndone\n")
	writeExecutable(t, binDir, "tesseract", "#!/bin/sh\nbase=$(basename \"$1\")\ncase \"$base\" in\n\tframe_0001.jpg)\n\t\tsleep 0.03\n\t\tprintf '%s\\n' 'recognized text frame one'\n\t\t;;\n\tframe_0002.jpg)\n\t\tsleep 0.01\n\t\tprintf '%s\\n' 'recognized text frame two'\n\t\t;;\n\tframe_0003.jpg)\n\t\tsleep 0.02\n\t\tprintf '%s\\n' 'recognized text frame three'\n\t\t;;\n\tesac\n")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	yt := NewYouTube()
	got, err := yt.GrabVisual("video123", "en", "", 0.4, 0)
	if err != nil {
		t.Fatalf("GrabVisual returned error: %v", err)
	}

	expectedOrder := []string{
		"recognized text frame one",
		"recognized text frame two",
		"recognized text frame three",
	}

	lastIndex := -1
	for _, want := range expectedOrder {
		currentIndex := strings.Index(got, want)
		if currentIndex == -1 {
			t.Fatalf("expected output to include %q, got %q", want, got)
		}
		if currentIndex <= lastIndex {
			t.Fatalf("expected %q to appear after the previous frame text, got %q", want, got)
		}
		lastIndex = currentIndex
	}

	if count := strings.Count(got, "-->"); count != 3 {
		t.Fatalf("expected three visual cues, got %d in %q", count, got)
	}
}

func TestGrabVisualRejectsDangerousYTDlpArgs(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix shell fixture")
	}

	binDir := t.TempDir()
	writeExecutable(t, binDir, "yt-dlp", "#!/bin/sh\nprintf '%s\\n' 'yt-dlp should not run' >&2\nexit 99\n")
	writeExecutable(t, binDir, "ffmpeg", "#!/bin/sh\nexit 0\n")
	writeExecutable(t, binDir, "tesseract", "#!/bin/sh\nexit 0\n")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	yt := NewYouTube()
	_, err := yt.GrabVisual("video123", "en", "--exec-before-download echo hacked", 0.4, 0)
	if err == nil {
		t.Fatal("expected invalid yt-dlp arguments error")
	}
	if !strings.Contains(err.Error(), "invalid yt-dlp arguments") {
		t.Fatalf("expected yt-dlp validation error, got %q", err.Error())
	}
	if strings.Contains(err.Error(), "yt-dlp should not run") {
		t.Fatalf("expected GrabVisual to reject args before invoking yt-dlp, got %q", err.Error())
	}
}

func writeExecutable(t *testing.T, dir, name, content string) {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
