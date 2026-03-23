package catalog

import (
	"dops/internal/domain"
	"io/fs"
	"os"
	"testing"
)

type fakeDirEntry struct {
	name  string
	isDir bool
}

func (f fakeDirEntry) Name() string               { return f.name }
func (f fakeDirEntry) IsDir() bool                 { return f.isDir }
func (f fakeDirEntry) Type() fs.FileMode           { return 0 }
func (f fakeDirEntry) Info() (fs.FileInfo, error)  { return nil, nil }

type fakeFS struct {
	files map[string][]byte
	dirs  map[string][]os.DirEntry
}

func newFakeFS() *fakeFS {
	return &fakeFS{
		files: make(map[string][]byte),
		dirs:  make(map[string][]os.DirEntry),
	}
}

func (f *fakeFS) ReadFile(path string) ([]byte, error) {
	data, ok := f.files[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return data, nil
}

func (f *fakeFS) WriteFile(string, []byte, fs.FileMode) error { return nil }
func (f *fakeFS) MkdirAll(string, fs.FileMode) error          { return nil }
func (f *fakeFS) Stat(string) (os.FileInfo, error)            { return nil, nil }

func (f *fakeFS) ReadDir(path string) ([]os.DirEntry, error) {
	entries, ok := f.dirs[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return entries, nil
}

const helloWorldYAML = `id: "default.hello-world"
name: "hello-world"
description: "Prints a hello world message"
version: "1.0.0"
risk_level: "low"
script: "./script.sh"
parameters:
  - name: "greeting"
    type: "string"
    required: true
    scope: "global"
    secret: false
    description: "The greeting message"
`

const highRiskYAML = `id: "default.danger-zone"
name: "danger-zone"
description: "A high risk runbook"
version: "1.0.0"
risk_level: "high"
script: "./script.sh"
`

func setupTestFS() *fakeFS {
	ffs := newFakeFS()
	ffs.dirs["/catalogs/default"] = []os.DirEntry{
		fakeDirEntry{name: "hello-world", isDir: true},
		fakeDirEntry{name: "danger-zone", isDir: true},
	}
	ffs.files["/catalogs/default/hello-world/runbook.yaml"] = []byte(helloWorldYAML)
	ffs.files["/catalogs/default/danger-zone/runbook.yaml"] = []byte(highRiskYAML)
	return ffs
}

func TestDiskCatalogLoader_LoadAll(t *testing.T) {
	ffs := setupTestFS()
	loader := NewDiskLoader(ffs)

	catalogs := []domain.Catalog{
		{Name: "default", Path: "/catalogs/default", Active: true, Policy: domain.CatalogPolicy{MaxRiskLevel: domain.RiskMedium}},
	}

	result, err := loader.LoadAll(catalogs, domain.RiskMedium)
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 catalog group, got %d", len(result))
	}

	// Only "hello-world" (low risk) should be loaded. "danger-zone" (high) exceeds medium ceiling.
	if len(result[0].Runbooks) != 1 {
		t.Fatalf("expected 1 runbook (filtered), got %d", len(result[0].Runbooks))
	}

	rb := result[0].Runbooks[0]
	if rb.ID != "default.hello-world" {
		t.Errorf("runbook ID = %q, want default.hello-world", rb.ID)
	}
	if rb.Name != "hello-world" {
		t.Errorf("runbook Name = %q, want hello-world", rb.Name)
	}
	if len(rb.Parameters) != 1 {
		t.Errorf("expected 1 parameter, got %d", len(rb.Parameters))
	}
}

func TestDiskCatalogLoader_SkipsInactiveCatalogs(t *testing.T) {
	ffs := setupTestFS()
	loader := NewDiskLoader(ffs)

	catalogs := []domain.Catalog{
		{Name: "default", Path: "/catalogs/default", Active: false},
	}

	result, err := loader.LoadAll(catalogs, domain.RiskMedium)
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected 0 catalogs (inactive), got %d", len(result))
	}
}

func TestDiskCatalogLoader_FindByID(t *testing.T) {
	ffs := setupTestFS()
	loader := NewDiskLoader(ffs)

	catalogs := []domain.Catalog{
		{Name: "default", Path: "/catalogs/default", Active: true, Policy: domain.CatalogPolicy{MaxRiskLevel: domain.RiskCritical}},
	}

	_, err := loader.LoadAll(catalogs, domain.RiskCritical)
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}

	t.Run("found", func(t *testing.T) {
		rb, cat, err := loader.FindByID("default.hello-world")
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}
		if rb.ID != "default.hello-world" {
			t.Errorf("runbook ID = %q", rb.ID)
		}
		if cat.Name != "default" {
			t.Errorf("catalog name = %q", cat.Name)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, _, err := loader.FindByID("unknown.runbook")
		if err == nil {
			t.Error("expected error for unknown ID")
		}
	})
}

func TestRiskFilter(t *testing.T) {
	tests := []struct {
		name     string
		level    domain.RiskLevel
		ceiling  domain.RiskLevel
		included bool
	}{
		{"low under medium", domain.RiskLow, domain.RiskMedium, true},
		{"medium at medium", domain.RiskMedium, domain.RiskMedium, true},
		{"high over medium", domain.RiskHigh, domain.RiskMedium, false},
		{"critical over medium", domain.RiskCritical, domain.RiskMedium, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := !tt.level.Exceeds(tt.ceiling)
			if got != tt.included {
				t.Errorf("RiskLevel(%q).Exceeds(%q) filter = %v, want included=%v", tt.level, tt.ceiling, !got, tt.included)
			}
		})
	}
}
