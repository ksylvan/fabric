package restapi

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/danielmiessler/fabric/internal/i18n"
	"github.com/danielmiessler/fabric/internal/plugins/db/fsdb"
	"github.com/gin-gonic/gin"
)

func captureServerLogs() (*bytes.Buffer, func()) {
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	return &buf, func() {
		slog.SetDefault(prev)
	}
}

func TestStorageHandler_RejectsPathTraversal(t *testing.T) {
	if _, err := i18n.Init("en"); err != nil {
		t.Fatalf("i18n.Init() error = %v", err)
	}
	gin.SetMode(gin.TestMode)

	root := t.TempDir()
	contextsDir := filepath.Join(root, "contexts")
	if err := os.MkdirAll(contextsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(root, "keep.txt")
	if err := os.WriteFile(marker, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := gin.New()
	NewContextsHandler(r, &fsdb.ContextsEntity{
		StorageEntity: &fsdb.StorageEntity{Label: "Contexts", Dir: contextsDir},
	})

	for _, path := range []string{"/contexts/..", "/contexts/%2e%2e"} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, path, nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("%s: got %d, want 400", path, w.Code)
		}
		if _, err := os.Stat(marker); err != nil {
			t.Fatalf("%s: parent was deleted: %v", path, err)
		}
		if _, err := os.Stat(contextsDir); err != nil {
			t.Fatalf("%s: contexts dir was deleted: %v", path, err)
		}
	}
}

func TestStorageHandler_LogsRejectedTraversal(t *testing.T) {
	if _, err := i18n.Init("en"); err != nil {
		t.Fatalf("i18n.Init() error = %v", err)
	}
	gin.SetMode(gin.TestMode)

	buf, restore := captureServerLogs()
	defer restore()

	r := gin.New()
	NewContextsHandler(r, &fsdb.ContextsEntity{
		StorageEntity: &fsdb.StorageEntity{Label: "Contexts", Dir: t.TempDir()},
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/contexts/%2e%2e", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want 400", w.Code)
	}
	if !strings.Contains(buf.String(), "Rejected invalid storage name") {
		t.Fatalf("missing security log entry: %q", buf.String())
	}
}

// Every storage route that takes a name must reject a traversal name, not
// only DELETE: all of them share storageError through the fsdb layer.
func TestStorageHandler_RejectsTraversalOnAllRoutes(t *testing.T) {
	if _, err := i18n.Init("en"); err != nil {
		t.Fatalf("i18n.Init() error = %v", err)
	}
	gin.SetMode(gin.TestMode)

	r := gin.New()
	NewContextsHandler(r, &fsdb.ContextsEntity{
		StorageEntity: &fsdb.StorageEntity{Label: "Contexts", Dir: t.TempDir()},
	})

	for _, tc := range []struct{ method, path string }{
		{http.MethodGet, "/contexts/%2e%2e"},
		{http.MethodDelete, "/contexts/%2e%2e"},
		{http.MethodPost, "/contexts/%2e%2e"},
		{http.MethodPut, "/contexts/rename/%2e%2e/ok"},
		{http.MethodPut, "/contexts/rename/ok/%2e%2e"},
	} {
		var body io.Reader
		if tc.method == http.MethodPost {
			body = strings.NewReader("x")
		}
		w := httptest.NewRecorder()
		req := httptest.NewRequest(tc.method, tc.path, body)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("%s %s: got %d, want 400", tc.method, tc.path, w.Code)
		}
	}
}

