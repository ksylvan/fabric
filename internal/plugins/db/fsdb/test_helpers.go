package fsdb

import "testing"

// setupTestStorage creates a StorageEntity with a temporary directory for testing.
// The directory is automatically cleaned up when the test completes.
func setupTestStorage(t *testing.T) *StorageEntity {
	t.Helper()
	dir := t.TempDir()
	return &StorageEntity{Dir: dir}
}

// setupTestStorageWithExtension creates a StorageEntity with a temporary directory
// and file extension for testing. The directory is automatically cleaned up when
// the test completes.
func setupTestStorageWithExtension(t *testing.T, ext string) *StorageEntity {
	t.Helper()
	dir := t.TempDir()
	return &StorageEntity{Dir: dir, FileExtension: ext}
}

// setupTestDB creates a Db instance with a temporary directory for testing.
// The directory is automatically cleaned up when the test completes.
func setupTestDB(t *testing.T) *Db {
	t.Helper()
	dir := t.TempDir()
	return NewDb(dir)
}

// setupTestContexts creates a ContextsEntity with a temporary directory for testing.
// The directory is automatically cleaned up when the test completes.
func setupTestContexts(t *testing.T) *ContextsEntity {
	t.Helper()
	dir := t.TempDir()
	return &ContextsEntity{
		StorageEntity: &StorageEntity{Dir: dir},
	}
}

// setupTestSessions creates a SessionsEntity with a temporary directory for testing.
// The directory is automatically cleaned up when the test completes.
func setupTestSessions(t *testing.T) *SessionsEntity {
	t.Helper()
	dir := t.TempDir()
	return &SessionsEntity{
		StorageEntity: &StorageEntity{Dir: dir, FileExtension: ".json"},
	}
}
