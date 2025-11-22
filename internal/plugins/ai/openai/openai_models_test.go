package openai

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Ensures we can list models when a provider returns a direct array of models
// instead of the standard OpenAI list response structure.
func TestListModels_FallbackToDirectFetch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/models", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"id":"github-model"}]`))
	}))
	defer srv.Close()

	client := NewClient()
	client.ApiKey.Value = "test-key"
	client.ApiBaseURL.Value = srv.URL

	err := client.configure()
	assert.NoError(t, err)

	models, err := client.ListModels()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(models))
	assert.Equal(t, "github-model", models[0])
}
