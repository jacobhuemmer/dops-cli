package config

import (
	"dops/internal/domain"
	"encoding/json"
	"io/fs"
	"os"
	"testing"
)

type fakeFS struct {
	files map[string][]byte
	dirs  map[string]bool
}

func newFakeFS() *fakeFS {
	return &fakeFS{
		files: make(map[string][]byte),
		dirs:  make(map[string]bool),
	}
}

func (f *fakeFS) ReadFile(path string) ([]byte, error) {
	data, ok := f.files[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return data, nil
}

func (f *fakeFS) WriteFile(path string, data []byte, _ fs.FileMode) error {
	f.files[path] = data
	return nil
}

func (f *fakeFS) MkdirAll(path string, _ fs.FileMode) error {
	f.dirs[path] = true
	return nil
}

func TestFileConfigStore_RoundTrip(t *testing.T) {
	ffs := newFakeFS()
	store := NewFileStore(ffs, "/fake/.dops/config.json")

	original := &domain.Config{
		Theme:    "tokyomidnight",
		Defaults: domain.Defaults{MaxRiskLevel: domain.RiskMedium},
		Vars: domain.Vars{
			Global: map[string]any{"region": "us-east-1"},
		},
	}

	if err := store.Save(original); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.Theme != original.Theme {
		t.Errorf("Theme = %q, want %q", loaded.Theme, original.Theme)
	}
	if loaded.Defaults.MaxRiskLevel != original.Defaults.MaxRiskLevel {
		t.Errorf("MaxRiskLevel = %q, want %q", loaded.Defaults.MaxRiskLevel, original.Defaults.MaxRiskLevel)
	}
	if loaded.Vars.Global["region"] != "us-east-1" {
		t.Errorf("region = %v, want us-east-1", loaded.Vars.Global["region"])
	}
}

func TestFileConfigStore_LoadMissing(t *testing.T) {
	ffs := newFakeFS()
	store := NewFileStore(ffs, "/fake/.dops/config.json")

	_, err := store.Load()
	if err == nil {
		t.Fatal("expected error loading nonexistent config")
	}
}

func TestFileConfigStore_EnsureDefaults(t *testing.T) {
	ffs := newFakeFS()
	store := NewFileStore(ffs, "/fake/.dops/config.json")

	cfg, err := store.EnsureDefaults()
	if err != nil {
		t.Fatalf("EnsureDefaults: %v", err)
	}

	if cfg.Theme != "tokyomidnight" {
		t.Errorf("default theme = %q, want tokyomidnight", cfg.Theme)
	}

	data, ok := ffs.files["/fake/.dops/config.json"]
	if !ok {
		t.Fatal("config file not written to disk")
	}

	var ondisk domain.Config
	if err := json.Unmarshal(data, &ondisk); err != nil {
		t.Fatalf("unmarshal written config: %v", err)
	}
	if ondisk.Theme != "tokyomidnight" {
		t.Errorf("on-disk theme = %q, want tokyomidnight", ondisk.Theme)
	}
}

func TestFileConfigStore_EnsureDefaults_ExistingFile(t *testing.T) {
	ffs := newFakeFS()
	store := NewFileStore(ffs, "/fake/.dops/config.json")

	existing := &domain.Config{Theme: "dracula"}
	data, _ := json.MarshalIndent(existing, "", "  ")
	ffs.files["/fake/.dops/config.json"] = data

	cfg, err := store.EnsureDefaults()
	if err != nil {
		t.Fatalf("EnsureDefaults: %v", err)
	}

	if cfg.Theme != "dracula" {
		t.Errorf("theme = %q, want dracula (existing)", cfg.Theme)
	}
}
