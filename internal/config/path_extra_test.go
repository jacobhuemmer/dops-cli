package config

import (
	"dops/internal/domain"
	"encoding/json"
	"fmt"
	"io/fs"
	"testing"
)

// --- Tests for setDefaults (0% covered) ---

func TestSetDefaults_MaxRiskLevel_Valid(t *testing.T) {
	levels := []string{"low", "medium", "high", "critical"}
	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			cfg := &domain.Config{}
			err := Set(cfg, "defaults.max_risk_level", level)
			if err != nil {
				t.Fatalf("Set defaults.max_risk_level=%q: %v", level, err)
			}
			if string(cfg.Defaults.MaxRiskLevel) != level {
				t.Errorf("MaxRiskLevel = %q, want %q", cfg.Defaults.MaxRiskLevel, level)
			}
		})
	}
}

func TestSetDefaults_MaxRiskLevel_InvalidString(t *testing.T) {
	cfg := &domain.Config{}
	err := Set(cfg, "defaults.max_risk_level", "unknown-level")
	if err == nil {
		t.Error("expected error for invalid risk level")
	}
}

func TestSetDefaults_MaxRiskLevel_NonString(t *testing.T) {
	cfg := &domain.Config{}
	err := Set(cfg, "defaults.max_risk_level", 42)
	if err == nil {
		t.Error("expected error when value is not a string")
	}
}

func TestSetDefaults_IncompletePath(t *testing.T) {
	cfg := &domain.Config{}
	err := Set(cfg, "defaults", "value")
	if err == nil {
		t.Error("expected error for incomplete defaults path")
	}
}

func TestSetDefaults_UnknownKey(t *testing.T) {
	cfg := &domain.Config{}
	err := Set(cfg, "defaults.nonexistent", "value")
	if err == nil {
		t.Error("expected error for unknown defaults key")
	}
}

// --- Tests for uncovered Get branches ---

func TestGet_ThemeNestedPath(t *testing.T) {
	cfg := newTestConfig()
	_, err := Get(cfg, "theme.nested")
	if err == nil {
		t.Error("expected error for nested theme path")
	}
}

func TestGet_DefaultsIncompletePath(t *testing.T) {
	cfg := newTestConfig()
	_, err := Get(cfg, "defaults")
	if err == nil {
		t.Error("expected error for incomplete defaults path")
	}
}

func TestGet_DefaultsUnknownKey(t *testing.T) {
	cfg := newTestConfig()
	_, err := Get(cfg, "defaults.nonexistent")
	if err == nil {
		t.Error("expected error for unknown defaults key")
	}
}

func TestGet_VarsIncompletePath(t *testing.T) {
	cfg := newTestConfig()
	_, err := Get(cfg, "vars")
	if err == nil {
		t.Error("expected error for incomplete vars path")
	}
}

func TestGet_VarsUnknownScope(t *testing.T) {
	cfg := newTestConfig()
	_, err := Get(cfg, "vars.unknown.key")
	if err == nil {
		t.Error("expected error for unknown vars scope")
	}
}

func TestGet_CatalogVarsIncompletePath(t *testing.T) {
	cfg := newTestConfig()
	_, err := Get(cfg, "vars.catalog")
	if err == nil {
		t.Error("expected error for incomplete catalog vars path")
	}
}

func TestGet_RunbookVarsIncompletePath(t *testing.T) {
	cfg := newTestConfig()
	_, err := Get(cfg, "vars.catalog.default.runbooks")
	if err == nil {
		t.Error("expected error for incomplete runbook vars path")
	}
}

func TestGet_RunbookVarsMissingRunbook(t *testing.T) {
	cfg := newTestConfig()
	_, err := Get(cfg, "vars.catalog.default.runbooks.missing.key")
	if err == nil {
		t.Error("expected error for missing runbook")
	}
}

func TestGet_RunbookVarsMissingKey(t *testing.T) {
	cfg := newTestConfig()
	_, err := Get(cfg, "vars.catalog.default.runbooks.hello-world.missing")
	if err == nil {
		t.Error("expected error for missing runbook var key")
	}
}

func TestGet_CatalogVarMissingKey(t *testing.T) {
	cfg := newTestConfig()
	_, err := Get(cfg, "vars.catalog.default.missing_key")
	if err == nil {
		t.Error("expected error for missing catalog var key")
	}
}

func TestGet_GlobalVarsNilMap(t *testing.T) {
	cfg := &domain.Config{
		Vars: domain.Vars{Global: nil},
	}
	_, err := Get(cfg, "vars.global.any")
	if err == nil {
		t.Error("expected error when global vars map is nil")
	}
}