// A non-validation failure stays a 500, and its body must not leak the
// filesystem path that the wrapped *os.PathError carries.
func TestStorageHandler_GenericErrorHidesPaths(t *testing.T) {
	if _, err := i18n.Init("en"); err != nil {
		t.Fatalf("i18n.Init() error = %v", err)
	}
	gin.SetMode(gin.TestMode)

	dir := t.TempDir()
	r := gin.New()
	NewContextsHandler(r, &fsdb.ContextsEntity{
		StorageEntity: &fsdb.StorageEntity{Label: "Contexts", Dir: dir},
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/contexts/missing", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("got %d, want 500", w.Code)
	}
	if strings.Contains(w.Body.String(), dir) {
		t.Fatalf("500 body leaks the storage path: %s", w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "internal error") {
		t.Fatalf("500 body is not the generic message: %s", w.Body.String())
	}
}

func TestPatternsHandler_RejectsPathTraversalSave(t *testing.T) {
	if _, err := i18n.Init("en"); err != nil {
		t.Fatalf("i18n.Init() error = %v", err)
	}
	gin.SetMode(gin.TestMode)

	root := t.TempDir()
	patternsDir := filepath.Join(root, "patterns")
	if err := os.MkdirAll(patternsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	r := gin.New()
	NewPatternsHandler(r, &fsdb.PatternsEntity{
		StorageEntity:     &fsdb.StorageEntity{Label: "Patterns", Dir: patternsDir, ItemIsDir: true},
		SystemPatternFile: "system.md",
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/patterns/%2e%2e", strings.NewReader("pwned"))
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("POST /patterns/%%2e%%2e: got %d, want 400", w.Code)
	}
	if _, err := os.Stat(filepath.Join(root, "system.md")); err == nil {
		t.Fatal("wrote system.md in the parent directory")
	}
}

func TestChatHandler_RejectsFilePathPatternName(t *testing.T) {
	if _, err := i18n.Init("en"); err != nil {
		t.Fatalf("i18n.Init() error = %v", err)
	}
	gin.SetMode(gin.TestMode)

	// The zero-value handler is a deliberate seam: a request that passes
	// pre-validation nil-panics inside HandleChat, so these tests cannot
	// false-pass on a request that was never validated.
	r := gin.New()
	r.POST("/chat", (&ChatHandler{}).HandleChat)

	w := httptest.NewRecorder()
	body := `{"prompts":[{"patternName":"/etc/hosts","userInput":"x"}]}`
	req := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want 400", w.Code)
	}
}

func TestChatHandler_RejectsUnsafeNames(t *testing.T) {
	if _, err := i18n.Init("en"); err != nil {
		t.Fatalf("i18n.Init() error = %v", err)
	}
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.POST("/chat", (&ChatHandler{}).HandleChat)

	cases := map[string]string{
		// Every prefix class loadPattern treats as a filesystem path
		"absolute pattern name":  `{"prompts":[{"patternName":"/etc/hosts","userInput":"x"}]}`,
		"home pattern name":      `{"prompts":[{"patternName":"~/secret.md","userInput":"x"}]}`,
		"relative pattern name":  `{"prompts":[{"patternName":"./secret.md","userInput":"x"}]}`,
		"backslash pattern name": `{"prompts":[{"patternName":"\\secret.md","userInput":"x"}]}`,
		// Context and session names get the same pre-validation
		"traversal context name": `{"prompts":[{"userInput":"x","contextName":"../keep.txt"}]}`,
		"traversal session name": `{"prompts":[{"userInput":"x","sessionName":".."}]}`,
		// The rejection loop bails on the first bad name, however deep it sits
		"second prompt path-like": `{"prompts":[{"userInput":"x"},{"patternName":"/etc/hosts","userInput":"y"}]}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("got status %d, want 400", w.Code)
			}
		})
	}
}

func TestChatHandler_LogsRejectedUnsafeNames(t *testing.T) {
	if _, err := i18n.Init("en"); err != nil {
		t.Fatalf("i18n.Init() error = %v", err)
	}
	gin.SetMode(gin.TestMode)

	t.Run("pattern name", func(t *testing.T) {
		buf, restore := captureServerLogs()
		defer restore()

		r := gin.New()
		r.POST("/chat", (&ChatHandler{}).HandleChat)

		w := httptest.NewRecorder()
		body := `{"prompts":[{"patternName":"/etc/hosts","userInput":"x"}]}`
		req := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("got status %d, want 400", w.Code)
		}
		if !strings.Contains(buf.String(), "Rejected invalid pattern name") {
			t.Fatalf("missing pattern-name security log entry: %q", buf.String())
		}
	})

	t.Run("context name", func(t *testing.T) {
		buf, restore := captureServerLogs()
		defer restore()

		r := gin.New()
		r.POST("/chat", (&ChatHandler{}).HandleChat)

		w := httptest.NewRecorder()
		body := `{"prompts":[{"userInput":"x","contextName":"../keep.txt"}]}`
		req := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("got status %d, want 400", w.Code)
		}
		if !strings.Contains(buf.String(), "Rejected invalid storage name") {
			t.Fatalf("missing storage-name security log entry: %q", buf.String())
		}
	})
}

func TestPatternsHandler_RejectsUnsafeNamesOnReadRoutes(t *testing.T) {
	if _, err := i18n.Init("en"); err != nil {
		t.Fatalf("i18n.Init() error = %v", err)
	}
	gin.SetMode(gin.TestMode)

	r := gin.New()
	NewPatternsHandler(r, &fsdb.PatternsEntity{
		StorageEntity:     &fsdb.StorageEntity{Label: "Patterns", Dir: t.TempDir(), ItemIsDir: true},
		SystemPatternFile: "system.md",
	})

	for _, req := range []*http.Request{
		httptest.NewRequest(http.MethodGet, "/patterns/%2e%2e", nil),
		httptest.NewRequest(http.MethodPost, "/patterns/%2e%2e/apply", strings.NewReader(`{"input":"x"}`)),
		// Names that fail only ValidateStorageName, not the file-path check
		httptest.NewRequest(http.MethodGet, "/patterns/foo:bar", nil),
		httptest.NewRequest(http.MethodGet, "/patterns/NUL", nil),
		httptest.NewRequest(http.MethodPost, "/patterns/foo:bar/apply", strings.NewReader(`{"input":"x"}`)),
	} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("%s %s: got %d, want 400", req.Method, req.URL.Path, w.Code)
		}
	}
}

// A benign request must clear pre-validation. With the zero-value handler
// seam, clearing validation nil-panics, which Recovery turns into a 500;
// a 400 would mean validation rejected valid names.
func TestChatHandler_AcceptsBenignNames(t *testing.T) {
	if _, err := i18n.Init("en"); err != nil {
		t.Fatalf("i18n.Init() error = %v", err)
	}
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(gin.Recovery())
	r.POST("/chat", (&ChatHandler{}).HandleChat)

	w := httptest.NewRecorder()
	body := `{"prompts":[{"userInput":"x","patternName":"summarize","contextName":"myctx","sessionName":"mysession"}]}`
	req := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code == http.StatusBadRequest {
		t.Fatalf("benign request was rejected with 400: %s", w.Body.String())
	}
}

func TestAPIKeyMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(APIKeyMiddleware("secret"))
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	t.Run("missing key", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("got %d, want 401", w.Code)
		}
	})

	t.Run("wrong key", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		req.Header.Set(APIKeyHeader, "wrong")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("got %d, want 401", w.Code)
		}
	})

	t.Run("valid key", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		req.Header.Set(APIKeyHeader, "secret")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("got %d, want 200", w.Code)
		}
	})
}

func TestAPIKeyMiddleware_LogsUnauthorizedAttempts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	buf, restore := captureServerLogs()
	defer restore()

	r := gin.New()
	r.Use(APIKeyMiddleware("secret"))
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set(APIKeyHeader, "wrong")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("got %d, want 401", w.Code)
	}

	logOutput := buf.String()
	if !strings.Contains(logOutput, "API key authentication failed") {
		t.Fatalf("missing auth failure log entry: %q", logOutput)
	}
	if !strings.Contains(logOutput, "invalid_api_key") {
		t.Fatalf("missing auth failure reason: %q", logOutput)
	}
	if strings.Contains(logOutput, "wrong") || strings.Contains(logOutput, "secret") {
		t.Fatalf("log leaked API key material: %q", logOutput)
	}
}
