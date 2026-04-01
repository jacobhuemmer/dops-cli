package cmd

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"dops/internal/domain"
	"dops/internal/vault"
)

func TestSplitError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantTitle  string
		wantDetail string
	}{
		{
			name:       "error with colon",
			err:        errors.New("config: file not found"),
			wantTitle:  "Config",
			wantDetail: "file not found",
		},
		{
			name:       "error without colon",
			err:        errors.New("something went wrong"),
			wantTitle:  "something went wrong",
			wantDetail: "",
		},
		{
			name:       "empty error",
			err:        errors.New(""),
			wantTitle:  "",
			wantDetail: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, detail := splitError(tt.err)
			if title != tt.wantTitle {
				t.Errorf("title = %q, want %q", title, tt.wantTitle)
			}
			if detail != tt.wantDetail {
				t.Errorf("detail = %q, want %q", detail, tt.wantDetail)
			}
		})
	}
}

func TestTitleCase(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "normal string", in: "hello", want: "Hello"},
		{name: "empty string", in: "", want: ""},
		{name: "single char", in: "a", want: "A"},
		{name: "already capitalized", in: "Hello", want: "Hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := titleCase(tt.in)
			if got != tt.want {
				t.Errorf("titleCase(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// stubVault implements domain.VaultStore for testing migrateVarsToVault.
type stubVault struct {
	exists bool
	saved  *domain.Vars
}

func (v *stubVault) Load() (*domain.Vars, error) { return v.saved, nil }
func (v *stubVault) Save(vars *domain.Vars) error { v.saved = vars; return nil }
func (v *stubVault) Exists() bool                 { return v.exists }

// stubFS implements config.FileSystem for testing migrateVarsToVault.
type stubFS struct {
	files map[string][]byte
}

func (f *stubFS) ReadFile(path string) ([]byte, error) {
	data, ok := f.files[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return data, nil
}

func (f *stubFS) WriteFile(path string, data []byte, perm fs.FileMode) error {
	f.files[path] = data
	return nil
}

func (f *stubFS) MkdirAll(path string, perm fs.FileMode) error {
	return nil
}

func TestMigrateVarsToVault(t *testing.T) {
	configWithVars := func() []byte {
		data, _ := json.Marshal(map[string]any{
			"theme": "tokyonight",
			"vars": map[string]any{
				"global":  map[string]any{"region": "us-east-1"},
				"catalog": map[string]any{},
			},
		})
		return data
	}

	configWithoutVars := func() []byte {
		data, _ := json.Marshal(map[string]any{
			"theme": "tokyonight",
		})
		return data
	}

	configWithEmptyVars := func() []byte {
		data, _ := json.Marshal(map[string]any{
			"theme": "tokyonight",
			"vars": map[string]any{
				"global":  map[string]any{},
				"catalog": map[string]any{},
			},
		})
		return data
	}

	tests := []struct {
		name       string
		vlt        *stubVault
		fs         *stubFS
		configPath string
		wantSaved  bool // whether vars should be saved to vault
	}{
		{
			name:       "vault already exists",
			vlt:        &stubVault{exists: true},
			fs:         &stubFS{files: map[string][]byte{}},
			configPath: "/fake/config.json",
			wantSaved:  false,
		},
		{
			name:       "no config file",
			vlt:        &stubVault{exists: false},
			fs:         &stubFS{files: map[string][]byte{}},
			configPath: "/fake/config.json",
			wantSaved:  false,
		},
		{
			name:       "config with vars",
			vlt:        &stubVault{exists: false},
			fs:         &stubFS{files: map[string][]byte{"/fake/config.json": configWithVars()}},
			configPath: "/fake/config.json",
			wantSaved:  true,
		},
		{
			name:       "config without vars",
			vlt:        &stubVault{exists: false},
			fs:         &stubFS{files: map[string][]byte{"/fake/config.json": configWithoutVars()}},
			configPath: "/fake/config.json",
			wantSaved:  false,
		},
		{
			name:       "empty vars",
			vlt:        &stubVault{exists: false},
			fs:         &stubFS{files: map[string][]byte{"/fake/config.json": configWithEmptyVars()}},
			configPath: "/fake/config.json",
			wantSaved:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := migrateVarsToVault(tt.configPath, tt.vlt, tt.fs)
			if err != nil {
				t.Fatalf("migrateVarsToVault() error = %v", err)
			}
			if tt.wantSaved && tt.vlt.saved == nil {
				t.Error("expected vars to be saved to vault, but nothing was saved")
			}
			if !tt.wantSaved && tt.vlt.saved != nil {
				t.Error("expected no save to vault, but vars were saved")
			}
		})
	}
}

// TestMigrateVarsToVault_CleanedConfig verifies that the vars key is removed
// from config.json after migration using real vault encryption.
func TestMigrateVarsToVault_CleanedConfig(t *testing.T) {
	dir := t.TempDir()
	dopsDir := filepath.Join(dir, ".dops")
	os.MkdirAll(filepath.Join(dopsDir, "keys"), 0o700)

	configPath := filepath.Join(dopsDir, "config.json")
	configData, _ := json.MarshalIndent(map[string]any{
		"theme": "tokyonight",
		"vars": map[string]any{
			"global":  map[string]any{"region": "us-east-1"},
			"catalog": map[string]any{},
		},
	}, "", "  ")
	os.WriteFile(configPath, configData, 0o644)

	vaultPath := filepath.Join(dopsDir, "vault.json")
	keysDir := filepath.Join(dopsDir, "keys")
	vlt := vault.New(vaultPath, keysDir)

	// Use real OS filesystem.
	realFS := &osFS{}
	err := migrateVarsToVault(configPath, vlt, realFS)
	if err != nil {
		t.Fatalf("migrateVarsToVault() error = %v", err)
	}

	// Verify vault was created with vars.
	vars, err := vlt.Load()
	if err != nil {
		t.Fatalf("load vault: %v", err)
	}
	if vars.Global["region"] != "us-east-1" {
		t.Errorf("vault region = %v, want us-east-1", vars.Global["region"])
	}

	// Verify config.json no longer has vars key.
	cleaned, _ := os.ReadFile(configPath)
	var raw map[string]json.RawMessage
	json.Unmarshal(cleaned, &raw)
	if _, ok := raw["vars"]; ok {
		t.Error("config.json should not contain vars after migration")
	}
}

// osFS implements config.FileSystem using the real OS filesystem.
type osFS struct{}

func (f *osFS) ReadFile(path string) ([]byte, error)              { return os.ReadFile(path) }
func (f *osFS) WriteFile(path string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(path, data, perm)
}
func (f *osFS) MkdirAll(path string, perm fs.FileMode) error { return os.MkdirAll(path, perm) }
