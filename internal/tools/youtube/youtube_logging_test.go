package youtube

import (
	"bytes"
	"os"
	"strings"
	"testing"

	debuglog "github.com/danielmiessler/fabric/internal/log"
)

func TestSanitizeYTArgsRedactsPasswordAndHeaderValues(t *testing.T) {
	args := []string{
		"yt-dlp",
		"--cookies-from-browser", "brave:Default",
		"--password=super-secret-password",
		"--add-header", "Authorization: Bearer abc123",
		"https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		"https://cdn.example.com/video?token=super-secret-token",
	}

	got := strings.Join(sanitizeYTArgs(args), " ")

	for _, secret := range []string{
		"brave:Default",
		"super-secret-password",
		"Authorization: Bearer abc123",
		"youtube.com",
		"cdn.example.com",
		"super-secret-token",
	} {
		if strings.Contains(got, secret) {
			t.Fatalf("expected sanitized args to redact %q, got %q", secret, got)
		}
	}
	if !strings.Contains(got, "<redacted>") {
		t.Fatalf("expected sanitized args to include redaction marker, got %q", got)
	}
	if strings.Count(got, "<redacted-url>") != 2 {
		t.Fatalf("expected both URLs to be redacted, got %q", got)
	}
}

func TestDetectErrorTraceLoggingRedactsURLs(t *testing.T) {
	oldLevel := debuglog.GetLevel()
	defer debuglog.SetLevel(oldLevel)
	debuglog.SetLevel(debuglog.Trace)

	var buf bytes.Buffer
	debuglog.SetOutput(&buf)
	defer debuglog.SetOutput(os.Stderr)

	err := detectError(strings.NewReader("https://cdn.example.com/video?token=super-secret-token\n429\n"))
	if err == nil {
		t.Fatal("expected detectError to return a mapped error")
	}

	logged := buf.String()
	if strings.Contains(logged, "super-secret-token") || strings.Contains(logged, "cdn.example.com") {
		t.Fatalf("expected trace log output to redact signed URL, got %q", logged)
	}
	if !strings.Contains(logged, "<redacted-url>") {
		t.Fatalf("expected trace log output to include redacted URL marker, got %q", logged)
	}
}
