package fsdb

import (
	"os"
	"path/filepath"
	"testing"
)

func TestContexts_GetContext(t *testing.T) {
	contexts := setupTestContexts(t)
	contextName := "testContext"
	contextPath := filepath.Join(contexts.Dir, contextName)
	contextContent := "test content"
	err := os.WriteFile(contextPath, []byte(contextContent), 0644)
	if err != nil {
		t.Fatalf("failed to write context file: %v", err)
	}
	context, err := contexts.Get(contextName)
	if err != nil {
		t.Fatalf("failed to get context: %v", err)
	}
	expectedContext := &Context{Name: contextName, Content: contextContent}
	if *context != *expectedContext {
		t.Errorf("expected %v, got %v", expectedContext, context)
	}
}
