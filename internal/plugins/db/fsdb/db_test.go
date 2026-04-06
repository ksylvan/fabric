package fsdb

import (
	"os"
	"runtime"
	"testing"
)

func TestDb_Configure(t *testing.T) {
	dir := t.TempDir()
	db := NewDb(dir)
	err := db.Configure()
	if err == nil {
		t.Fatalf("db is configured, but must not be at empty dir: %v", dir)
	}
	if db.IsEnvFileExists() {
		t.Fatalf("db file exists, but must not be at empty dir: %v", dir)
	}

	err = db.SaveEnv("")
	if err != nil {
		t.Fatalf("db can't save env for empty conf.: %v", err)
	}

	err = db.Configure()
	if err != nil {
		t.Fatalf("db is not configured, but shall be after save: %v", err)
	}
}

func TestDb_LoadEnvFile(t *testing.T) {
	dir := t.TempDir()
	db := NewDb(dir)
	content := "KEY=VALUE\n"
	err := os.WriteFile(db.EnvFilePath, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("failed to write .env file: %v", err)
	}
	err = db.LoadEnvFile()
	if err != nil {
		t.Errorf("failed to load .env file: %v", err)
	}
	assertSecureEnvFilePerms(t, db.EnvFilePath)
}

func TestDb_SaveEnv(t *testing.T) {
	dir := t.TempDir()
	db := NewDb(dir)
	content := "KEY=VALUE\n"
	err := db.SaveEnv(content)
	if err != nil {
		t.Errorf("failed to save .env file: %v", err)
	}
	if _, err := os.Stat(db.EnvFilePath); os.IsNotExist(err) {
		t.Errorf("expected .env file to be saved")
	}
	assertSecureEnvFilePerms(t, db.EnvFilePath)
}

func assertSecureEnvFilePerms(t *testing.T, path string) {
	t.Helper()

	if runtime.GOOS == "windows" {
		return
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat %s: %v", path, err)
	}
	if got := info.Mode().Perm(); got != EnvFilePerms {
		t.Fatalf("permissions for %s = %o, want %o", path, got, EnvFilePerms)
	}
}
