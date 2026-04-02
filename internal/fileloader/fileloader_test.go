package fileloader

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type testLoadable struct {
	Name string `yaml:"Name"`
	Path string `yaml:"Path"`
}

func (t testLoadable) Id() string {
	return t.Name
}

func (t testLoadable) Validate() error {
	return nil
}

func (t testLoadable) Filepath() string {
	return t.Path
}

func TestSaveAllFlatFilesCreatesParentDirectories(t *testing.T) {
	basePath := t.TempDir()

	data := map[string]testLoadable{
		"one": {
			Name: "one",
			Path: "nested/example.yaml",
		},
	}

	saved, err := SaveAllFlatFiles(basePath, data)
	if err != nil {
		t.Fatalf("SaveAllFlatFiles() error = %v", err)
	}

	if saved != 1 {
		t.Fatalf("SaveAllFlatFiles() saved = %d, want 1", saved)
	}

	if _, err := os.Stat(filepath.Join(basePath, "nested", "example.yaml")); err != nil {
		t.Fatalf("Stat(saved file) error = %v", err)
	}
}

func TestSaveAllFlatFilesReturnsErrorInsteadOfPanicking(t *testing.T) {
	basePath := t.TempDir()

	data := map[string]testLoadable{
		"bad": {
			Name: "bad",
			Path: "nested/example.txt",
		},
	}

	saved, err := SaveAllFlatFiles(basePath, data)
	if err == nil {
		t.Fatal("SaveAllFlatFiles() error = nil, want unsupported file type error")
	}

	if saved != 0 {
		t.Fatalf("SaveAllFlatFiles() saved = %d, want 0", saved)
	}

	if !strings.Contains(err.Error(), "unsupported file type") {
		t.Fatalf("SaveAllFlatFiles() error = %v, want unsupported file type", err)
	}
}
