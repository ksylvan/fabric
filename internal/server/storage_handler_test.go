package restapi

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danielmiessler/fabric/internal/plugins/db/fsdb"
	"github.com/gin-gonic/gin"
)

type repeatedByteReader byte

func (r repeatedByteReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = byte(r)
	}
	return len(p), nil
}

func TestStorageHandler_GetRejectsInvalidName(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	contexts := &fsdb.ContextsEntity{
		StorageEntity: &fsdb.StorageEntity{Dir: t.TempDir()},
	}
	NewStorageHandler(router, "contexts", contexts)

	req := httptest.NewRequest(http.MethodGet, "/contexts/..", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestStorageHandler_SaveRejectsOversizedRequestBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	contexts := &fsdb.ContextsEntity{
		StorageEntity: &fsdb.StorageEntity{Dir: t.TempDir()},
	}
	NewStorageHandler(router, "contexts", contexts)

	req := httptest.NewRequest(
		http.MethodPost,
		"/contexts/example",
		io.LimitReader(repeatedByteReader('a'), int64(maxRequestBodyBytes)+1),
	)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status %d, got %d", http.StatusRequestEntityTooLarge, recorder.Code)
	}
	if body := recorder.Body.String(); !strings.Contains(body, errRequestBodyTooLarge.Error()) {
		t.Fatalf("expected oversized body error, got %q", body)
	}
}
