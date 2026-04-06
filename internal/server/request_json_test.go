package restapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDecodeStrictJSONRejectsMultipleJSONValues(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"first"}{"name":"second"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	var payload struct {
		Name string `json:"name"`
	}
	err := decodeStrictJSON(ctx, &payload)
	if !errors.Is(err, errMultipleJSONValues) {
		t.Fatalf("expected %v, got %v", errMultipleJSONValues, err)
	}
}

func TestDecodeStrictJSONUsesJSONNumberForUntypedNumbers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"options":{"num_ctx":9007199254740993}}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	var payload struct {
		Options map[string]any `json:"options"`
	}
	if err := decodeStrictJSON(ctx, &payload); err != nil {
		t.Fatalf("expected decode to succeed, got %v", err)
	}

	number, ok := payload.Options["num_ctx"].(json.Number)
	if !ok {
		t.Fatalf("expected num_ctx to decode as json.Number, got %T", payload.Options["num_ctx"])
	}
	if number.String() != "9007199254740993" {
		t.Fatalf("expected original numeric string to be preserved, got %q", number.String())
	}
}
