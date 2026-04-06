package restapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/danielmiessler/fabric/internal/chat"
	"github.com/danielmiessler/fabric/internal/core"
	"github.com/danielmiessler/fabric/internal/domain"
	debuglog "github.com/danielmiessler/fabric/internal/log"
	"github.com/danielmiessler/fabric/internal/plugins"
	"github.com/danielmiessler/fabric/internal/plugins/ai"
	"github.com/danielmiessler/fabric/internal/plugins/db/fsdb"
	"github.com/danielmiessler/fabric/internal/tools"
	"github.com/gin-gonic/gin"
)

func TestBuildPromptChatRequest_PreservesStrategyAndUserInput(t *testing.T) {
	prompt := PromptRequest{
		UserInput:    "user input",
		Vendor:       "TestVendor",
		Model:        "test-model",
		ContextName:  "ctx",
		PatternName:  "pattern",
		StrategyName: "strategy",
		SessionName:  "session",
		Variables: map[string]string{
			"topic": "pipelines",
		},
	}

	request := buildPromptChatRequest(prompt, "en")

	if request.Message == nil {
		t.Fatal("expected request message to be set")
	}
	if request.Message.Content != "user input" {
		t.Fatalf("expected user input to stay unchanged, got %q", request.Message.Content)
	}
	if request.StrategyName != "strategy" {
		t.Fatalf("expected strategy name to be preserved, got %q", request.StrategyName)
	}
	if request.PatternName != "pattern" {
		t.Fatalf("expected pattern name to be preserved, got %q", request.PatternName)
	}
	if request.ContextName != "ctx" {
		t.Fatalf("expected context name to be preserved, got %q", request.ContextName)
	}
	if request.SessionName != "session" {
		t.Fatalf("expected session name to be preserved, got %q", request.SessionName)
	}
	if request.Language != "en" {
		t.Fatalf("expected language to be preserved, got %q", request.Language)
	}
	if got := request.PatternVariables["topic"]; got != "pipelines" {
		t.Fatalf("expected variables to be preserved, got %q", got)
	}
	if !request.RestrictTemplateFeatures {
		t.Fatal("expected server chat requests to restrict unsafe template features")
	}
}

func TestWriteSSEResponse_FormatsEventStreamChunk(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	response := StreamResponse{
		Type:    "content",
		Format:  "markdown",
		Content: "hello",
	}

	if err := writeSSEResponse(ctx.Writer, response); err != nil {
		t.Fatalf("expected SSE write to succeed, got %v", err)
	}

	payload := recorder.Body.String()
	if !strings.HasPrefix(payload, "data: ") {
		t.Fatalf("expected SSE payload prefix, got %q", payload)
	}

	var got StreamResponse
	raw := strings.TrimSuffix(strings.TrimPrefix(payload, "data: "), "\n\n")
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatalf("expected valid JSON payload, got %v", err)
	}
	if got != response {
		t.Fatalf("expected %#v, got %#v", response, got)
	}
}

func TestWriteSSEResponse_EscapesEmbeddedEventContent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	response := StreamResponse{
		Type:    "content",
		Format:  "markdown",
		Content: "line one\n\ndata: forged-event",
	}

	if err := writeSSEResponse(ctx.Writer, response); err != nil {
		t.Fatalf("expected SSE write to succeed, got %v", err)
	}

	payload := recorder.Body.String()
	if strings.Count(payload, "\n\n") != 1 {
		t.Fatalf("expected a single SSE frame terminator, got %q", payload)
	}
	if strings.Contains(payload, "\n\ndata: forged-event\n\n") {
		t.Fatalf("expected embedded content to stay JSON-escaped, got %q", payload)
	}
}