// --- Tests for Set uncovered branches ---

func TestSet_ThemeWithSpaces(t *testing.T) {
	cfg := &domain.Config{}
	err := Set(cfg, "theme", "has space")
	if err == nil {
		t.Error("expected error for theme with spaces")
	}
}

func TestSet_ThemeWithTab(t *testing.T) {
	cfg := &domain.Config{}
	err := Set(cfg, "theme", "has\ttab")
	if err == nil {
		t.Error("expected error for theme with tab")
	}
}

func TestSet_ThemeNonString(t *testing.T) {
	cfg := &domain.Config{}
	err := Set(cfg, "theme", 123)
	if err == nil {
		t.Error("expected error for non-string theme")
	}
}

func TestSet_ThemeNestedPath(t *testing.T) {
	cfg := &domain.Config{}
	err := Set(cfg, "theme.nested", "value")
	if err == nil {
		t.Error("expected error for nested theme path")
	}
}

func TestSet_UnknownTopLevel(t *testing.T) {
	cfg := &domain.Config{}
	err := Set(cfg, "bogus", "value")
	if err == nil {
		t.Error("expected error for unknown top-level key")
	}
}

func TestSet_VarsIncompletePath(t *testing.T) {
	cfg := &domain.Config{}
	err := Set(cfg, "vars", "value")
	if err == nil {
		t.Error("expected error for incomplete vars path")
	}
}

func TestSet_VarsUnknownScope(t *testing.T) {
	cfg := &domain.Config{}
	err := Set(cfg, "vars.unknown.key", "value")
	if err == nil {
		t.Error("expected error for unknown vars scope")
	}
}

func TestSet_CatalogRunbookIncompletePath(t *testing.T) {
	cfg := newTestConfig()
	err := Set(cfg, "vars.catalog.default.runbooks", "value")
	if err == nil {
		t.Error("expected error for incomplete runbook path")
	}
}

func TestSet_CatalogIncompletePath(t *testing.T) {
	cfg := &domain.Config{}
	err := Set(cfg, "vars.catalog", "value")
	if err == nil {
		t.Error("expected error for incomplete catalog path")
	}
}

// --- Tests for Unset uncovered branches ---

func TestUnset_NonVarsPath(t *testing.T) {
	cfg := newTestConfig()
	err := Unset(cfg, "theme")
	if err == nil {
		t.Error("expected error for non-vars unset path")
	}
}

func TestUnset_IncompletePath(t *testing.T) {
	cfg := newTestConfig()
	err := Unset(cfg, "vars.global")
	if err == nil {
		t.Error("expected error for incomplete unset path")
	}
}

func TestUnset_UnknownScope(t *testing.T) {
	cfg := newTestConfig()
	err := Unset(cfg, "vars.unknown.key")
	if err == nil {
		t.Error("expected error for unknown vars scope in unset")
	}
}

func TestUnset_NilGlobalMap(t *testing.T) {
	cfg := &domain.Config{Vars: domain.Vars{Global: nil}}
	err := Unset(cfg, "vars.global.key")
	if err == nil {
		t.Error("expected error for nil global map")
	}
}

func TestUnset_MissingCatalog(t *testing.T) {
	cfg := newTestConfig()
	err := Unset(cfg, "vars.catalog.missing.key")
	if err == nil {
		t.Error("expected error for missing catalog in unset")
	}
}

func TestUnset_MissingCatalogVar(t *testing.T) {
	cfg := newTestConfig()
	err := Unset(cfg, "vars.catalog.default.nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent catalog var")
	}
}

func TestUnset_RunbookIncompletePath(t *testing.T) {
	cfg := newTestConfig()
	err := Unset(cfg, "vars.catalog.default.runbooks")
	if err == nil {
		t.Error("expected error for incomplete runbook unset path")
	}
}

func TestUnset_RunbookMissing(t *testing.T) {
	cfg := newTestConfig()
	err := Unset(cfg, "vars.catalog.default.runbooks.missing.key")
	if err == nil {
		t.Error("expected error for missing runbook in unset")
	}
}

func TestUnset_RunbookVarMissing(t *testing.T) {
	cfg := newTestConfig()
	err := Unset(cfg, "vars.catalog.default.runbooks.hello-world.missing")
	if err == nil {
		t.Error("expected error for missing var in runbook unset")
	}
}

