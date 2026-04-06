package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestEnsureEnvFileCreatesSecureEnvFile(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", homeDir)
	}

	if err := ensureEnvFile(); err != nil {
		t.Fatalf("ensureEnvFile() error = %v", err)
	}

	envPath := filepath.Join(homeDir, ".config", "fabric", ".env")
	info, err := os.Stat(envPath)
	if err != nil {
		t.Fatalf("failed to stat %s: %v", envPath, err)
	}

	if runtime.GOOS != "windows" && info.Mode().Perm() != EnvFilePerms {
		t.Fatalf("permissions for %s = %o, want %o", envPath, info.Mode().Perm(), EnvFilePerms)
	}
}
