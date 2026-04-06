package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/danielmiessler/fabric/internal/plugins/db/fsdb"
)

func TestEnsureEnvFileSecuresConfigPath(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", homeDir)
	}

	configDir := filepath.Join(homeDir, ".config", "fabric")
	envPath := filepath.Join(configDir, ".env")

	t.Run("creates secure config path", func(t *testing.T) {
		if err := ensureEnvFile(); err != nil {
			t.Fatalf("ensureEnvFile() error = %v", err)
		}

		assertSecureConfigDirPerms(t, configDir)
		assertSecureEnvFilePerms(t, envPath)
	})

	t.Run("repairs existing insecure permissions", func(t *testing.T) {
		if err := os.Chmod(configDir, 0o755); err != nil {
			t.Fatalf("failed to chmod %s: %v", configDir, err)
		}
		if err := os.Chmod(envPath, 0o644); err != nil {
			t.Fatalf("failed to chmod %s: %v", envPath, err)
		}

		if err := ensureEnvFile(); err != nil {
			t.Fatalf("ensureEnvFile() error = %v", err)
		}

		assertSecureConfigDirPerms(t, configDir)
		assertSecureEnvFilePerms(t, envPath)
	})
}

func assertSecureConfigDirPerms(t *testing.T, path string) {
	t.Helper()

	if runtime.GOOS == "windows" {
		return
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat %s: %v", path, err)
	}
	if got := info.Mode().Perm(); got != fsdb.ConfigDirPerms {
		t.Fatalf("permissions for %s = %o, want %o", path, got, fsdb.ConfigDirPerms)
	}
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
	if got := info.Mode().Perm(); got != fsdb.EnvFilePerms {
		t.Fatalf("permissions for %s = %o, want %o", path, got, fsdb.EnvFilePerms)
	}
}