func TestUnset_CatalogIncompletePath(t *testing.T) {
	cfg := newTestConfig()
	err := Unset(cfg, "vars.catalog.default")
	if err == nil {
		t.Error("expected error for incomplete catalog unset path")
	}
}

// --- Tests for EnsureDefaults uncovered branches ---

type errorFS struct {
	fakeFS
	readErr  error
	writeErr error
}

func (e *errorFS) ReadFile(path string) ([]byte, error) {
	if e.readErr != nil {
		return nil, e.readErr
	}
	return e.fakeFS.ReadFile(path)
}

func (e *errorFS) WriteFile(path string, data []byte, perm fs.FileMode) error {
	if e.writeErr != nil {
		return e.writeErr
	}
	return e.fakeFS.WriteFile(path, data, perm)
}

func TestEnsureDefaults_NonExistError(t *testing.T) {
	// ReadFile returns a non-ErrNotExist error.
	efs := &errorFS{
		fakeFS:  *newFakeFS(),
		readErr: fmt.Errorf("read config: %w", fmt.Errorf("disk failure")),
	}
	store := NewFileStore(efs, "/fake/config.json")

	_, err := store.EnsureDefaults()
	if err == nil {
		t.Fatal("expected error for non-ErrNotExist read failure")
	}
}

func TestEnsureDefaults_SaveFails(t *testing.T) {
	efs := &errorFS{
		fakeFS:   *newFakeFS(),
		writeErr: fmt.Errorf("disk full"),
	}
	store := NewFileStore(efs, "/fake/config.json")

	_, err := store.EnsureDefaults()
	if err == nil {
		t.Fatal("expected error when save fails during EnsureDefaults")
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	ffs := newFakeFS()
	ffs.files["/fake/config.json"] = []byte("not json")
	store := NewFileStore(ffs, "/fake/config.json")

	_, err := store.Load()
	if err == nil {
		t.Fatal("expected error for invalid JSON in config file")
	}
}

func TestDefaultConfig_Values(t *testing.T) {
	cfg := defaultConfig()
	if cfg.Theme != "github" {
		t.Errorf("default theme = %q, want github", cfg.Theme)
	}
	if cfg.Defaults.MaxRiskLevel != domain.RiskMedium {
		t.Errorf("default MaxRiskLevel = %q, want medium", cfg.Defaults.MaxRiskLevel)
	}
	if cfg.Catalogs == nil {
		t.Error("default catalogs should not be nil")
	}
}

// --- EnsureDefaults with corrupt existing file ---

func TestEnsureDefaults_CorruptExistingFile(t *testing.T) {
	ffs := newFakeFS()
	ffs.files["/fake/config.json"] = []byte("{invalid json")
	store := NewFileStore(ffs, "/fake/config.json")

	_, err := store.EnsureDefaults()
	if err == nil {
		t.Fatal("expected error for corrupt existing config file")
	}
}

// --- Save error paths ---

type mkdirErrorFS struct {
	fakeFS
}

func (e *mkdirErrorFS) MkdirAll(_ string, _ fs.FileMode) error {
	return fmt.Errorf("permission denied")
}

func TestSave_MkdirFails(t *testing.T) {
	efs := &mkdirErrorFS{fakeFS: *newFakeFS()}
	store := NewFileStore(efs, "/fake/nested/config.json")

	cfg := &domain.Config{Theme: "test"}
	err := store.Save(cfg)
	if err == nil {
		t.Fatal("expected error when mkdir fails")
	}
}

// Verify Save with invalid config does not produce error (json.MarshalIndent handles all domain types).
func TestSave_RoundTripWithVarsExcluded(t *testing.T) {
	ffs := newFakeFS()
	store := NewFileStore(ffs, "/fake/config.json")

	cfg := &domain.Config{
		Theme:    "nord",
		Defaults: domain.Defaults{MaxRiskLevel: domain.RiskHigh},
		Catalogs: []domain.Catalog{
			{Name: "test", Path: "/tmp/test", Active: true},
		},
	}

	if err := store.Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	data := ffs.files["/fake/config.json"]
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}

	if _, ok := raw["vars"]; ok {
		t.Error("vars should not appear in saved config")
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Theme != "nord" {
		t.Errorf("theme = %q, want nord", loaded.Theme)
	}
	if loaded.Defaults.MaxRiskLevel != domain.RiskHigh {
		t.Errorf("MaxRiskLevel = %q, want high", loaded.Defaults.MaxRiskLevel)
	}
	if len(loaded.Catalogs) != 1 {
		t.Errorf("catalogs len = %d, want 1", len(loaded.Catalogs))
	}
}
