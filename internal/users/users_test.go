package users

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/configs"
)

func configureUserDataPathForTest(t *testing.T) string {
	t.Helper()

	basePath := t.TempDir()
	if err := configs.AddOverlayOverrides(map[string]any{
		"FilePaths.DataFiles": basePath,
	}); err != nil {
		t.Fatalf("AddOverlayOverrides() error = %v", err)
	}

	return basePath
}

func TestSearchOfflineUsersReturnsMalformedYamlError(t *testing.T) {
	basePath := configureUserDataPathForTest(t)
	usersPath := filepath.Join(basePath, "users")
	if err := os.MkdirAll(usersPath, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	if err := os.WriteFile(filepath.Join(usersPath, "broken.yaml"), []byte("UserId: ["), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	err := SearchOfflineUsers(func(u *UserRecord) bool {
		t.Fatalf("search callback should not run for malformed yaml: %#v", u)
		return true
	})
	if err == nil {
		t.Fatal("SearchOfflineUsers() error = nil, want YAML error")
	}
}

func TestSearchOfflineUsersMissingDirectoryIsNotError(t *testing.T) {
	configureUserDataPathForTest(t)

	if err := SearchOfflineUsers(func(u *UserRecord) bool { return true }); err != nil {
		t.Fatalf("SearchOfflineUsers() error = %v, want nil when users directory is missing", err)
	}
}

func TestLoadUserReturnsYamlError(t *testing.T) {
	basePath := configureUserDataPathForTest(t)
	usersPath := filepath.Join(basePath, "users")
	if err := os.MkdirAll(usersPath, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	idx := NewUserIndex()
	if err := idx.Create(); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := idx.AddUser(1, "alice"); err != nil {
		t.Fatalf("AddUser() error = %v", err)
	}

	if err := os.WriteFile(filepath.Join(usersPath, "1.yaml"), []byte("Username: alice\nCharacter: ["), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if _, err := LoadUser("alice"); err == nil {
		t.Fatal("LoadUser() error = nil, want YAML error")
	}
}
