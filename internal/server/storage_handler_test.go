package restapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielmiessler/fabric/internal/plugins/db/fsdb"
	"github.com/gin-gonic/gin"
)

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
