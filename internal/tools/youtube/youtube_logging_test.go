package youtube

import (
	"bytes"
	"os"
	"path/filepath"
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

func TestBuildSafeYTDlpArgsRejectsDangerousFlags(t *testing.T) {
	t.Parallel()

	testCases := []string{
		"--exec=sh -c 'echo hacked'",
		"--exec-before-download echo hacked",
		"--config-locations /tmp/yt-dlp.conf",
		"--plugin-dirs=/tmp/plugins",
		"--alias safe '--exec echo hacked'",
	}

	for _, additionalArgs := range testCases {
		additionalArgs := additionalArgs
		t.Run(additionalArgs, func(t *testing.T) {
			t.Parallel()

			_, err := buildSafeYTDlpArgs([]string{"--get-url"}, additionalArgs)
			if err == nil {
				t.Fatalf("expected %q to be rejected", additionalArgs)
			}
			if !strings.Contains(err.Error(), "invalid yt-dlp arguments") {
				t.Fatalf("expected yt-dlp validation error, got %q", err.Error())
			}
		})
	}
}

func TestBuildSafeYTDlpArgsAllowsExpectedAuthenticationFlags(t *testing.T) {
	t.Parallel()

	got, err := buildSafeYTDlpArgs(
		[]string{"--write-auto-subs"},
		"--cookies-from-browser brave --sleep-requests 2",
	)
	if err != nil {
		t.Fatalf("buildSafeYTDlpArgs returned error: %v", err)
	}

	want := []string{
		"--ignore-config",
		"--write-auto-subs",
		"--cookies-from-browser",
		"brave",
		"--sleep-requests",
		"2",
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d args, got %d (%q)", len(want), len(got), strings.Join(got, " "))
	}
	for i, wantArg := range want {
		if got[i] != wantArg {
			t.Fatalf("expected arg %d to be %q, got %q", i, wantArg, got[i])
		}
	}
}

func TestReadAndCleanVTTFileReadErrorOmitsTempPath(t *testing.T) {
	t.Parallel()

	yt := NewYouTube()
	missingFile := filepath.Join(t.TempDir(), "missing.vtt")

	_, err := yt.readAndCleanVTTFile(missingFile)
	if err == nil {
		t.Fatal("expected VTT read error")
	}
	if strings.Contains(err.Error(), missingFile) {
		t.Fatalf("expected read error to omit internal path, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "open:") {
		t.Fatalf("expected sanitized filesystem error context, got %q", err.Error())
	}
}
