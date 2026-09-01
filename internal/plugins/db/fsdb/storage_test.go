package fsdb

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/danielmiessler/fabric/internal/i18n"
)

func TestStorage_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	storage := &StorageEntity{Dir: dir}
	name := "test"
	content := []byte("test content")
	if err := storage.Save(name, content); err != nil {
		t.Fatalf("failed to save content: %v", err)
	}
	loadedContent, err := storage.Load(name)
	if err != nil {
		t.Fatalf("failed to load content: %v", err)
	}
	if string(loadedContent) != string(content) {
		t.Errorf("expected %v, got %v", string(content), string(loadedContent))
	}
}

func TestStorage_Exists(t *testing.T) {
	dir := t.TempDir()
	storage := &StorageEntity{Dir: dir}
	name := "test"
	if storage.Exists(name) {
		t.Errorf("expected file to not exist")
	}
	if err := storage.Save(name, []byte("test content")); err != nil {
		t.Fatalf("failed to save content: %v", err)
	}
	if !storage.Exists(name) {
		t.Errorf("expected file to exist")
	}
}

func TestStorage_Delete(t *testing.T) {
	dir := t.TempDir()
	storage := &StorageEntity{Dir: dir}
	name := "test"
	if err := storage.Save(name, []byte("test content")); err != nil {
		t.Fatalf("failed to save content: %v", err)
	}
	if err := storage.Delete(name); err != nil {
		t.Fatalf("failed to delete content: %v", err)
	}
	if storage.Exists(name) {
		t.Errorf("expected file to be deleted")
	}
}

// invalidStorageNames are names ValidateStorageName must reject on every
// platform. Shared by the storage and pattern traversal tests so a new
// attack name is pinned everywhere at once. The backslash cases pin the
// `\` half of the separator check, which is the Windows-only escape guard
// a "simplify to filepath.Base" refactor would silently drop. The colon,
// reserved-name, and trailing dot/space cases pin the Windows edge cases
// (NTFS alternate data streams, DOS device names, silently stripped name
// suffixes).
var invalidStorageNames = []string{
	"..", "../keep.txt", "/etc/passwd", "foo/../../keep.txt", ".", "",
	`foo\bar`, `..\x`,
	"foo:bar", "NUL", "con.txt", "CON.tar.gz", "foo.", "foo ",
}

// newTraversalFixture returns a storage entity confined under a temp root
// with a marker file outside the entity dir, plus a check that fails the
// test if the marker or the entity dir is gone.
func newTraversalFixture(t *testing.T) (storage *StorageEntity, checkSurvived func()) {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "contexts")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(root, "keep.txt")
	if err := os.WriteFile(marker, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	storage = &StorageEntity{Dir: dir, Label: "Contexts"}
	checkSurvived = func() {
		t.Helper()
		if _, err := os.Stat(marker); err != nil {
			t.Fatalf("parent marker was removed: %v", err)
		}
		if _, err := os.Stat(dir); err != nil {
			t.Fatalf("storage dir was removed: %v", err)
		}
	}
	return
}

func TestStorage_RejectsPathTraversal(t *testing.T) {
	if _, err := i18n.Init("en"); err != nil {
		t.Fatalf("i18n.Init() error = %v", err)
	}

	storage, checkSurvived := newTraversalFixture(t)
	for _, name := range invalidStorageNames {
		t.Run(name, func(t *testing.T) {
			if err := storage.Delete(name); err == nil {
				t.Fatalf("Delete(%q) succeeded, want error", name)
			}
			if err := storage.Save(name, []byte("pwned")); err == nil {
				t.Fatalf("Save(%q) succeeded, want error", name)
			}
			if _, err := storage.Load(name); err == nil {
				t.Fatalf("Load(%q) succeeded, want error", name)
			}
			if storage.Exists(name) {
				t.Fatalf("Exists(%q) is true, want false", name)
			}
		})
	}

	checkSurvived()
}

func TestInvalidStorageNameError_DefaultMessage(t *testing.T) {
	if _, err := i18n.Init("en"); err != nil {
		t.Fatalf("i18n.Init() error = %v", err)
	}
	err := &InvalidStorageNameError{Name: "bad:name"}
	if got, want := err.Error(), `invalid name: "bad:name"`; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}

func TestStorage_Rename(t *testing.T) {
	dir := t.TempDir()
	storage := &StorageEntity{Dir: dir, Label: "Contexts"}
	if err := storage.Save("old", []byte("content")); err != nil {
		t.Fatalf("failed to save content: %v", err)
	}
	if err := storage.Rename("old", "new"); err != nil {
		t.Fatalf("failed to rename: %v", err)
	}
	if storage.Exists("old") {
		t.Errorf("expected old name to be gone")
	}
	loaded, err := storage.Load("new")
	if err != nil {
		t.Fatalf("failed to load renamed content: %v", err)
	}
	if string(loaded) != "content" {
		t.Errorf("expected %q, got %q", "content", string(loaded))
	}
}

func TestStorage_RenameRejectsPathTraversal(t *testing.T) {
	if _, err := i18n.Init("en"); err != nil {
		t.Fatalf("i18n.Init() error = %v", err)
	}

	storage, checkSurvived := newTraversalFixture(t)
	if err := storage.Save("ok", []byte("content")); err != nil {
		t.Fatalf("failed to save content: %v", err)
	}

	for _, name := range invalidStorageNames {
		t.Run(name, func(t *testing.T) {
			if err := storage.Rename("ok", name); err == nil {
				t.Fatalf("Rename(%q, %q) succeeded, want error", "ok", name)
			}
			if err := storage.Rename(name, "ok"); err == nil {
				t.Fatalf("Rename(%q, %q) succeeded, want error", name, "ok")
			}
		})
	}

	checkSurvived()
	if !storage.Exists("ok") {
		t.Fatalf("legitimate entry was moved or deleted")
	}
}