func TestHandleChat_SetsEventStreamHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest("POST", "/chat", strings.NewReader(`{"prompts":[],"language":"en"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler := &ChatHandler{}
	handler.HandleChat(ctx)

	if got := recorder.Header().Get("Content-Type"); got != "text/event-stream" {
		t.Fatalf("expected SSE content type, got %q", got)
	}
	if got := recorder.Header().Get("Cache-Control"); got != chatSSECacheControl {
		t.Fatalf("expected cache-control header %q, got %q", chatSSECacheControl, got)
	}
	if got := recorder.Header().Get("Connection"); got != "keep-alive" {
		t.Fatalf("expected keep-alive header, got %q", got)
	}
	if got := recorder.Header().Get("Pragma"); got != "no-cache" {
		t.Fatalf("expected pragma header %q, got %q", "no-cache", got)
	}
	if got := recorder.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("expected nosniff header, got %q", got)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected no CORS header without a matching origin, got %q", got)
	}
}

func TestHandleChat_AllowsLocalDevCORSOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(`{"prompts":[],"language":"en"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Request.Header.Set("Origin", localDevCORSOrigin)

	handler := &ChatHandler{}
	handler.HandleChat(ctx)

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != localDevCORSOrigin {
		t.Fatalf("expected local dev origin to be allowed, got %q", got)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(got, http.MethodPost) {
		t.Fatalf("expected allow methods to include POST, got %q", got)
	}
}

func TestHandleChat_DoesNotReflectUnexpectedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(`{"prompts":[],"language":"en"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Request.Header.Set("Origin", "https://example.com")

	handler := &ChatHandler{}
	handler.HandleChat(ctx)

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected unexpected origin to receive no CORS header, got %q", got)
	}
}

func TestHandleChat_RejectsUnknownJSONFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(`{"prompts":[],"language":"en","unexpected":true}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler := &ChatHandler{}
	handler.HandleChat(ctx)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
	if body := recorder.Body.String(); !strings.Contains(body, "unknown field") {
		t.Fatalf("expected unknown field error, got %q", body)
	}
}

func TestHandleChatOptions_AllowsLocalDevPreflight(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	NewChatHandler(router, nil, nil)

	req := httptest.NewRequest(http.MethodOptions, "/chat", nil)
	req.Header.Set("Origin", localDevCORSOrigin)
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != localDevCORSOrigin {
		t.Fatalf("expected local dev origin to be allowed, got %q", got)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Headers"); !strings.Contains(got, "Content-Type") {
		t.Fatalf("expected allow headers to include Content-Type, got %q", got)
	}
}

func TestHandleChat_RequestLoggingRespectsDebugLevel(t *testing.T) {
	gin.SetMode(gin.TestMode)

	requestBody := `{"prompts":[],"language":"en"}`

	t.Run("debug off", func(t *testing.T) {
		oldLevel := debuglog.GetLevel()
		defer debuglog.SetLevel(oldLevel)
		debuglog.SetLevel(debuglog.Off)

		var buf bytes.Buffer
		debuglog.SetOutput(&buf)
		defer debuglog.SetOutput(os.Stderr)

		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest("POST", "/chat", strings.NewReader(requestBody))
		ctx.Request.Header.Set("Content-Type", "application/json")

		handler := &ChatHandler{}
		handler.HandleChat(ctx)

		if got := buf.String(); got != "" {
			t.Fatalf("expected no debug request logging with --debug=0, got %q", got)
		}
	})

	t.Run("debug basic", func(t *testing.T) {
		oldLevel := debuglog.GetLevel()
		defer debuglog.SetLevel(oldLevel)
		debuglog.SetLevel(debuglog.Basic)

		var buf bytes.Buffer
		debuglog.SetOutput(&buf)
		defer debuglog.SetOutput(os.Stderr)

		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest("POST", "/chat", strings.NewReader(requestBody))
		ctx.Request.Header.Set("Content-Type", "application/json")

		handler := &ChatHandler{}
		handler.HandleChat(ctx)

		if got := buf.String(); !strings.Contains(got, `Received chat request - language="en" prompts=0`) {
			t.Fatalf("expected debug request logging at basic level, got %q", got)
		}
	})
}

func TestHandleChat_RedactsBackendErrorsFromSSEStream(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &ChatHandler{
		registry: newStreamingErrorRegistry(t, errors.New("backend failed with sk-test-secret from /tmp/fabric-secret")),
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(
		http.MethodPost,
		"/chat",
		strings.NewReader(`{"prompts":[{"userInput":"hello"}],"language":"en"}`),
	)
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.HandleChat(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, `"type":"error"`) {
		t.Fatalf("expected error event in SSE stream, got %q", body)
	}
	if strings.Contains(body, "sk-test-secret") || strings.Contains(body, "/tmp/fabric-secret") {
		t.Fatalf("expected SSE stream to redact backend details, got %q", body)
	}
	if !strings.Contains(body, clientSafeChatErrorMessage) {
		t.Fatalf("expected generic chat error message, got %q", body)
	}
	if strings.Contains(body, `"type":"complete"`) {
		t.Fatalf("expected errored stream to stop without a completion event, got %q", body)
	}
}

type streamingErrorVendor struct {
	err error
}

func (v *streamingErrorVendor) GetName() string                       { return "TestVendor" }
func (v *streamingErrorVendor) GetSetupDescription() string           { return "TestVendor" }
func (v *streamingErrorVendor) IsConfigured() bool                    { return true }
func (v *streamingErrorVendor) Configure() error                      { return nil }
func (v *streamingErrorVendor) Setup() error                          { return nil }
func (v *streamingErrorVendor) SetupFillEnvFileContent(*bytes.Buffer) {}
func (v *streamingErrorVendor) ListModels(context.Context) ([]string, error) {
	return []string{"test-model"}, nil
}
func (v *streamingErrorVendor) SendStream(_ context.Context, _ []*chat.ChatCompletionMessage, _ *domain.ChatOptions, updates chan domain.StreamUpdate) error {
	close(updates)
	return v.err
}
func (v *streamingErrorVendor) Send(context.Context, []*chat.ChatCompletionMessage, *domain.ChatOptions) (string, error) {
	return "", v.err
}
func (v *streamingErrorVendor) NeedsRawMode(string) bool { return false }

func newStreamingErrorRegistry(t *testing.T, sendErr error) *core.PluginRegistry {
	t.Helper()

	db := fsdb.NewDb(t.TempDir())
	vendorManager := ai.NewVendorsManager()
	vendorManager.AddVendors(&streamingErrorVendor{err: sendErr})

	return &core.PluginRegistry{
		Db:            db,
		VendorManager: vendorManager,
		Defaults: &tools.Defaults{
			PluginBase:         &plugins.PluginBase{},
			Vendor:             &plugins.Setting{Value: "TestVendor"},
			Model:              &plugins.SetupQuestion{Setting: &plugins.Setting{Value: "test-model"}},
			ModelContextLength: &plugins.SetupQuestion{Setting: &plugins.Setting{Value: "0"}},
		},
	}
}
