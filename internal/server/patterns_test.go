package restapi

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/danielmiessler/fabric/internal/plugins/db/fsdb"
	"github.com/gin-gonic/gin"
)

func TestPatternsHandler_ApplyPatternDoesNotTreatPatternNameAsFilePath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	workDir := t.TempDir()
	t.Chdir(workDir)

	if err := os.WriteFile(filepath.Join(workDir, ".env"), []byte("TOP_SECRET"), 0o644); err != nil {
		t.Fatalf("failed to write sentinel file: %v", err)
	}

	patternsDir := t.TempDir()
	patterns := &fsdb.PatternsEntity{
		StorageEntity: &fsdb.StorageEntity{
			Dir:       patternsDir,
			Label:     "patterns",
			ItemIsDir: true,
		},
		SystemPatternFile: "system.md",
	}

	router := gin.New()
	NewPatternsHandler(router, patterns)

	req := httptest.NewRequest(http.MethodPost, "/patterns/.env/apply", strings.NewReader(`{"input":"hello"}`))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}

	if strings.Contains(recorder.Body.String(), "TOP_SECRET") {
		t.Fatalf("expected response body to avoid disclosing local file contents, got %q", recorder.Body.String())
	}
}

func TestPatternsHandler_ApplyPatternRejectsUnknownJSONFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	patternsDir := t.TempDir()
	patterns := &fsdb.PatternsEntity{
		StorageEntity: &fsdb.StorageEntity{
			Dir:       patternsDir,
			Label:     "patterns",
			ItemIsDir: true,
		},
		SystemPatternFile: "system.md",
	}

	router := gin.New()
	NewPatternsHandler(router, patterns)

	req := httptest.NewRequest(http.MethodPost, "/patterns/example/apply", strings.NewReader(`{"input":"hello","unexpected":true}`))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
	if body := recorder.Body.String(); !strings.Contains(body, "unknown field") {
		t.Fatalf("expected unknown field error, got %q", body)
	}
}

func TestPatternsHandler_SaveRejectsUnsafeTemplateDirectives(t *testing.T) {
	gin.SetMode(gin.TestMode)

	patternsDir := t.TempDir()
	patterns := &fsdb.PatternsEntity{
		StorageEntity: &fsdb.StorageEntity{
			Dir:       patternsDir,
			Label:     "patterns",
			ItemIsDir: true,
		},
		SystemPatternFile: "system.md",
	}

	router := gin.New()
	NewPatternsHandler(router, patterns)

	req := httptest.NewRequest(http.MethodPost, "/patterns/example", strings.NewReader("{{ext:demo:run:1}}"))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
	if body := recorder.Body.String(); !strings.Contains(body, "template extensions are disabled") {
		t.Fatalf("expected extension restriction error, got %q", body)
	}
}

func TestPatternsHandler_ApplyPatternRejectsUnsafeStoredPatternDirective(t *testing.T) {
	gin.SetMode(gin.TestMode)

	patternsDir := t.TempDir()
	patterns := &fsdb.PatternsEntity{
		StorageEntity: &fsdb.StorageEntity{
			Dir:       patternsDir,
			Label:     "patterns",
			ItemIsDir: true,
		},
		SystemPatternFile: "system.md",
	}

	if err := os.MkdirAll(filepath.Join(patternsDir, "unsafe-pattern"), 0o755); err != nil {
		t.Fatalf("failed to create pattern directory: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(patternsDir, "unsafe-pattern", "system.md"),
		[]byte("{{plugin:sys:env:HOME}}"),
		0o644,
	); err != nil {
		t.Fatalf("failed to write pattern: %v", err)
	}

	router := gin.New()
	NewPatternsHandler(router, patterns)

	req := httptest.NewRequest(http.MethodPost, "/patterns/unsafe-pattern/apply", strings.NewReader(`{"input":"hello"}`))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
	if body := recorder.Body.String(); !strings.Contains(body, "template plugin") || !strings.Contains(body, "disabled for remote pattern execution") {
		t.Fatalf("expected unsafe plugin error, got %q", body)
	}
}

func TestPatternsHandler_ApplyPatternRejectsTemplateDirectiveInjectionInVariables(t *testing.T) {
	gin.SetMode(gin.TestMode)

	patternsDir := t.TempDir()
	patterns := &fsdb.PatternsEntity{
		StorageEntity: &fsdb.StorageEntity{
			Dir:       patternsDir,
			Label:     "patterns",
			ItemIsDir: true,
		},
		SystemPatternFile: "system.md",
	}

	if err := os.MkdirAll(filepath.Join(patternsDir, "safe-pattern"), 0o755); err != nil {
		t.Fatalf("failed to create pattern directory: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(patternsDir, "safe-pattern", "system.md"),
		[]byte("Role: {{role}}"),
		0o644,
	); err != nil {
		t.Fatalf("failed to write pattern: %v", err)
	}

	router := gin.New()
	NewPatternsHandler(router, patterns)

	req := httptest.NewRequest(
		http.MethodPost,
		"/patterns/safe-pattern/apply",
		strings.NewReader(`{"input":"hello","variables":{"role":"{{plugin:sys:env:HOME}}"}}`),
	)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
	if body := recorder.Body.String(); !strings.Contains(body, "cannot contain nested template directives") {
		t.Fatalf("expected nested directive rejection, got %q", body)
	}
}
