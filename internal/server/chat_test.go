package restapi

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

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
	if got := recorder.Header().Get("Cache-Control"); got != "no-cache" {
		t.Fatalf("expected no-cache header, got %q", got)
	}
	if got := recorder.Header().Get("Connection"); got != "keep-alive" {
		t.Fatalf("expected keep-alive header, got %q", got)
	}
}
