package fsdb

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
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

func TestStorage_RejectsTraversalNames(t *testing.T) {
	dir := t.TempDir()
	storage := &StorageEntity{Dir: dir}

	if err := storage.Save("../escape", []byte("secret")); !errors.Is(err, ErrInvalidStorageName) {
		t.Fatalf("expected invalid storage name error from Save, got %v", err)
	}

	if _, err := storage.Load("../escape"); !errors.Is(err, ErrInvalidStorageName) {
		t.Fatalf("expected invalid storage name error from Load, got %v", err)
	}

	if err := storage.Save("safe", []byte("content")); err != nil {
		t.Fatalf("failed to seed safe file: %v", err)
	}

	if err := storage.Rename("safe", "../escape"); !errors.Is(err, ErrInvalidStorageName) {
		t.Fatalf("expected invalid storage name error from Rename, got %v", err)
	}

	if err := storage.Delete("../escape"); !errors.Is(err, ErrInvalidStorageName) {
		t.Fatalf("expected invalid storage name error from Delete, got %v", err)
	}

	if storage.Exists("../escape") {
		t.Fatal("expected Exists to report false for invalid storage names")
	}

	if _, err := os.Stat(filepath.Join(dir, "..", "escape")); !os.IsNotExist(err) {
		t.Fatalf("expected traversal target to remain absent, stat err=%v", err)
	}
}
