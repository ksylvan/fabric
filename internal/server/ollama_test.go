package restapi

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBuildFabricChatURL(t *testing.T) {
	tests := []struct {
		name    string
		addr    string
		want    string
		wantErr bool
	}{
		{
			name:    "empty address",
			addr:    "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "valid http URL",
			addr:    "http://localhost:8080",
			want:    "http://localhost:8080",
			wantErr: false,
		},
		{
			name:    "valid https URL",
			addr:    "https://api.example.com",
			want:    "https://api.example.com",
			wantErr: false,
		},
		{
			name:    "http URL with trailing slash",
			addr:    "http://localhost:8080/",
			want:    "http://localhost:8080",
			wantErr: false,
		},
		{
			name:    "malformed URL - missing host",
			addr:    "http://",
			want:    "",
			wantErr: true,
		},
		{
			name:    "malformed URL - port only with http",
			addr:    "https://:8080",
			want:    "",
			wantErr: true,
		},
		{
			name:    "colon-prefixed port",
			addr:    ":8080",
			want:    "http://127.0.0.1:8080",
			wantErr: false,
		},
		{
			name:    "bare host:port",
			addr:    "localhost:8080",
			want:    "http://localhost:8080",
			wantErr: false,
		},
		{
			name:    "bare hostname",
			addr:    "localhost",
			want:    "http://localhost",
			wantErr: false,
		},
		{
			name:    "IP address with port",
			addr:    "192.168.1.1:3000",
			want:    "http://192.168.1.1:3000",
			wantErr: false,
		},
		{
			name:    "bare address with path - invalid",
			addr:    "localhost:8080/some/path",
			want:    "",
			wantErr: true,
		},
		{
			name:    "bare hostname with path - invalid",
			addr:    "localhost/api",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildFabricChatURL(tt.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildFabricChatURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("buildFabricChatURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseOllamaNumCtx(t *testing.T) {
	tests := []struct {
		name    string
		options map[string]any
		want    int
		wantErr bool
		errMsg  string
	}{
		// --- Valid inputs ---
		{
			name:    "nil options",
			options: nil,
			want:    0,
			wantErr: false,
		},
		{
			name:    "empty options",
			options: map[string]any{},
			want:    0,
			wantErr: false,
		},
		{
			name:    "num_ctx not present",
			options: map[string]any{"other_key": 123},
			want:    0,
			wantErr: false,
		},
		{
			name:    "num_ctx is null",
			options: map[string]any{"num_ctx": nil},
			want:    0,
			wantErr: false,
		},
		{
			name:    "valid int",
			options: map[string]any{"num_ctx": 4096},
			want:    4096,
			wantErr: false,
		},
		{
			name:    "valid float64 (whole number)",
			options: map[string]any{"num_ctx": float64(8192)},
			want:    8192,
			wantErr: false,
		},
		{
			name:    "valid float32 (whole number)",
			options: map[string]any{"num_ctx": float32(2048)},
			want:    2048,
			wantErr: false,
		},
		{
			name:    "valid json.Number",
			options: map[string]any{"num_ctx": json.Number("16384")},
			want:    16384,
			wantErr: false,
		},
		{
			name:    "valid string",
			options: map[string]any{"num_ctx": "32768"},
			want:    32768,
			wantErr: false,
		},
		{
			name:    "valid int64",
			options: map[string]any{"num_ctx": int64(65536)},
			want:    65536,
			wantErr: false,
		},
		// --- Invalid inputs ---
		{
			name:    "float64 with fractional part",
			options: map[string]any{"num_ctx": 4096.5},
			want:    0,
			wantErr: true,
			errMsg:  "num_ctx must be an integer, got float with fractional part",
		},
		{
			name:    "float32 with fractional part",
			options: map[string]any{"num_ctx": float32(2048.75)},
			want:    0,
			wantErr: true,
			errMsg:  "num_ctx must be an integer, got float with fractional part",
		},
		{
			name:    "negative int",
			options: map[string]any{"num_ctx": -100},
			want:    0,
			wantErr: true,
			errMsg:  "num_ctx must be positive",
		},
		{
			name:    "zero int",
			options: map[string]any{"num_ctx": 0},
			want:    0,
			wantErr: true,
			errMsg:  "num_ctx must be positive",
		},
		{
			name:    "negative float64",
			options: map[string]any{"num_ctx": float64(-500)},
			want:    0,
			wantErr: true,
			errMsg:  "num_ctx must be positive",
		},
		{
			name:    "negative float32",
			options: map[string]any{"num_ctx": float32(-250)},
			want:    0,
			wantErr: true,
			errMsg:  "num_ctx must be positive",
		},
		{
			name:    "non-numeric string",
			options: map[string]any{"num_ctx": "not-a-number"},
			want:    0,
			wantErr: true,
			errMsg:  "num_ctx must be a valid number",
		},
		{
			name:    "invalid json.Number",
			options: map[string]any{"num_ctx": json.Number("invalid")},
			want:    0,
			wantErr: true,
			errMsg:  "num_ctx must be a valid number",
		},
		{
			name:    "exceeds maximum allowed value",
			options: map[string]any{"num_ctx": 2000000},
			want:    0,
			wantErr: true,
			errMsg:  "num_ctx exceeds maximum allowed value",
		},
		{
			name:    "unsupported type (bool)",
			options: map[string]any{"num_ctx": true},
			want:    0,
			wantErr: true,
			errMsg:  "num_ctx must be a number, got invalid type",
		},
		{
			name:    "unsupported type (slice)",
			options: map[string]any{"num_ctx": []int{1, 2, 3}},
			want:    0,
			wantErr: true,
			errMsg:  "num_ctx must be a number, got invalid type",
		},
		// --- Edge cases ---
		{
			name:    "minimum valid value",
			options: map[string]any{"num_ctx": 1},
			want:    1,
			wantErr: false,
		},
		{
			name:    "maximum allowed value",
			options: map[string]any{"num_ctx": 1000000},
			want:    1000000,
			wantErr: false,
		},
		{
			name:    "very large float64 (overflow)",
			options: map[string]any{"num_ctx": float64(math.MaxFloat64)},
			want:    0,
			wantErr: true,
			errMsg:  "num_ctx value out of range",
		},
		{
			name:    "large int64 exceeding maxInt on 32-bit",
			options: map[string]any{"num_ctx": int64(1 << 40)},
			want:    0,
			wantErr: true,
			errMsg:  "num_ctx", // either "too large" or "exceeds maximum"
		},
		{
			name:    "long string gets truncated in error",
			options: map[string]any{"num_ctx": "this-is-a-very-long-string-that-should-be-truncated-in-the-error-message"},
			want:    0,
			wantErr: true,
			errMsg:  "num_ctx must be a valid number",
		},
		// --- Special float values ---
		{
			name:    "float64 NaN",
			options: map[string]any{"num_ctx": math.NaN()},
			want:    0,
			wantErr: true,
			errMsg:  "num_ctx must be a finite number",
		},
		{
			name:    "float64 positive infinity",
			options: map[string]any{"num_ctx": math.Inf(1)},
			want:    0,
			wantErr: true,
			errMsg:  "num_ctx must be a finite number",
		},
		{
			name:    "float64 negative infinity",
			options: map[string]any{"num_ctx": math.Inf(-1)},
			want:    0,
			wantErr: true,
			errMsg:  "num_ctx must be a finite number",
		},
		{
			name:    "float32 NaN",
			options: map[string]any{"num_ctx": float32(math.NaN())},
			want:    0,
			wantErr: true,
			errMsg:  "num_ctx must be a finite number",
		},
		{
			name:    "float32 positive infinity",
			options: map[string]any{"num_ctx": float32(math.Inf(1))},
			want:    0,
			wantErr: true,
			errMsg:  "num_ctx must be a finite number",
		},
		{
			name:    "float32 negative infinity",
			options: map[string]any{"num_ctx": float32(math.Inf(-1))},
			want:    0,
			wantErr: true,
			errMsg:  "num_ctx must be a finite number",
		},
		// --- Negative int64 (32-bit wraparound prevention) ---
		{
			name:    "negative int64",
			options: map[string]any{"num_ctx": int64(-1000)},
			want:    0,
			wantErr: true,
			errMsg:  "num_ctx must be positive",
		},
		{
			name:    "negative json.Number",
			options: map[string]any{"num_ctx": json.Number("-500")},
			want:    0,
			wantErr: true,
			errMsg:  "num_ctx must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseOllamaNumCtx(tt.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseOllamaNumCtx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("parseOllamaNumCtx() error message = %q, want to contain %q", err.Error(), tt.errMsg)
				}
			}
			if got != tt.want {
				t.Errorf("parseOllamaNumCtx() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOllamaChatRejectsUnknownJSONFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/chat", strings.NewReader(`{"model":"pattern:latest","messages":[],"stream":false,"unexpected":true}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	APIConvert{}.ollamaChat(ctx)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
	if body := recorder.Body.String(); !strings.Contains(body, "unknown field") {
		t.Fatalf("expected unknown field error, got %q", body)
	}
}
